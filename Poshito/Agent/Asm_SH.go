//go:build !asm

package main

func executeAssembly(chatId int64, data []byte, assemblyArgs []string, runtime string) { sinkhole(chatId) }
func executeAssemblyByHash(chatId int64, hash string, assemblyArgs []string, runtime string) { sinkhole(chatId) }
func executePowershell(chatId int64, assemblyArgs []string, runtime string) {sinkhole(chatId)}
