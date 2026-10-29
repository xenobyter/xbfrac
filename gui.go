package main

import (
	"errors"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	imgWidth       = 900
	imgHeight      = 600
	sidebarWidth   = 200
	defaultXMin    = -2.2
	defaultXMax    = 1.0
	defaultYMin    = -1.2
	defaultYMax    = 1.2
	defaultMaxIter = 200
)

// params bündelt die Eingaben der Sidebar.
type params struct {
	xmin, xmax, ymin, ymax float64
	maxIter                int
}

func parseFloatField(name, s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(s, ",", ".")), 64)
	if err != nil {
		return 0, errors.New(name + ": keine gültige Zahl")
	}
	return v, nil
}

func parseParams(xmin, xmax, ymin, ymax, iter string) (params, error) {
	var p params
	var err error
	if p.xmin, err = parseFloatField("xmin", xmin); err != nil {
		return p, err
	}
	if p.xmax, err = parseFloatField("xmax", xmax); err != nil {
		return p, err
	}
	if p.ymin, err = parseFloatField("ymin", ymin); err != nil {
		return p, err
	}
	if p.ymax, err = parseFloatField("ymax", ymax); err != nil {
		return p, err
	}
	if p.maxIter, err = strconv.Atoi(strings.TrimSpace(iter)); err != nil {
		return p, errors.New("maxIter: keine gültige ganze Zahl")
	}
	if p.xmin >= p.xmax || p.ymin >= p.ymax {
		return p, errors.New("min muss kleiner als max sein")
	}
	if p.maxIter < 1 {
		return p, errors.New("maxIter muss mindestens 1 sein")
	}
	return p, nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// runGUI zeigt das Fraktal in einem Fenster mit Parameter-Sidebar an.
func runGUI(p Palette) error {
	grid, err := Compute(imgWidth, imgHeight, defaultXMin, defaultXMax, defaultYMin, defaultYMax, defaultMaxIter)
	if err != nil {
		return err
	}
	img := canvas.NewImageFromImage(Render(grid, defaultMaxIter, p))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest

	a := app.New()
	w := a.NewWindow("xbfrac")

	xminE, xmaxE := widget.NewEntry(), widget.NewEntry()
	yminE, ymaxE := widget.NewEntry(), widget.NewEntry()
	iterE := widget.NewEntry()
	setDefaults := func() {
		xminE.SetText(formatFloat(defaultXMin))
		xmaxE.SetText(formatFloat(defaultXMax))
		yminE.SetText(formatFloat(defaultYMin))
		ymaxE.SetText(formatFloat(defaultYMax))
		iterE.SetText(strconv.Itoa(defaultMaxIter))
	}
	setDefaults()

	var calcBtn, resetBtn *widget.Button
	recalc := func() {
		pr, err := parseParams(xminE.Text, xmaxE.Text, yminE.Text, ymaxE.Text, iterE.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		calcBtn.Disable()
		resetBtn.Disable()
		go func() {
			g, err := Compute(imgWidth, imgHeight, pr.xmin, pr.xmax, pr.ymin, pr.ymax, pr.maxIter)
			fyne.Do(func() {
				defer calcBtn.Enable()
				defer resetBtn.Enable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				img.Image = Render(g, pr.maxIter, p)
				img.Refresh()
			})
		}()
	}
	calcBtn = widget.NewButton("Neu berechnen", recalc)
	resetBtn = widget.NewButton("Zurücksetzen", func() {
		setDefaults()
		recalc()
	})

	form := container.NewVBox(
		widget.NewLabel("xmin"), xminE,
		widget.NewLabel("xmax"), xmaxE,
		widget.NewLabel("ymin"), yminE,
		widget.NewLabel("ymax"), ymaxE,
		widget.NewLabel("maxIter"), iterE,
		calcBtn, resetBtn,
	)
	sidebar := container.NewBorder(nil, nil, nil, widget.NewSeparator(),
		container.NewPadded(form))
	fixed := container.NewGridWrap(fyne.NewSize(sidebarWidth, imgHeight), sidebar)

	w.SetContent(container.NewBorder(nil, nil, fixed, nil, img))
	w.Resize(fyne.NewSize(imgWidth+sidebarWidth, imgHeight))
	w.ShowAndRun()
	return nil
}
