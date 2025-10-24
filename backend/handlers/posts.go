package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/backend/db"
	"social-network/backend/utils"
	"strconv"
	"strings"
)

// normalizeURL converts Windows-style paths to web-friendly URLs
func normalizeURL(path string) string {
	if path == "" {
		return ""
	}
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.ReplaceAll(path, "//", "/")
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + strings.TrimLeft(path, "/")
	}
	return path
}

// CreatePostHandler handles creating a new post from a JSON payload
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)

	var payload struct {
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
		Privacy  string `json:"privacy"`
		Allowed  string `json:"allowed"` // comma-separated ids for private
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.Content == "" {
		utils.Error(w, http.StatusBadRequest, "Post content cannot be empty")
		return
	}

	if payload.Privacy == "" {
		payload.Privacy = "public"
	}

	imagePath := normalizeURL(payload.ImageURL)

	_, err := db.DB.Exec("INSERT INTO posts (author_id, content, image_url, privacy, allowed_user_ids) VALUES (?, ?, ?, ?, ?)", userID, payload.Content, imagePath, payload.Privacy, payload.Allowed)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	utils.JSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// ListFeedHandler returns posts visible to the requester
func ListFeedHandler(w http.ResponseWriter, r *http.Request) {
	// optional ?user_id to list a user's posts
	viewer := utils.GetUserIDFromContext(r)
	if viewer == "" {
		viewer = utils.GetUserIDFromSession(w, r)
	}
	var viewerID int64
	if viewer != "" {
		viewerID, _ = strconv.ParseInt(viewer, 10, 64)
	}

	qUser := r.URL.Query().Get("user_id")
	var rows *sql.Rows
	var err error
	if qUser != "" {
		// list posts by a specific user, but apply privacy
		tid, _ := strconv.ParseInt(qUser, 10, 64)
		rows, err = db.DB.Query(`
			SELECT p.id, p.author_id, p.content, p.image_url, p.privacy, p.allowed_user_ids, p.created_at, u.nickname 
			FROM posts p JOIN users u ON p.author_id = u.id 
			WHERE p.author_id = ? 
			ORDER BY p.created_at DESC`, tid)
	} else {
		// feed: show public posts + posts from followed users + own private posts where allowed
		rows, err = db.DB.Query(`
			SELECT p.id, p.author_id, p.content, p.image_url, p.privacy, p.allowed_user_ids, p.created_at, u.nickname 
			FROM posts p JOIN users u ON p.author_id = u.id 
			ORDER BY p.created_at DESC`)
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to load posts")
		return
	}
	defer rows.Close()

	type P struct {
		ID             int64        `json:"id"`
		AuthorID       int64        `json:"author_id"`
		AuthorNickname string       `json:"author_nickname"`
		Content        string       `json:"content"`
		ImageURL       string       `json:"image_url"`
		Privacy        string       `json:"privacy"`
		Allowed        string       `json:"allowed_user_ids"`
		Created        string       `json:"created_at"`
		Comments       []commentDTO `json:"comments"`
		CommentCount   int          `json:"comment_count"`
	}
	var out []P
	for rows.Next() {
		var p P
		var allowed sql.NullString
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.ImageURL, &p.Privacy, &allowed, &p.Created, &p.AuthorNickname); err != nil {
			continue
		}
		p.Allowed = allowed.String
		p.ImageURL = normalizeURL(p.ImageURL)
		// privacy enforcement: minimalistic
		visible := false
		if p.Privacy == "public" {
			visible = true
		} else if p.Privacy == "followers" {
			if viewerID > 0 {
				var cnt int
				db.DB.QueryRow("SELECT COUNT(1) FROM followers WHERE follower_id=? AND followed_id=?", viewerID, p.AuthorID).Scan(&cnt)
				if cnt > 0 || viewerID == p.AuthorID {
					visible = true
				}
			}
		} else if p.Privacy == "private" {
			if viewerID == p.AuthorID {
				visible = true
			} else if p.Allowed != "" {
				parts := strings.Split(p.Allowed, ",")
				for _, s := range parts {
					if s == "" {
						continue
					}
					if vid, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64); vid == viewerID {
						visible = true
						break
					}
				}
			}
		}
		if visible {
			out = append(out, p)
		}
	}

	if len(out) > 0 {
		for i := range out {
			comments, err := loadComments(out[i].ID)
			if err != nil {
				continue
			}
			out[i].Comments = comments
			out[i].CommentCount = len(comments)
		}
	}
	utils.JSON(w, http.StatusOK, out)
}

// AddCommentHandler adds a comment to a post (respecting post visibility implicitly by assuming front-end only shows allowed posts)
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	uid := utils.GetUserIDFromContext(r)
	if uid == "" {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := strconv.ParseInt(uid, 10, 64)
	var payload struct {
		PostID   int64  `json:"post_id"`
		Content  string `json:"content"`
		ImageURL string `json:"image_url,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid input")
		return
	}
	imagePath := normalizeURL(payload.ImageURL)
	_, err := db.DB.Exec("INSERT INTO comments (post_id, user_id, content, image_url) VALUES (?, ?, ?, ?)", payload.PostID, userID, payload.Content, imagePath)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type commentDTO struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Nickname  string `json:"nickname"`
	Content   string `json:"content"`
	ImageURL  string `json:"image_url,omitempty"`
	CreatedAt string `json:"created_at"`
}

func loadComments(postID int64) ([]commentDTO, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, c.content, c.image_url, c.created_at, IFNULL(u.nickname, '')
		FROM comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []commentDTO
	for rows.Next() {
		var c commentDTO
		var image sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &image, &c.CreatedAt, &c.Nickname); err != nil {
			continue
		}
		c.ImageURL = normalizeURL(image.String)
		comments = append(comments, c)
	}
	return comments, nil
}
