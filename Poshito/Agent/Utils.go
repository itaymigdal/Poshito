package main

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"
)

func calcSleepTime(timeframe int, jitterPercent int) int {
	rand.Seed(time.Now().UnixNano())

	jitterRange := (jitterPercent * timeframe) / 100
	jitterRandom := rand.Intn(jitterRange+1) - jitterRange/2

	return timeframe + jitterRandom
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
