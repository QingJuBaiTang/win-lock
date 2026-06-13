package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// normal and shifted key rows (⌫ handled specially)
var (
	kbNormal = [][]string{
		{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "<-"},
		{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p"},
		{"a", "s", "d", "f", "g", "h", "j", "k", "l"},
		{"z", "x", "c", "v", "b", "n", "m", ".", "@", "-"},
	}
	kbShifted = [][]string{
		{"!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "<-"},
		{"Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P"},
		{"A", "S", "D", "F", "G", "H", "J", "K", "L"},
		{"Z", "X", "C", "V", "B", "N", "M", "_", "+", "="},
	}
)

const (
	keyW float32 = 54
	keyH float32 = 50
)

func sized(w, h float32, btn *widget.Button) fyne.CanvasObject {
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(w, h)), btn)
}

// buildKeyboard returns a touch-friendly QWERTY keyboard.
// onChar is called with the typed string, onBack removes last char, onEnter submits.
func buildKeyboard(onChar func(string), onBack, onEnter func()) fyne.CanvasObject {
	shifted := false

	type keyPair struct {
		btn    *widget.Button
		normal string
		shift  string
	}
	var allKeys []keyPair

	syncLabels := func() {
		for _, k := range allKeys {
			if shifted {
				k.btn.SetText(k.shift)
			} else {
				k.btn.SetText(k.normal)
			}
		}
	}

	makeRow := func(normals, shifts []string) *fyne.Container {
		row := container.NewHBox()
		for i, n := range normals {
			n, s := n, shifts[i]

			if n == "<-" {
				bsBtn := widget.NewButton("<-", onBack)
				row.Add(sized(keyW*1.5, keyH, bsBtn))
				continue
			}

			kp := keyPair{normal: n, shift: s}
			kp.btn = widget.NewButton(n, func() {
				if shifted {
					onChar(s)
				} else {
					onChar(n)
				}
				shifted = false
				syncLabels()
			})
			allKeys = append(allKeys, kp)
			row.Add(sized(keyW, keyH, kp.btn))
		}
		return row
	}

	rows := make([]*fyne.Container, len(kbNormal))
	for i := range kbNormal {
		rows[i] = makeRow(kbNormal[i], kbShifted[i])
	}

	shiftBtn := widget.NewButton("Shift", func() {
		shifted = !shifted
		syncLabels()
	})
	spaceBtn := widget.NewButton("Space", func() { onChar(" ") })
	enterBtn := widget.NewButton("Enter", onEnter)
	enterBtn.Importance = widget.HighImportance

	bottomRow := container.NewHBox(
		sized(keyW*2, keyH, shiftBtn),
		sized(keyW*4, keyH, spaceBtn),
		sized(keyW*2, keyH, enterBtn),
	)

	return container.NewVBox(
		container.NewCenter(rows[0]),
		container.NewCenter(rows[1]),
		container.NewCenter(rows[2]),
		container.NewCenter(rows[3]),
		container.NewCenter(bottomRow),
	)
}
