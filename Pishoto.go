package main

import (
	"io"
	"fmt"
	"log"
	"bytes"
	"strconv"
	"net/http"
	"encoding/json"
	_ "embed"
)

var (
	//go:embed bot_token
	botToken string
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

// SendMessage sends a message to the specified chat ID
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
	log.Printf("SendMessage Response: %s", body)

	return nil
}

// GetUpdates fetches updates from the Telegram API
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
			fmt.Printf("Received message: %s\n", text)

			// Echo the received message back to the user
			responseText := "You said: " + text
			err := SendMessage(chatID, responseText)
			if err != nil {
				log.Fatalf("Error sending message: %v", err)
			}
		}
	}
}
