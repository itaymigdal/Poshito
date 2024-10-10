package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	//go:embed Config/pass_md5
	passMd5 string
	//go:embed Config/bot_token
	botToken string
	// holds the approved sessions
	chatIDs []int64
	// Telegram message offset
	offset = 0
	// Telegram APIs
	baseURL        = "https://api.telegram.org/bot" + botToken + "/"
	getFileURL     = baseURL + "getFile?file_id="
	getFileURL2    = "https://api.telegram.org/file/bot" + botToken + "/"
	sendfileURL    = baseURL + "sendDocument"
	sendMessageURL = baseURL + "sendMessage"
	getUpdatesURL  = baseURL + "getUpdates?offset="
	// Telegram's maximum message length
	MaxMessageLength = 4096
)

// Message structure to handle Telegram API messages
type Message struct {
	Chat     Chat      `json:"chat"`
	Text     string    `json:"text"`
	Document *Document `json:"document,omitempty"`
	Caption  string    `json:"caption"`
}

// Document structure to hold File structure
type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
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

func splitMessage(text string, maxLength int) []string {
	var messages []string
	for len(text) > 0 {
		if len(text) <= maxLength {
			messages = append(messages, text)
			break
		}

		// Find the last space within the maxLength
		lastSpace := strings.LastIndex(text[:maxLength], " ")
		if lastSpace == -1 {
			// If no space found, just cut at maxLength
			lastSpace = maxLength
		}

		messages = append(messages, text[:lastSpace])
		text = text[lastSpace:]
	}
	return messages
}

func SendMessage(chatID int64, text string) error {
	if chatID == 0 {
		return nil
	}
	// Split the message if it's too long
	messages := splitMessage(text, MaxMessageLength)

	for _, msg := range messages {
		payload := map[string]interface{}{
			"chat_id": chatID,
			"text":    msg,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		resp, err := http.Post(sendMessageURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return err
		}

		if !result["ok"].(bool) {
			return fmt.Errorf("failed to send message: %v", result["description"])
		}
	}

	return nil
}

func sendFile(chatID int64, fileName string, fileData []byte) error {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add chatID to form data
	err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	if err != nil {
		return err
	}

	// Create form file
	part, err := writer.CreateFormFile("document", fileName)
	if err != nil {
		return err
	}

	// Write file data to form file
	_, err = part.Write(fileData)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", sendfileURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}

func GetUpdates(offset int) (Response, error) {
	url := getUpdatesURL + strconv.Itoa(offset)

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

func downloadFile(doc *Document, path string) error {

	// First, get the file path
	resp, err := http.Get(getFileURL + doc.FileID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Ok {
		return fmt.Errorf("failed to get file path")
	}

	// Now download the file
	resp, err = http.Get(getFileURL2 + result.Result.FilePath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func downloadFileBytes(doc *Document) ([]byte, error) {

	// First, get the file path
	resp, err := http.Get(getFileURL + doc.FileID)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []byte{}, err
	}

	if !result.Ok {
		return []byte{}, err
	}

	// Now download the file
	resp, err = http.Get(getFileURL2 + result.Result.FilePath)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	// Convert to byte array
	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return fileBytes, nil
}
