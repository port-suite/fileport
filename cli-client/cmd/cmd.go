package cmd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Airbag65/fileport/cli-client/fs"
	fpNet "github.com/Airbag65/fileport/cli-client/net"
	"github.com/fatih/color"
)

func (c *HelpCommand) Execute() {
	title, err := fs.GetTitle()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	fpYellow.Println(title)
	fmt.Println("Usage: fileport <command> [arguments]\nCOMMANDS:")
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, " status\tCheck login status")
	fmt.Fprintln(w, " login\tLogin to the port suite")
	fmt.Fprintln(w, " signout\tSign out from the port suite")
	fmt.Fprintln(w, " register\tRegister a new account")
	fmt.Fprintln(w, " list [-r|--recursive] [<directory>]\tList files stored in fileport")
	fmt.Fprintln(w, " get <file-name>\tDownload a file from fileport")
	fmt.Fprintln(w, " upload <source> <destination>\tUpload a source file to a destination in fileport")
	fmt.Fprintln(w, " remove <target>\tDelete a target file from fileport")
	fmt.Fprintln(w, " mkdir <directory>\tCreate a directory in fileport")
	fmt.Fprintln(w, " rmdir <directory>\tRemove a directory in fileport")
	fmt.Fprintln(w, " move <target> <destination>\tMove or rename a file in fileport")
	fmt.Fprintln(w, " copy <source> <destination>\tCopy a source file in fileport to a destination")
	fmt.Fprintln(w, " \t")
	fmt.Fprintln(w, " version\tDisplay the current fileport version")
	fmt.Fprintln(w, " help\tList all possible commands and their usage")
	w.Flush()
}

func (c *StatusCommand) Execute() {
	ip, err := fs.GetCofigIP()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
		color.Red("Could not connect to the server")
		fmt.Printf("Using IP: %s\n", ip)
		return
	}

	auth, err := fs.GetLocalAuth()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	if auth.AuthToken == "" {
		color.Red("You are not signed in to fileport!")
		color.Red("Run 'fileport login' to sign in")
		fmt.Printf("Using IP: %s\n", ip)
		return
	}
	code, err := fpNet.ValidateUserToken(auth.Email, auth.AuthToken)
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	if code != fpNet.OK {
		color.Red("You are not signed in to fileport!")
		color.Red("Run 'fileport login' to sign in")
		fmt.Printf("Using IP: %s\n", ip)
		return
	}
	color.Green("You are signed in to fileport! fileport is ready to use")
	fmt.Println("Your credentials:")
	fmt.Println("-----------------")
	fmt.Printf("Name: %s %s\n", auth.Name, auth.Surname)
	fmt.Printf("Email: %s\n", auth.Email)
	fmt.Printf("Using IP: %s\n", ip)

}

func (c *LoginCommad) Execute() {
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
		color.Red("Could not connect to the server")
		return
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\033[H\033[2J")
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	email = strings.TrimSuffix(email, "\n")
	fmt.Print("Password: ")
	password := GetPassword()
	response, err := fpNet.Login(email, encryptPassword(password))
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	switch fpNet.ResponseCode(response.ResponseCode) {
	case fpNet.NotFound:
		color.Yellow("Account with with email '%s' does not exist\n", email)
	case fpNet.ImATeapot:
		color.Yellow("Already logged with email '%s'\n", email)
	case fpNet.Unauthorized:
		color.Red("Incorrect password!")
	case fpNet.OK:
		if err = fs.SaveLocalAuth(response.Name, response.Surname, email, response.AuthToken); err != nil {
			color.Red("Something went wrong")
			return
		}
		color.Green("You are now logged in as '%s %s'\n", response.Name, response.Surname)
	}
}

func (c *SignOutCommand) Execute() {
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
		color.Red("Could not connect to the server")
		return
	}
	localAuth, err := fs.GetLocalAuth()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	if localAuth.Email == "" {
		color.Yellow("You were already signed out")
		return
	}
	responseCode, err := fpNet.SignOut(localAuth.Email)
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	switch responseCode {
	case fpNet.OK:
		if err = fs.SaveLocalAuth("", "", "", ""); err != nil {
			color.Red("Something went wrong")
			return
		}
		color.Green("You are now signed out")
		return
	case fpNet.NotModified:
		color.Yellow("You were already signed out")
		return
	}
}

func (c *RegisterCommand) Execute() {
	if authStatus, _ := fpNet.AuthServiceIsUp(); !authStatus {
		color.Red("Could not connect to the server")
		return
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	if auth.AuthToken != "" {
		res, err := fpNet.ValidateUserToken(auth.Email, auth.AuthToken)
		if err != nil {
			color.Red("Something went wrong")
			return
		}
		if res == fpNet.OK {
			color.Yellow("Cannot create new user while signed in")
			return
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\033[H\033[2J")
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	email = strings.TrimSuffix(email, "\n")
	fmt.Print("Name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	name = strings.TrimSuffix(name, "\n")
	fmt.Print("Surname: ")
	surname, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Something went wrong")
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
		color.Red("Passwords must match")
	}
	responseCode, err := fpNet.RegisterUser(email, name, surname, encryptPassword(password))
	switch responseCode {
	case fpNet.ImATeapot:
		color.Yellow("User with email '%s' already exists\n", email)
		return
	case fpNet.OK:
		color.Green("Created new user '%s %s' with email '%s'\n", name, surname, email)
	}

	loginRes, err := fpNet.Login(email, encryptPassword(password))
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	if err = fs.SaveLocalAuth(name, surname, email, loginRes.AuthToken); err != nil {
		color.Red("Something went wrong")
		return
	}
	fmt.Println()
	color.Green("You are now logged in as '%s %s'\n", name, surname)
}

func isAuthorized() bool {
	ip, err := fs.GetCofigIP()
	retVal := true
	if err != nil {
		retVal = false
	}
	auth, err := fs.GetLocalAuth()
	if err != nil {
		retVal = false
	}
	reqObj, err := json.Marshal(map[string]string{
		"email":      auth.Email,
		"auth_token": auth.AuthToken,
	})
	if err != nil {
		retVal = false
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s:8000/valid", ip), bytes.NewBuffer(reqObj))
	if err != nil {
		retVal = false
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := fpNet.Client.Do(request)
	if err != nil {
		retVal = false
	}
	if response.StatusCode != 200 {
		retVal = false
	}
	if !retVal {
		color.Red("You need to be signed in to use this feature")
		color.Red("Run 'fileport login' to sign in")
	}
	return retVal
}

func (c *ListCommand) Execute() {
	if !isAuthorized() {
		return
	}
	dir, err := fpNet.GetFilesList(c.Path, c.Recursive)
	if err != nil {
		color.Red("Something went wrong")
		fmt.Println(err)
		return
	}
	if dir == nil {
		color.Yellow("'%s': no such file or directory\n", c.Path)
		return
	}
	dir.Print()
}

func (c *GetCommand) Execute() {
	if !isAuthorized() {
		return
	}
	response, err := fpNet.GetFile(c.Path)
	if err != nil {
		color.Red("Something went wrong")
		return
	}

	switch response.ResponseCode {
	case 401:
		color.Red("Must be signed in")
		return
	case 404:
		color.Yellow("File '%s' does not exist\n", c.Path)
		return
	case 400:
		color.Red("Something went wrong")
		return
	}
	ip, err := fs.GetCofigIP()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, response.PortNumber))
	buff := new(bytes.Buffer)
	var (
		size int64
	)
	binary.Read(conn, binary.LittleEndian, &size)
	_, err = io.CopyN(buff, conn, size)
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	if err = os.WriteFile(response.FileName, buff.Bytes(), 0766); err != nil {
		color.Red("Could not save file '%s'. Try again later!\n", response.FileName)
		return
	}
	color.Green("Downloaded 1 file from fileport: %s\n", c.Path)
}

func (c *UploadCommand) Execute() {
	if !isAuthorized() {
		return
	}
	curDir, err := os.Getwd()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	filePath := fmt.Sprintf("%s/%s", curDir, c.FileName)
	fileStats, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		color.Yellow("File: '%s'. No such file\n", c.FileName)
		return
	}
	var fileName string
	if strings.Contains(c.FileName, "/") {
		fileName = strings.Split(c.FileName, "/")[len(strings.Split(c.FileName, "/"))-1]
	} else {
		fileName = c.FileName
	}
	response, err := fpNet.UploadFile(fileName, c.DestinationPath)
	if err != nil {
		fmt.Println("sending file")
		color.Red("Something went wrong")
		return
	}
	switch response.ResponseCode {
	case 400:
		color.Red("Something went wrong")
		return
	case 401:
		color.Red("Must be signed in")
		return
	case 500:
		color.Red("Something went wrong")
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	ip, err := fs.GetCofigIP()
	if err != nil {
		color.Red("Something went wrong")
		return
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, response.PortNumber))
	if err != nil {
		color.Red("Could not connect to file server")
		return
	}
	binary.Write(conn, binary.LittleEndian, fileStats.Size())
	_, err = io.CopyN(conn, file, fileStats.Size())
	if err != nil {
		color.Red("Could not upload file '%s'. Please try again later!\n", fileName)
		return
	}
	color.Green("Uploaded 1 file to fileport: %s\n", fileName)
}

func (c *MkdirCommand) Execute() {
	if !isAuthorized() {
		return
	}
	if err := fpNet.Mkdir(c.DirName); err != nil {
		status, ok := errors.AsType[*fpNet.StatusNotOK](err)
		if !ok {
			color.Red("Something went wrong")
			return
		}
		fmt.Printf("Status was: %d\n", status.StatusCode)
		return
	}
	color.Green("Created directory: %s\n", c.DirName)
}

func (c *RemoveCommand) Execute() {
	if !isAuthorized() {
		return
	}
	if err := fpNet.Remove(c.FileName); err != nil {
		status, ok := errors.AsType[*fpNet.StatusNotOK](err)
		if !ok {
			color.Red("Something went wrong")
			return
		}
		fmt.Printf("Status was: %d\n", status.StatusCode)
		return
	}
	color.Green("Deleted file: %s\n", c.FileName)
}

func (c *RmdirCommand) Execute() {
	if !isAuthorized() {
		return
	}
	if err := fpNet.Rmdir(c.DirName); err != nil {
		status, ok := errors.AsType[*fpNet.StatusNotOK](err)
		if !ok {
			color.Red("Something went wrong")
			return
		}
		fmt.Printf("Status was: %d\n", status.StatusCode)
		return
	}
	color.Green("Deleted directory: %s\n", c.DirName)
}

func (c *VersionCommand) Execute() {
	fmt.Println("fileport version v0.2.1")
}

func (c *MoveCommand) Execute() {
	if !isAuthorized() {
		return
	}
	if err := fpNet.MoveOrCopy(c.Target, c.Destination, fpNet.MOVE_MODE); err != nil {
		intervensionRes, ok := errors.AsType[*fpNet.IntervensionResultError](err)
		if ok {
			if intervensionRes.IntervensionResult != "OK" {
				color.Red("Something went wrong")
				return
			}
			if intervensionRes.PerformedMove {
				goto NoErr
			} else {
				color.Green("No move executed")
				return
			}
		}
		status, ok := errors.AsType[*fpNet.StatusNotOK](err)
		if !ok {
			color.Red("Something went wrong")
			return
		}
		fmt.Printf("Status was: %d\n", status.StatusCode)
	}
NoErr:
	color.Green("Moved '%s' to '%s'\n", c.Target, c.Destination)
}

func (c *CopyCommand) Execute() {
	if !isAuthorized() {
		return
	}
	if err := fpNet.MoveOrCopy(c.Source, c.Destination, fpNet.COPY_MODE); err != nil {
		intervensionRes, ok := errors.AsType[*fpNet.IntervensionResultError](err)
		if ok {
			if intervensionRes.IntervensionResult != "OK" {
				color.Red("Something went wrong")
				return
			}
			if intervensionRes.PerformedMove {
				goto NoErr
			} else {
				color.Green("No copy executed")
				return
			}
		}
		status, ok := errors.AsType[*fpNet.StatusNotOK](err)
		if !ok {
			color.Red("Something went wrong")
			return
		}
		fmt.Printf("Status was: %d\n", status.StatusCode)
	}
NoErr:
	color.Green("Copied '%s' to '%s'\n", c.Source, c.Destination)
}
