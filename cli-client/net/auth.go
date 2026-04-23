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
	healthRes, err := Client.Do(healthCheck)
	if err != nil || healthRes.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

func ValidateUserToken(email, token string) (ResponseCode, error) {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return Nil, err
	}
	reqObj := ValidateTokenReq{
		Email:     email,
		AuthToken: token,
	}
	reqBody, err := json.Marshal(reqObj)
	if err != nil {
		return Nil, err
	}
	validTokenReq, err := http.NewRequest("POST", fmt.Sprintf("http://%s:8000/valid", ip), bytes.NewBuffer(reqBody))
	validTokenReq.Header.Set("Content-Type", "application/json")
	res, err := Client.Do(validTokenReq)
	if err != nil {
		return Nil, err
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
	res, err := Client.Do(loginReq)
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

func SignOut(email string) (ResponseCode, error) {
	reqObj := &SignOutReq{
		Email:            email,
		ClientIdentifier: "cli",
		IpAddr:           GetOutboundIP(),
	}
	reqBody, err := json.Marshal(reqObj)
	if err != nil {
		return Nil, err
	}
	ip, err := fs.GetCofigIP()
	if err != nil {
		return Nil, err
	}
	signOutReq, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:8000/signOut", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return Nil, err
	}
	signOutReq.Header.Set("Content-Type", "application/json")
	response, err := Client.Do(signOutReq)
	if err != nil {
		return Nil, err
	}
	return ResponseCode(response.StatusCode), nil
}

func RegisterUser(email, name, surname, password string) (ResponseCode, error) {
	reqObj := &RegisterRequest{
		Email:    email,
		Name:     name,
		Surname:  surname,
		Password: password,
	}
	reqBody, err := json.Marshal(reqObj)
	if err != nil {
		return Nil, err
	}
	ip, err := fs.GetCofigIP()
	if err != nil {
		return Nil, err
	}
	registerReq, err := http.NewRequest("POST", fmt.Sprintf("http://%s:8000/new", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return Nil, err
	}
	registerReq.Header.Set("Content-Type", "application/json")
	res, err := Client.Do(registerReq)
	if err != nil {
		return Nil, err
	}
	return ResponseCode(res.StatusCode), nil
}
