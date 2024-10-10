package main

import (
	"fmt"
	"os/exec"
)

func executeCommand(chatID int64, commandParts []string) {

	// Execute the command using the first element as the command and the rest as arguments
	cmd := exec.Command(commandParts[0], commandParts[1:]...)

	// Will hold the response
	responseStr := ""

	// Get the combined output (stdout + stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		responseStr = fmt.Sprintf("Error: %v", err)
		responseStr += "\n" + string(output)
	} else {
		responseStr = string(output)
	}
	// Send message to server
	SendMessage(chatID, responseStr)
}
