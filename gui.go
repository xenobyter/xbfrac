package main

import (
	"errors"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
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
	minDragSize    = 4
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

func fittedImageRect(widgetSize, imageSize fyne.Size) (fyne.Position, fyne.Size, bool) {
	if widgetSize.Width <= 0 || widgetSize.Height <= 0 || imageSize.Width <= 0 || imageSize.Height <= 0 {
		return fyne.Position{}, fyne.Size{}, false
	}
	scaleX := widgetSize.Width / imageSize.Width
	scaleY := widgetSize.Height / imageSize.Height
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	drawWidth := imageSize.Width * scale
	drawHeight := imageSize.Height * scale
	return fyne.NewPos((widgetSize.Width-drawWidth)/2, (widgetSize.Height-drawHeight)/2),
		fyne.NewSize(drawWidth, drawHeight), true
}

func positionInRect(position, origin fyne.Position, size fyne.Size) bool {
	return position.X >= origin.X && position.X <= origin.X+size.Width &&
		position.Y >= origin.Y && position.Y <= origin.Y+size.Height
}

func clampToRect(position, origin fyne.Position, size fyne.Size) fyne.Position {
	if position.X < origin.X {
		position.X = origin.X
	} else if position.X > origin.X+size.Width {
		position.X = origin.X + size.Width
	}
	if position.Y < origin.Y {
		position.Y = origin.Y
	} else if position.Y > origin.Y+size.Height {
		position.Y = origin.Y + size.Height
	}
	return position
}

func dragSelectionRect(start, end fyne.Position, widgetSize, imageSize fyne.Size) (fyne.Position, fyne.Size, bool) {
	imagePosition, imageDrawSize, ok := fittedImageRect(widgetSize, imageSize)
	if !ok || !positionInRect(start, imagePosition, imageDrawSize) {
		return fyne.Position{}, fyne.Size{}, false
	}
	end = clampToRect(end, imagePosition, imageDrawSize)
	left, right := start.X, end.X
	top, bottom := start.Y, end.Y
	if left > right {
		left, right = right, left
	}
	if top > bottom {
		top, bottom = bottom, top
	}
	return fyne.NewPos(left, top), fyne.NewSize(right-left, bottom-top), true
}

func mapTapToComplex(position fyne.Position, widgetSize, imageSize fyne.Size, bounds params) (complex128, bool) {
	if bounds.xmin >= bounds.xmax || bounds.ymin >= bounds.ymax {
		return 0, false
	}
	imagePosition, imageDrawSize, ok := fittedImageRect(widgetSize, imageSize)
	if !ok || !positionInRect(position, imagePosition, imageDrawSize) {
		return 0, false
	}
	xFraction := (position.X - imagePosition.X) / imageDrawSize.Width
	yFraction := (position.Y - imagePosition.Y) / imageDrawSize.Height
	real := bounds.xmin + float64(xFraction)*(bounds.xmax-bounds.xmin)
	imag := bounds.ymin + float64(yFraction)*(bounds.ymax-bounds.ymin)
	return complex(real, imag), true
}

func mapDragToBounds(start, end fyne.Position, widgetSize, imageSize fyne.Size, bounds params) (params, bool) {
	selectionPosition, selectionSize, ok := dragSelectionRect(start, end, widgetSize, imageSize)
	if !ok || selectionSize.Width < minDragSize || selectionSize.Height < minDragSize {
		return bounds, false
	}
	first, ok := mapTapToComplex(selectionPosition, widgetSize, imageSize, bounds)
	if !ok {
		return bounds, false
	}
	lastPosition := selectionPosition.Add(fyne.NewPos(selectionSize.Width, selectionSize.Height))
	last, ok := mapTapToComplex(lastPosition, widgetSize, imageSize, bounds)
	if !ok {
		return bounds, false
	}
	zoomed := bounds
	zoomed.xmin, zoomed.xmax = real(first), real(last)
	zoomed.ymin, zoomed.ymax = imag(first), imag(last)
	return zoomed, true
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

func zoomOut(bounds params, center complex128) (params, error) {
	halfWidth := bounds.xmax - bounds.xmin
	halfHeight := bounds.ymax - bounds.ymin
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
	image          *canvas.Image
	selection      *canvas.Rectangle
	imageSize      fyne.Size
	onTap          func(fyne.Position, fyne.Size)
	onSecondaryTap func(fyne.Position, fyne.Size)
	onDrag         func(fyne.Position, fyne.Position, fyne.Size)
	dragStart      fyne.Position
	dragEnd        fyne.Position
	pointerDown    bool
	dragging       bool
	suppressTap    bool
}

func newTappableImage(image *canvas.Image, imageSize fyne.Size, onTap, onSecondaryTap func(fyne.Position, fyne.Size), onDrag func(fyne.Position, fyne.Position, fyne.Size)) *tappableImage {
	tappable := &tappableImage{
		image: image, imageSize: imageSize, onTap: onTap, onSecondaryTap: onSecondaryTap, onDrag: onDrag,
		selection: &canvas.Rectangle{
			FillColor:   color.NRGBA{R: 0, G: 0, B: 0, A: 24},
			StrokeColor: color.White,
			StrokeWidth: 1,
		},
	}
	tappable.selection.Hide()
	tappable.ExtendBaseWidget(tappable)
	return tappable
}

func (t *tappableImage) CreateRenderer() fyne.WidgetRenderer {
	return &tappableImageRenderer{widget: t}
}

type tappableImageRenderer struct {
	widget *tappableImage
}

func (r *tappableImageRenderer) Layout(size fyne.Size) {
	r.widget.image.Resize(size)
	if !r.widget.dragging {
		r.widget.selection.Hide()
		return
	}
	position, selectionSize, ok := dragSelectionRect(r.widget.dragStart, r.widget.dragEnd, size, r.widget.imageSize)
	if !ok {
		r.widget.selection.Hide()
		return
	}
	r.widget.selection.Show()
	r.widget.selection.Move(position)
	r.widget.selection.Resize(selectionSize)
}

func (r *tappableImageRenderer) MinSize() fyne.Size {
	return r.widget.image.MinSize()
}

func (r *tappableImageRenderer) Refresh() {
	r.Layout(r.widget.Size())
	r.widget.image.Refresh()
	r.widget.selection.Refresh()
}

func (r *tappableImageRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.widget.image, r.widget.selection}
}

func (r *tappableImageRenderer) Destroy() {}

func (t *tappableImage) MouseDown(event *desktop.MouseEvent) {
	if event.Button != desktop.MouseButtonPrimary {
		return
	}
	t.suppressTap = false
	if t.onDrag == nil {
		return
	}
	imagePosition, imageDrawSize, ok := fittedImageRect(t.Size(), t.imageSize)
	if !ok || !positionInRect(event.Position, imagePosition, imageDrawSize) {
		return
	}
	t.pointerDown = true
	t.dragging = false
	t.dragStart = event.Position
	t.dragEnd = event.Position
}

func (t *tappableImage) Dragged(event *fyne.DragEvent) {
	if !t.pointerDown {
		return
	}
	t.dragEnd = event.Position
	if !t.dragging && (absFloat(t.dragEnd.X-t.dragStart.X) >= minDragSize || absFloat(t.dragEnd.Y-t.dragStart.Y) >= minDragSize) {
		t.dragging = true
	}
	if t.dragging {
		t.Refresh()
	}
}

func (t *tappableImage) DragEnd() {}

func (t *tappableImage) MouseUp(event *desktop.MouseEvent) {
	if event.Button != desktop.MouseButtonPrimary || !t.pointerDown {
		return
	}
	t.pointerDown = false
	t.dragEnd = event.Position
	if !t.dragging && (absFloat(t.dragEnd.X-t.dragStart.X) >= minDragSize || absFloat(t.dragEnd.Y-t.dragStart.Y) >= minDragSize) {
		t.dragging = true
	}
	if !t.dragging {
		return
	}
	t.suppressTap = true
	start, end := t.dragStart, t.dragEnd
	t.dragging = false
	t.Refresh()
	if t.onDrag != nil {
		t.onDrag(start, end, t.Size())
	}
}

func (t *tappableImage) Tapped(event *fyne.PointEvent) {
	if t.suppressTap {
		t.suppressTap = false
		return
	}
	if t.onTap != nil {
		t.onTap(event.Position, t.Size())
	}
}

func (t *tappableImage) TappedSecondary(event *fyne.PointEvent) {
	if t.onSecondaryTap != nil {
		t.onSecondaryTap(event.Position, t.Size())
	}
}

func absFloat(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
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
	zoomAt := func(position fyne.Position, size fyne.Size, zoom func(params, complex128) (params, error)) {
		if calculating {
			return
		}
		center, ok := mapTapToComplex(position, size, fyne.NewSize(imgWidth, imgHeight), displayed)
		if !ok {
			return
		}
		zoomed, err := zoom(displayed, center)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		xminE.SetText(formatFloat(zoomed.xmin))
		xmaxE.SetText(formatFloat(zoomed.xmax))
		yminE.SetText(formatFloat(zoomed.ymin))
		ymaxE.SetText(formatFloat(zoomed.ymax))
		recalc()
	}
	imageSize := fyne.NewSize(imgWidth, imgHeight)
	imageWidget := newTappableImage(img, imageSize,
		func(position fyne.Position, size fyne.Size) {
			zoomAt(position, size, zoomIn)
		},
		func(position fyne.Position, size fyne.Size) {
			zoomAt(position, size, zoomOut)
		},
		func(start, end fyne.Position, size fyne.Size) {
			if calculating {
				return
			}
			zoomed, ok := mapDragToBounds(start, end, size, imageSize, displayed)
			if !ok {
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
