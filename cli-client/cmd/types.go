package cmd

import (
	"fmt"
	"slices"

	"github.com/Airbag65/fileport/cli-client/fs"
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

type GetCommand struct {
	Path string
}

type UploadCommand struct {
	FileName        string
	DestinationPath string
}

type MkdirCommand struct {
	DirName string
}

type RemoveCommand struct {
	FileName string
}

type RmdirCommand struct {
	DirName string
}

type MoveCommand struct {
	Target      string
	Destination string
}

type VersionCommand struct{}

type CopyCommand struct {
	Source      string
	Destination string
}

type InitCommand struct{}

type AliasCommans struct {
	Command string
	Alias   string
}

type ConfigCommand struct{}

type ViewCommand struct{}

type StatCommand struct {
	Target string
}

var (
	fpYellow = color.RGB(255, 249, 87)
)

func GenerateCommand(args []string) Command {
	aliases, err := fs.GetConfigAliases()
	if err != nil {
		color.Red("Could not fetch aliases")
		return nil
	}
	if len(args) < 1 {
		fmt.Println("Usage: fileport <command>")
		color.Yellow("Run 'fileport help' for further instructions")
		return nil
	}
	cmd := args[0]
	switch {
	case aliases.Help.Contains(cmd), cmd == "help":
		return &HelpCommand{}
	case aliases.Status.Contains(cmd), cmd == "status":
		return &StatusCommand{}
	case aliases.Login.Contains(cmd), cmd == "login":
		return &LoginCommad{}
	case aliases.SignOut.Contains(cmd), cmd == "signout":
		return &SignOutCommand{}
	case aliases.Register.Contains(cmd), cmd == "register":
		return &RegisterCommand{}
	case aliases.List.Contains(cmd), cmd == "list":
		if len(args) > 3 {
			fmt.Println("fileport: Invalid argument")
			return nil
		}
		rec := false
		path := "."
		recAlternatives := []string{"-r", "--recursive"}
		switch len(args) {
		case 2:
			if args[1][0] == '-' {
				if slices.Contains(recAlternatives, args[1]) {
					rec = true
				} else {
					fmt.Printf("fileport: Invalid option '%s'. Options: [-r|--recursive]\n", args[1])
					return nil
				}
			} else {
				path = args[1]
			}
		case 3:
			if slices.Contains(recAlternatives, args[2]) {
				rec = true
				path = args[1]
			} else {
				fmt.Printf("Usage: fileport %s [path] [-r|--recursive]\n", args[0])
				return nil
			}
		}
		return &ListCommand{
			Recursive: rec,
			Path:      path,
		}
	case aliases.Get.Contains(cmd), cmd == "get":
		if len(args) != 2 {
			fmt.Printf("Usage: fileport %s <file-name>\n", args[0])
			return nil
		}
		return &GetCommand{
			Path: args[1],
		}
	case aliases.Upload.Contains(cmd), cmd == "upload":
		if len(args) != 3 {
			fmt.Printf("Usage: fileport %s <file> <destination-path>\n", args[0])
			return nil
		}
		return &UploadCommand{
			FileName:        args[1],
			DestinationPath: args[2],
		}
	case aliases.Mkdir.Contains(cmd), cmd == "mkdir":
		if len(args) != 2 {
			fmt.Printf("Usage: fileport %s <directory-name>\n", args[0])
			return nil
		}
		return &MkdirCommand{
			DirName: args[1],
		}
	case aliases.Remove.Contains(cmd), cmd == "remove":
		if len(args) != 2 {
			fmt.Printf("Usage: fileport %s <file>\n", args[0])
			return nil
		}
		return &RemoveCommand{
			FileName: args[1],
		}
	case aliases.Rmdir.Contains(cmd), cmd == "rmdir":
		if len(args) != 2 {
			fmt.Printf("Usage: fileport %s <dir>\n", args[0])
			return nil
		}
		return &RmdirCommand{
			DirName: args[1],
		}
	case aliases.Version.Contains(cmd), cmd == "version":
		return &VersionCommand{}
	case cmd == "move":
		if len(args) != 3 {
			fmt.Printf("Usage: fileport %s <target-file> <destination>\n", args[0])
			return nil
		}
		return &MoveCommand{
			Target:      args[1],
			Destination: args[2],
		}
	case aliases.Copy.Contains(cmd), cmd == "copy":
		if len(args) != 3 {
			fmt.Printf("Usage: fileport %s <target-file> <destination>\n", args[0])
			return nil
		}
		return &CopyCommand{
			Source:      args[1],
			Destination: args[2],
		}
	case aliases.Init.Contains(cmd), cmd == "init":
		return &InitCommand{}
	case aliases.Alias.Contains(cmd), cmd == "alias":
		if len(args) != 3 {
			fmt.Printf("Usage: fileport %s <command> <alias>\n", args[0])
			return nil
		}
		return &AliasCommans{
			Command: args[1],
			Alias:   args[2],
		}
	case aliases.Config.Contains(cmd), cmd == "config":
		return &ConfigCommand{}
	case aliases.View.Contains(cmd), cmd == "view":
		return &ViewCommand{}
	case aliases.Stat.Contains(cmd), cmd == "stat":
		if len(args) != 2 {
			fmt.Printf("Usage: fileport %s <target>\n", args[0])
			return nil
		}
		return &StatCommand{
			Target: args[1],
		}
	default:
		fmt.Println("fileport: Invalid argument")
		return nil
	}
}
