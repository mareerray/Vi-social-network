package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/backend/bus"
	"social-network/backend/db"
	"social-network/backend/models"
	"social-network/backend/utils"
	"strconv"
)

// CreateNotification inserts a notification into DB for recipient. actorID may be 0.
func CreateNotification(recipientID int64, actorID int64, ntype string, data string) error {
	_, err := db.DB.Exec("INSERT INTO notifications (recipient_id, actor_id, type, data, is_read, created_at) VALUES (?, ?, ?, ?, 0, CURRENT_TIMESTAMP)", recipientID, actorID, ntype, data)
	return err
}

// Notify builds a consistent JSON payload, persists the notification, and
// publishes a realtime copy to the in-memory bus so connected websocket
// clients receive it.
func Notify(recipientID int64, actorID int64, ntype string, payload map[string]interface{}) error {

	// Add actor info (nickname, avatar) to the payload
	if actorID > 0 {
		var nickname, avatar sql.NullString
		err := db.DB.QueryRow("SELECT nickname, avatar FROM users WHERE id = ?", actorID).Scan(&nickname, &avatar)
		if err == nil {
			payload["actor_nickname"] = nickname.String
			payload["actor_avatar"] = utils.AbsURL(nil, avatar.String)
		}
	}
	// ensure payload is JSON string
	dataBytes, _ := json.Marshal(payload)
	dataStr := string(dataBytes)

	if err := CreateNotification(recipientID, actorID, ntype, dataStr); err != nil {
		log.Println("CreateNotification error:", err)
		// still try to publish realtime for a best-effort UX
	}

	// publish to bus for realtime delivery (best-effort)
	notif := map[string]interface{}{
		"type": ntype,
		"data": payload,
	}
	realtimeBytes, _ := json.Marshal(notif)
	bus.PublishNotification(recipientID, realtimeBytes)
	log.Printf("Published realtime notification type=%s to recipient=%d", ntype, recipientID)
	return nil
}

// GET /api/notifications - list recent notifications for current user
func ListNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rows, err := db.DB.Query("SELECT id, recipient_id, actor_id, type, data, is_read, created_at FROM notifications WHERE recipient_id=? ORDER BY created_at DESC LIMIT 50", userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query notifications")
		return
	}
	defer rows.Close()

	var out []models.Notification
	for rows.Next() {
		var n models.Notification
		var isRead int
		if err := rows.Scan(&n.ID, &n.RecipientID, &n.ActorID, &n.Type, &n.Data, &isRead, &n.CreatedAt); err != nil {
			continue
		}
		n.IsRead = isRead == 1
		out = append(out, n)
	}
	utils.JSON(w, http.StatusOK, out)
}

// POST /api/notifications/mark-read - mark notifications read (accepts optional id)
func MarkNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var payload struct {
		ID *int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		// allow empty body to mark all read
	}

	if payload.ID != nil {
		_, err := db.DB.Exec("UPDATE notifications SET is_read=1 WHERE id=? AND recipient_id=?", *payload.ID, userID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to mark read")
			return
		}
	} else {
		_, err := db.DB.Exec("UPDATE notifications SET is_read=1 WHERE recipient_id=?", userID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to mark read")
			return
		}
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
