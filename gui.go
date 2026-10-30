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
	for _, field := range []struct {
		name  string
		input string
		value *float64
	}{
		{"xmin", xmin, &p.xmin},
		{"xmax", xmax, &p.xmax},
		{"ymin", ymin, &p.ymin},
		{"ymax", ymax, &p.ymax},
	} {
		value, err := parseFloatField(field.name, field.input)
		if err != nil {
			return p, err
		}
		*field.value = value
	}
	maxIter, err := strconv.Atoi(strings.TrimSpace(iter))
	if err != nil {
		return p, errors.New("maxIter: keine gültige ganze Zahl")
	}
	p.maxIter = maxIter
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

func mapTapToComplex(position fyne.Position, widgetSize, imageSize fyne.Size, bounds params) (complex128, bool) {
	if widgetSize.Width <= 0 || widgetSize.Height <= 0 || imageSize.Width <= 0 || imageSize.Height <= 0 ||
		bounds.xmin >= bounds.xmax || bounds.ymin >= bounds.ymax {
		return 0, false
	}
	scaleX := widgetSize.Width / imageSize.Width
	scaleY := widgetSize.Height / imageSize.Height
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	drawWidth := imageSize.Width * scale
	drawHeight := imageSize.Height * scale
	left := (widgetSize.Width - drawWidth) / 2
	top := (widgetSize.Height - drawHeight) / 2
	if position.X < left || position.X > left+drawWidth || position.Y < top || position.Y > top+drawHeight {
		return 0, false
	}
	xFraction := (position.X - left) / drawWidth
	yFraction := (position.Y - top) / drawHeight
	real := bounds.xmin + float64(xFraction)*(bounds.xmax-bounds.xmin)
	imag := bounds.ymin + float64(yFraction)*(bounds.ymax-bounds.ymin)
	return complex(real, imag), true
}

func zoomIn(bounds params, center complex128) (params, error) {
	halfWidth := (bounds.xmax - bounds.xmin) / 4
	halfHeight := (bounds.ymax - bounds.ymin) / 4
	zoomed := bounds
	zoomed.xmin = real(center) - halfWidth
	zoomed.xmax = real(center) + halfWidth
	zoomed.ymin = imag(center) - halfHeight
	zoomed.ymax = imag(center) + halfHeight
	if zoomed.xmin >= zoomed.xmax || zoomed.ymin >= zoomed.ymax {
		return bounds, errors.New("Zoomgrenze durch Gleitkommagenauigkeit erreicht")
	}
	return zoomed, nil
}

type tappableImage struct {
	widget.BaseWidget
	image *canvas.Image
	onTap func(fyne.Position, fyne.Size)
}

func newTappableImage(image *canvas.Image, onTap func(fyne.Position, fyne.Size)) *tappableImage {
	tappable := &tappableImage{image: image, onTap: onTap}
	tappable.ExtendBaseWidget(tappable)
	return tappable
}

func (t *tappableImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.image)
}

func (t *tappableImage) Tapped(event *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap(event.Position, t.Size())
	}
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
	displayed := params{
		xmin: defaultXMin, xmax: defaultXMax,
		ymin: defaultYMin, ymax: defaultYMax,
		maxIter: defaultMaxIter,
	}

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
	calculating := false
	recalc := func() {
		if calculating {
			return
		}
		pr, err := parseParams(xminE.Text, xmaxE.Text, yminE.Text, ymaxE.Text, iterE.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		calculating = true
		calcBtn.Disable()
		resetBtn.Disable()
		go func() {
			g, err := Compute(imgWidth, imgHeight, pr.xmin, pr.xmax, pr.ymin, pr.ymax, pr.maxIter)
			fyne.Do(func() {
				calculating = false
				defer calcBtn.Enable()
				defer resetBtn.Enable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				displayed = pr
				img.Image = Render(g, pr.maxIter, p)
				img.Refresh()
			})
		}()
	}
	imageWidget := newTappableImage(img, func(position fyne.Position, size fyne.Size) {
		if calculating {
			return
		}
		center, ok := mapTapToComplex(position, size, fyne.NewSize(imgWidth, imgHeight), displayed)
		if !ok {
			return
		}
		zoomed, err := zoomIn(displayed, center)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		xminE.SetText(formatFloat(zoomed.xmin))
		xmaxE.SetText(formatFloat(zoomed.xmax))
		yminE.SetText(formatFloat(zoomed.ymin))
		ymaxE.SetText(formatFloat(zoomed.ymax))
		recalc()
	})
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

	w.SetContent(container.NewBorder(nil, nil, fixed, nil, imageWidget))
	w.Resize(fyne.NewSize(imgWidth+sidebarWidth, imgHeight))
	w.ShowAndRun()
	return nil
}
