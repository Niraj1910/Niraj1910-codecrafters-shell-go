package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type builtinCompleter struct {
	lastInput string
	tabCount  int
}

func (c *builtinCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])

	// Only autocomplete the first word
	if strings.Contains(input, " ") {
		fmt.Print("\x07")
		return nil, 0
	}
	compl := []string{}

	builtins := []string{"echo", "exit"}

	for _, cmd := range builtins {
		if strings.HasPrefix(cmd, input) {
			suffix := cmd[len(input):] + " "
			return [][]rune{[]rune(suffix)}, pos
		}
	}

	// Executable completion
	for _, execDir := range executablesInPATH() {
		if strings.HasPrefix(execDir, input) {
			compl = append(compl, execDir)
		}
	}

	// --------- STATE MANAGEMENT ------------
	if c.lastInput != input {
		c.lastInput = input
		c.tabCount = 1
	} else {
		c.tabCount++
	}

	return c.handleCompletions(compl, input, pos)
}

func (c *builtinCompleter) handleCompletions(compl []string, input string, pos int) ([][]rune, int) {

	if len(compl) == 0 {
		fmt.Print("\x07")
		return nil, 0
	}

	// Single match -> full completion
	if len(compl) == 1 {
		suffix := compl[0][len(input):] + " "
		return [][]rune{[]rune(suffix)}, pos
	}

	// Multiple matches -> try LCP
	lcp := longestCommonPrefix(compl)

	if len(lcp) > len(input) {
		suffix := lcp[len(input):]
		return [][]rune{[]rune(suffix)}, pos
	}

	// Multiple matches
	if c.tabCount == 1 {
		fmt.Print("\x07")
		return nil, 0
	}

	// Second tab  -> print options
	sort.Strings(compl)

	fmt.Print("\n")
	for i, c := range compl {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(c)
	}

	fmt.Print("\n$ " + input)

	return nil, 0
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]

	for _, s := range strs[1:] {
		i := 0
		for i < len(prefix) && i < len(s) && prefix[i] == s[i] {
			i++
		}
		prefix = prefix[:i]
		if prefix == "" {
			break
		}
	}
	return prefix
}

func executablesInPATH() []string {
	var result []string
	seen := make(map[string]bool)

	path := os.Getenv("PATH")
	dirs := strings.Split(path, ":")

	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			if info.Mode()&0111 != 0 {
				name := file.Name()
				if !seen[name] {
					seen[name] = true
					result = append(result, name)
				}
			}
		}
	}
	return result
}

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
		if i+3 < len(line) && line[i+1] == '>' && line[i+2] == '>' {
			k = i + 3
		} else {
			k = i + 2
		}
	} else {
		if i+2 < len(line) && line[i+1] == '>' {
			k = i + 2
		} else {
			k = i + 1
		}
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
	isStderr := false
	append := false

	// Case: 1>> or 2>>
	if i+2 < len(line) && (line[i] == '1' || line[i] == '2') && line[i+1] == '>' && line[i+2] == '>' {
		isStderr = (line[i] == '2')
		append = true
		return isStderr, append
	}

	// Case: >> (stdout)
	if i+1 < len(line) && line[i] == '>' && line[i+1] == '>' {
		isStderr = false
		append = true
		return isStderr, append
	}

	// Case: 1> or 2>
	if i+1 < len(line) && (line[i] == '1' || line[i] == '2') && line[i+1] == '>' {
		isStderr = (line[i] == '2')
		append = false
		return isStderr, append
	}

	// Case: >
	if line[i] == '>' {
		isStderr = false
		append = false
		return isStderr, append
	}

	return false, false
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

func splitPipeLine(line string) []string {
	var parts []string
	var curr strings.Builder
	inSingle := false
	inDouble := false
	escape := false

	for i := 0; i < len(line); i++ {
		ch := rune(line[i])

		if escape {
			curr.WriteRune(ch)
			escape = false
			continue
		}

		if ch == '\\' {
			escape = true
			curr.WriteRune(ch)
			continue
		}

		switch ch {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
			curr.WriteRune(ch)
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
			curr.WriteRune(ch)
		case '|':
			if !inSingle && !inDouble {
				parts = append(parts, strings.TrimSpace(curr.String()))
				curr.Reset()
			} else {
				curr.WriteRune(ch)
			}
		default:
			curr.WriteRune(ch)
		}
	}

	if curr.Len() > 0 {
		parts = append(parts, strings.TrimSpace(curr.String()))
	}

	return parts
}

func executePipeLine(parts []string) {
	var prevRead *os.File
	var cmds []*exec.Cmd

	// First, create all commands
	for i, part := range parts {
		args, _, _ := parseTokens(part)
		if len(args) == 0 {
			continue
		}

		if isBuiltin(args[0]) {
			// For builtins, we'll handle them specially
			if i == len(parts)-1 {
				// Last command - execute builtin with proper I/O
				runBuiltinInPipeline(args, prevRead, os.Stdout, os.Stderr)
			} else {
				// Middle command - need to create a pipe
				read, write, _ := os.Pipe()
				runBuiltinInPipeline(args, prevRead, write, os.Stderr)
				write.Close()
				prevRead = read
			}
		} else {
			// External command
			cmd := exec.Command(args[0], args[1:]...)
			cmds = append(cmds, cmd)

			// Set up stdin
			if prevRead != nil {
				cmd.Stdin = prevRead
			} else {
				cmd.Stdin = os.Stdin
			}

			// Set up stdout - create pipe if not last command
			if i < len(parts)-1 {
				read, write, _ := os.Pipe()
				cmd.Stdout = write
				prevRead = read
			} else {
				cmd.Stdout = os.Stdout
			}

			cmd.Stderr = os.Stderr
		}
	}

	// Start and wait for external commands
	for _, cmd := range cmds {
		cmd.Start()
	}

	// Wait for all commands to finish
	for _, cmd := range cmds {
		cmd.Wait()
	}

	// Close any remaining file descriptors
	if prevRead != nil {
		prevRead.Close()
	}
}

func runBuiltinInPipeline(args []string, in *os.File, out *os.File, errOut *os.File) {
	// Handle stdin if provided
	if in != nil {
		// For builtins in pipeline, they might need to process stdin
		switch args[0] {
		case "type":
			// type command ignores stdin - it only cares about its argument
			// Read and discard stdin to prevent it from being printed
			io.Copy(io.Discard, in)
		case "echo":
			// echo also ignores stdin
			io.Copy(io.Discard, in)
		default:
			// For other builtins, they might process stdin
			// We'll implement as needed
		}
	}

	var output string

	switch args[0] {
	case "echo":
		output = strings.Join(args[1:], " ") + "\n"

	case "type":
		if len(args) > 1 {
			output = commandInfo(args[1]) + "\n"
		}

	case "pwd":
		cwd, _ := os.Getwd()
		output = cwd + "\n"

	case "cd":
		if len(args) > 1 {
			output = changeDirs(args[1])
		}

	case "exit":
		// In pipeline, exit might not make sense, but we handle it
		output = ""
	}

	if out != nil {
		out.Write([]byte(output))
	}
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

	cfg := readline.Config{
		Prompt:       "$ ",
		AutoComplete: &builtinCompleter{},
	}

	rl, err := readline.NewEx(&cfg)
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()

		if err != nil {
			fmt.Printf("cound not read the command: %s", err)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		pipeline := splitPipeLine(line)
		if len(pipeline) > 1 {
			executePipeLine(pipeline)
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
