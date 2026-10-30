package main

import (
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
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
	maxScale       = 4
	minImageSize   = 50
	resizeDelay    = 300 * time.Millisecond
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

// fitBoundsToAspect erweitert den Bereich um sein Zentrum, sodass die Pixel quadratisch sind.
func fitBoundsToAspect(bounds params, width, height int) params {
	widthSpan, heightSpan := bounds.xmax-bounds.xmin, bounds.ymax-bounds.ymin
	if width <= 0 || height <= 0 || widthSpan <= 0 || heightSpan <= 0 {
		return bounds
	}
	target := float64(width) / float64(height)
	if widthSpan/heightSpan < target {
		widthSpan = heightSpan * target
	} else {
		heightSpan = widthSpan / target
	}
	centerX, centerY := (bounds.xmin+bounds.xmax)/2, (bounds.ymin+bounds.ymax)/2
	fitted := bounds
	fitted.xmin, fitted.xmax = centerX-widthSpan/2, centerX+widthSpan/2
	fitted.ymin, fitted.ymax = centerY-heightSpan/2, centerY+heightSpan/2
	return fitted
}

// clampSize skaliert die Größe unter Beibehaltung des Seitenverhältnisses in die erlaubten Grenzen.
func clampSize(width, height int) (int, int) {
	if width <= 0 || height <= 0 {
		return minImageSize, minImageSize
	}
	scale := 1.0
	if factor := float64(maxScale*imgWidth) / float64(width); factor < scale {
		scale = factor
	}
	if factor := float64(maxScale*imgHeight) / float64(height); factor < scale {
		scale = factor
	}
	width = int(math.Round(float64(width) * scale))
	height = int(math.Round(float64(height) * scale))
	if width < minImageSize {
		width = minImageSize
	}
	if height < minImageSize {
		height = minImageSize
	}
	return width, height
}

func validateSize(width, height int) error {
	if width < minImageSize || height < minImageSize {
		return fmt.Errorf("Größe muss mindestens %d×%d px sein", minImageSize, minImageSize)
	}
	if width > maxScale*imgWidth || height > maxScale*imgHeight {
		return fmt.Errorf("Größe darf höchstens %d×%d px sein", maxScale*imgWidth, maxScale*imgHeight)
	}
	return nil
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

// lockAspect passt end so an, dass das Rechteck das Seitenverhältnis des Bildes hat und im Bild bleibt.
func lockAspect(start, end, imagePosition fyne.Position, imageDrawSize fyne.Size) fyne.Position {
	aspect := imageDrawSize.Width / imageDrawSize.Height
	dx, dy := end.X-start.X, end.Y-start.Y
	signX, signY := float32(1), float32(1)
	maxWidth := imagePosition.X + imageDrawSize.Width - start.X
	if dx < 0 {
		signX = -1
		maxWidth = start.X - imagePosition.X
	}
	maxHeight := imagePosition.Y + imageDrawSize.Height - start.Y
	if dy < 0 {
		signY = -1
		maxHeight = start.Y - imagePosition.Y
	}
	width := absFloat(dx)
	if byHeight := absFloat(dy) * aspect; byHeight > width {
		width = byHeight
	}
	if width > maxWidth {
		width = maxWidth
	}
	if limit := maxHeight * aspect; width > limit {
		width = limit
	}
	return fyne.NewPos(start.X+signX*width, start.Y+signY*width/aspect)
}

func dragSelectionRect(start, end fyne.Position, widgetSize, imageSize fyne.Size, keepAspect bool) (fyne.Position, fyne.Size, bool) {
	imagePosition, imageDrawSize, ok := fittedImageRect(widgetSize, imageSize)
	if !ok || !positionInRect(start, imagePosition, imageDrawSize) {
		return fyne.Position{}, fyne.Size{}, false
	}
	end = clampToRect(end, imagePosition, imageDrawSize)
	if keepAspect {
		end = lockAspect(start, end, imagePosition, imageDrawSize)
	}
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

func mapDragToBounds(start, end fyne.Position, widgetSize, imageSize fyne.Size, bounds params, keepAspect bool) (params, bool) {
	selectionPosition, selectionSize, ok := dragSelectionRect(start, end, widgetSize, imageSize, keepAspect)
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
	onDrag         func(fyne.Position, fyne.Position, fyne.Size, bool)
	onScroll       func(float32)
	onResize       func(fyne.Size)
	lastSize       fyne.Size
	keepAspect     bool
	dragStart      fyne.Position
	dragEnd        fyne.Position
	pointerDown    bool
	dragging       bool
	suppressTap    bool
}

func newTappableImage(image *canvas.Image, imageSize fyne.Size, onTap, onSecondaryTap func(fyne.Position, fyne.Size), onDrag func(fyne.Position, fyne.Position, fyne.Size, bool), onScroll func(float32)) *tappableImage {
	tappable := &tappableImage{
		image: image, imageSize: imageSize, onTap: onTap, onSecondaryTap: onSecondaryTap, onDrag: onDrag, onScroll: onScroll,
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
	if size != r.widget.lastSize {
		r.widget.lastSize = size
		if r.widget.onResize != nil {
			r.widget.onResize(size)
		}
	}
	if !r.widget.dragging {
		r.widget.selection.Hide()
		return
	}
	position, selectionSize, ok := dragSelectionRect(r.widget.dragStart, r.widget.dragEnd, size, r.widget.imageSize, r.widget.keepAspect)
	if !ok {
		r.widget.selection.Hide()
		return
	}
	if r.widget.keepAspect {
		r.widget.selection.FillColor = color.NRGBA{R: 255, G: 140, B: 0, A: 48}
		r.widget.selection.StrokeColor = color.NRGBA{R: 255, G: 140, B: 0, A: 255}
		r.widget.selection.StrokeWidth = 2
	} else {
		r.widget.selection.FillColor = color.NRGBA{R: 0, G: 0, B: 0, A: 24}
		r.widget.selection.StrokeColor = color.White
		r.widget.selection.StrokeWidth = 1
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
	t.keepAspect = event.Modifier&fyne.KeyModifierShift != 0
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

// setKeepAspect wird von den Shift-Tasten-Events des Fensters aufgerufen.
func (t *tappableImage) setKeepAspect(keep bool) {
	if t.keepAspect == keep {
		return
	}
	t.keepAspect = keep
	if t.dragging {
		t.Refresh()
	}
}

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
	start, end, keep := t.dragStart, t.dragEnd, t.keepAspect
	t.dragging = false
	t.Refresh()
	if t.onDrag != nil {
		t.onDrag(start, end, t.Size(), keep)
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

func (t *tappableImage) Scrolled(event *fyne.ScrollEvent) {
	if t.onScroll != nil {
		t.onScroll(event.Scrolled.DY)
	}
}

func absFloat(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}

// runGUI zeigt das Fraktal in einem Fenster mit Parameter-Sidebar an.
func runGUI() error {
	renderW, renderH := imgWidth, imgHeight
	defaults := params{
		xmin: defaultXMin, xmax: defaultXMax,
		ymin: defaultYMin, ymax: defaultYMax,
		maxIter: defaultMaxIter,
	}
	start := fitBoundsToAspect(defaults, renderW, renderH)
	grid, err := Compute(renderW, renderH, start.xmin, start.xmax, start.ymin, start.ymax, start.maxIter)
	if err != nil {
		return err
	}
	paletteOptions := AvailablePalettes()
	paletteNames := make([]string, len(paletteOptions))
	for index, option := range paletteOptions {
		paletteNames[index] = option.Name
	}
	selectedPaletteIndex := 0
	selectedPalette := paletteOptions[selectedPaletteIndex].Palette
	img := canvas.NewImageFromImage(Render(grid, defaultMaxIter, selectedPalette))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	displayed := start

	a := app.New()
	w := a.NewWindow("xbfrac")

	xminE, xmaxE := widget.NewEntry(), widget.NewEntry()
	yminE, ymaxE := widget.NewEntry(), widget.NewEntry()
	iterE := widget.NewEntry()
	var paletteSelect *widget.Select
	calculating := false
	repaint := func() {
		img.Image = Render(grid, displayed.maxIter, selectedPalette)
		img.Refresh()
	}
	paletteSelect = widget.NewSelect(paletteNames, func(name string) {
		if calculating {
			return
		}
		for index, option := range paletteOptions {
			if option.Name == name {
				selectedPaletteIndex = index
				selectedPalette = option.Palette
				repaint()
				return
			}
		}
	})
	paletteSelect.SetSelected(paletteNames[selectedPaletteIndex])
	setBounds := func(b params) {
		xminE.SetText(formatFloat(b.xmin))
		xmaxE.SetText(formatFloat(b.xmax))
		yminE.SetText(formatFloat(b.ymin))
		ymaxE.SetText(formatFloat(b.ymax))
	}
	setBounds(start)
	iterE.SetText(strconv.Itoa(defaultMaxIter))

	imageSize := fyne.NewSize(float32(renderW), float32(renderH))
	var imageWidget *tappableImage
	follow := true
	lockedAspect := float64(renderW) / float64(renderH)
	resizeAgain := false
	widthE, heightE := widget.NewEntry(), widget.NewEntry()
	widthE.SetText(strconv.Itoa(renderW))
	heightE.SetText(strconv.Itoa(renderH))
	widthE.Disable()
	heightE.Disable()
	targetSize := func() (int, int, error) {
		if follow {
			size := imageWidget.Size()
			width, height := clampSize(int(size.Width), int(size.Height))
			return width, height, nil
		}
		width, err := strconv.Atoi(strings.TrimSpace(widthE.Text))
		if err != nil {
			return 0, 0, errors.New("Breite: keine gültige ganze Zahl")
		}
		height := int(math.Round(float64(width) / lockedAspect))
		return width, height, validateSize(width, height)
	}
	var applyWindowSize func()

	var calcBtn, resetBtn, saveBtn *widget.Button
	recalc := func() {
		if calculating {
			return
		}
		pr, err := parseParams(xminE.Text, xmaxE.Text, yminE.Text, ymaxE.Text, iterE.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		width, height, err := targetSize()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		calculating = true
		calcBtn.Disable()
		resetBtn.Disable()
		saveBtn.Disable()
		paletteSelect.Disable()
		go func() {
			g, err := Compute(width, height, pr.xmin, pr.xmax, pr.ymin, pr.ymax, pr.maxIter)
			fyne.Do(func() {
				calculating = false
				calcBtn.Enable()
				resetBtn.Enable()
				saveBtn.Enable()
				paletteSelect.Enable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				displayed = pr
				grid = g
				renderW, renderH = width, height
				imageSize = fyne.NewSize(float32(width), float32(height))
				imageWidget.imageSize = imageSize
				widthE.SetText(strconv.Itoa(width))
				heightE.SetText(strconv.Itoa(height))
				repaint()
				if resizeAgain {
					resizeAgain = false
					applyWindowSize()
				}
			})
		}()
	}
	zoomAt := func(position fyne.Position, size fyne.Size, zoom func(params, complex128) (params, error)) {
		if calculating {
			return
		}
		center, ok := mapTapToComplex(position, size, imageSize, displayed)
		if !ok {
			return
		}
		zoomed, err := zoom(displayed, center)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		setBounds(zoomed)
		recalc()
	}
	applyWindowSize = func() {
		if !follow {
			return
		}
		if calculating {
			resizeAgain = true
			return
		}
		width, height, _ := targetSize()
		if width == renderW && height == renderH {
			return
		}
		setBounds(fitBoundsToAspect(displayed, width, height))
		recalc()
	}
	imageWidget = newTappableImage(img, imageSize,
		func(position fyne.Position, size fyne.Size) {
			zoomAt(position, size, zoomIn)
		},
		func(position fyne.Position, size fyne.Size) {
			zoomAt(position, size, zoomOut)
		},
		func(start, end fyne.Position, size fyne.Size, keepAspect bool) {
			if calculating {
				return
			}
			zoomed, ok := mapDragToBounds(start, end, size, imageSize, displayed, keepAspect)
			if !ok {
				return
			}
			setBounds(zoomed)
			recalc()
		},
		func(deltaY float32) {
			if calculating || deltaY == 0 {
				return
			}
			if deltaY > 0 {
				selectedPaletteIndex = (selectedPaletteIndex + 1) % len(paletteOptions)
			} else {
				selectedPaletteIndex = (selectedPaletteIndex - 1 + len(paletteOptions)) % len(paletteOptions)
			}
			paletteSelect.SetSelected(paletteNames[selectedPaletteIndex])
		})
	var resizeTimer *time.Timer
	imageWidget.onResize = func(fyne.Size) {
		if !follow {
			return
		}
		if resizeTimer != nil {
			resizeTimer.Stop()
		}
		resizeTimer = time.AfterFunc(resizeDelay, func() { fyne.Do(applyWindowSize) })
	}
	followCheck := widget.NewCheck("Fenstergröße folgen", func(on bool) {
		follow = on
		if on {
			widthE.Disable()
			applyWindowSize()
			return
		}
		lockedAspect = float64(renderW) / float64(renderH)
		widthE.Enable()
	})
	followCheck.Checked = true
	if dc, ok := w.Canvas().(desktop.Canvas); ok {
		isShift := func(e *fyne.KeyEvent) bool {
			return e.Name == desktop.KeyShiftLeft || e.Name == desktop.KeyShiftRight
		}
		dc.SetOnKeyDown(func(e *fyne.KeyEvent) {
			if isShift(e) {
				imageWidget.setKeepAspect(true)
			}
		})
		dc.SetOnKeyUp(func(e *fyne.KeyEvent) {
			if isShift(e) {
				imageWidget.setKeepAspect(false)
			}
		})
	}
	calcBtn = widget.NewButton("Neu berechnen", recalc)
	resetBtn = widget.NewButton("Zurücksetzen", func() {
		if width, height, err := targetSize(); err == nil {
			setBounds(fitBoundsToAspect(defaults, width, height))
		}
		iterE.SetText(strconv.Itoa(defaultMaxIter))
		paletteSelect.SetSelected(paletteNames[0])
		recalc()
	})
	saveBtn = widget.NewButton("PNG speichern", func() {
		if calculating {
			return
		}
		exportImage := Render(grid, displayed.maxIter, selectedPalette)
		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return
			}
			encodeErr := png.Encode(writer, exportImage)
			closeErr := writer.Close()
			if encodeErr != nil {
				dialog.ShowError(encodeErr, w)
			} else if closeErr != nil {
				dialog.ShowError(closeErr, w)
			}
		}, w)
		saveDialog.SetFileName("mandelbrot.png")
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
		saveDialog.Show()
	})

	form := container.NewVBox(
		widget.NewLabel("xmin"), xminE,
		widget.NewLabel("xmax"), xmaxE,
		widget.NewLabel("ymin"), yminE,
		widget.NewLabel("ymax"), ymaxE,
		widget.NewLabel("maxIter"), iterE,
		widget.NewLabel("Größe (px)"), followCheck, widthE, heightE,
		widget.NewLabel("Farbpalette"), paletteSelect,
		calcBtn, resetBtn, saveBtn,
	)
	sidebar := container.NewBorder(nil, nil, nil, widget.NewSeparator(),
		container.NewPadded(form))
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(sidebarWidth, 0))
	fixed := container.NewStack(spacer, container.NewVScroll(sidebar))

	w.SetContent(container.NewBorder(nil, nil, fixed, nil, imageWidget))
	w.Resize(fyne.NewSize(imgWidth+sidebarWidth, imgHeight))
	w.ShowAndRun()
	return nil
}
