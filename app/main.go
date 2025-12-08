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

func parseInput(input string) (string, []string) {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func findExecutable(cmd string) (string, bool) {
	path := os.Getenv("PATH")        // get the full path
	dirs := strings.Split(path, " ") // get the all dirs

	for _, dir := range dirs {
		fullPath := filepath.Join(dir, cmd)
		info, err := os.Stat(fullPath)
		if err != nil && !info.IsDir() && info.Mode().Perm()&0111 != 0 {
			return fullPath, true
		}
	}

	return "", false
}

func isBuiltInCmds(args []string) string {
	builtInCommands := []string{"type", "echo", "exit"}

	// check for builtin commands
	for _, elem := range builtInCommands {
		if args[0] == elem {
			return fmt.Sprintf("%s is a shell builtin", args)
		}
	}

	// check for executable files
	fullPath, found := findExecutable(args[0])
	if found {
		return fmt.Sprintf("%s is %s", args[0], fullPath)
	}

	return fmt.Sprintf("%s not found", args[0])
}

func main() {

	for {
		fmt.Print("$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Printf("cound not read the command: %s", err)
		}
		input = strings.TrimSpace(input)

		command, arguments := parseInput(input)

		switch command {
		case "type":
			output := isBuiltInCmds(arguments)
			fmt.Println(output)
		case "echo":
			fmt.Println(arguments)
		case "exit":
			return
		default:
			fmt.Println(command + ": command not found")
		}
	}
}
