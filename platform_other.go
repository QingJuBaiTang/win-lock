//go:build !windows

package main

import "fyne.io/fyne/v2"

// 非 Windows 平台：无键盘钩子，仅基础全屏锁定（用于测试编译）

func installPlatformHooks(_ fyne.Window) {}

func uninstallPlatformHooks() {}

func bringToFront() {
	if lockWin != nil {
		lockWin.RequestFocus()
	}
}
