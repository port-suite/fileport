package fs

import (
	"encoding/json"
	"os"
	"slices"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Global Global `toml:"global"`
	Alias  *Alias `toml:"alias"`
}

type Global struct {
	IpAddr     string `toml:"ip_addr"`
	SourcePath string `toml:"source_path"`
}

type AliasList []string

type Alias struct {
	Help     AliasList `toml:"help"`
	Status   AliasList `toml:"status"`
	Login    AliasList `toml:"login"`
	SignOut  AliasList `toml:"signout"`
	Register AliasList `toml:"register"`
	List     AliasList `toml:"list"`
	Get      AliasList `toml:"get"`
	Upload   AliasList `toml:"upload"`
	Mkdir    AliasList `toml:"mkdir"`
	Remove   AliasList `toml:"remove"`
	Rmdir    AliasList `toml:"rmdir"`
	Move     AliasList `toml:"move"`
	Version  AliasList `toml:"version"`
	Alias    AliasList `toml:"alias"` // Implement alias command
	Init     AliasList `toml:"init"`
	Copy     AliasList `toml:"copy"`
	Config   AliasList `toml:"config"`
}

func (al *AliasList) Contains(command string) bool {
	return slices.Contains(*al, command)
}

func SaveConfiguration(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := homeDir + "/.fileport/config.toml"
	configFile, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer configFile.Close()
	err = toml.NewEncoder(configFile).Encode(config)
	if err != nil {
		return err
	}
	return nil
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

func NewAlias() *Alias {
	return &Alias{
		Help:     []string{},
		Status:   []string{},
		Login:    []string{},
		SignOut:  []string{},
		Register: []string{},
		List:     []string{},
		Get:      []string{},
		Upload:   []string{},
		Mkdir:    []string{},
		Remove:   []string{},
		Rmdir:    []string{},
		Move:     []string{},
		Version:  []string{},
		Alias:    []string{},
		Init:     []string{},
		Copy:     []string{},
	}
}

func GetConfigAliases() (*Alias, error) {
	config, err := GetConfiguration()
	if err != nil {
		return nil, err
	}
	if config.Alias == nil {
		config.Alias = NewAlias()
		SaveConfiguration(config)
	}
	return config.Alias, nil
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
