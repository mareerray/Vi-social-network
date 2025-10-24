package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"social-network/backend/db"
	"social-network/backend/utils"
	"strconv"
	"strings"
)

// POST /api/follow - send follow request (handles public/private profile logic)
func FollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		TargetID int64 `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check target profile type
	var profileType string
	err = db.DB.QueryRow("SELECT profile_type FROM users WHERE id = ?", payload.TargetID).Scan(&profileType)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "User not found")
		return
	}

	profileType = strings.ToLower(profileType)
	if profileType == "public" {
		// Auto-follow
		_, err := db.DB.Exec("INSERT OR IGNORE INTO followers (follower_id, followed_id, created_at) VALUES (?, ?, datetime('now'))", userID, payload.TargetID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to follow")
			return
		}
		// create notification for the target user about the new follower
		_ = Notify(payload.TargetID, userID, "new_follower", map[string]interface{}{"follower_id": userID, "url": fmt.Sprintf("/profile/%d", userID)})
		utils.JSON(w, http.StatusOK, map[string]string{"status": "followed"})
		return
	}
	// Private: create follow request
	_, err = db.DB.Exec("INSERT OR IGNORE INTO follow_requests (sender_id, receiver_id, status, created_at) VALUES (?, ?, 'pending', datetime('now'))", userID, payload.TargetID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to send request")
		return
	}
	// notify target about follow request
	_ = Notify(payload.TargetID, userID, "follow_request", map[string]interface{}{"requester_id": userID, "url": fmt.Sprintf("/profile/%d/requests", userID)})
	utils.JSON(w, http.StatusOK, map[string]string{"status": "requested"})
}

// POST /api/follow/accept - accept request
func AcceptFollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		SenderID int64 `json:"sender_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	// Accept request
	res, err := db.DB.Exec("UPDATE follow_requests SET status='accepted' WHERE sender_id=? AND receiver_id=? AND status='pending'", payload.SenderID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		utils.Error(w, http.StatusBadRequest, "No pending request")
		return
	}
	// Add to followers table
	_, err = db.DB.Exec("INSERT OR IGNORE INTO followers (follower_id, followed_id, created_at) VALUES (?, ?, datetime('now'))", payload.SenderID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	// notify sender their request was accepted
	_ = Notify(payload.SenderID, userID, "follow_request_accepted", map[string]interface{}{"follower_id": userID, "url": fmt.Sprintf("/profile/%d", userID)})
	utils.JSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

// POST /api/follow/decline - decline request
func DeclineFollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		SenderID int64 `json:"sender_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	res, err := db.DB.Exec("UPDATE follow_requests SET status='declined' WHERE sender_id=? AND receiver_id=? AND status='pending'", payload.SenderID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		utils.Error(w, http.StatusBadRequest, "No pending request")
		return
	}
	// notify sender their request was declined
	_ = Notify(payload.SenderID, userID, "follow_request_declined", map[string]interface{}{"follower_id": userID, "url": fmt.Sprintf("/profile/%d", userID)})
	utils.JSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

// POST /api/unfollow - unfollow a user
func UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := utils.GetUserIDFromContext(r)
	if userIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload struct {
		TargetID int64 `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	_, err = db.DB.Exec("DELETE FROM followers WHERE follower_id=? AND followed_id=?", userID, payload.TargetID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "unfollowed"})
}

// GET /api/follow/requests - list pending follow requests for current user
func ListRequests(w http.ResponseWriter, r *http.Request) {
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
	rows, err := db.DB.Query(`
		SELECT fr.id, fr.sender_id, fr.created_at, u.nickname, u.avatar
		FROM follow_requests fr
		JOIN users u ON u.id = fr.sender_id
		WHERE fr.receiver_id=? AND fr.status='pending'
		ORDER BY fr.created_at DESC`, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query requests")
		return
	}
	defer rows.Close()

	type req struct {
		ID             int64  `json:"id"`
		SenderID       int64  `json:"sender_id"`
		SenderNickname string `json:"sender_nickname"`
		SenderAvatar   string `json:"sender_avatar"`
		Created        string `json:"created_at"`
	}
	var out []req
	for rows.Next() {
		var ritem req
		var created sql.NullString
		var nickname, avatar sql.NullString
		if err := rows.Scan(&ritem.ID, &ritem.SenderID, &created, &nickname, &avatar); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to read request")
			return
		}
		ritem.Created = created.String
		ritem.SenderNickname = nickname.String
		ritem.SenderAvatar = utils.AbsURL(r, avatar.String)
		out = append(out, ritem)
	}
	utils.JSON(w, http.StatusOK, out)
}

// GET /api/follow/status?target_id=<id>
func FollowStatusHandler(w http.ResponseWriter, r *http.Request) {
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
	targetParam := r.URL.Query().Get("target_id")
	if targetParam == "" {
		utils.Error(w, http.StatusBadRequest, "Missing target_id")
		return
	}
	targetID, err := strconv.ParseInt(targetParam, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid target_id")
		return
	}

	status := map[string]bool{
		"following":       false,
		"request_pending": false,
	}

	if userID == targetID {
		utils.JSON(w, http.StatusOK, status)
		return
	}

	var count int
	if err := db.DB.QueryRow("SELECT COUNT(1) FROM followers WHERE follower_id=? AND followed_id=?", userID, targetID).Scan(&count); err == nil && count > 0 {
		status["following"] = true
	}

	if err := db.DB.QueryRow("SELECT COUNT(1) FROM follow_requests WHERE sender_id=? AND receiver_id=? AND status='pending'", userID, targetID).Scan(&count); err == nil && count > 0 {
		status["request_pending"] = true
	}

	utils.JSON(w, http.StatusOK, status)
}
