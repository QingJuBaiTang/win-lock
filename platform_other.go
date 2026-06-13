//go:build !windows

package main

import "fyne.io/fyne/v2"

func installPlatformHooks(_ fyne.Window) {}
func uninstallPlatformHooks()            {}
func showOnScreenKeyboard()              {}

func setupAutoStart() error      { return nil }
func removeAutoStart() error     { return nil }
func isAutoStartEnabled() bool   { return false }

func bringToFront() {
	if lockWin != nil {
		lockWin.RequestFocus()
	}
}
