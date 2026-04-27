package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
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
	portNum := GeneratePort()
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
	stat, err := os.Stat(userHome + req.DirName)
	if err == nil {
		if stat.IsDir() {
			WriteCustom(w, 304, fmt.Sprintf("'%s' already exists", req.DirName))
			return
		}
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
	userHome := GetUserDir(email)
	if userHome == "" {
		slog.Error("something went wrong while getting user directory")
		InternalServerError(w)
		return
	}
	var reqBody RemoveRequest
	if err = json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		slog.Info("could not decode request body")
		BadRequest(w)
		return
	}
	if reqBody.FileName[0] != '/' {
		userHome += "/"
	}

	fullPath := userHome + reqBody.FileName
	stat, err := os.Stat(fullPath)
	if err != nil {
		slog.Error(fmt.Sprintf("could not read stats for '%s'", fullPath), "error", err)
		InternalServerError(w)
		return
	}

	if stat.IsDir() {
		slog.Info(fmt.Sprintf("'%s' is not a file", fullPath))
		WriteCustom(w, 304, fmt.Sprintf("'%s' is a directory", reqBody.FileName))
		return
	}
	if err = os.Remove(fullPath); err != nil {
		slog.Info("no such file")
		NotFound(w)
		return
	}
	resObj := map[string]any{
		"message": "OK",
		"status":  200,
	}
	WriteJSON(w, resObj)
}

func rmdirHandler(w http.ResponseWriter, r *http.Request) {
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
	userHome := GetUserDir(email)
	var req RmdirRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w)
		return
	}
	if req.DirName[0] != '/' {
		userHome += "/"
	}
	fullPath := userHome + req.DirName
	stat, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		WriteCustom(w, 304, fmt.Sprintf("'%s' no such directory", req.DirName))
		return
	} else if !os.IsNotExist(err) && err != nil {
		InternalServerError(w)
		return
	}
	if !stat.IsDir() {
		WriteCustom(w, 304, fmt.Sprintf("'%s' no such directory", req.DirName))
		return
	}
	if err = os.RemoveAll(fullPath); err != nil {
		InternalServerError(w)
		return
	}
	WriteJSON(w, map[string]any{
		"message": "OK",
		"status":  200,
	})
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
	userHome := GetUserDir(email)
	var req MoveRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if req.Target[0] != '/' {
		req.Target = "/" + req.Target
	}
	if req.Destination[0] != '/' {
		req.Destination = "/" + req.Destination
	}
	_, err = os.Stat(userHome + req.Target)
	if os.IsNotExist(err) {
		WriteCustom(w, 304, fmt.Sprintf("'%s' target does not exist", req.Target[1:]))
		slog.Info("target does not exist", "target", req.Target[1:])
		return
	}
	_, err = os.Stat(userHome + req.Destination)
	if os.IsNotExist(err) {
		// File/Dir does not exist
		if err = os.Rename(userHome+req.Target, userHome+req.Destination); err != nil {
			InternalServerError(w)
			slog.Error("something went wrong", "error", err)
			return
		}
		WriteJSON(w, map[string]any{
			"status":  200,
			"message": "OK",
		})
		return
	}
	portNum := GeneratePort()
	needsIntervensionRes := map[string]any{
		"message":  "Intervension needed",
		"status":   303,
		"port_num": portNum,
	}
	w.WriteHeader(303)
	json.NewEncoder(w).Encode(needsIntervensionRes)
	rc := http.NewResponseController(w)
	rc.Flush()
	respch := make(chan string)
	go StartIntervensionServer(portNum, respch)

	intervResp := <-respch
	fmt.Println(intervResp)
	if strings.ToLower(intervResp) == "n" {
		respch <- string(DONE)
		return
	} else if strings.ToLower(intervResp) != "y" {
		respch <- string(INVALID_RESPONSE)
		return
	}
	if err = os.Rename(userHome+req.Target, userHome+req.Destination); err != nil {
		slog.Error("something went wrong", "error", err)
		respch <- string(FAILED)
		return
	}
	respch <- string(DONE)
}

func StartIntervensionServer(portNum int, respch chan string) {
	slog.Info("starting intervension server on", "port", portNum)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", portNum))
	if err != nil {
		return
	}
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	respByte := make([]byte, 1)
	_, err = conn.Read(respByte)
	if err != nil {
		return
	}
	slog.Info("read byte", "byte", string(respByte))
	respch <- string(respByte)
	actionDone := <-respch
	if ChanAction(actionDone) == DONE {
		conn.Write([]byte("OK\n"))
	} else if ChanAction(actionDone) == FAILED {
		conn.Write([]byte("FAILED\n"))
	} else if ChanAction(actionDone) == INVALID_RESPONSE {
		conn.Write([]byte("INVALID RESPONSE\n"))
	}
}
