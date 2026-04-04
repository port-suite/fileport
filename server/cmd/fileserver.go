package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

type FileServer struct {
	Path string
	port int
}

type ConnMode int

const (
	Read ConnMode = iota
	Write
)

func NewFileServer(filePath string, port int) *FileServer {
	return &FileServer{
		Path: filePath,
		port: port,
	}
}

func (fs *FileServer) Start(mode ConnMode) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", fs.port))
	if err != nil {
		slog.Error("Could not start listener", "error", err)
		return
	}

	fmt.Println("Starting File Server")
	conn, err := listener.Accept()
	if err != nil {
		slog.Error("Could not accept connection", "error", err)
		return
	}
	switch mode {
	case Read:
		err = fs.retrieveFile(conn)
	case Write:
		err = fs.sendFile(conn)
	}
	if err != nil {
		slog.Error("file transfer error", "mode", mode, "error", err)
	}
	if err = conn.Close(); err != nil {
		slog.Error("could not close connection", "error", err)
	}
}

func (fs *FileServer) retrieveFile(conn net.Conn) error {
	buff := new(bytes.Buffer)
	var (
		size     int64
		fileName string
	)
	binary.Read(conn, binary.LittleEndian, &size)
	binary.Read(conn, binary.LittleEndian, &fileName)
	n, err := io.CopyN(buff, conn, size)
	if err != nil {
		slog.Error("could not read to buffer", "error", err)
		return err
	}
	slog.Info("read bytes", "number of bytes", n)

	// TODO: Save file in file system on target location
	// Assumption: fs.Path includes the users email address
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
	binary.Write(conn, binary.LittleEndian, fileStat.Name())
	n, err := io.CopyN(conn, file, fileStat.Size())
	if err != nil {
		slog.Error("could not write file over connection", "error", err)
		return err
	}
	slog.Info("wrote bytes", "number of bytes", n)
	return nil
}
