package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func handleRedirectStdout(filePath string, flagAppend bool) *os.File {
	// check file exists
	_, fileErr := os.Stat(filePath)
	var file *os.File
	var err error
	if os.IsNotExist(fileErr) {
		// create a new file with write permissions
		if flagAppend {
			file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		}

		if err != nil {
			fmt.Printf("err: can not create file: %s \n", err)
			return nil
		}
		return file
	}
	// File exist -> open it for writing + truncate
	if flagAppend {
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	}

	if err != nil {
		fmt.Printf("err: can not open file: %s", err)
		return nil
	}
	return file
}

func extractFilePath(r rune, i int, line string) string {
	var k int
	// Move index to first character after ">" or "1>" or "2>"
	if r == '1' || r == '2' {
		k = i + 2
	} else {
		k = i + 1
	}

	// skip whitespace
	for k < len(line) && (line[k] == ' ' || line[k] == '\t') {
		k++
	}

	// extract ONE filename token
	start := k
	for k < len(line) && line[k] != ' ' && line[k] != '\t' {
		k++
	}

	filePath := strings.TrimSpace(line[start:k])
	return filePath
}

func detectRedirectOrAppend(i int, line string) (bool, bool) {
	var isStderr, append bool

	// Case 1: 1>> or 2>>
	if i+2 < len(line) && (line[i] == '1' || line[i] == '2' && line[i+1] == '>' && line[i+2] == '>') {
		isStderr = (line[i] == '2')
		append = true
	}

	// Case 2: 1> or 2>
	if i+1 < len(line) && (line[i] == '1' || line[i] == '2' && line[i+1] == '>') {
		isStderr = (line[i] == '2')
		append = false
	}

	// Case 3: >> (stdout)
	if i+1 < len(line) && line[i] == '>' && line[i+1] == '>' {
		isStderr = (line[i] == '2')
		append = true
	}

	// /Case 4: > (stdout)
	if line[i] == '>' {
		isStderr = false
		append = false
	}

	return isStderr, append
}

func parseTokens(line string) ([]string, *os.File, *os.File) {
	var args []string
	var cur strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	var redirectStdoutFile *os.File
	var stdoutErrFile *os.File

	for i := 0; i < len(line); i++ {
		r := rune(line[i])

		if r == '\n' || r == '\r' {
			continue
		}

		if r == '>' || (r == '1' || r == '2') && i+1 < len(line) && line[i+1] == '>' {

			isStderr, appendFlag := detectRedirectOrAppend(i, line)

			filePath := extractFilePath(r, i, line)
			file := handleRedirectStdout(filePath, appendFlag)

			if isStderr {
				stdoutErrFile = file
			} else {
				redirectStdoutFile = file
			}
			// add the token built so far
			token := strings.TrimSpace(cur.String())
			if len(token) > 0 {
				args = append(args, token)
			}
			break
		}

		// BACKSLASH HANDLING
		if r == '\\' {
			// Backslash inside single quotes = literal backslash
			if inSingleQuote {
				cur.WriteRune('\\')
				continue
			}

			// Inside double quotes, only \" and \\ are special
			if inDoubleQuote {
				if i+1 < len(line) {
					next := line[i+1]
					if next == '"' || next == '\\' {
						cur.WriteByte(next)
						i++
						continue
					}
				}
				// Otherwise, backslash is literal inside double-quote
				cur.WriteRune('\\')
				continue
			}

			// Outside quotes: escape ANY character
			if i+1 < len(line) {
				cur.WriteByte(line[i+1])
				i++
			}
			continue
		}

		// DOUBLE QUOTE HANDLING
		if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}
		// SINGLE QUOTE HANDLING
		if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}

		if !inSingleQuote && !inDoubleQuote && (r == ' ' || r == '\t') {
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
			continue
		}

		cur.WriteRune(r)
	}

	if cur.Len() > 0 {
		args = append(args, cur.String())
	}

	// for idx, elem := range args {
	// 	fmt.Printf("%d - %s \n", idx, elem)
	// }

	return args, redirectStdoutFile, stdoutErrFile
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
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		tokens, redirectStdoutFile, stdoutErrFile := parseTokens(line)
		command := tokens[0]
		arguments := tokens[1:]

		switch command {
		case "type":
			output := commandInfo(arguments[0])
			fmt.Println(output)

		case "echo":
			output := strings.Join(arguments, " ") + "\n"

			if redirectStdoutFile != nil {
				redirectStdoutFile.Write([]byte(output))
			} else {
				fmt.Print(output)
			}

		case "pwd":
			cwd, _ := os.Getwd()
			fmt.Println(cwd)

		case "cd":
			if len(arguments) == 0 {
				continue
			}
			target := arguments[0]
			dir := changeDirs(target)
			fmt.Print(dir)
			continue

		case "exit":
			return

		default:
			_, found := findExecutable(command)

			if !found {
				fmt.Println(command + ": command not found")
				continue
			}

			// fmt.Print(arguments)

			cmd := exec.Command(command, arguments...)
			cmd.Stdin = os.Stdin

			if redirectStdoutFile != nil {
				cmd.Stdout = redirectStdoutFile
			} else {
				cmd.Stdout = os.Stdout
			}

			if stdoutErrFile != nil {
				cmd.Stderr = stdoutErrFile
			} else {
				cmd.Stderr = os.Stderr
			}

			cmd.Run()

		}

		if redirectStdoutFile != nil {
			redirectStdoutFile.Close()
		}
		if stdoutErrFile != nil {
			stdoutErrFile.Close()
		}
	}
}
