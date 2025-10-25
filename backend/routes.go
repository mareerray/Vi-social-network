package main

import (
	"net/http"
	"os"
	"path/filepath"
	"social-network/backend/handlers"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux) {
	// Choose whether to serve production or dev folder
	staticDir := "./frontend/public"
	if _, err := os.Stat("./frontend/dist"); err == nil {
		staticDir = "./frontend/dist"
	}

	// Serve API, websocket, upload routes etc. (your existing handlers)
	mux.Handle("/ws", AuthMiddleware(http.HandlerFunc(HandleWebSocket)))
	mux.Handle("/api/messages/history", AuthMiddleware(http.HandlerFunc(handlers.GetMessageHistory)))
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/check-session", handlers.CheckSessionHandler)
	mux.Handle("/api/follow", AuthMiddleware(http.HandlerFunc(handlers.FollowHandler)))
	mux.Handle("/api/unfollow", AuthMiddleware(http.HandlerFunc(handlers.UnfollowHandler)))
	mux.Handle("/api/follow/accept", AuthMiddleware(http.HandlerFunc(handlers.AcceptFollowHandler)))
	mux.Handle("/api/follow/decline", AuthMiddleware(http.HandlerFunc(handlers.DeclineFollowHandler)))
	mux.Handle("/api/follow/requests", AuthMiddleware(http.HandlerFunc(handlers.ListRequests)))
	mux.Handle("/api/follow/status", AuthMiddleware(http.HandlerFunc(handlers.FollowStatusHandler)))
	mux.HandleFunc("/api/profile/", handlers.GetProfileHandler)
	mux.Handle("/api/profile/update", AuthMiddleware(http.HandlerFunc(handlers.UpdateProfileHandler)))
	mux.Handle("/api/profile/followers", AuthMiddleware(http.HandlerFunc(handlers.GetFollowersHandler)))
	mux.Handle("/api/profile/following", AuthMiddleware(http.HandlerFunc(handlers.GetFollowingHandler)))
	mux.Handle("/api/profile/privacy", AuthMiddleware(http.HandlerFunc(handlers.TogglePrivacyHandler)))
	mux.Handle("/api/posts/create", AuthMiddleware(http.HandlerFunc(handlers.CreatePostHandler)))
	mux.HandleFunc("/api/posts", handlers.ListFeedHandler)
	mux.HandleFunc("/api/users", handlers.PublicUsersHandler)
	mux.Handle("/api/notifications", AuthMiddleware(http.HandlerFunc(handlers.ListNotificationsHandler)))
	mux.Handle("/api/notifications/mark-read", AuthMiddleware(http.HandlerFunc(handlers.MarkNotificationsReadHandler)))
	mux.Handle("/api/group/create", AuthMiddleware(http.HandlerFunc(handlers.CreateGroupHandler)))
	mux.HandleFunc("/api/groups", handlers.ListGroupsHandler)
	mux.HandleFunc("/api/group", handlers.GetGroupHandler)
	mux.Handle("/api/group/invite", AuthMiddleware(http.HandlerFunc(handlers.InviteHandler)))
	mux.Handle("/api/group/invite/respond", AuthMiddleware(http.HandlerFunc(handlers.RespondInviteHandler)))
	mux.Handle("/api/group/membership", AuthMiddleware(http.HandlerFunc(handlers.CheckMembershipHandler)))
	mux.Handle("/api/group/request", AuthMiddleware(http.HandlerFunc(handlers.RequestToJoinHandler)))
	mux.Handle("/api/group/request/respond", AuthMiddleware(http.HandlerFunc(handlers.RespondRequestHandler)))
	mux.Handle("/api/group/requests", AuthMiddleware(http.HandlerFunc(handlers.ListRequestsHandler)))
	mux.Handle("/api/group/request/status", AuthMiddleware(http.HandlerFunc(handlers.GetRequestStatusHandler)))
	mux.Handle("/api/group/post/create", AuthMiddleware(http.HandlerFunc(handlers.CreateGroupPostHandler)))
	mux.HandleFunc("/api/group/posts", handlers.ListGroupPostsHandler)
	mux.Handle("/api/group/messages", AuthMiddleware(http.HandlerFunc(handlers.ListGroupMessagesHandler)))
	mux.Handle("/api/group/comment", AuthMiddleware(http.HandlerFunc(handlers.AddGroupCommentHandler)))
	mux.Handle("/api/group/comments", AuthMiddleware(http.HandlerFunc(handlers.ListGroupCommentsHandler)))
	mux.Handle("/api/group/event/create", AuthMiddleware(http.HandlerFunc(handlers.CreateEventHandler)))
	mux.Handle("/api/group/event/vote", AuthMiddleware(http.HandlerFunc(handlers.VoteEventHandler)))
	mux.Handle("/api/group/events", AuthMiddleware(http.HandlerFunc(handlers.ListEventsHandler)))
	mux.Handle("/api/posts/comment", AuthMiddleware(http.HandlerFunc(handlers.AddCommentHandler)))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("backend/uploads"))))
	mux.Handle("/api/upload", AuthMiddleware(http.HandlerFunc(handlers.UploadHandler)))

	// === SPA fallback handler for Vue Router ===
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Exclude API, WebSocket, and uploads routes
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/ws") ||
			strings.HasPrefix(r.URL.Path, "/uploads/") {
			return
		}

		// Resolve actual file path inside the staticDir
		requestedPath := filepath.Join(staticDir, r.URL.Path)
		if info, err := os.Stat(requestedPath); err == nil && !info.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
			// Serve the static file directly (e.g. /assets/main.css)
			fileServer.ServeHTTP(w, r)
			return
		}

		// For any other path, serve index.html so the SPA router can take over
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})
}
