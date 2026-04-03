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
	if authStatus, _ := net.AuthServiceIsUp(); !authStatus {
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
	if authStatus, _ := net.AuthServiceIsUp(); !authStatus {
		red.Println("Could not connect to the server")
		return
	}
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

func (c *SignOutCommand) Execute() {
	if authStatus, _ := net.AuthServiceIsUp(); !authStatus {
		red.Println("Could not connect to the server")
		return
	}
	localAuth, err := fs.GetLocalAuth()
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if localAuth.Email == "" {
		yellow.Println("You were already signed out")
		return
	}
	responseCode, err := net.SignOut(localAuth.Email)
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	switch responseCode {
	case net.OK:
		if err = fs.SaveLocalAuth("", "", "", ""); err != nil {
			red.Println("Something went wrong")
			return
		}
		green.Println("You are now signed out")
		return
	case net.NotModified:
		yellow.Println("You were already signed out")
		return
	}
}

func (c *RegisterCommand) Execute() {
	if authStatus, _ := net.AuthServiceIsUp(); !authStatus {
		red.Println("Could not connect to the server")
		return
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if auth.AuthToken != "" {
		res, err := net.ValidateUserToken(auth.Email, auth.AuthToken)
		if err != nil {
			red.Println("Something went wrong")
			return
		}
		if res == net.OK {
			yellow.Println("Cannot create new user while signed in")
			return
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\033[H\033[2J")
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	email = strings.TrimSuffix(email, "\n")
	fmt.Print("Name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	name = strings.TrimSuffix(name, "\n")
	fmt.Print("Surname: ")
	surname, err := reader.ReadString('\n')
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	surname = strings.TrimSuffix(surname, "\n")
	var password string
	var confirmPassword string
	for {
		fmt.Print("Password: ")
		password = GetPassword()
		fmt.Print("Confirm password: ")
		confirmPassword = GetPassword()
		if password == confirmPassword {
			break
		}
		red.Println("Passwords must match")
	}
	responseCode, err := net.RegisterUser(email, name, surname, encryptPassword(password))
	switch responseCode {
	case net.ImATeapot:
		yellow.Printf("User with email '%s' already exists\n", email)
		return
	case net.OK:
		green.Printf("Created new user '%s %s' with email '%s'\n", name, surname, email)
	}

	loginRes, err := net.Login(email, encryptPassword(password))
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if err = fs.SaveLocalAuth(name, surname, email, loginRes.AuthToken); err != nil {
		red.Println("Something went wrong")
		return
	}
	fmt.Println()
	green.Printf("You are now logged in as '%s %s'\n", name, surname)
}

func (c *ListCommand) Execute() {
	dir, err := net.GetFilesList(c.Path, c.Recursive)
	if err != nil {
		red.Println("Something went wrong")
		fmt.Println(err)
		return
	}
	dir.Print()
}
