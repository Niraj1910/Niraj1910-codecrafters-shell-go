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
		case "exit":
			return
		case "echo":
			fmt.Println(argument)
		default:
			fmt.Println(command + ": command not found")
		}
	}
}
