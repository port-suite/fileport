package net

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

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
	response, err := Client.Do(requset)
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
	response, err := Client.Do(request)
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
	response, err := Client.Do(request)
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

func Mkdir(dirName string) error {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return err
	}
	reqBody, err := json.Marshal(&MkdirRequest{
		DirName: dirName,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s:8001/files/mkdir", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		return err
	}
	AddHeadersJSON(request, auth.AuthToken)
	response, err := Client.Do(request)
	if err != nil {
		return err
	}
	if ResponseCode(response.StatusCode) != OK {
		return &StatusNotOK{response.StatusCode}
	}
	return nil
}

func Remove(fileName string) error {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return err
	}
	reqBody, err := json.Marshal(&RemoveRequest{
		FileName: fileName,
	})
	request, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:8001/files/delete", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		return err
	}
	AddHeadersJSON(request, auth.AuthToken)
	response, err := Client.Do(request)
	if err != nil {
		return err
	}
	if ResponseCode(response.StatusCode) != OK {
		return &StatusNotOK{response.StatusCode}
	}
	return nil
}

func Rmdir(dirName string) error {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return err
	}
	reqBody, err := json.Marshal(map[string]string{
		"dir_name": dirName,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:8001/files/rmdir", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		return err
	}
	AddHeadersJSON(request, auth.AuthToken)
	response, err := Client.Do(request)
	if err != nil {
		return err
	}
	if ResponseCode(response.StatusCode) != OK {
		return &StatusNotOK{response.StatusCode}
	}
	return nil
}

func Move(target, destination string) error {
	ip, err := fs.GetCofigIP()
	if err != nil {
		return err
	}
	reqBody, err := json.Marshal(&MoveRequest{
		Target:      target,
		Destination: destination,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:8001/files/move", ip), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		return err
	}
	AddHeadersJSON(request, auth.AuthToken)
	response, err := Client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == 200 {
		return nil
	}
	if response.StatusCode != 303 {
		return &StatusNotOK{response.StatusCode}
	}
	var resBody IntervensionResponse
	if err = json.NewDecoder(response.Body).Decode(&resBody); err != nil {
		return err
	}
	fmt.Printf("'%s' already exists\n", destination)
	fmt.Printf("Do you want to override '%s' with the content of '%s'? [Y/n] ", destination, target)
	confirmation := "y"
	fmt.Scanln(&confirmation)
	confirmation = strings.ToLower(confirmation)
	if strings.ToLower(confirmation) != "y" {
		confirmation = "n"
	}
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", resBody.PortNum))
	conn.Write([]byte(confirmation))
	tcpRes, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return err
	}
	if tcpRes[len(tcpRes)-1] == 0x0a {
		tcpRes = tcpRes[:len(tcpRes)-1]
	}
	conn.Close()
	var moveDone bool
	if confirmation == "y" {
		moveDone = true
	} else {
		moveDone = false
	}
	return &IntervensionResultError{
		IntervensionResult: string(tcpRes),
		PerformedMove:      moveDone,
	}
}
