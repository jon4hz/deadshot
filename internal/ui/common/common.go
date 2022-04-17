package common

import (
	"github.com/muesli/reflow/wordwrap"
	te "github.com/muesli/termenv"
)

const wrapAt = 60

var (
	// Color wraps termenv.ColorProfile.Color, which produces a termenv color
	// for use in termenv styling.
	Color func(string) te.Color = te.ColorProfile().Color

	// HasDarkBackground stores whether or not the terminal has a dark
	// background.
	HasDarkBackground = te.HasDarkBackground()
)

// Colors for dark and light backgrounds.
var (
	Indigo       = NewColorPair("#7571F9", "#5A56E0")
	SubtleIndigo = NewColorPair("#514DC1", "#7D79F6")
	Cream        = NewColorPair("#FFFDF5", "#FFFDF5")
	YellowGreen  = NewColorPair("#ECFD65", "#04B575")
	Yellow       = NewColorPair("#ECFD65", "#CCB12C")
	Fuschia      = NewColorPair("#EE6FF8", "#EE6FF8")
	Green        = NewColorPair("#04B575", "#04B575")
	Red          = NewColorPair("#ED567A", "#FF4672")
	FaintRed     = NewColorPair("#C74665", "#FF6F91")
	SpinnerColor = NewColorPair("#747373", "#8E8E8E")
	NoColor      = NewColorPair("", "")
)

type ColorPair struct {
	Dark  string
	Light string
}

// NewColorPair is a helper function for creating a ColorPair.
func NewColorPair(dark, light string) ColorPair {
	return ColorPair{dark, light}
}

// Color returns the appropriate termenv.Color for the terminal background.
func (c ColorPair) Color() te.Color {
	if HasDarkBackground {
		return Color(c.Dark)
	}

	return Color(c.Light)
}

func (c ColorPair) String() string {
	if HasDarkBackground {
		return c.Dark
	}
	return c.Light
}

// Wrap wraps lines at a predefined width via package muesli/reflow.
func Wrap(s string) string {
	return wordwrap.String(s, wrapAt)
}

// Subtle applies formatting to strings intended to be "subtle".
func Subtle(s string) string {
	return te.String(s).Foreground(NewColorPair("#5C5C5C", "#9B9B9B").Color()).String()
}
