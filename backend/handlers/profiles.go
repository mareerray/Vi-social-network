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

// GET /api/profile/<id> - if id is omitted, returns current user's profile (requires auth cookie)
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	// The user making the request. Can be empty if not logged in.
	requestingUserIDStr := utils.GetUserIDFromContext(r)
	if requestingUserIDStr == "" {
		requestingUserIDStr = utils.GetUserIDFromSession(w, r)
	}
	var requestingID int64
	if requestingUserIDStr != "" {
		requestingID, _ = strconv.ParseInt(requestingUserIDStr, 10, 64)
	}

	// The user profile being requested.
	pathSuffix := strings.TrimPrefix(r.URL.Path, "/api/profile")
	targetUserIDStr := strings.Trim(pathSuffix, "/")
	if targetUserIDStr == "" {
		targetUserIDStr = strings.TrimSpace(r.URL.Query().Get("id"))
	}

	var targetID int64
	var err error

	// If no ID is in the URL, it means the user is requesting their own profile.
	if targetUserIDStr == "" {
		if requestingID == 0 {
			utils.Error(w, http.StatusUnauthorized, "Not logged in")
			return
		}
		targetID = requestingID
	} else {
		// An ID is in the URL, so parse it.
		targetID, err = strconv.ParseInt(targetUserIDStr, 10, 64)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
	}

	// At this point, we have the ID of the profile we want to view (targetID).
	// Now, let's fetch that user's basic info and privacy setting.
	var userBase struct {
		ID          int64
		Nickname    sql.NullString
		Avatar      sql.NullString
		ProfileType sql.NullString
	}

	err = db.DB.QueryRow(`SELECT id, nickname, avatar, profile_type FROM users WHERE id = ?`, targetID).
		Scan(&userBase.ID, &userBase.Nickname, &userBase.Avatar, &userBase.ProfileType)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.Error(w, http.StatusNotFound, "User not found")
			return
		}
		fmt.Println("error getting profile 1:", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	profileType := strings.ToLower(strings.TrimSpace(userBase.ProfileType.String))
	if profileType == "" {
		profileType = "public"
	}

	// Authorization Logic
	isOwnProfile := requestingID != 0 && targetID == requestingID
	canViewProfile := false

	// User is looking at their own profile
	if isOwnProfile {
		canViewProfile = true
	} else if profileType == "public" {
		// Profile is public, anyone can view
		canViewProfile = true
	} else if profileType == "private" {
		// Profile is private, check if the requester is an accepted follower.
		// This requires the requester to be logged in (requestingID != 0).
		if requestingID != 0 {
			var dummy int
			err := db.DB.QueryRow(`SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ?`, requestingID, targetID).Scan(&dummy)
			if err == nil {
				canViewProfile = true
			} else if err != sql.ErrNoRows {
				fmt.Println("error getting profile 2:", err)
				utils.Error(w, http.StatusInternalServerError, "Failed to check follow status")
				return
			}
		}
		// If requester is not logged in, canViewProfile remains false for private profiles.
	}

	if !canViewProfile {
		// User cannot view the full profile, send limited data
		limitedProfile := map[string]interface{}{
			"id":            userBase.ID,
			"nickname":      userBase.Nickname.String,
			"avatar":        utils.AbsURL(r, userBase.Avatar.String),
			"profile_type":  profileType,
			"is_accessible": false,
		}
		utils.JSON(w, http.StatusOK, limitedProfile)
		return
	}

	// If we get here, the user is authorized to see the full profile.
	var fullProfile struct {
		ID          int64
		Email       sql.NullString
		FirstName   sql.NullString
		LastName    sql.NullString
		DateOfBirth sql.NullString
		Avatar      sql.NullString
		Nickname    sql.NullString
		About       sql.NullString
		ProfileType sql.NullString
		// CreatedAt   sql.NullString
	}

	err = db.DB.QueryRow(`SELECT id, email, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_type FROM users WHERE id = ?`, targetID).
		Scan(&fullProfile.ID, &fullProfile.Email, &fullProfile.FirstName, &fullProfile.LastName, &fullProfile.DateOfBirth, &fullProfile.Avatar, &fullProfile.Nickname, &fullProfile.About, &fullProfile.ProfileType)
		// , created_at
		//  &fullProfile.CreatedAt
	if err != nil {
		fmt.Println("error getting profile 3:", err)
		utils.Error(w, http.StatusInternalServerError, "Failed to load profile data")
		return
	}

	if v := strings.ToLower(strings.TrimSpace(fullProfile.ProfileType.String)); v != "" {
		profileType = v
	}

	resp := map[string]interface{}{
		"id":            fullProfile.ID,
		"first_name":    fullProfile.FirstName.String,
		"last_name":     fullProfile.LastName.String,
		"date_of_birth": fullProfile.DateOfBirth.String,
		"avatar":        utils.AbsURL(r, fullProfile.Avatar.String),
		"nickname":      fullProfile.Nickname.String,
		"about":         fullProfile.About.String,
		"profile_type":  profileType,
		// "created_at":    fullProfile.CreatedAt.String,
		"is_accessible": true,
	}

	if isOwnProfile && fullProfile.Email.Valid {
		resp["email"] = fullProfile.Email.String
	}

	utils.JSON(w, http.StatusOK, resp)
}

// PUT /api/profile/update
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var payload struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DateOfBirth string `json:"date_of_birth"`
		Avatar      string `json:"avatar"`
		Nickname    string `json:"nickname"`
		About       string `json:"about"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	_, err := db.DB.Exec(`
		UPDATE users 
		SET first_name = ?, last_name = ?, date_of_birth = ?, avatar = ?, nickname = ?, about_me = ?
		WHERE id = ?`,
		payload.FirstName, payload.LastName, payload.DateOfBirth, payload.Avatar, payload.Nickname, payload.About, uid,
	)

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /api/profile/followers?id=<id>
func GetFollowersHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		utils.Error(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	rows, err := db.DB.Query(`
		SELECT u.id, u.nickname, u.avatar 
		FROM users u
		JOIN followers f ON u.id = f.follower_id
		WHERE f.followed_id = ?
		ORDER BY f.created_at DESC`,
		idParam,
	)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query followers")
		return
	}
	defer rows.Close()

	var followers []map[string]interface{}
	for rows.Next() {
		var id int
		var nickname, avatar sql.NullString
		if err := rows.Scan(&id, &nickname, &avatar); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to scan follower")
			return
		}
		followers = append(followers, map[string]interface{}{
			"id":       id,
			"nickname": nickname.String,
			"avatar":   utils.AbsURL(r, avatar.String),
		})
	}

	utils.JSON(w, http.StatusOK, followers)
}

// GET /api/profile/following?id=<id>
func GetFollowingHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		utils.Error(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	rows, err := db.DB.Query(`
		SELECT u.id, u.nickname, u.avatar 
		FROM users u
		JOIN followers f ON u.id = f.followed_id
		WHERE f.follower_id = ?
		ORDER BY f.created_at DESC`,
		idParam,
	)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to query following")
		return
	}
	defer rows.Close()

	var following []map[string]interface{}
	for rows.Next() {
		var id int
		var nickname, avatar sql.NullString
		if err := rows.Scan(&id, &nickname, &avatar); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to scan following")
			return
		}
		following = append(following, map[string]interface{}{
			"id":       id,
			"nickname": nickname.String,
			"avatar":   utils.AbsURL(r, avatar.String),
		})
	}

	utils.JSON(w, http.StatusOK, following)
}

// POST /api/profile/privacy
func TogglePrivacyHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var payload struct {
		ProfileType string `json:"profile_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.ProfileType != "public" && payload.ProfileType != "private" {
		utils.Error(w, http.StatusBadRequest, "Invalid profile type")
		return
	}

	_, err := db.DB.Exec(`UPDATE users SET profile_type = ? WHERE id = ?`, payload.ProfileType, uid)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update privacy")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"status": "success", "profile_type": payload.ProfileType})
}
