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

	tempCopy := input
	parts := strings.Fields(tempCopy)

	args := strings.SplitAfter(input, parts[0])

	// args[0] = strings.Trim(args[0], "[] ")

	if len(parts) == 0 {
		return "", nil
	}
	return args[0], args[1:]
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

func changeDirs(target string) string {
	// ~ (tilde) for Home directory
	if target == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("cd: %s: No such file or directory\n", target)
		}
		target = home
	}

	// Convert relative paths to absolute
	absPath, err := filepath.Abs(target)
	if err != nil {
		return fmt.Sprintf("cd: %s: No such file or directory\n", target)

	}

	// check if the path exists and is a driectory
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		return fmt.Sprintf("cd: %s: No such file or directory\n", target)

	}

	// change directory
	err = os.Chdir(absPath)
	if err != nil {
		return fmt.Sprintf("cd: %s: No such file or directory \n", target)
	}
	return ""
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
		// line = strings.TrimSpace(line)
		// if line == "" {
		// 	continue
		// }

		command, arguments := parseInput(line)

		switch command {
		case "type":
			output := commandInfo(arguments[0])
			fmt.Println(output)

		case "echo":
			// fmt.Println("arguments -> ", arguments)

			arg := strings.Split(arguments[0], "")

			// fmt.Println(arg)

			var result string
			quotIsClosed := true
			checkForSpace := false
			for _, w := range arg {

				if w == "'" {

					if quotIsClosed {
						quotIsClosed = false
						continue
					} else {
						quotIsClosed = true
						continue
					}
				}

				if !quotIsClosed {
					result += w
				} else if w != " " {
					result += w
					checkForSpace = true
				} else if checkForSpace {
					result += " "
					checkForSpace = false
				}
			}
			fmt.Print(result)

		case "pwd":
			cwd, _ := os.Getwd()
			fmt.Println(cwd)

		case "cd":
			if len(arguments) == 0 {
				continue
			}
			target := arguments[0]
			fmt.Println(changeDirs(target))

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
