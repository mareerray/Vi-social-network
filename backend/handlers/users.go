package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"social-network/backend/db"
	"social-network/backend/utils"
)

// PublicUsersHandler returns a sanitized list of users with relationship flags relative to the requester.
func PublicUsersHandler(w http.ResponseWriter, r *http.Request) {
	requesterIDStr := utils.GetUserIDFromContext(r)
	if requesterIDStr == "" {
		requesterIDStr = utils.GetUserIDFromSession(w, r)
	}

	if requesterIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	requesterID, err := strconv.ParseInt(requesterIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	followedIDs := make(map[int64]bool)
	rows, err := db.DB.Query("SELECT followed_id FROM followers WHERE follower_id = ?", requesterID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id sql.NullInt64
			if err := rows.Scan(&id); err == nil && id.Valid {
				followedIDs[id.Int64] = true
			}
		}
	}

	pendingIDs := make(map[int64]bool)
	reqRows, err := db.DB.Query("SELECT receiver_id FROM follow_requests WHERE sender_id = ? AND status = 'pending'", requesterID)
	if err == nil {
		defer reqRows.Close()
		for reqRows.Next() {
			var id sql.NullInt64
			if err := reqRows.Scan(&id); err == nil && id.Valid {
				pendingIDs[id.Int64] = true
			}
		}
	}

	type publicUser struct {
		ID             int64  `json:"id"`
		Nickname       string `json:"nickname"`
		DisplayName    string `json:"display_name"`
		Avatar         string `json:"avatar"`
		ProfileType    string `json:"profile_type"`
		IsSelf         bool   `json:"is_self"`
		IsFollowing    bool   `json:"is_following"`
		RequestPending bool   `json:"request_pending"`
	}

	usersRows, err := db.DB.Query(`
		SELECT id, nickname, first_name, last_name, avatar, profile_type
		FROM users
		ORDER BY LOWER(nickname), LOWER(first_name), LOWER(last_name)
	`)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	defer usersRows.Close()

	var result []publicUser
	for usersRows.Next() {
		var (
			id          int64
			nickname    sql.NullString
			firstName   sql.NullString
			lastName    sql.NullString
			avatar      sql.NullString
			profileType sql.NullString
		)

		if err := usersRows.Scan(&id, &nickname, &firstName, &lastName, &avatar, &profileType); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to read user")
			return
		}

		displayName := strings.TrimSpace(nickname.String)
		if displayName == "" {
			displayName = strings.TrimSpace(strings.Join([]string{firstName.String, lastName.String}, " "))
		}
		if displayName == "" {
			displayName = "Member"
		}

		user := publicUser{
			ID:             id,
			Nickname:       nickname.String,
			DisplayName:    displayName,
			Avatar:         utils.AbsURL(r, avatar.String),
			ProfileType:    strings.ToLower(profileType.String),
			IsSelf:         id == requesterID,
			IsFollowing:    followedIDs[id],
			RequestPending: pendingIDs[id],
		}

		result = append(result, user)
	}

	utils.JSON(w, http.StatusOK, result)
}
