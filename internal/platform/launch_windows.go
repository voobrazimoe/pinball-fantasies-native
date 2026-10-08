//go:build windows

package platform

import (
	"runtime"
	"syscall"
	"unsafe"
)

var shell32 = syscall.NewLazyDLL("shell32.dll")
var ole32 = syscall.NewLazyDLL("ole32.dll")

type browseInfo struct {
	Owner, Root        uintptr
	DisplayName, Title *uint16
	Flags              uint32
	Callback, Param    uintptr
	Image              int32
}

// A thread-local CBT hook renames the standard Yes/No/Cancel buttons as the
// dialog activates, so the plain MessageBox offers the three launch choices.
var launchHook uintptr
var launchHookCallback = syscall.NewCallback(func(code, wParam, lParam uintptr) uintptr {
	if int32(code) == 5 { // HCBT_ACTIVATE
		for id, text := range map[uintptr]string{6: "Import game…", 7: "Play demo", 2: "Quit"} {
			up("SetDlgItemTextW").Call(wParam, id, uintptr(unsafe.Pointer(utf(text))))
		}
		up("UnhookWindowsHookEx").Call(launchHook)
		launchHook = 0
	}
	r, _, _ := up("CallNextHookEx").Call(0, code, wParam, lParam)
	return r
})

func askLaunch() launchChoice {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	thread, _, _ := kernel32.NewProc("GetCurrentThreadId").Call()
	launchHook, _, _ = up("SetWindowsHookExW").Call(5, launchHookCallback, 0, thread) // WH_CBT
	// MB_YESNOCANCEL | MB_ICONINFORMATION | MB_SETFOREGROUND
	r, _, _ := up("MessageBoxW").Call(0, uintptr(unsafe.Pointer(utf(launchText))), uintptr(unsafe.Pointer(utf("Pinball Fantasies"))), 0x3|0x40|0x10000)
	if launchHook != 0 {
		up("UnhookWindowsHookEx").Call(launchHook)
		launchHook = 0
	}
	switch r {
	case 6:
		return launchImport
	case 7:
		return launchDemo
	}
	return launchQuit
}

func pickFolder() (string, bool) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// COINIT_APARTMENTTHREADED, required by the resizable folder dialog.
	if r, _, _ := ole32.NewProc("CoInitializeEx").Call(0, 2); r == 0 || r == 1 {
		defer ole32.NewProc("CoUninitialize").Call()
	}
	name := make([]uint16, 260)
	info := browseInfo{DisplayName: &name[0], Title: utf("Choose the original Pinball Fantasies folder (with INTRO.PRG and TABLE1-4)"),
		Flags: 0x1 | 0x40 | 0x200} // BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE | BIF_NONEWFOLDERBUTTON
	list, _, _ := shell32.NewProc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&info)))
	if list == 0 {
		return "", false
	}
	defer ole32.NewProc("CoTaskMemFree").Call(list)
	path := make([]uint16, 32768)
	if r, _, _ := shell32.NewProc("SHGetPathFromIDListEx").Call(list, uintptr(unsafe.Pointer(&path[0])), uintptr(len(path)), 0); r == 0 {
		return "", false
	}
	return syscall.UTF16ToString(path), true
}

func showMessage(text string) {
	up("MessageBoxW").Call(0, uintptr(unsafe.Pointer(utf(text))), uintptr(unsafe.Pointer(utf("Pinball Fantasies"))), 0x30|0x10000)
}
