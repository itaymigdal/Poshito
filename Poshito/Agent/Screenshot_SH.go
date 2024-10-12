//go:build !screen

package main

func takeScreenshots(chatId int64) { sinkhole(chatId) }
