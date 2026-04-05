package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Airbag65/fileport/cli-client/fs"
)

func GetFilesList(path string, recursive bool) (fs.Inode, error) {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return nil, err
	}
	requset, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8001/files/list", ip), &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		return nil, err
	}
	requset.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth.AuthToken))
	requset.Header.Set("target", path)
	if recursive {
		requset.Header.Set("recursive", "true")
	}
	response, err := client.Do(requset)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, nil
	}
	var m map[string]any
	err = json.NewDecoder(response.Body).Decode(&m)
	if err != nil {
		return nil, err
	}
	dir := fs.MapToDirectoryInodeR(m)
	return dir, nil
}

func GetFile(path string) (*GetFileResponse, error) {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8001/files/get?path=%s", ip, path), nil)
	if err != nil {
		return nil, err
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth.AuthToken))
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return &GetFileResponse{
			ResponseCode: response.StatusCode,
			PortNumber:   -1,
			FileName:     "",
		}, nil
	}
	var getFileRes GetFileResponse
	if err = json.NewDecoder(response.Body).Decode(&getFileRes); err != nil {
		return nil, err
	}
	return &getFileRes, nil
}

func UploadFile(fileName, destPath string) (*UploadFileResponse, error) {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return nil, err
	}
	reqBody, err := json.Marshal(&UploadFileRequest{
		FileName:    fileName,
		Destination: destPath,
	})
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s:8001/files/upload", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth.AuthToken))
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return &UploadFileResponse{
			ResponseCode: response.StatusCode,
			PortNumber:   -1,
		}, nil
	}
	var uploadFileRes UploadFileResponse
	if err = json.NewDecoder(response.Body).Decode(&uploadFileRes); err != nil {
		return nil, err
	}
	return &uploadFileRes, nil
}
