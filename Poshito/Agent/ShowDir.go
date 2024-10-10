//go:build dir

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func showDir(chatID int64, dirpath string) {
	var result string

	// Check if the path is valid
	if len(dirpath) == 0 {
		SendMessage(chatID, "Empty path")
		return
	}

	// Check if the path exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		SendMessage(chatID, "Path does not exist")
		return
	}

	// Check if the path is a directory
	fileInfo, err := os.Stat(dirpath)
	if err != nil {
		SendMessage(chatID, fmt.Sprintf("Could not read path: %v", err))
		return
	}
	if !fileInfo.IsDir() {
		SendMessage(chatID, "Path leads to a file, not a directory")
		return
	}

	// Read the directory
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		SendMessage(chatID, fmt.Sprintf("Could not read directory: %v", err))
		return
	}

	// Prepare text for directories and files
	var textDirs, textFiles string
	dirCount, fileCount := 0, 0
	const maxItems = 50 // Limit to 10 directories and 10 files

	for _, entry := range entries {
		entryPath := filepath.Join(dirpath, entry.Name())
		if entry.IsDir() {
			if dirCount < maxItems {
				textDirs += fmt.Sprintf("📂 %s\n\n", entryPath)
			}
			dirCount++
		} else {
			if fileCount < maxItems {
				// Get file size in MB
				fileInfo, err := os.Stat(entryPath)
				var sizeInfo string
				if err != nil {
					sizeInfo = "[size unknown]"
				} else {
					fileSize := fileInfo.Size()
					if fileSize >= 1024*1024 {
						// Show in MB if size >= 1 MB
						sizeMB := float64(fileSize) / (1024 * 1024)
						sizeInfo = fmt.Sprintf("[%.2f MB]", sizeMB)
					} else {
						// Show in KB if size < 1 MB
						sizeKB := float64(fileSize) / 1024
						sizeInfo = fmt.Sprintf("[%.2f KB]", sizeKB)
					}
				}
				textFiles += fmt.Sprintf("📄 %s  %s\n\n", entryPath, sizeInfo)
			}
			fileCount++
		}
	}
	// Append "..." if there are more than maxItems directories or files
	if dirCount > maxItems {
		textDirs += "📂 ...\n\n"
	}
	if fileCount > maxItems {
		textFiles += "📄 ...\n\n"
	}
	// Combine text
	result = textDirs + textFiles

	if len(result) == 0 {
		SendMessage(chatID, "Empty directory")
		return
	}
	const MESSAGE_MAX_SIZE = 3500
	if len(result) > MESSAGE_MAX_SIZE {
		SendMessage(chatID, fmt.Sprintf("[message is too long, showing first %d directories and %d files]\n\n%s", maxItems, maxItems, result[:MESSAGE_MAX_SIZE]))
		return
	}
	SendMessage(chatID, result)

}
