package main

import (
	"encoding/json"
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

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if !ensureJSON(w, r) {
		slog.Info("bad requsest. Content-Type!=application/json")
		return
	}
	email, err := verifyToken(r)
	if err != nil {
		slog.Info("not authorized")
		Unauthorized(w)
		return
	}
	var req UploadFileRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Info("could not decode request body")
		BadRequest(w)
		return
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("could not get home dir", "error", err)
		InternalServerError(w)
		return
	}
	fullPath := fmt.Sprintf("%s/.fileport/users/%s/%s", homeDir, email, req.Destination)
	portNum := 8000 + rand.Intn(1000-100) + 100
	fs := NewFileServer(fullPath, portNum)
	response := &SendFileReponse{
		ResponseCode: 200,
		PortNumber:   portNum,
	}
	go fs.Start(MODE_READ)
	fs.msgch <- req.FileName
	WriteJSON(w, response)
}

func mkdirHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("did we even get here?")
	if !ensureJSON(w, r) {
		slog.Info("bad requsest. Content-Type!=application/json")
		return
	}
	email, err := verifyToken(r)
	if err != nil {
		slog.Info("not authorized")
		Unauthorized(w)
		return
	}
	var req MkdirRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Info("could not decode request body")
		BadRequest(w)
		return
	}
	userHome := GetUserDir(email)
	if userHome == "" {
		slog.Error("something went wrong while getting user directory")
		InternalServerError(w)
		return
	}
	if req.DirName[0] != '/' {
		userHome += "/"
	}
	if err = os.MkdirAll(fmt.Sprintf("%s%s", userHome, req.DirName), 0755); err != nil {
		slog.Error("could not mkdir", "error", err)
		InternalServerError(w)
		return
	}
	resObj := map[string]any{
		"message": "OK",
		"status":  200,
	}
	WriteJSON(w, resObj)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if !ensureJSON(w, r) {
		slog.Info("bad requsest. Content-Type!=application/json")
		return
	}
	email, err := verifyToken(r)
	if err != nil {
		slog.Info("not authorized")
		Unauthorized(w)
		return
	}
	fmt.Println(email)
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	if !ensureJSON(w, r) {
		slog.Info("bad requsest. Content-Type!=application/json")
		return
	}
	email, err := verifyToken(r)
	if err != nil {
		slog.Info("not authorized")
		Unauthorized(w)
		return
	}
	fmt.Println(email)
}
