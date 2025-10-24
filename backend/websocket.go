package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"social-network/backend/db"
	"social-network/backend/handlers"
	"social-network/backend/models"
	"social-network/backend/utils"

	"github.com/gorilla/websocket"
)

const throttleRate = 500 * time.Millisecond

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients      = make(map[string]*Client)
	clientsMutex sync.RWMutex
)

type Client struct {
	ID       string
	Nickname string
	Conn     *websocket.Conn
	Send     chan []byte
	lastSent time.Time
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(utils.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Get user nickname for the client
	var nickname string
	err = db.DB.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&nickname)
	if err != nil {
		log.Println("Error fetching user nickname:", err)
		nickname = userID // fallback
	}

	client := &Client{
		ID:       userID,
		Nickname: nickname,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	clientsMutex.Lock()
	if oldClient, exists := clients[userID]; exists {
		oldClient.Conn.Close() // Triggers cleanup in readPump()
	}
	clients[userID] = client
	clientsMutex.Unlock()

	fmt.Println("User connected:", userID, "Nickname:", nickname)

	_, err = db.DB.Exec("UPDATE users SET online_status = 1 WHERE id = ?", userID)
	if err != nil {
		log.Println("Error updating user status:", err)
	}

	sendOnlineUsers("")
	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.Conn.Close()
		clientsMutex.Lock()
		delete(clients, c.ID)
		clientsMutex.Unlock()
		db.DB.Exec("UPDATE users SET online_status = 0 WHERE id = ?", c.ID)
		sendOnlineUsers("")
	}()

	for {
		_, msgBytes, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// decode into a lightweight struct to support group messages
		var raw struct {
			Type       string `json:"type"`
			ReceiverID string `json:"receiver_id"`
			GroupID    int64  `json:"group_id"`
			Content    string `json:"content"`
		}
		if err := json.Unmarshal(msgBytes, &raw); err != nil {
			log.Println("Message unmarshal error:", err)
			continue
		}

		// basic emoji shortcode expansion (small set)
		emojiMap := map[string]string{
			":smile:":     "😄",
			":heart:":     "❤️",
			":thumbs_up:": "👍",
			":laugh:":     "😂",
			":cry:":       "😢",
		}
		for k, v := range emojiMap {
			if strings.Contains(raw.Content, k) {
				raw.Content = strings.ReplaceAll(raw.Content, k, v)
			}
		}

		// DM (direct message)
		if raw.Type == "message" {
			// enforce allowed users: either follows the other
			senderIDInt, _ := strconv.ParseInt(c.ID, 10, 64)
			receiverIDInt, errConv := strconv.ParseInt(raw.ReceiverID, 10, 64)
			if errConv != nil {
				// invalid receiver id
				continue
			}

			var relCount int
			err := db.DB.QueryRow("SELECT COUNT(1) FROM followers WHERE (follower_id=? AND followed_id=?) OR (follower_id=? AND followed_id=?)", senderIDInt, receiverIDInt, receiverIDInt, senderIDInt).Scan(&relCount)
			if err != nil {
				log.Println("relation check error:", err)
				continue
			}
			if relCount == 0 {
				// not allowed to DM
				errMsg := models.Message{Type: "error", Content: "You are not allowed to message this user."}
				payload, _ := json.Marshal(errMsg)
				c.Send <- payload
				continue
			}

			// insert DM
			result, err := db.DB.Exec("INSERT INTO messages (sender_id, receiver_id, content, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)", senderIDInt, receiverIDInt, raw.Content)
			if err != nil {
				log.Println("DB insert error:", err)
				continue
			}
			msgID, _ := result.LastInsertId()
			var createdAt string
			db.DB.QueryRow("SELECT created_at FROM messages WHERE id = ?", msgID).Scan(&createdAt)

			// prepare outgoing message
			out := models.Message{
				ID:         int(msgID),
				Type:       "message",
				Content:    raw.Content,
				SenderID:   c.ID,
				SenderName: c.Nickname,
				ReceiverID: raw.ReceiverID,
			}

			// parse createdAt (DB returns string) and set CreatedAt on outgoing message
			if parsed, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
				out.CreatedAt = parsed
			} else if parsedRFC, err2 := time.Parse(time.RFC3339, createdAt); err2 == nil {
				out.CreatedAt = parsedRFC
			}
			encoded, _ := json.Marshal(out)

			clientsMutex.RLock()
			receiverClient, ok := clients[raw.ReceiverID]
			clientsMutex.RUnlock()
			if ok {
				receiverClient.Send <- encoded
				// lightweight realtime notification
				notification := models.Message{Type: "new_message_notification", SenderID: out.SenderID, SenderName: out.SenderName, Content: out.Content}
				notifPayload, _ := json.Marshal(notification)
				receiverClient.Send <- notifPayload
			}
			// persist & publish structured notification (store preview only)
			preview := raw.Content
			if len(preview) > 140 {
				preview = preview[:140]
			}
			_ = handlers.Notify(receiverIDInt, senderIDInt, "new_message", map[string]interface{}{"message_id": msgID, "conversation_id": receiverIDInt, "preview": preview, "url": "/chat"})

			// echo back to sender
			c.Send <- encoded
			continue
		}

		// Group message
		if raw.Type == "group_message" {
			senderIDInt, _ := strconv.ParseInt(c.ID, 10, 64)
			// check membership
			var memberCount int
			err := db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", raw.GroupID, senderIDInt).Scan(&memberCount)
			if err != nil {
				log.Println("group membership check error:", err)
				continue
			}
			if memberCount == 0 {
				errMsg := models.Message{Type: "error", Content: "You are not a member of this group."}
				payload, _ := json.Marshal(errMsg)
				c.Send <- payload
				continue
			}

			// persist group message
			res, err := db.DB.Exec("INSERT INTO group_messages (group_id, sender_id, content, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)", raw.GroupID, senderIDInt, raw.Content)
			if err != nil {
				log.Println("group message insert error:", err)
				continue
			}
			gmID, _ := res.LastInsertId()

			// build outgoing payload
			out := map[string]interface{}{
				"id":          gmID,
				"type":        "group_message",
				"group_id":    raw.GroupID,
				"content":     raw.Content,
				"sender_id":   c.ID,
				"sender_name": c.Nickname,
			}
			encoded, _ := json.Marshal(out)

			// notify group members (both realtime and persistent)
			rows, err := db.DB.Query("SELECT user_id FROM group_members WHERE group_id=? AND user_id != ?", raw.GroupID, senderIDInt)
			if err != nil {
				log.Println("group members query error:", err)
				continue
			}
			defer rows.Close()
			var uid int64
			var recipients []int64
			for rows.Next() {
				if err := rows.Scan(&uid); err == nil {
					recipients = append(recipients, uid)
				}
			}
			// send to connected members
			for _, rid := range recipients {
				ridStr := strconv.FormatInt(rid, 10)
				clientsMutex.RLock()
				memberClient, ok := clients[ridStr]
				clientsMutex.RUnlock()
				if ok {
					memberClient.Send <- encoded
				}
				// persist & publish structured group_message notification (preview + link)
				preview := raw.Content
				if len(preview) > 140 {
					preview = preview[:140]
				}
				_ = handlers.Notify(rid, senderIDInt, "group_message", map[string]interface{}{"message_id": gmID, "group_id": raw.GroupID, "preview": preview, "url": fmt.Sprintf("/groups/%d", raw.GroupID)})
			}
			// also echo to sender
			c.Send <- encoded
			_ = handlers.Notify(senderIDInt, senderIDInt, "group_message_sent", map[string]interface{}{"message_id": gmID, "group_id": raw.GroupID})
			continue
		}

		if raw.Type == "typing" {
			clientsMutex.RLock()
			receiver, ok := clients[raw.ReceiverID]
			clientsMutex.RUnlock()

			if ok {
				fmt.Println("Forwarding typing notification from", c.ID, "to", receiver.ID)
				typingNotification := models.Message{
					Type:       "typing",
					SenderID:   c.ID,
					SenderName: c.Nickname,
					ReceiverID: raw.ReceiverID,
				}
				payload, _ := json.Marshal(typingNotification)
				receiver.Send <- payload
			}
			continue
		}

		if raw.Type == "stop_typing" {
			clientsMutex.RLock()
			receiver, ok := clients[raw.ReceiverID]
			clientsMutex.RUnlock()

			if ok {
				stopTypingNotification := models.Message{
					Type:       "stop_typing",
					SenderID:   c.ID,
					ReceiverID: raw.ReceiverID,
				}
				payload, _ := json.Marshal(stopTypingNotification)
				receiver.Send <- payload
			}
			continue
		}

		if raw.Type == "user_list_request" {
			sendOnlineUsers(c.ID)
			continue
		}

		log.Println("Unknown message type:", raw.Type)
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		// Extract the message type
		var raw struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &raw); err != nil {
			log.Println("Failed to parse message type:", err)
			continue
		}

		// Apply throttling only for typing events
		if raw.Type == "typing" || raw.Type == "stop_typing" {
			if !c.canSendMessage() {
				log.Println("Throttled message:", string(msg))
				continue
			}
		}

		// Send message
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("WebSocket write error for user %s: %v", c.ID, err)
			break
		}
	}
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Client) canSendMessage() bool {
	if time.Since(c.lastSent) < throttleRate {
		return false
	}
	c.lastSent = time.Now()
	return true
}

func sendOnlineUsers(_ string) {
	// Query once for the full user list (avoids running many parallel DB queries
	// which can cause 'database is locked' errors under SQLite).
	rows, err := db.DB.Query(`
SELECT u.id,
	u.nickname,
	IFNULL(u.avatar, ''),
	CASE WHEN u.online_status = 1 THEN 1 ELSE 0 END AS is_online,
	MAX(m.created_at) as last_msg
FROM users u
LEFT JOIN messages m ON (
	(u.id = m.sender_id AND m.receiver_id = ?) OR
	(u.id = m.receiver_id AND m.sender_id = ?)
)
WHERE u.id != ?
GROUP BY u.id
ORDER BY
	is_online DESC,
	last_msg DESC NULLS LAST,
	u.nickname COLLATE NOCASE ASC
	`, "0", "0", "0") // placeholders; ordering is independent of requester

	if err != nil {
		log.Println("User fetch error:", err)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, nickname, avatar string
		var isOnline int
		var lastMsg sql.NullString

		if err := rows.Scan(&id, &nickname, &avatar, &isOnline, &lastMsg); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":        id,
			"nickname":  nickname,
			"avatar":    avatar,
			"is_online": isOnline == 1,
		})
	}

	jsonUsers, _ := json.Marshal(users)
	update := models.Message{Type: "user_list", Content: string(jsonUsers)}
	payload, _ := json.Marshal(update)

	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	for _, client := range clients {
		// broadcast same payload to all clients
		sendToClient(client, payload)
	}
}

// sendToClient tries to send a payload to a client's Send channel without blocking
// the caller. It will attempt a non-blocking send first, then fall back to a
// goroutine that will time out after a short period. This avoids closing the
// connection for transient backpressure while still protecting the broadcaster.
func sendToClient(c *Client, payload []byte) {
	// fast path: if channel has capacity, send immediately
	select {
	case c.Send <- payload:
		return
	default:
	}

	// otherwise attempt a timed send in a goroutine
	go func() {
		select {
		case c.Send <- payload:
			return
		case <-time.After(600 * time.Millisecond):
			// timed out sending to slow client; drop the message silently
			return
		}
	}()
}
