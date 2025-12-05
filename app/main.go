package main

import (
	"bufio"
	"fmt"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage

	for {
		fmt.Print("$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')

		// command = strings.TrimSpace(command)

		if command == "exit" {
			break
		}

		if err != nil {
			fmt.Printf("cound not read the command: %s", err)
		}

		fmt.Println(command[:len(command)-1] + ": command not found")

	}
}
