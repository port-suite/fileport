package cmd

import (
	"fmt"
	"slices"

	"github.com/fatih/color"
)

type Command interface {
	Execute()
}

type HelpCommand struct{}

type StatusCommand struct{}

type LoginCommad struct{}

type SignOutCommand struct{}

type RegisterCommand struct{}

type ListCommand struct {
	Recursive bool
	Path      string
}

var (
	red      = color.RGB(255, 0, 0)
	green    = color.RGB(0, 255, 0)
	fpYellow = color.RGB(255, 249, 87)
	yellow   = color.RGB(255, 255, 0)
)

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
	case "signout":
		return &SignOutCommand{}
	case "register":
		return &RegisterCommand{}
	case "list":
		if len(args) > 3 {
			fmt.Println("fileport: Invalid argument")
			return nil
		}
		rec := false
		path := "."
		recAlternatives := []string{"-r", "--recursive"}
		switch len(args) {
		case 2:
			if slices.Contains(recAlternatives, args[1]) {
				rec = true
			} else {
				path = args[1]
			}
		case 3:
			if slices.Contains(recAlternatives, args[2]) {
				rec = true
				path = args[1]
			} else {
				fmt.Printf("Usage: fileport %s [path] [-r --recursive]", args[0])
				return nil
			}
		}
		return &ListCommand{
			Recursive: rec,
			Path:      path,
		}
	default:
		fmt.Println("fileport: Invalid argument")
		return nil
	}
}
