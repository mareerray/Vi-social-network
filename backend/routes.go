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
			mux.ServeHTTP(w, r)
			return
		}

		// Resolve actual file path
		requestedPath := filepath.Join(staticDir, r.URL.Path)
		if _, err := os.Stat(requestedPath); err == nil && !strings.HasSuffix(r.URL.Path, "/") {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Serve index.html for all other routes (SPA history mode)
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})
}

/*
func rrrregisterRoutes(mux *http.ServeMux) {
	// Serve production build if present, otherwise the dev public folder
	if _, err := os.Stat("./frontend/dist"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))
	} else {
		mux.Handle("/", http.FileServer(http.Dir("./frontend/public")))
	}

	// Websocket endpoint (protected by auth middleware so context contains user ID)
	mux.Handle("/ws", AuthMiddleware(http.HandlerFunc(HandleWebSocket)))

	// chat message history
	mux.Handle("/api/messages/history", AuthMiddleware(http.HandlerFunc(handlers.GetMessageHistory)))

	// API endpoints
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/check-session", handlers.CheckSessionHandler)
	// follower endpoints
	mux.Handle("/api/follow", AuthMiddleware(http.HandlerFunc(handlers.FollowHandler)))     // POST to follow
	mux.Handle("/api/unfollow", AuthMiddleware(http.HandlerFunc(handlers.UnfollowHandler))) // POST to unfollow
	mux.Handle("/api/follow/accept", AuthMiddleware(http.HandlerFunc(handlers.AcceptFollowHandler)))
	mux.Handle("/api/follow/decline", AuthMiddleware(http.HandlerFunc(handlers.DeclineFollowHandler)))
	mux.Handle("/api/follow/requests", AuthMiddleware(http.HandlerFunc(handlers.ListRequests))) // GET list pending requests
	mux.Handle("/api/follow/status", AuthMiddleware(http.HandlerFunc(handlers.FollowStatusHandler)))

	// profile endpoints
	// This handles /api/profile/ (for self) and /api/profile/<id> for others
	mux.HandleFunc("/api/profile/", handlers.GetProfileHandler)
	mux.Handle("/api/profile/update", AuthMiddleware(http.HandlerFunc(handlers.UpdateProfileHandler)))
	mux.Handle("/api/profile/followers", AuthMiddleware(http.HandlerFunc(handlers.GetFollowersHandler)))
	mux.Handle("/api/profile/following", AuthMiddleware(http.HandlerFunc(handlers.GetFollowingHandler)))
	mux.Handle("/api/profile/privacy", AuthMiddleware(http.HandlerFunc(handlers.TogglePrivacyHandler)))

	// posts
	mux.Handle("/api/posts/create", AuthMiddleware(http.HandlerFunc(handlers.CreatePostHandler)))
	mux.HandleFunc("/api/posts", handlers.ListFeedHandler)

	// notifications
	// sanitized user list endpoint
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
	// group messages history
	mux.Handle("/api/group/messages", AuthMiddleware(http.HandlerFunc(handlers.ListGroupMessagesHandler)))
	mux.Handle("/api/group/comment", AuthMiddleware(http.HandlerFunc(handlers.AddGroupCommentHandler)))
	mux.Handle("/api/group/event/create", AuthMiddleware(http.HandlerFunc(handlers.CreateEventHandler)))
	mux.Handle("/api/group/event/vote", AuthMiddleware(http.HandlerFunc(handlers.VoteEventHandler)))
	mux.Handle("/api/group/events", AuthMiddleware(http.HandlerFunc(handlers.ListEventsHandler)))
	mux.Handle("/api/posts/comment", AuthMiddleware(http.HandlerFunc(handlers.AddCommentHandler)))

	// serve uploaded images
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("backend/uploads"))))

	// upload endpoint
	mux.Handle("/api/upload", AuthMiddleware(http.HandlerFunc(handlers.UploadHandler)))
}
*/
