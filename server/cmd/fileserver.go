package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
)

type FileServer struct {
	Path  string
	Port  int
	msgch chan string
}

type ConnMode int

const (
	MODE_READ ConnMode = iota
	MODE_WRITE
)

func NewFileServer(filePath string, port int) *FileServer {
	return &FileServer{
		Path:  filePath,
		Port:  port,
		msgch: make(chan string, 2),
	}
}

func (fs *FileServer) Start(mode ConnMode) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", fs.Port))
	if err != nil {
		slog.Error("Could not start listener", "error", err)
		return
	}
	fmt.Printf("Starting File Server on port :%d\n", fs.Port)
	conn, err := listener.Accept()
	if err != nil {
		slog.Error("Could not accept connection", "error", err)
		return
	}
	switch mode {
	case MODE_READ:
		err = fs.retrieveFile(conn)
	case MODE_WRITE:
		err = fs.sendFile(conn)
	}
	if err != nil {
		slog.Error("file transfer error", "mode", mode, "error", err)
	}
	if err = conn.Close(); err != nil {
		slog.Error("could not close connection", "error", err)
	}
}

func extractDirectory(path string) string {
	pathParts := strings.Split(path, "/")
	dirPath := "/"
	for i, part := range pathParts {
		if i == len(pathParts)-1 {
			break
		}
		dirPath = fmt.Sprintf("%s%s/", dirPath, part)
	}
	return dirPath
}

func (fs *FileServer) retrieveFile(conn net.Conn) error {
	buff := new(bytes.Buffer)
	var (
		size int64
	)
	binary.Read(conn, binary.LittleEndian, &size)
	n, err := io.CopyN(buff, conn, size)
	if err != nil {
		slog.Error("could not read to buffer", "error", err)
		return err
	}
	slog.Info("read bytes", "number of bytes", n)

	if err = os.MkdirAll(extractDirectory(fs.Path), 0755); err != nil {
		slog.Error("could not mkdir", "error", err)
		return err
	}
	stat, err := os.Stat(fs.Path)
	if err != nil {
		slog.Error("could not get file stats", "error", err)
		return err
	}
	if stat.IsDir() {
		if fs.Path[len(fs.Path)-1] != '/' {
			fs.Path = fs.Path + "/"
		}
		select {
		case name := <-fs.msgch:
			fs.Path = fs.Path + name
		}
		// fs.Path = fs.Path + name
		// d1, d2, d3 := rand.Intn(10), rand.Intn(10), rand.Intn(10)
		// fs.Path = fmt.Sprintf("%supload_%d%d%d.file", fs.Path, d1, d2, d3)
	}
	if err = os.WriteFile(fs.Path, buff.Bytes(), 0755); err != nil {
		slog.Error("could not write file", "error", err)
		return err
	}
	return nil
}

func (fs *FileServer) sendFile(conn net.Conn) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("could now fetch home dir", "error", err)
		return err
	}
	fullPath := homeDir + "/.fileport/users/" + fs.Path
	file, err := os.Open(fullPath)
	if err != nil {
		slog.Error("could not open file", "error", err)
		return err
	}
	fileStat, err := file.Stat()
	if err != nil {
		slog.Error("could not view stats on file", "error", err)
		return err
	}
	binary.Write(conn, binary.LittleEndian, fileStat.Size())
	n, err := io.CopyN(conn, file, fileStat.Size())
	if err != nil {
		slog.Error("could not write file over connection", "error", err)
		return err
	}
	slog.Info("wrote bytes", "number of bytes", n)
	return nil
}
