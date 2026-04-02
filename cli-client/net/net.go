package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Airbag65/fileport/cli-client/fs"
)

// AuthServiceIsUp polls the authentication service (authport) to see whether or not it
// is running. Returns (true, nil) of all is good, (false, nil) if no error occured but
// the service is down and (false, err) if somehow the Ip address of the server could not
// be fetched from the configuration file
func AuthServiceIsUp() (bool, error) {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return false, err
	}
	healthCheck, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8000/health", ip), &bytes.Buffer{})
	healthRes, err := client.Do(healthCheck)
	if err != nil || healthRes.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

func ValidateUserToken(email, token string) (ResponseCode, error) {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return -1, err
	}
	reqObj := ValidateTokenReq{
		Email:     email,
		AuthToken: token,
	}
	reqBody, err := json.Marshal(reqObj)
	if err != nil {
		return -1, err
	}
	validTokenReq, err := http.NewRequest("POST", fmt.Sprintf("http://%s:8000/valid", ip), bytes.NewBuffer(reqBody))
	validTokenReq.Header.Set("Content-Type", "application/json")
	res, err := client.Do(validTokenReq)
	if err != nil {
		return -1, err
	}
	return ResponseCode(res.StatusCode), nil
}

func Login(email, password string) (*LoginResponse, error) {
	reqObj := LoginRequset{
		Email:            email,
		Password:         password,
		RemoteAddr:       GetOutboundIP(),
		ClientIdentifier: "cli",
	}
	ip, err := fs.GetCofigIP()
	if err != nil {
		return nil, err
	}
	reqBody, err := json.Marshal(reqObj)
	if err != nil {
		return nil, err
	}
	loginReq, err := http.NewRequest("POST", fmt.Sprintf("http://%s:8000/login", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	loginReq.Header.Set("Content-Type", "application/json")
	res, err := client.Do(loginReq)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return &LoginResponse{ResponseCode: res.StatusCode}, nil
	}
	var response LoginResponse
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}
