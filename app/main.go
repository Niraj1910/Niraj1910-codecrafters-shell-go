package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func getCmdAndArg(input string) (string, string) {
	cmd, arg, _ := strings.Cut(input, " ")
	return cmd, arg
}

func isBuiltInCmds(cmd string) string {
	builtInCommands := []string{"type", "echo", "exit"}

	for _, elem := range builtInCommands {
		if cmd == elem {
			return fmt.Sprintf("%s is a shell builtin", cmd)
		}
	}
	return fmt.Sprintf("%s not found", cmd)
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
