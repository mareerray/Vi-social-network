package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"social-network/backend/db"
	"social-network/backend/utils"
)

// ListGroupMessagesHandler returns recent messages for a group (membership required)
func ListGroupMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(utils.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	gidStr := r.URL.Query().Get("group_id")
	if gidStr == "" {
		http.Error(w, "group_id required", http.StatusBadRequest)
		return
	}
	gid, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid group_id", http.StatusBadRequest)
		return
	}

	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		http.Error(w, "invalid user", http.StatusInternalServerError)
		return
	}

	// verify membership
	var cnt int
	err = db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", gid, uid).Scan(&cnt)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if cnt == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// pagination params
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	if limit > 200 {
		limit = 200
	}

	beforeIDStr := r.URL.Query().Get("before_id")
	var rows *sql.Rows
	if beforeIDStr != "" {
		beforeID, errConv := strconv.ParseInt(beforeIDStr, 10, 64)
		if errConv != nil {
			http.Error(w, "invalid before_id", http.StatusBadRequest)
			return
		}
		rows, err = db.DB.Query(`SELECT gm.id, gm.sender_id, gm.content, gm.created_at, u.nickname FROM group_messages gm JOIN users u ON u.id = gm.sender_id WHERE gm.group_id = ? AND gm.id < ? ORDER BY gm.id DESC LIMIT ?`, gid, beforeID, limit)
	} else {
		rows, err = db.DB.Query(`SELECT gm.id, gm.sender_id, gm.content, gm.created_at, u.nickname FROM group_messages gm JOIN users u ON u.id = gm.sender_id WHERE gm.group_id = ? ORDER BY gm.id DESC LIMIT ?`, gid, limit)
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type msg struct {
		ID         int64          `json:"id"`
		SenderID   int64          `json:"sender_id"`
		Content    string         `json:"content"`
		CreatedAt  sql.NullString `json:"created_at"`
		SenderName string         `json:"sender_name"`
	}

	var out []msg
	for rows.Next() {
		var m msg
		if err := rows.Scan(&m.ID, &m.SenderID, &m.Content, &m.CreatedAt, &m.SenderName); err == nil {
			out = append(out, m)
		}
	}

	// reverse to oldest-first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	utils.JSON(w, http.StatusOK, out)
}
