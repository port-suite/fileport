package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"

	"golang.org/x/term"
)

func GetPassword() string {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return ""
	}
	defer term.Restore(fd, oldState)

	var password []byte
	for {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			return ""
		}
		char := b[0]
		if char == '\r' || char == '\n' {
			break
		}
		// handle backspace
		if char == 8 || char == 127 {
			if len(password) > 0 {
				password = password[:len(password)-1]
				fmt.Print("\b \b")
			}
			continue
		}
		password = append(password, char)
		fmt.Print("*")
	}
	fmt.Print("\r\n")
	return string(password)
}

func extractSalt(pwd string) (string, string) {
	saltBeginning := ""
	for i := len(pwd) - 2; i < len(pwd); i++ {
		saltBeginning = fmt.Sprintf("%s%c", saltBeginning, pwd[i])
	}

	saltEnd := ""
	for _, c := range pwd[:5] {
		saltEnd = fmt.Sprintf("%s%c", saltEnd, c)
	}
	return saltBeginning, saltEnd
}

func encryptPassword(origPwd string) string {
	encPassword := origPwd
	saltBeginning, saltEnd := extractSalt(encPassword)

	sha256Encoder := sha256.New()
	sha256Encoder.Write([]byte(encPassword))
	encPassword = fmt.Sprintf("%x", sha256Encoder.Sum(nil))

	encPassword = fmt.Sprintf("%s%s%s", saltBeginning, encPassword, saltEnd)

	sha256Encoder.Write([]byte(encPassword))
	encPassword = fmt.Sprintf("%x", sha256Encoder.Sum(nil))
	return encPassword
}
