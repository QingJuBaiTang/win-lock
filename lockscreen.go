package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// lockWindowTitle is used by the Windows platform hook to find the HWND by title.
const lockWindowTitle = "ClassLock-LOCK"

var lockWin fyne.Window

func showLockScreen() {
	if lockWin != nil {
		bringToFront()
		return
	}

	w := fyneApp.NewWindow(lockWindowTitle)
	w.SetFullScreen(true)
	lockWin = w

	bg := canvas.NewRectangle(color.NRGBA{R: 10, G: 15, B: 35, A: 255})

	timeText := canvas.NewText(time.Now().Format("15:04"), color.White)
	timeText.TextSize = 80
	timeText.Alignment = fyne.TextAlignCenter
	timeText.TextStyle = fyne.TextStyle{Bold: true}

	dateText := canvas.NewText(formatDate(time.Now()), color.NRGBA{R: 170, G: 178, B: 200, A: 255})
	dateText.TextSize = 22
	dateText.Alignment = fyne.TextAlignCenter

	hint := canvas.NewText("Screen locked  -  Enter password to unlock", color.NRGBA{R: 110, G: 120, B: 150, A: 255})
	hint.TextSize = 14
	hint.Alignment = fyne.TextAlignCenter

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Password...")

	errText := canvas.NewText("", color.NRGBA{R: 255, G: 90, B: 90, A: 255})
	errText.TextSize = 13
	errText.Alignment = fyne.TextAlignCenter

	unlockFn := func() {
		if checkPassword(passEntry.Text) {
			closeLockScreen()
		} else {
			errText.Text = "Wrong password, try again"
			errText.Refresh()
			passEntry.SetText("")
		}
	}

	unlockBtn := widget.NewButton("Unlock", unlockFn)
	unlockBtn.Importance = widget.HighImportance
	passEntry.OnSubmitted = func(_ string) { unlockFn() }

	kbBtn := widget.NewButton("Keyboard", func() {
		showOnScreenKeyboard()
	})

	entryWrap := container.New(layout.NewGridWrapLayout(fyne.NewSize(300, 40)), passEntry)

	panel := container.NewVBox(
		timeText,
		dateText,
		widget.NewSeparator(),
		hint,
		widget.NewSeparator(),
		entryWrap,
		errText,
		container.NewCenter(container.NewHBox(unlockBtn, kbBtn)),
	)

	content := container.NewStack(
		bg,
		container.NewCenter(panel),
	)

	w.SetContent(content)
	w.SetCloseIntercept(func() {})

	stopClock := make(chan struct{})
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				n := time.Now()
				timeText.Text = n.Format("15:04")
				dateText.Text = formatDate(n)
				timeText.Refresh()
				dateText.Refresh()
			case <-stopClock:
				return
			}
		}
	}()

	w.SetOnClosed(func() {
		close(stopClock)
		lockWin = nil
	})

	w.Show()
	w.Canvas().Focus(passEntry)
	installPlatformHooks(w)
}

func closeLockScreen() {
	if lockWin == nil {
		return
	}
	uninstallPlatformHooks()
	w := lockWin
	lockWin = nil
	w.Close()
}

func formatDate(t time.Time) string {
	return t.Format("Mon, 02 Jan 2006")
}
