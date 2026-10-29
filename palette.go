package main

import "image/color"

// Palette ordnet einer Iterationszahl eine Farbe zu.
type Palette interface {
	Color(iter, maxIter int) color.Color
}

// GradientPalette interpoliert zyklisch zwischen Stützfarben; Punkte der Menge sind Inside.
type GradientPalette struct {
	Stops  []color.RGBA
	Period int // Iterationen pro Durchlauf des Verlaufs
	Inside color.RGBA
}

// NewClassicPalette liefert die klassische blau-weiß-orange-schwarze Palette.
func NewClassicPalette() *GradientPalette {
	return &GradientPalette{
		Stops: []color.RGBA{
			{0, 7, 100, 255},
			{32, 107, 203, 255},
			{237, 255, 255, 255},
			{255, 170, 0, 255},
			{0, 2, 0, 255},
		},
		Period: 64,
		Inside: color.RGBA{0, 0, 0, 255},
	}
}

func (p *GradientPalette) Color(iter, maxIter int) color.Color {
	if iter >= maxIter || len(p.Stops) == 0 || p.Period <= 0 {
		return p.Inside
	}
	n := len(p.Stops)
	pos := float64(iter%p.Period) / float64(p.Period) * float64(n)
	i := int(pos)
	return interpolateColor(p.Stops[i%n], p.Stops[(i+1)%n], pos-float64(i))
}

// interpolateColor mischt a und b entsprechend dem Anteil t, indem es ihre RGB-Kanäle
// interpoliert und eine deckende Farbe zurückgibt. Bei t = 0 übernimmt das Ergebnis
// die RGB-Werte von a, bei t = 1 die von b.
func interpolateColor(a, b color.RGBA, t float64) color.RGBA {
	mix := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}
