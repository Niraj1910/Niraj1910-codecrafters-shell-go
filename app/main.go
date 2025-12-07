package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func getCmdAndArg(input string) (string, string) {
	cmd, arg, _ := strings.Cut(input, " ")
	return cmd, arg
}

func isBuiltInCmds(arg string) string {
	builtInCommands := []string{"type", "echo", "exit"}

	// check for builtin commands
	for _, elem := range builtInCommands {
		if arg == elem {
			return fmt.Sprintf("%s is a shell builtin", arg)
		}
	}

	// check for executable files
	path := os.Getenv("PATH")
	dirs := strings.Split(path, ":")

	for _, dir := range dirs {
		fullPath := filepath.Join(dir, arg)
		fileInfo, err := os.Stat(fullPath)
		if err == nil && !fileInfo.IsDir() {
			mode := fileInfo.Mode()
			// check execute permission
			if mode&0111 != 0 {
				return fmt.Sprintf("%s is %s", arg, fullPath)
			}
		}
	}

	return fmt.Sprintf("%s not found", arg)
}

func main() {
	// TODO: Uncomment the code below to pass the first stage

	for {
		fmt.Print("$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Printf("cound not read the command: %s", err)
		}
		input = strings.TrimSpace(input)

		command, argument := getCmdAndArg(input)

		switch command {
		case "type":
			output := isBuiltInCmds(argument)
			fmt.Println(output)
		case "echo":
			fmt.Println(argument)
		case "exit":
			return
		default:
			fmt.Println(command + ": command not found")
		}
	}
}
