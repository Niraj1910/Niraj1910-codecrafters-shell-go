package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage

	for {
		fmt.Print("$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')

		command = strings.TrimSpace(command)
		// fmt.Println(len(command))
		// fmt.Println(command)

		if command == "exit" {
			break
		}

		if err != nil {
			fmt.Printf("cound not read the command: %s", err)
		}

		fmt.Println(command + ": command not found")

	}
}
