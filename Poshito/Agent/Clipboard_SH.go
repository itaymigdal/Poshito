//go:build !clip

package main

func getClipboard(chatId int64) { sinkhole(chatId) }