package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"social-network/backend/db"
	"social-network/backend/utils"
)

// AuthMiddleware validates the session cookie and places the user ID into the request context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userIDInt int64
		var expiry time.Time
		err = db.DB.QueryRow("SELECT user_id, expiry FROM sessions WHERE cookie_token = ?", cookie.Value).
			Scan(&userIDInt, &expiry)
		if err != nil || time.Now().After(expiry) {
			// remove cookie client-side
			http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1})
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// store user id as string in context for consistency with handlers
		ctx := context.WithValue(r.Context(), utils.UserIDKey, strconv.FormatInt(userIDInt, 10))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
