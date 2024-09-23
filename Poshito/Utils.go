package main

import (
	"crypto/md5"
	"encoding/hex"
)


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
