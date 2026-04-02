package fs

import (
	"encoding/json"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Global Global `toml:"global"`
}

type Global struct {
	IpAddr     string `toml:"ip_addr"`
	SourcePath string `toml:"source_path"`
}

func GetConfiguration() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := homeDir + "/.fileport/config.toml"
	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	_, err = toml.Decode(string(configFile), &config)
	return &config, nil
}

func GetTitle() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := homeDir + "/.fileport/fileport_title.txt"

	content, err := os.ReadFile(path)
	return string(content), err
}

type LocalAuth struct {
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	Email     string `json:"email"`
	AuthToken string `json:"auth_token"`
}

func GetLocalAuth() (*LocalAuth, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := homeDir + "/.portsuite/authentication.json"
	authFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var auth LocalAuth
	if err = json.Unmarshal(authFile, &auth); err != nil {
		return nil, err
	}
	return &auth, nil
}

func SaveLocalAuth(name, surname, email, authToken string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := homeDir + "/.portsuite/authentication.json"
	localAuth := &LocalAuth{
		Name:      name,
		Surname:   surname,
		Email:     email,
		AuthToken: authToken,
	}
	fileBytes, err := json.Marshal(localAuth)
	if err != nil {
		return err
	}
	return os.WriteFile(path, fileBytes, 0644)
}

func GetCofigIP() (string, error) {
	conf, err := GetConfiguration()
	if err != nil {
		return "", err
	}
	return conf.Global.IpAddr, nil
}
