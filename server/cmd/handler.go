package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func listDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	email, err := verifyToken(r)
	if err != nil {
		slog.Error("verify token error", "err", err)
		Unauthorized(w)
		return
	}
	var recursive bool
	if r.Header.Get("recursive") == "true" {
		recursive = true
	} else {
		recursive = false
	}
	targetDir := r.Header.Get("target")
	if targetDir == "" {
		targetDir = "."
	}

	_, err = GetUserDirPath(email)
	if err != nil {
		InternalServerError(w)
		return
	}
	var dir *DirectoryINode
	path := fmt.Sprintf("%s/%s", email, targetDir)
	if recursive {
		dir, err = GetDirectoryContentR(path)
	} else {
		dir, err = GetDirectoryContent(path)
	}
	if err != nil {
		InternalServerError(w)
		return
	}
	if dir == nil {
		NotFound(w)
		return
	}

	WriteJSON(w, dir)
}

func getFileHandler(w http.ResponseWriter, r *http.Request) {
	email, err := verifyToken(r)
	if err != nil {
		slog.Error("invalid token error", "error", err)
		Unauthorized(w)
		return
	}
	fmt.Println(email)
	fmt.Println(r.URL.Query())
	WriteJSON(w, "Hello")
}
