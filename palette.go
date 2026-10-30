package main

import "image/color"

// Palette ordnet einem Iterationswert eine Farbe zu und kennzeichnet damit Punkte innerhalb der Menge.
type Palette interface {
	// Color liefert die Farbe für iter; Werte ab maxIter gelten als innerhalb der Menge.
	Color(iter, maxIter int) color.Color
}

// GradientPalette interpoliert zyklisch zwischen Stützfarben; Punkte innerhalb der Menge erhalten Inside.
type GradientPalette struct {
	Stops []color.RGBA // Farben, zwischen denen der Verlauf interpoliert.
	Period int          // Iterationszahl pro vollständigem Verlaufszyklus.
	Inside color.RGBA   // Farbe für Punkte innerhalb der Menge.
}

// PaletteOption verbindet den Anzeigenamen einer Palette mit ihrer Implementierung.
type PaletteOption struct {
	Name    string  // Name, der in der Oberfläche angezeigt wird.
	Palette Palette // Implementierung, die Iterationswerte einfärbt.
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

// AvailablePalettes liefert die in der Oberfläche angebotenen Farbpaletten.
func AvailablePalettes() []PaletteOption {
	return []PaletteOption{
		{Name: "Classic", Palette: NewClassicPalette()},
		{Name: "Fire", Palette: &GradientPalette{
			Stops:  []color.RGBA{{20, 0, 40, 255}, {180, 20, 0, 255}, {255, 220, 40, 255}, {255, 255, 220, 255}},
			Period: 80,
			Inside: color.RGBA{0, 0, 0, 255},
		}},
		{Name: "Ocean", Palette: &GradientPalette{
			Stops:  []color.RGBA{{0, 18, 60, 255}, {0, 120, 160, 255}, {160, 245, 220, 255}, {10, 40, 100, 255}},
			Period: 80,
			Inside: color.RGBA{0, 0, 0, 255},
		}},
		{Name: "Grayscale", Palette: &GradientPalette{
			Stops:  []color.RGBA{{0, 0, 0, 255}, {255, 255, 255, 255}},
			Period: 64,
			Inside: color.RGBA{0, 0, 0, 255},
		}},
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
