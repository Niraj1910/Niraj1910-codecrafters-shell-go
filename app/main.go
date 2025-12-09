package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
	dirs := strings.Split(path, ":") // get the all dirs

	for _, dir := range dirs {
		fullPath := filepath.Join(dir, cmd)

		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return fullPath, true
		}
	}

	return "", false
}

func isBuiltin(cmd string) bool {
	builtins := map[string]bool{
		"type": true,
		"echo": true,
		"exit": true,
		"pwd":  true,
		"cd":   true,
	}
	return builtins[cmd]
}

func commandInfo(cmd string) string {
	// check for builin
	if isBuiltin(cmd) {
		return fmt.Sprintf("%s is a shell builtin", cmd)
	}
	//  check for executbale files
	if exe, found := findExecutable(cmd); found {
		return fmt.Sprintf("%s is %s", cmd, exe)
	}

	return fmt.Sprintf("%s not found", cmd)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("cound not read the command: %s", err)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		command, arguments := parseInput(line)

		switch command {
		case "type":
			output := commandInfo(arguments[0])
			fmt.Println(output)

		case "echo":
			fmt.Println(strings.Join(arguments, " "))

		case "pwd":
			cwd, _ := os.Getwd()
			fmt.Println(cwd)

		case "cd":
			if len(arguments) == 0 {
				continue
			}
			target := arguments[0]

			// ~ (tilde) for Home directory
			if target == "~" {
				home, _ := os.UserHomeDir()
				fmt.Printf("cd: %s: No such file or directory\n", target)
				target = home
			}

			// if relative path, join it with the current working directory
			if !filepath.IsAbs(target) {
				cwd, _ := os.Getwd()
				target = filepath.Join(cwd, target)
			}

			// change directory
			err := os.Chdir(target)
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory \n", target)
			}

		case "exit":
			return

		default:

			_, found := findExecutable(command)

			if !found {
				fmt.Println(command + ": command not found")
				continue
			}

			cmd := exec.Command(command, arguments...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			cmd.Run()

		}
	}
}
