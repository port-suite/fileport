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
