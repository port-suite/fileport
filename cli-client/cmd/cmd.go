package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Airbag65/fileport/cli-client/fs"
	"github.com/Airbag65/fileport/cli-client/net"
)

func (c *HelpCommand) Execute() {
	title, err := fs.GetTitle()
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	fpYellow.Println(title)
}

func (c *StatusCommand) Execute() {
	ip, err := fs.GetCofigIP()
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	authStatus, err := net.AuthServiceIsUp()
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if !authStatus {
		red.Println("Could not connect to the server")
		fmt.Printf("Using IP: %s\n", ip)
		return
	}

	auth, err := fs.GetLocalAuth()
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if auth.AuthToken == "" {
		red.Println("You are not signed in to fileport!")
		red.Println("Run 'fileport login' to sign in")
		fmt.Printf("Using IP: %s\n", ip)
		return
	}
	code, err := net.ValidateUserToken(auth.Email, auth.AuthToken)
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if code != net.OK {
		red.Println("You are not signed in to fileport!")
		red.Println("Run 'fileport login' to sign in")
		fmt.Printf("Using IP: %s\n", ip)
		return
	}
	green.Println("You are signed in to fileport! fileport is ready to use")
	fmt.Println("Your credentials:")
	fmt.Println("-----------------")
	fmt.Printf("Name: %s %s\n", auth.Name, auth.Surname)
	fmt.Printf("Email: %s\n", auth.Email)
	fmt.Printf("Using IP: %s\n", ip)

}

func (c *LoginCommad) Execute() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\033[H\033[2J")
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	email = strings.TrimSuffix(email, "\n")
	fmt.Print("Password: ")
	password := GetPassword()
	response, err := net.Login(email, encryptPassword(password))
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	switch net.ResponseCode(response.ResponseCode) {
	case net.NotFound:
		yellow.Printf("Account with with email '%s' does not exist\n", email)
	case net.ImATeapot:
		yellow.Printf("Already logged with email '%s'\n", email)
	case net.Unauthorized:
		red.Println("Incorrect password!")
	case net.OK:
		if err = fs.SaveLocalAuth(response.Name, response.Surname, email, response.AuthToken); err != nil {
			red.Println("Something went wrong")
			return
		}
		green.Printf("You are now logged in as '%s %s'\n", response.Name, response.Surname)
	}
}

func GetCommand(args []string) Command {
	if len(args) < 1 {
		fmt.Println("Usage: fileport <command>")
		yellow.Println("Run 'fileport help' for further instructions")
		return nil
	}
	switch args[0] {
	case "help":
		return &HelpCommand{}
	case "status":
		return &StatusCommand{}
	case "login":
		return &LoginCommad{}
	default:
		fmt.Println("fileport: Invalid argument")
		return nil
	}
}
