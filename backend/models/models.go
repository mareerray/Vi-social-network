package models

import "time"

// Register and responses
type RegisterRequest struct {
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Avatar      string `json:"avatar"`
	About       string `json:"about_me"`
	ProfileType string `json:"profile_type"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // email or nickname
	Password   string `json:"password"`
}

type LoginResponse struct {
	UserID string `json:"user_id"`
}

// Shared Models
type User struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"-"` // don’t expose in JSON
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth string    `json:"date_of_birth"`
	Avatar      string    `json:"avatar,omitempty"`
	Nickname    string    `json:"nickname"`
	About       string    `json:"about,omitempty"`
	ProfileType string    `json:"profile_type"` // public/private
	CreatedAt   time.Time `json:"created_at"`
}

type Post struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	ImageURL       string    `json:"image_url,omitempty"`
	AuthorID       int       `json:"author_id"`
	AuthorNickname string    `json:"author_nickname"`
	CreatedAt      time.Time `json:"-"`
	CreatedAtHuman string    `json:"created_at"`
	Categories     []string  `json:"categories"`
}

type Comment struct {
	ID             int       `json:"id"`
	PostID         int       `json:"post_id"`
	UserID         string    `json:"user_id"`
	Content        string    `json:"content"`
	ImageURL       string    `json:"image_url,omitempty"`
	ParentID       *int      `json:"parent_id,omitempty"`
	CreatedAt      time.Time `json:"-"`
	CreatedAtHuman string    `json:"created_at"`
	Nickname       string    `json:"nickname"`
	ReplyCount     int       `json:"reply_count"`
	LikeCount      int       `json:"like_count"`
	DislikeCount   int       `json:"dislike_count"`
	Replies        []Comment `json:"replies,omitempty"`
}

type CommentRequest struct {
	PostID  int    `json:"post_id"`
	Content string `json:"content"`
	UserID  string `json:"user_id"`
}

type CommentResponse struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID         int       `json:"id"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	ReceiverID string    `json:"receiver_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Session struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	CookieToken string    `json:"cookie_token"`
	Expiry      time.Time `json:"expiry"`
}

type Follower struct {
	ID         int64 `json:"id"`
	FollowerID int64 `json:"follower_id"`
	FollowedID int64 `json:"followed_id"`
}

type FollowRequest struct {
	ID         int64  `json:"id"`
	SenderID   int64  `json:"sender_id"`
	ReceiverID int64  `json:"receiver_id"`
	Status     string `json:"status"` // pending, accepted, declined
}

type Notification struct {
	ID          int64  `json:"id"`
	RecipientID int64  `json:"recipient_id"`
	ActorID     int64  `json:"actor_id,omitempty"`
	Type        string `json:"type"`
	Data        string `json:"data,omitempty"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
}
