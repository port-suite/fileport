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
