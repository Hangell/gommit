//go:build windows
// +build windows

package platform

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleOutputCP = kernel32.NewProc("GetConsoleOutputCP")
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
	procGetStdHandle       = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode     = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode     = kernel32.NewProc("SetConsoleMode")
	procGetConsoleCP       = kernel32.NewProc("GetConsoleCP")
	procSetConsoleCP       = kernel32.NewProc("SetConsoleCP")
)

const (
	// Handles padrão
	STD_INPUT_HANDLE  = ^uintptr(9) + 1  // -10
	STD_OUTPUT_HANDLE = ^uintptr(10) + 1 // -11
	STD_ERROR_HANDLE  = ^uintptr(11) + 1 // -12

	// Flags de console
	ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	ENABLE_VIRTUAL_TERMINAL_INPUT      = 0x0200

	// Code page
	CP_UTF8 = 65001
)

func init() {
	setupWindowsConsole()
}

func setupWindowsConsole() {
	setConsoleOutputCP(CP_UTF8)
	setConsoleInputCP(CP_UTF8)          // <- entrada também em UTF-8
	enableVTOnHandle(STD_OUTPUT_HANDLE) // stdout com ANSI/VT100
	enableVTOnHandle(STD_ERROR_HANDLE)  // stderr idem
	enableVTInput()
}

func enableVTInput() {
	h, _, _ := procGetStdHandle.Call(STD_INPUT_HANDLE)
	if h == 0 || h == uintptr(syscall.InvalidHandle) {
		return
	}
	var mode uint32
	if r1, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); r1 != 0 {
		procSetConsoleMode.Call(h, uintptr(mode|ENABLE_VIRTUAL_TERMINAL_INPUT))
	}
}

func setConsoleOutputCP(codePage uint32) {
	procSetConsoleOutputCP.Call(uintptr(codePage))
}

func getConsoleOutputCP() uint32 {
	ret, _, _ := procGetConsoleOutputCP.Call()
	return uint32(ret)
}

func setConsoleInputCP(codePage uint32) {
	procSetConsoleCP.Call(uintptr(codePage))
}

func getConsoleInputCP() uint32 {
	ret, _, _ := procGetConsoleCP.Call()
	return uint32(ret)
}

func enableVTOnHandle(handleKind uintptr) {
	h, _, _ := procGetStdHandle.Call(handleKind)
	if h == 0 || h == uintptr(syscall.InvalidHandle) {
		return
	}
	var mode uint32
	if r1, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); r1 == 0 {
		return
	}
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
	procSetConsoleMode.Call(h, uintptr(mode))
}

// SupportsUnicode indica se o console está OK para UTF-8/emoji
func SupportsUnicode() bool {
	if os.Getenv("NO_EMOJI") == "1" || os.Getenv("TERM") == "dumb" {
		return false
	}
	// Terminais modernos no Windows
	if os.Getenv("WT_SESSION") != "" || // Windows Terminal
		os.Getenv("ConEmuPID") != "" || // ConEmu/Cmder
		os.Getenv("VSCODE_INJECTION") == "1" { // VSCode
		return true
	}
	// Se entrada/saída estão em UTF-8, consideramos suportado
	return getConsoleOutputCP() == CP_UTF8 && getConsoleInputCP() == CP_UTF8
}
