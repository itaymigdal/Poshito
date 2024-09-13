package main

import (
	_ "embed"
	"io"
	"os"
	"fmt"
	"bytes"
	"strconv"
	"net/http"
	"encoding/json"
	"path/filepath"
	"mime/multipart"
)

var (
	//go:embed pass_md5
	passMd5 string
	//go:embed bot_token
	botToken string
	// holds the approved sessions
	chatIDs  []int64
	// Telegram message offset
	offset = 0
	// Telegram APIs
	baseURL  = "https://api.telegram.org/bot" + botToken + "/"
	getFileURL = baseURL + "getFile?file_id="
	getFileURL2 = "https://api.telegram.org/file/bot" + botToken + "/"
	sendfileURL = baseURL + "sendDocument"
	sendMessageURL = baseURL + "sendMessage"
	getUpdatesURL = baseURL + "getUpdates?offset="

)

// Message structure to handle Telegram API messages
type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
	Document *Document `json:"document,omitempty"`
	Caption  string `json:"caption"`
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

func SendMessage(chatID int64, text string) error {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(sendMessageURL, "application/json", bytes.NewBuffer(jsonPayload))
	fmt.Println(resp)
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

func downloadFile(doc *Document, caption string) error {

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
    dir := filepath.Dir(caption)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Create the file
    out, err := os.Create(caption)
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