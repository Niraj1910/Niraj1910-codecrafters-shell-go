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

		command, argument := getCmdAndArg(input)

		if command == "exit" {
			return
		} else if command == "echo" {
			fmt.Println(argument)
		} else {
			fmt.Println(command + ": command not found")
		}
	}
}
