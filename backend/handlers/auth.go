package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social-network/backend/db"
	"social-network/backend/models"
	"social-network/backend/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid input"}`, http.StatusBadRequest)
		return
	}

	// Require essential fields; nickname may be omitted and will be auto-generated
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, `{"error":"Missing required fields"}`, http.StatusBadRequest)
		return
	}

	// Auto-generate a nickname from email local-part if none provided
	if req.Nickname == "" {
		local := strings.Split(req.Email, "@")[0]
		// sanitize local part
		local = strings.ToLower(strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
				return r
			}
			return -1
		}, local))
		if local == "" {
			local = "user"
		}
		base := local
		// ensure uniqueness by appending numeric suffix if needed
		suffix := 0
		for {
			candidate := base
			if suffix > 0 {
				candidate = fmt.Sprintf("%s%d", base, suffix)
			}
			var exInt int
			err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE nickname = ?)", candidate).Scan(&exInt)
			if err != nil {
				log.Println("DB check error while generating nickname:", err)
				http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
				return
			}
			if exInt == 0 {
				req.Nickname = candidate
				break
			}
			suffix++
		}
	}

	// Ensure profile_type is valid; default to 'public' when empty/invalid
	if req.ProfileType == "" {
		req.ProfileType = "public"
	} else {
		t := strings.ToLower(strings.TrimSpace(req.ProfileType))
		if t != "public" && t != "private" {
			req.ProfileType = "public"
		} else {
			req.ProfileType = t
		}
	}

	var existsInt int
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ? OR nickname = ?)", req.Email, req.Nickname).Scan(&existsInt)
	if err != nil {
		log.Println("DB check error:", err)
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if existsInt != 0 {
		http.Error(w, `{"error":"Email or nickname already in use"}`, http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password hash error:", err)
		http.Error(w, `{"error":"Server error"}`, http.StatusInternalServerError)
		return
	}

	result, err := db.DB.Exec(`
		INSERT INTO users (email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		strings.ToLower(req.Email), hashedPassword, req.FirstName, req.LastName, req.DateOfBirth,
		req.Avatar, req.Nickname, req.About, req.ProfileType,
	)
	if err != nil {
		// Detailed log for debugging
		log.Printf("User creation error: %v; params: email=%s, nickname=%s, dob=%s", err, req.Email, req.Nickname, req.DateOfBirth)
		// return generic message to client
		http.Error(w, `{"error":"Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve inserted user ID: %v", err)
		http.Error(w, `{"error":"Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("User registered successfully with ID: %d", userID)

	// Create session for the newly registered user (auto-login)
	sessionToken := uuid.New().String()
	expiry := time.Now().Add(24 * time.Hour)
	_, err = db.DB.Exec(
		`INSERT INTO sessions (user_id, cookie_token, expiry) VALUES (?, ?, ?)`,
		userID, sessionToken, expiry,
	)
	if err != nil {
		log.Printf("Session creation error after registration: %v", err)
		// still return success for user creation, but log session error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.RegisterResponse{Message: "Registration successful"})
		return
	}

	// Send cookie to browser
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful", "user_id": strconv.FormatInt(userID, 10)})
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Identifier == "" || req.Password == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var userID int64
	var hashedPassword string
	err = db.DB.QueryRow(`
		SELECT id, password FROM users WHERE email = ? OR nickname = ?`,
		req.Identifier, req.Identifier).Scan(&userID, &hashedPassword)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Login query error: %v", err)
		http.Error(w, `{"error":"Server error"}`, http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Clear old sessions for this user
	_, _ = db.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)

	// Create new session
	sessionToken := uuid.New().String()
	expiry := time.Now().Add(24 * time.Hour)
	_, err = db.DB.Exec(
		`INSERT INTO sessions (user_id, cookie_token, expiry) VALUES (?, ?, ?)`,
		userID, sessionToken, expiry,
	)
	if err != nil {
		log.Printf("Session creation error: %v", err)
		http.Error(w, `{"error":"Server error"}`, http.StatusInternalServerError)
		return
	}

	// Set user online status
	_, err = db.DB.Exec("UPDATE users SET online_status = 1 WHERE id = ?", userID)
	if err != nil {
		log.Printf("Failed to update online status for user %d: %v", userID, err)
		// Non-fatal error, so we don't abort the login
	}

	// Send cookie to browser
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   r.TLS != nil, // only secure in HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.LoginResponse{UserID: strconv.FormatInt(userID, 10)})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Get user ID before deleting the session
		var userID int64
		err := db.DB.QueryRow("SELECT user_id FROM sessions WHERE cookie_token = ?", cookie.Value).Scan(&userID)

		// Delete the session
		db.DB.Exec("DELETE FROM sessions WHERE cookie_token = ?", cookie.Value)

		// Update online status if user ID was found
		if err == nil {
			_, err = db.DB.Exec("UPDATE users SET online_status = 0 WHERE id = ?", userID)
			if err != nil {
				log.Printf("Failed to update online status on logout for user %d: %v", userID, err)
			}
		}

		// Expire the cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "session_token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}
	w.WriteHeader(http.StatusOK)
}

func CheckSessionHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, `{"error":"No session"}`, http.StatusUnauthorized)
		return
	}

	var userIDInt int64
	var expiry time.Time
	err = db.DB.QueryRow("SELECT user_id, expiry FROM sessions WHERE cookie_token = ?", cookie.Value).
		Scan(&userIDInt, &expiry)
	if err != nil || time.Now().After(expiry) {
		http.Error(w, `{"error":"Invalid or expired session"}`, http.StatusUnauthorized)
		return
	}

	// fetch basic profile for the user
	var nickname, avatar sql.NullString
	err = db.DB.QueryRow("SELECT nickname, avatar FROM users WHERE id = ?", userIDInt).Scan(&nickname, &avatar)
	if err != nil {
		// If profile fetch fails, still return the user id
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"user_id": strconv.FormatInt(userIDInt, 10)})
		return
	}

	resp := map[string]string{"user_id": strconv.FormatInt(userIDInt, 10)}
	if nickname.Valid {
		resp["nickname"] = nickname.String
	}
	if avatar.Valid {
		resp["avatar"] = utils.AbsURL(r, avatar.String)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func CleanupSessions() {
	_, err := db.DB.Exec("DELETE FROM sessions WHERE expiry < ?", time.Now())
	if err != nil {
		log.Printf("Session cleanup error: %v", err)
	}
}
