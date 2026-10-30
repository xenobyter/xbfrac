package main

import (
	"math"
	"testing"
)

func TestFitBoundsToAspect(t *testing.T) {
	base := params{xmin: -2, xmax: 2, ymin: -1, ymax: 1, maxIter: 50}
	tests := []struct {
		name          string
		width, height int
	}{
		{"breiter", 800, 200},
		{"höher", 200, 800},
		{"gleich", 400, 200},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fitBoundsToAspect(base, tc.width, tc.height)
			ratio := (got.xmax - got.xmin) / (got.ymax - got.ymin)
			if want := float64(tc.width) / float64(tc.height); math.Abs(ratio-want) > 1e-12 {
				t.Errorf("Seitenverhältnis = %v, erwartet %v", ratio, want)
			}
			if cx := (got.xmin + got.xmax) / 2; math.Abs(cx) > 1e-12 {
				t.Errorf("Zentrum x = %v", cx)
			}
			if cy := (got.ymin + got.ymax) / 2; math.Abs(cy) > 1e-12 {
				t.Errorf("Zentrum y = %v", cy)
			}
			if got.xmax-got.xmin < base.xmax-base.xmin-1e-12 || got.ymax-got.ymin < base.ymax-base.ymin-1e-12 {
				t.Error("Bereich darf nicht schrumpfen")
			}
			if got.maxIter != base.maxIter {
				t.Error("maxIter muss erhalten bleiben")
			}
		})
	}
}

func TestClampSize(t *testing.T) {
	tests := []struct {
		inW, inH, wantW, wantH int
	}{
		{900, 600, 900, 600},
		{7200, 4800, maxScale * imgWidth, maxScale * imgHeight},
		{10, 10, minImageSize, minImageSize},
		{0, 0, minImageSize, minImageSize},
	}
	for _, tc := range tests {
		w, h := clampSize(tc.inW, tc.inH)
		if w != tc.wantW || h != tc.wantH {
			t.Errorf("clampSize(%d, %d) = %d×%d, erwartet %d×%d", tc.inW, tc.inH, w, h, tc.wantW, tc.wantH)
		}
	}
}

func TestValidateSize(t *testing.T) {
	tests := []struct {
		w, h    int
		wantErr bool
	}{
		{900, 600, false},
		{maxScale * imgWidth, maxScale * imgHeight, false},
		{maxScale*imgWidth + 1, 600, true},
		{900, maxScale*imgHeight + 1, true},
		{minImageSize - 1, 600, true},
		{900, minImageSize - 1, true},
	}
	for _, tc := range tests {
		if err := validateSize(tc.w, tc.h); (err != nil) != tc.wantErr {
			t.Errorf("validateSize(%d, %d) Fehler = %v", tc.w, tc.h, err)
		}
	}
}
