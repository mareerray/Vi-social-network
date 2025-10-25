package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"social-network/backend/db"
	"social-network/backend/utils"
	"strconv"
	"strings"
)

// CreateGroupHandler - POST { name, description }
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	res, err := db.DB.Exec("INSERT INTO groups (owner_id, name, description) VALUES (?, ?, ?)", userID, payload.Name, payload.Description)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create group")
		return
	}
	id, _ := res.LastInsertId()
	// add owner as member
	db.DB.Exec("INSERT OR IGNORE INTO group_members (group_id, user_id, role) VALUES (?, ?, 'owner')", id, userID)
	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": "created", "group_id": id})
}

// ListGroupsHandler - GET /api/groups
func ListGroupsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, owner_id, name, description, created_at FROM groups ORDER BY created_at DESC")
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list groups")
		return
	}
	defer rows.Close()
	type G struct {
		ID          int64  `json:"id"`
		OwnerID     int64  `json:"owner_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Created     string `json:"created_at"`
	}
	var out []G
	for rows.Next() {
		var g G
		rows.Scan(&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Created)
		out = append(out, g)
	}
	utils.JSON(w, http.StatusOK, out)
}

// GetGroupHandler - GET /api/group?id=<id>
func GetGroupHandler(w http.ResponseWriter, r *http.Request) {
	// Add panic recovery
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Panic in GetGroupHandler: %v", rec)
			utils.Error(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Check database connection
	if db.DB == nil {
		utils.Error(w, http.StatusInternalServerError, "Database not initialized")
		return
	}
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		utils.Error(w, http.StatusBadRequest, "Missing id")
		return
	}
	gid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var g struct {
		ID          int64          `json:"id"`
		OwnerID     int64          `json:"owner_id"`
		Name        string         `json:"name"`
		Description sql.NullString `json:"description"`
		Created     string         `json:"created_at"`
	}

	err = db.DB.QueryRow("SELECT id, owner_id, name, description, created_at FROM groups WHERE id = ?", gid).Scan(&g.ID, &g.OwnerID, &g.Name, &g.Description, &g.Created)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Group not found")
		return
	}
	// get members count
	var memberCount int
	err = db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id = ?", gid).Scan(&memberCount)
	if err != nil {
		// Handle the error - either log it or set default
		log.Printf("Failed to get member count for group %d: %v", gid, err)
		memberCount = 0 // Set safe default
	}
	// load members list (id, nickname, full_name, avatar) so frontend can filter invite candidates
	membersRows, mErr := db.DB.Query(`
		SELECT u.id, u.nickname, u.first_name, u.last_name, u.avatar
		FROM users u
		INNER JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = ?
	`, gid)
	var membersList []map[string]interface{}
	if mErr == nil {
		defer membersRows.Close()
		for membersRows.Next() {
			var id int64
			var nick, firstName, lastName, avatar sql.NullString
			if scanErr := membersRows.Scan(&id, &nick, &firstName, &lastName, &avatar); scanErr == nil {
				full := strings.TrimSpace(firstName.String + " " + lastName.String)
				membersList = append(membersList, map[string]interface{}{
					"id":        id,
					"nickname":  nick.String,
					"full_name": full,
					"avatar":    utils.AbsURL(r, avatar.String),
				})
			}
		}
	} else {
		log.Printf("Failed to load members for group %d: %v", gid, mErr)
	}
	// convert sql.NullString to plain string for JSON response
	groupObj := map[string]interface{}{
		"id":          g.ID,
		"owner_id":    g.OwnerID,
		"name":        g.Name,
		"description": g.Description.String,
		"created_at":  g.Created,
	}
	resp := map[string]interface{}{
		"group":        groupObj,
		"members":      memberCount,
		"members_list": membersList,
	}
	utils.JSON(w, http.StatusOK, resp)
}

// InviteHandler - POST { group_id, invitee_id }
func InviteHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	inviter, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		GroupID   int64 `json:"group_id"`
		InviteeID int64 `json:"invitee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// check if inviter is a member
	var memberCount int
	if err := db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id = ? AND user_id = ?",
		payload.GroupID, inviter).Scan(&memberCount); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}
	if memberCount == 0 {
		utils.Error(w, http.StatusForbidden, "Only group members can invite")
		return
	}
	// deduplicate pending invites
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_invites WHERE group_id=? AND invitee_id=? AND status='pending'", payload.GroupID, payload.InviteeID).Scan(&cnt)
	if cnt > 0 {
		utils.JSON(w, http.StatusOK, map[string]string{"status": "already_pending"})
		return
	}
	res, err := db.DB.Exec("INSERT INTO group_invites (group_id, inviter_id, invitee_id) VALUES (?, ?, ?)", payload.GroupID, inviter, payload.InviteeID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to invite")
		return
	}
	id, _ := res.LastInsertId()
	// notify the invitee about the invite
	_ = Notify(payload.InviteeID, inviter, "group_invite", map[string]interface{}{"invite_id": id, "group_id": payload.GroupID, "url": fmt.Sprintf("/groups/%d", payload.GroupID)})
	utils.JSON(w, http.StatusOK, map[string]string{"status": "invited"})
}

// RespondInviteHandler - POST { invite_id, action: accept|decline }
func RespondInviteHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		InviteID int64  `json:"invite_id"`
		Action   string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	var invite struct {
		GroupID   int64
		InviteeID int64
		InviterID int64
	}
	err := db.DB.QueryRow("SELECT group_id, invitee_id, inviter_id FROM group_invites WHERE id = ? AND status = 'pending'", payload.InviteID).Scan(&invite.GroupID, &invite.InviteeID, &invite.InviterID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invite")
		return
	}
	if invite.InviteeID != userID {
		utils.Error(w, http.StatusForbidden, "Not allowed")
		return
	}
	if payload.Action == "accept" {
		db.DB.Exec("UPDATE group_invites SET status='accepted' WHERE id=?", payload.InviteID)
		db.DB.Exec("INSERT OR IGNORE INTO group_members (group_id, user_id) VALUES (?,?)", invite.GroupID, userID)
		// notify inviter that invite was accepted
		_ = Notify(invite.InviterID, userID, "group_invite_response", map[string]interface{}{"invite_id": payload.InviteID, "status": "accepted", "group_id": invite.GroupID, "url": fmt.Sprintf("/groups/%d", invite.GroupID)})
		utils.JSON(w, http.StatusOK, map[string]string{"status": "accepted"})
		return
	}
	db.DB.Exec("UPDATE group_invites SET status='declined' WHERE id=?", payload.InviteID)
	// notify inviter that invite was declined
	_ = Notify(invite.InviterID, userID, "group_invite_response", map[string]interface{}{"invite_id": payload.InviteID, "status": "declined", "group_id": invite.GroupID, "url": fmt.Sprintf("/groups/%d", invite.GroupID)})
	utils.JSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

// CreateGroupPostHandler - POST multipart/form with content & optional image
func CreateGroupPostHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	r.ParseMultipartForm(10 << 20)
	gidStr := r.FormValue("group_id")
	gid, _ := strconv.ParseInt(gidStr, 10, 64)
	content := r.FormValue("content")
	imageURL := ""
	file, fh, err := r.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()
		// save to backend/uploads/posts to match upload endpoint and served static path
		saveDir := filepath.Join("backend", "uploads", "posts")
		os.MkdirAll(saveDir, 0755)
		fname := fmt.Sprintf("group_%d_%s", gid, filepath.Base(fh.Filename))
		dstPath := filepath.Join(saveDir, fname)
		dst, _ := os.Create(dstPath)
		defer dst.Close()
		io.Copy(dst, file)
		imageURL = "/uploads/posts/" + fname
	}
	// ensure user is a member
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", gid, userID).Scan(&cnt)
	if cnt == 0 {
		utils.Error(w, http.StatusForbidden, "Not a member")
		return
	}
	_, err = db.DB.Exec("INSERT INTO group_posts (group_id, author_id, content, image_url) VALUES (?, ?, ?, ?)", gid, userID, content, imageURL)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "created"})
}

// ListGroupPostsHandler - GET /api/group/posts?group_id=<id>
func ListGroupPostsHandler(w http.ResponseWriter, r *http.Request) {
	gidStr := r.URL.Query().Get("group_id")
	if gidStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing group_id")
		return
	}
	gid, _ := strconv.ParseInt(gidStr, 10, 64)
	rows, err := db.DB.Query("SELECT id, group_id, author_id, content, image_url, created_at FROM group_posts WHERE group_id = ? ORDER BY created_at DESC", gid)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	defer rows.Close()
	type P struct {
		ID       int64  `json:"id"`
		GroupID  int64  `json:"group_id"`
		AuthorID int64  `json:"author_id"`
		Content  string `json:"content"`
		Image    string `json:"image_url"`
		Created  string `json:"created_at"`
	}
	var out []P
	for rows.Next() {
		var p P
		rows.Scan(&p.ID, &p.GroupID, &p.AuthorID, &p.Content, &p.Image, &p.Created)
		out = append(out, p)
	}
	utils.JSON(w, http.StatusOK, out)
}

// AddGroupCommentHandler - POST { post_id, content }
func AddGroupCommentHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		PostID  int64  `json:"post_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// check membership by looking up post's group
	var gid int64
	err := db.DB.QueryRow("SELECT group_id FROM group_posts WHERE id = ?", payload.PostID).Scan(&gid)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid post")
		return
	}
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", gid, userID).Scan(&cnt)
	if cnt == 0 {
		utils.Error(w, http.StatusForbidden, "Not a member")
		return
	}
	_, err = db.DB.Exec("INSERT INTO group_comments (post_id, user_id, content) VALUES (?, ?, ?)", payload.PostID, userID, payload.Content)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateEventHandler - POST { group_id, title, description, event_time }
func CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		GroupID     int64  `json:"group_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		EventTime   string `json:"event_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// ensure creator is member
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", payload.GroupID, userID).Scan(&cnt)
	if cnt == 0 {
		utils.Error(w, http.StatusForbidden, "Not a member")
		return
	}
	_, err := db.DB.Exec("INSERT INTO events (group_id, creator_id, title, description, event_time) VALUES (?, ?, ?, ?, ?)", payload.GroupID, userID, payload.Title, payload.Description, payload.EventTime)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	// notify group members about the new event (persist notifications)
	rows, err := db.DB.Query("SELECT user_id FROM group_members WHERE group_id = ? AND user_id != ?", payload.GroupID, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var mid int64
			if err := rows.Scan(&mid); err == nil {
				// create a simple JSON payload with event info
				data := map[string]interface{}{"group_id": payload.GroupID, "title": payload.Title, "event_time": payload.EventTime, "url": fmt.Sprintf("/groups/%d", payload.GroupID)}
				_ = Notify(mid, userID, "group_event", data)
			}
		}
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "created"})
}

// VoteEventHandler - POST { event_id, vote }
func VoteEventHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		EventID int64  `json:"event_id"`
		Vote    string `json:"vote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// upsert vote
	db.DB.Exec("INSERT OR REPLACE INTO event_votes (id, event_id, user_id, vote) VALUES ((SELECT id FROM event_votes WHERE event_id=? AND user_id=?), ?, ?, ?)", payload.EventID, userID, payload.EventID, userID, payload.Vote)
	utils.JSON(w, http.StatusOK, map[string]string{"status": "voted"})
}

// ListEventsHandler - GET /api/group/events?group_id=<id>
// Returns events for a group including aggregated vote counts and current user's vote
func ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	gidStr := r.URL.Query().Get("group_id")
	if gidStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing group_id")
		return
	}
	gid, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid group_id")
		return
	}
	// ensure user is a member of the group
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", gid, userID).Scan(&cnt)
	if cnt == 0 {
		utils.Error(w, http.StatusForbidden, "Not a member")
		return
	}

	rows, err := db.DB.Query("SELECT id, group_id, creator_id, title, description, event_time, created_at FROM events WHERE group_id = ? ORDER BY created_at DESC", gid)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query events")
		return
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var id int64
		var groupID int64
		var creatorID int64
		var title string
		var description sql.NullString
		var eventTime sql.NullString
		var created string
		rows.Scan(&id, &groupID, &creatorID, &title, &description, &eventTime, &created)

		// aggregate votes
		voteRows, _ := db.DB.Query("SELECT vote, COUNT(1) FROM event_votes WHERE event_id = ? GROUP BY vote", id)
		votes := map[string]int{}
		for voteRows.Next() {
			var v string
			var c int
			voteRows.Scan(&v, &c)
			votes[v] = c
		}
		voteRows.Close()

		// current user's vote
		var myVote sql.NullString
		db.DB.QueryRow("SELECT vote FROM event_votes WHERE event_id=? AND user_id=?", id, userID).Scan(&myVote)

		out = append(out, map[string]interface{}{"id": id, "group_id": groupID, "creator_id": creatorID, "title": title, "description": description.String, "event_time": eventTime.String, "created_at": created, "votes": votes, "my_vote": myVote.String})
	}
	utils.JSON(w, http.StatusOK, out)
}

// CheckMembershipHandler - GET /api/group/membership?group_id=<id>
func CheckMembershipHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	gidStr := r.URL.Query().Get("group_id")
	if gidStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing group_id")
		return
	}
	gid, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid group_id")
		return
	}
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", gid, userID).Scan(&cnt)
	utils.JSON(w, http.StatusOK, map[string]bool{"is_member": cnt > 0})
}

// RequestToJoinHandler - POST { group_id }
func RequestToJoinHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		GroupID int64 `json:"group_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	// ensure group exists
	var ownerID int64
	err := db.DB.QueryRow("SELECT owner_id FROM groups WHERE id = ?", payload.GroupID).Scan(&ownerID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid group")
		return
	}
	// ensure not already a member
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_members WHERE group_id=? AND user_id=?", payload.GroupID, userID).Scan(&cnt)
	if cnt > 0 {
		utils.Error(w, http.StatusBadRequest, "Already a member")
		return
	}
	// insert request (unique constraint prevents duplicates)
	_, err = db.DB.Exec("INSERT OR IGNORE INTO group_requests (group_id, requester_id) VALUES (?,?)", payload.GroupID, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to request to join")
		return
	}
	// notify owner
	_ = Notify(ownerID, userID, "group_join_request", map[string]interface{}{"group_id": payload.GroupID, "requester_id": userID, "url": fmt.Sprintf("/groups/%d/requests", payload.GroupID)})
	utils.JSON(w, http.StatusOK, map[string]string{"status": "requested"})
}

// RespondRequestHandler - POST { request_id, action: accept|decline }
func RespondRequestHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		RequestID int64  `json:"request_id"`
		Action    string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	var req struct {
		GroupID     int64
		RequesterID int64
	}
	err := db.DB.QueryRow("SELECT group_id, requester_id FROM group_requests WHERE id = ? AND status = 'pending'", payload.RequestID).Scan(&req.GroupID, &req.RequesterID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// ensure current user is the owner of the group
	var ownerID int64
	err = db.DB.QueryRow("SELECT owner_id FROM groups WHERE id = ?", req.GroupID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		utils.Error(w, http.StatusForbidden, "Not allowed")
		return
	}
	if payload.Action == "accept" {
		db.DB.Exec("UPDATE group_requests SET status='accepted' WHERE id=?", payload.RequestID)
		db.DB.Exec("INSERT OR IGNORE INTO group_members (group_id, user_id) VALUES (?,?)", req.GroupID, req.RequesterID)
		_ = Notify(req.RequesterID, userID, "group_join_response", map[string]interface{}{"request_id": payload.RequestID, "status": "accepted", "group_id": req.GroupID, "url": fmt.Sprintf("/groups/%d", req.GroupID)})
		utils.JSON(w, http.StatusOK, map[string]string{"status": "accepted"})
		return
	}
	db.DB.Exec("UPDATE group_requests SET status='declined' WHERE id=?", payload.RequestID)
	_ = Notify(req.RequesterID, userID, "group_join_response", map[string]interface{}{"request_id": payload.RequestID, "status": "declined", "group_id": req.GroupID, "url": fmt.Sprintf("/groups/%d", req.GroupID)})
	utils.JSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

// ListRequestsHandler - GET /api/group/requests?group_id=<id> (owner only)
func ListRequestsHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	gidStr := r.URL.Query().Get("group_id")
	if gidStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing group_id")
		return
	}
	gid, _ := strconv.ParseInt(gidStr, 10, 64)
	// ensure current user is owner
	var ownerID int64
	err := db.DB.QueryRow("SELECT owner_id FROM groups WHERE id = ?", gid).Scan(&ownerID)
	if err != nil || ownerID != userID {
		utils.Error(w, http.StatusForbidden, "Not allowed")
		return
	}
	rows, err := db.DB.Query(`SELECT gr.id, gr.requester_id, u.nickname, u.avatar, gr.status, gr.created_at
		FROM group_requests gr
		JOIN users u ON gr.requester_id = u.id
		WHERE gr.group_id = ? AND gr.status = 'pending'
		ORDER BY gr.created_at DESC`, gid)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query requests")
		return
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id int64
		var requester int64
		var nickname sql.NullString
		var avatar sql.NullString
		var status string
		var created string
		rows.Scan(&id, &requester, &nickname, &avatar, &status, &created)
		out = append(out, map[string]interface{}{"id": id, "requester_id": requester, "nickname": nickname.String, "avatar": utils.AbsURL(r, avatar.String), "status": status, "created_at": created})
	}
	utils.JSON(w, http.StatusOK, out)
}

// GetRequestStatusHandler - GET /api/group/request/status?group_id=<id>
// returns { has_pending: true|false }
func GetRequestStatusHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	gidStr := r.URL.Query().Get("group_id")
	if gidStr == "" {
		utils.Error(w, http.StatusBadRequest, "Missing group_id")
		return
	}
	gid, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid group_id")
		return
	}
	var cnt int
	db.DB.QueryRow("SELECT COUNT(1) FROM group_requests WHERE group_id=? AND requester_id=? AND status='pending'", gid, userID).Scan(&cnt)
	utils.JSON(w, http.StatusOK, map[string]bool{"has_pending": cnt > 0})
}
