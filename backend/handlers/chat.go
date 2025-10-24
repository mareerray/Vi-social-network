package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/backend/db"
	"social-network/backend/models"
	"social-network/backend/utils"
	"strconv"
)

// GetAllUsers - Returns users sorted by: online first, then by last message time, then alphabetically
// This is REQUIRED by the project specs: "organized by the last message sent (just like discord)"
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(utils.UserIDKey).(string)

	// This query implements the Discord-like sorting requirement
	rows, err := db.DB.Query(`
		SELECT u.id, u.nickname,
		    CASE WHEN u.online_status = 1 THEN 1 ELSE 0 END AS is_online
		FROM users u
		LEFT JOIN messages m ON (u.id = m.sender_id OR u.id = m.receiver_id)
		    AND (m.sender_id = ? OR m.receiver_id = ?)
		WHERE u.id != ?
		GROUP BY u.id
		ORDER BY 
		    is_online DESC,                    -- Online users first
		    MAX(m.created_at) DESC,           -- Then by last message time
		    u.nickname COLLATE NOCASE ASC     -- Then alphabetically
		`, userID, userID, userID)

	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
		IsOnline bool   `json:"is_online"`
	}

	for rows.Next() {
		var u struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
			IsOnline bool   `json:"is_online"`
		}
		var isOnline int

		if err := rows.Scan(&u.ID, &u.Nickname, &isOnline); err == nil {
			u.IsOnline = isOnline == 1
			users = append(users, u)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetMessageHistory - Returns message history with proper pagination
// "Reload the last 10 messages and when scrolled up to see more messages"
func GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(utils.UserIDKey).(string)
	otherUserID := r.URL.Query().Get("user_id")
	offsetStr := r.URL.Query().Get("offset")

	if otherUserID == "" {
		http.Error(w, `{"error":"Missing user_id parameter"}`, http.StatusBadRequest)
		return
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsed
		}
	}

	// CRITICAL: Must use DESC order for pagination to work correctly
	// Frontend will reverse for display
	rows, err := db.DB.Query(`
		SELECT m.id, m.sender_id, u.nickname, m.receiver_id, m.content, m.created_at
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) 
			OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT 10 OFFSET ?`,
		userID, otherUserID, otherUserID, userID, offset)

	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.SenderName,
			&msg.ReceiverID,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	// Reverse so frontend gets oldest-first for display
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	json.NewEncoder(w).Encode(messages)
}
