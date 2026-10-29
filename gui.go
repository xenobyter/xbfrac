package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const (
	imgWidth  = 900
	imgHeight = 600
	maxIter   = 200
)

// runGUI zeigt das Fraktal in einem Fenster an.
func runGUI(p Palette) error {
	grid, err := Compute(imgWidth, imgHeight, -2.2, 1.0, -1.2, 1.2, maxIter)
	if err != nil {
		return err
	}
	img := canvas.NewImageFromImage(Render(grid, maxIter, p))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest

	a := app.New()
	w := a.NewWindow("xbfrac")
	w.SetContent(img)
	w.Resize(fyne.NewSize(imgWidth, imgHeight))
	w.ShowAndRun()
	return nil
}
