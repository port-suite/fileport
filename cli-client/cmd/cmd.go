package cmd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/Airbag65/fileport/cli-client/fs"
	fpNet "github.com/Airbag65/fileport/cli-client/net"
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
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
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
	code, err := fpNet.ValidateUserToken(auth.Email, auth.AuthToken)
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if code != fpNet.OK {
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
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
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
	response, err := fpNet.Login(email, encryptPassword(password))
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	switch fpNet.ResponseCode(response.ResponseCode) {
	case fpNet.NotFound:
		yellow.Printf("Account with with email '%s' does not exist\n", email)
	case fpNet.ImATeapot:
		yellow.Printf("Already logged with email '%s'\n", email)
	case fpNet.Unauthorized:
		red.Println("Incorrect password!")
	case fpNet.OK:
		if err = fs.SaveLocalAuth(response.Name, response.Surname, email, response.AuthToken); err != nil {
			red.Println("Something went wrong")
			return
		}
		green.Printf("You are now logged in as '%s %s'\n", response.Name, response.Surname)
	}
}

func (c *SignOutCommand) Execute() {
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
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
	responseCode, err := fpNet.SignOut(localAuth.Email)
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	switch responseCode {
	case fpNet.OK:
		if err = fs.SaveLocalAuth("", "", "", ""); err != nil {
			red.Println("Something went wrong")
			return
		}
		green.Println("You are now signed out")
		return
	case fpNet.NotModified:
		yellow.Println("You were already signed out")
		return
	}
}

func (c *RegisterCommand) Execute() {
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
		red.Println("Could not connect to the server")
		return
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	if auth.AuthToken != "" {
		res, err := fpNet.ValidateUserToken(auth.Email, auth.AuthToken)
		if err != nil {
			red.Println("Something went wrong")
			return
		}
		if res == fpNet.OK {
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
	responseCode, err := fpNet.RegisterUser(email, name, surname, encryptPassword(password))
	switch responseCode {
	case fpNet.ImATeapot:
		yellow.Printf("User with email '%s' already exists\n", email)
		return
	case fpNet.OK:
		green.Printf("Created new user '%s %s' with email '%s'\n", name, surname, email)
	}

	loginRes, err := fpNet.Login(email, encryptPassword(password))
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
	dir, err := fpNet.GetFilesList(c.Path, c.Recursive)
	if err != nil {
		red.Println("Something went wrong")
		fmt.Println(err)
		return
	}
	if dir == nil {
		yellow.Printf("'%s': no such file or directory\n", c.Path)
		return
	}
	dir.Print()
}

func (c *GetCommand) Execute() {
	response, err := fpNet.GetFile(c.Path)
	if err != nil {
		red.Println("Something went wrong")
		return
	}

	switch response.ResponseCode {
	case 401:
		red.Println("Must be signed in")
		return
	case 404:
		yellow.Printf("File '%s' does not exist\n", c.Path)
		return
	case 400:
		red.Println("Something went wrong")
		return
	}
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", response.PortNumber))
	buff := new(bytes.Buffer)
	var (
		size int64
	)
	binary.Read(conn, binary.LittleEndian, &size)
	_, err = io.CopyN(buff, conn, size)
	if err != nil {
		red.Println("Something went wrong")
		return
	}
	// TODO: Save file in file system on target location
	// Assumption: fs.Path includes the users email address
	if err = os.WriteFile(response.FileName, buff.Bytes(), 0766); err != nil {
		red.Printf("Could not save file '%s'. Try again later!\n", response.FileName)
		return
	}
	green.Printf("Downloaded 1 file from fileport: %s\n", c.Path)
}
