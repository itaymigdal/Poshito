//go:build drm

package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

//go:embed Config/marker
var marker string

const GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS = 0x00000004

func getMachineGuid() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()

	machineGuid, _, err := key.GetStringValue("MachineGuid")
	if err != nil {
		return ""
	}

	return machineGuid
}

func getCurrentModulePath() (string, error) {
	var module windows.Handle
	dummy := func() {}

	ret, _, err := procGetModuleHandleExA.Call(
		uintptr(GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS),
		uintptr(unsafe.Pointer(*(**uintptr)(unsafe.Pointer(&dummy)))),
		uintptr(unsafe.Pointer(&module)),
	)
	if ret == 0 {
		return "", fmt.Errorf("GetModuleHandleExA failed: %v", err)
	}
	defer windows.CloseHandle(module)

	buf := make([]byte, windows.MAX_PATH)
	size, _, err := procGetModuleFileNameA.Call(
		uintptr(module),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if size == 0 {
		return "", fmt.Errorf("GetModuleFileNameA failed: %v", err)
	}

	return string(buf[:size]), nil
}

func drm() bool {
	machineGuid := getMachineGuid()
	if len(machineGuid) == 0 {
		// failed to get MachineGuid
		return false
	}

	// Create machine Id and prepare the string to append
	machineId := md5Hash(strings.TrimSpace(string(machineGuid)))
	toAppend := machineId + marker

	// Get the path of the current executable
	poshitoPath, err := getCurrentModulePath()
	if err != nil {
		// failed to retrieve Poshito path
		return false
	}

	file, err := os.Open(poshitoPath)
	if err != nil {
		// failed to open executable
		return false
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		// failed to get file stats
		return false
	}

	size := stat.Size()
	readSize := int64(32 + len(marker))
	buffer := make([]byte, readSize)
	_, err = file.ReadAt(buffer, size-readSize)
	if err != nil {
		// failed to read machine ID & marker from executable
		return false
	}

	// Check if marker already exists
	if strings.HasSuffix(string(buffer), marker) {
		// Marker already exists
		if strings.HasPrefix(string(buffer), machineId) {
			// Same machine, keep running
			return true
		} else {
			// Different machine - exit
			os.Exit(0)
		}
	}

	// Append machine ID & marker to the executable
	f, err := os.OpenFile(poshitoPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// failed to open executable for writing
		return false
	}
	defer f.Close()

	_, err = f.WriteString(toAppend)
	if err != nil {
		// failed to append machine ID & marker to executable
		return false
	}

	// machine ID & Marker appended successfully
	return true
}
