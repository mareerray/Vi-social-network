package utils

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

var db *sql.DB

// SetDB injects the database instance into the utils package
func SetDB(database *sql.DB) {
	db = database
}

// contextKey is a private type for context keys in session utils.
type contextKey string

// UserIDKey is used to store/retrieve the user ID in request context.
const UserIDKey contextKey = "userID"

// GetUserIDFromSession extracts the user ID from the session cookie
func GetUserIDFromSession(w http.ResponseWriter, r *http.Request) string {
	// use the same cookie name as the auth handlers: session_token
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return ""
	}

	var userIDInt int64
	// migrations create cookie_token and expiry columns; query cookie_token
	err = db.QueryRow("SELECT user_id FROM sessions WHERE cookie_token = ?", cookie.Value).Scan(&userIDInt)
	if err != nil {
		expireCookie(w, "session_token")
		return ""
	}

	return strconv.FormatInt(userIDInt, 10)
}

func expireCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// GetUserIDFromContext reads the user id string placed into the request context by AuthMiddleware
func GetUserIDFromContext(r *http.Request) string {
	if v := r.Context().Value(UserIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
