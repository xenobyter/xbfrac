# xbfrac

`xbfrac` ist ein interaktiver Mandelbrot-Viewer in Go. Die Fyne-Oberfläche zeigt das Fraktal an und erlaubt es, Ausschnitte zu vergrößern sowie Darstellungsparameter und Farbpaletten anzupassen.

## Funktionen

- Berechnung des Mandelbrot-Satzes für einen konfigurierbaren Ausschnitt und eine einstellbare Iterationsgrenze
- Vier Farbpaletten: Classic, Fire, Ocean und Grayscale
- Vergrößern per Linksklick oder durch Auswahl eines Rechtecks
- Verkleinern per Rechtsklick
- Palettenwechsel mit dem Mausrad
- Export der aktuellen Ansicht als PNG
- Eingabefelder für die Grenzen `xmin`, `xmax`, `ymin`, `ymax` und `maxIter`
- Zurücksetzen auf den Standardausschnitt und die Classic-Palette

Beim Ziehen mit gedrückter Umschalttaste bleibt das Auswahlrechteck im Seitenverhältnis des Bildes. Die Standardansicht reicht von `xmin=-2.2` bis `xmax=1.0` und von `ymin=-1.2` bis `ymax=1.2`; die Iterationsgrenze beträgt 200.

## Voraussetzungen

- Go in der in `go.mod` angegebenen Version (Go 1.26.8) oder neuer
- Für die Fyne-Desktopoberfläche die plattformspezifischen nativen Abhängigkeiten von Fyne. Unter Linux gehören dazu eine C-Toolchain sowie die Entwicklungsbibliotheken für OpenGL und X11.

## Starten

Im Projektverzeichnis ausführen:

```sh
go run .
```

## Tests

```sh
go test ./...
```

## Aufbau

- `mandelbrot.go` berechnet die Iterationswerte und das zweidimensionale Raster.
- `palette.go` stellt Farbpaletten und Farbinterpolation bereit.
- `render.go` wandelt das Raster in ein RGBA-Bild um.
- `gui.go` enthält die Fyne-Oberfläche, Parametervalidierung, Maus- und Zoominteraktionen sowie den PNG-Export.
