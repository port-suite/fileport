package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func verifyToken(r *http.Request) (string, error) {
	request, err := http.NewRequest("GET", "http://127.0.0.1:8000/validate", &bytes.Buffer{})
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", r.Header.Get("Authorization"))
	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", &InvalidTokenError{}
	}
	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return "", err
	}
	return user.Email, nil
}

func GetUserDir(email string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("could not get home dir", "error", err)
		return ""
	}
	return fmt.Sprintf("%s/.fileport/users/%s", homeDir, email)
}

/* --- CUSTOM ERRORS --- */
type InvalidTokenError struct{}

func (e *InvalidTokenError) Error() string {
	return "InvalidTokenError"
}
