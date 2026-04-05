package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
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
	query := r.URL.Query()
	if !query.Has("path") {
		slog.Info("no path provided")
		BadRequest(w)
		return
	}
	var path string
	path = query.Get("path")
	if path == "" {
		slog.Info("no such file")
		NotFound(w)
		return
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		InternalServerError(w)
		return
	}
	path = fmt.Sprintf("%s/%s", email, path)
	if _, err = os.Stat(fmt.Sprintf("%s/.fileport/users/%s", homeDir, path)); os.IsNotExist(err) {
		NotFound(w)
		return
	}
	portNum := 8000 + rand.Intn(1000-100) + 100
	response := &GetFileResponse{
		ResponseCode: 200,
		PortNumber:   portNum,
	}
	fs := NewFileServer(path, portNum)
	go fs.Start(MODE_WRITE)
	response.FileName = strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
	WriteJSON(w, response)
}
