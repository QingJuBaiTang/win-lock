package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/bcrypt"
)

const appID = "com.classroom.winlock"

var (
	fyneApp fyne.App
	cfg     Config
	cfgPath string
)

type Config struct {
	PasswordHash string `json:"password_hash"`
}

func main() {
	fyneApp = app.NewWithID(appID)
	fyneApp.Settings().SetTheme(theme.DarkTheme())

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".win-lock")
	os.MkdirAll(dir, 0700)
	cfgPath = filepath.Join(dir, "config.json")
	loadConfig()

	if cfg.PasswordHash == "" {
		showFirstRunSetup()
	} else {
		startApp()
		showLockScreen() // 每次启动直接锁屏
	}

	fyneApp.Run()
}

func loadConfig() {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &cfg)
}

func saveConfig() error {
	data, err := json.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, data, 0600)
}

func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte(pw)) == nil
}

func showFirstRunSetup() {
	w := fyneApp.NewWindow("ClassLock - First Run Setup")
	w.Resize(fyne.NewSize(380, 250))
	w.CenterOnScreen()

	p1 := widget.NewPasswordEntry()
	p1.SetPlaceHolder("Password")
	p2 := widget.NewPasswordEntry()
	p2.SetPlaceHolder("Confirm password")
	errLbl := widget.NewLabel("")

	btn := widget.NewButton("Set Password & Start", func() {
		if p1.Text == "" {
			errLbl.SetText("Password cannot be empty")
			return
		}
		if p1.Text != p2.Text {
			errLbl.SetText("Passwords do not match")
			return
		}
		hash, err := hashPassword(p1.Text)
		if err != nil {
			errLbl.SetText("Error: " + err.Error())
			return
		}
		cfg.PasswordHash = hash
		if err := saveConfig(); err != nil {
			errLbl.SetText("Save failed: " + err.Error())
			return
		}
		setupAutoStart()
		w.Close()
		startApp()
		showLockScreen()
	})
	btn.Importance = widget.HighImportance

	w.SetCloseIntercept(func() {})
	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("ClassLock - Set your unlock password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		p1, p2,
		errLbl,
		btn,
	))
	w.Show()
}

func showChangePassword() {
	w := fyneApp.NewWindow("Change Password")
	w.Resize(fyne.NewSize(360, 270))
	w.CenterOnScreen()

	old := widget.NewPasswordEntry()
	old.SetPlaceHolder("Current password")
	p1 := widget.NewPasswordEntry()
	p1.SetPlaceHolder("New password")
	p2 := widget.NewPasswordEntry()
	p2.SetPlaceHolder("Confirm new password")
	errLbl := widget.NewLabel("")

	saveBtn := widget.NewButton("Save", func() {
		if !checkPassword(old.Text) {
			errLbl.SetText("Current password is incorrect")
			return
		}
		if p1.Text == "" {
			errLbl.SetText("New password cannot be empty")
			return
		}
		if p1.Text != p2.Text {
			errLbl.SetText("Passwords do not match")
			return
		}
		hash, err := hashPassword(p1.Text)
		if err != nil {
			errLbl.SetText("Error: " + err.Error())
			return
		}
		cfg.PasswordHash = hash
		saveConfig()
		dialog.ShowInformation("Success", "Password updated", w)
		w.Close()
	})
	saveBtn.Importance = widget.HighImportance

	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("Change Unlock Password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		old, p1, p2,
		errLbl,
		saveBtn,
	))
	w.Show()
}

func startApp() {
	// System tray only works reliably on Windows with a proper binary.
	// On macOS/Linux, show a small control window instead.
	if runtime.GOOS != "windows" {
		showFallbackWindow()
		return
	}

	desk, ok := fyneApp.(desktop.App)
	if !ok {
		showFallbackWindow()
		return
	}

	desk.SetSystemTrayIcon(makeTrayIcon())
	desk.SetSystemTrayMenu(fyne.NewMenu("ClassLock",
		fyne.NewMenuItem("Lock Screen", showLockScreen),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Change Password", showChangePassword),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", fyneApp.Quit),
	))
}

func showFallbackWindow() {
	w := fyneApp.NewWindow("ClassLock")
	w.Resize(fyne.NewSize(280, 170))
	w.CenterOnScreen()

	lockBtn := widget.NewButton("Lock Screen", showLockScreen)
	lockBtn.Importance = widget.HighImportance

	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("ClassLock", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		lockBtn,
		widget.NewButton("Change Password", showChangePassword),
	))
	w.SetCloseIntercept(func() { fyneApp.Quit() })
	w.Show()
}
