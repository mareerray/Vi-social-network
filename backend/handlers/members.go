package handlers

import (
	"net/http"
	"social-network/backend/db"
	"social-network/backend/utils"
	"strconv"
	"strings"
)

// GetGroupMembersHandler returns a list of all members in a group
func GetGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	// Extract group ID from path (/api/groups/{id}/members)
	groupIDStr := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	groupIDStr = strings.TrimSuffix(groupIDStr, "/members")

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	// Query members
	rows, err := db.DB.Query(`
		SELECT u.id, u.nickname, u.first_name, u.last_name, u.avatar
		FROM users u
		INNER JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = ?
	`, groupID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch group members")
		return
	}
	defer rows.Close()

	type Member struct {
		ID       int64  `json:"id"`
		Nickname string `json:"nickname"`
		FullName string `json:"full_name"`
		Avatar   string `json:"avatar"`
	}

	var members []Member
	for rows.Next() {
		var m Member
		var nickname, firstName, lastName, avatar string
		if err := rows.Scan(&m.ID, &nickname, &firstName, &lastName, &avatar); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to read member data")
			return
		}
		m.Nickname = nickname
		m.FullName = strings.TrimSpace(firstName + " " + lastName)
		m.Avatar = utils.AbsURL(r, avatar)
		members = append(members, m)
	}

	utils.JSON(w, http.StatusOK, members)
}
