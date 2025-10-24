package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"social-network/backend/utils"
	"strings"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// The user making the request. Must be logged in to upload.
	requestingUserIDStr := utils.GetUserIDFromContext(r)
	if requestingUserIDStr == "" {
		utils.Error(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	// Parse the multipart form, with a 10MB max file size
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Could not parse multipart form")
		return
	}

	// Get the file from the form data
	file, handler, err := r.FormFile("file")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Could not get uploaded file")
		return
	}
	defer file.Close()

	// Check the file type
	uploadType := r.FormValue("type") // "avatar" or "post"
	if uploadType != "avatar" && uploadType != "post" {
		utils.Error(w, http.StatusBadRequest, "Invalid upload type specified")
		return
	}

	// Create a unique filename to avoid collisions
	ext := filepath.Ext(handler.Filename)
	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), requestingUserIDStr, ext)

	// Define the path to save the file
	var savePath string
	if uploadType == "avatar" {
		savePath = filepath.Join("backend", "uploads", "avatars", filename)
	} else {
		savePath = filepath.Join("backend", "uploads", "posts", filename)
	}

	// Create the destination file and copy the uploaded bytes directly
	os.MkdirAll(filepath.Dir(savePath), 0755)
	dst, err := os.Create(savePath)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Could not create file on server")
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Could not save file")
		return
	}

	// Return the URL path to the file (always forward slashes, prefixed with /)
	relPath := strings.TrimPrefix(savePath, "backend")
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	relPath = strings.ReplaceAll(relPath, "//", "/")
	relPath = strings.TrimSpace(relPath)
	if relPath == "" {
		relPath = "/uploads"
	}
	if !strings.HasPrefix(relPath, "/") {
		relPath = "/" + strings.TrimLeft(relPath, "/")
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"url": relPath,
	})
}
