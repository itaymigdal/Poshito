//go:build !bof

package main

func executeBof(chatId int64, bofBytes []byte, bofArgs []string) { sinkhole(chatId) }
