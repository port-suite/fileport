package net

import (
	"fmt"
	"net"
	"net/http"
)

var (
	client = &http.Client{}
)

type ValidateTokenReq struct {
	AuthToken string `json:"auth_token"`
	Email     string `json:"email"`
}

type ResponseCode int

const (
	OK                  ResponseCode = 200
	BadRequset          ResponseCode = 400
	NotFound            ResponseCode = 404
	Unauthorized        ResponseCode = 401
	NotModified         ResponseCode = 304
	ImATeapot           ResponseCode = 418
	InternalServerError ResponseCode = 500
	Nil                 ResponseCode = -1
)

type LoginRequset struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ClientIdentifier string `json:"client_identifier"`
	RemoteAddr       string `json:"remote_addr"`
}

type LoginResponse struct {
	ResponseCode    int    `json:"response_code"`
	ResponseMessage string `json:"response_message"`
	Name            string `json:"name"`
	Surname         string `json:"surname"`
	AuthToken       string `json:"auth_token"`
}

type SignOutReq struct {
	Email            string `json:"email"`
	ClientIdentifier string `json:"client_identifier"`
	IpAddr           string `json:"ip_addr"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Password string `json:"password"`
}

type GetFileResponse struct {
	ResponseCode int    `json:"response_code"`
	PortNumber   int    `json:"port_number"`
	FileName     string `json:"file_name"`
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer conn.Close()

	ip := conn.LocalAddr().(*net.UDPAddr)
	return ip.IP.String()
}

type UploadFileRequest struct {
	FileName    string `json:"file_name"`
	Destination string `json:"destination"`
}

type UploadFileResponse struct {
	ResponseCode int `json:"response_code"`
	PortNumber   int `json:"port_number"`
}

type MkdirRequest struct {
	DirName string `json:"dir_name"`
}
