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

type historyKeeper struct {
	historyList []string
}

func (h *historyKeeper) push(cmd string) {
	h.historyList = append(h.historyList, cmd)
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

func isBuiltin(cmd string) bool {
	builtins := map[string]bool{
		"type":    true,
		"echo":    true,
		"exit":    true,
		"pwd":     true,
		"cd":      true,
		"history": true,
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

func splitPipeLine(line string) []string {

	var parts []string
	var curr strings.Builder
	inSingle := false
	inDouble := false

	for i := 0; i < len(line); i++ {
		ch := rune(line[i])

		switch ch {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '|':
			if !inSingle && !inDouble {
				parts = append(parts, strings.TrimSpace(curr.String()))
				curr.Reset()
				continue
			}
		}
		curr.WriteRune(ch)
	}

	if curr.Len() > 0 {
		parts = append(parts, strings.TrimSpace(curr.String()))
		curr.Reset()
	}

	return parts
}

func executePipeLine(parts []string) {
	var prevRead *os.File
	var cmds []*exec.Cmd

	for i, part := range parts {
		args, _, _ := parseTokens(part)
		if len(args) == 0 {
			return
		}

		isLast := i == len(parts)-1
		var readEnd, writeEnd *os.File

		if !isLast {
			readEnd, writeEnd, _ = os.Pipe()
		}

		if isBuiltin(args[0]) {
			stdin := os.Stdin
			if prevRead != nil {
				stdin = prevRead
			}

			stdout := os.Stdout
			if writeEnd != nil {
				stdout = writeEnd
			}

			runBuiltin(args[0], args[1:], stdin, stdout)

		} else {
			cmd := exec.Command(args[0], args[1:]...)

			if prevRead != nil {
				cmd.Stdin = prevRead
			}

			if writeEnd != nil {
				cmd.Stdout = writeEnd
			} else {
				cmd.Stdout = os.Stdout
			}

			cmd.Stderr = os.Stderr
			cmd.Start()
			cmds = append(cmds, cmd)
		}

		if prevRead != nil {
			prevRead.Close()
		}
		if writeEnd != nil {
			writeEnd.Close()
		}

		prevRead = readEnd
	}

	// WAIT ONLY AFTER ALL COMMANDS START
	for _, cmd := range cmds {
		cmd.Wait()
	}
}

func runBuiltin(cmd string, args []string, stdin io.Reader, stdout io.Writer) {
	switch cmd {
	case "echo":
		fmt.Fprintln(stdout, strings.Join(args, " "))

	case "type":
		if len(args) > 0 {
			fmt.Fprintln(stdout, commandInfo(args[0]))
		}

	case "pwd":
		cwd, _ := os.Getwd()
		fmt.Fprintln(stdout, cwd)
	}
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

	// initialise history
	cmdRecords := historyKeeper{}

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

		// append the cmd in the history
		cmdRecords.push(line)

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

		case "history":

			for i, h := range cmdRecords.historyList {
				fmt.Printf("%d %s \n", i+1, h)
			}
			fmt.Println(cmdRecords.historyList)

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
