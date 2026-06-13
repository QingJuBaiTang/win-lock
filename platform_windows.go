//go:build windows

package main

import (
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"golang.org/x/sys/windows/registry"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procFindWindowW         = user32.NewProc("FindWindowW")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procGetMessage          = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessage     = user32.NewProc("DispatchMessageW")
	procPostThreadMessage   = user32.NewProc("PostThreadMessageW")
	procGetCurrentThreadId  = kernel32.NewProc("GetCurrentThreadId")
)

const (
	whKeyboardLL = 13
	wmKeydown    = 0x0100
	wmSyskeydown = 0x0104
	wmQuit       = 0x0012
	vkLWin       = 0x5B
	vkRWin       = 0x5C
	vkTab        = 0x09
	vkF4         = 0x73
	llkhfAltdown = 0x20
	hwndTopmost  = ^uintptr(0) // HWND_TOPMOST = -1
	swpNomove    = 0x0002
	swpNosize    = 0x0001
)

type kbdllhookstruct struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type winMsg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX     int32
	PtY     int32
}

var (
	hookHandle   uintptr
	hookThreadID uint32
	stopTopmostC chan struct{}

	// 包级变量保持 callback 不被 GC
	kbHookCb = syscall.NewCallback(keyboardHookProc)
)

func keyboardHookProc(nCode int32, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 {
		kbd := (*kbdllhookstruct)(unsafe.Pointer(lParam))
		switch wParam {
		case wmKeydown, wmSyskeydown:
			// 屏蔽 Windows 键
			if kbd.VkCode == vkLWin || kbd.VkCode == vkRWin {
				return 1
			}
			// 屏蔽 Alt+Tab
			if kbd.VkCode == vkTab && (kbd.Flags&llkhfAltdown != 0) {
				return 1
			}
			// 屏蔽 Alt+F4
			if kbd.VkCode == vkF4 && (kbd.Flags&llkhfAltdown != 0) {
				return 1
			}
		}
	}
	ret, _, _ := procCallNextHookEx.Call(hookHandle, uintptr(nCode), wParam, lParam)
	return ret
}

func installPlatformHooks(_ fyne.Window) {
	stopTopmostC = make(chan struct{})
	go runKeyboardHook()
	go keepWindowTopmost(stopTopmostC)
}

func uninstallPlatformHooks() {
	if stopTopmostC != nil {
		close(stopTopmostC)
		stopTopmostC = nil
	}
	if hookHandle != 0 {
		procUnhookWindowsHookEx.Call(hookHandle)
		hookHandle = 0
	}
	// 发送 WM_QUIT 让消息泵退出
	tid := atomic.LoadUint32(&hookThreadID)
	if tid != 0 {
		procPostThreadMessage.Call(uintptr(tid), wmQuit, 0, 0)
		atomic.StoreUint32(&hookThreadID, 0)
	}
}

func showOnScreenKeyboard() {
	exec.Command("osk.exe").Start()
}

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func setupAutoStart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue("ClassLock", `"`+exe+`"`)
}

func removeAutoStart() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.DeleteValue("ClassLock")
}

func isAutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue("ClassLock")
	return err == nil
}

func bringToFront() {
	title, _ := syscall.UTF16PtrFromString(lockWindowTitle)
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd != 0 {
		procSetForegroundWindow.Call(hwnd)
	}
	_ = title // keep alive
}

// runKeyboardHook 安装全局键盘钩子并维护消息泵。
// 必须锁定 OS 线程，低级钩子依赖线程消息队列。
func runKeyboardHook() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tid, _, _ := procGetCurrentThreadId.Call()
	atomic.StoreUint32(&hookThreadID, uint32(tid))

	h, _, _ := procSetWindowsHookExW.Call(whKeyboardLL, kbHookCb, 0, 0)
	hookHandle = h

	var msg winMsg
	for {
		r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if r == 0 || r == ^uintptr(0) { // WM_QUIT 或错误
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// keepWindowTopmost 每 300ms 将锁屏窗口强制置顶并抢回焦点
func keepWindowTopmost(stop chan struct{}) {
	title, _ := syscall.UTF16PtrFromString(lockWindowTitle)

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
			if hwnd != 0 {
				procSetWindowPos.Call(hwnd, hwndTopmost, 0, 0, 0, 0, swpNomove|swpNosize)
				procSetForegroundWindow.Call(hwnd)
			}
		}
	}
}
