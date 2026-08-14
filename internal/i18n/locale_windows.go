//go:build windows

package i18n

import (
	"syscall"
	"unsafe"
)

func systemLocale() string {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserDefaultLocaleName")
	buf := make([]uint16, 85)
	result, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if result == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
