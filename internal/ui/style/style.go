package style

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	te "github.com/muesli/termenv"
)

var (
	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	MainStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#FF4672"})
	SecondStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#ECFD65", Dark: "#ECFD65"})
	ErrStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#FF4672"})
	SubtleStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})
)

var HasDarkBackground = te.HasDarkBackground()

var profile te.Profile

func init() {
	profile = te.ColorProfile()
}

type color map[te.Profile]lipgloss.AdaptiveColor

func (c color) get() lipgloss.AdaptiveColor {
	color, ok := c[profile]
	if !ok {
		return lipgloss.AdaptiveColor{}
	}
	return color
}

var subtleColor = color{
	te.ANSI:      {Light: "#9B9B9B", Dark: "#5C5C5C"},
	te.TrueColor: {Light: "#9B9B9B", Dark: "#5C5C5C"},
}

var creamColor = color{
	te.ANSI:      {Light: "#FFFDF5", Dark: "#FFFDF5"},
	te.TrueColor: {Light: "#FFFDF5", Dark: "#FFFDF5"},
}

var redColor = color{
	te.ANSI:      {Light: "#800000", Dark: "#800000"},
	te.TrueColor: {Light: "#ED567A", Dark: "#FF4672"},
}

func GetListTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(MainStyle.GetForeground()).Foreground(lipgloss.Color("#FFF"))
}

func GetMainColor() lipgloss.TerminalColor {
	return MainStyle.GetForeground()
}

func GetSecondColor() lipgloss.TerminalColor {
	return SecondStyle.GetForeground()
}

func GetActiveColor() lipgloss.TerminalColor {
	return GetMainColor()
}

func GetErrColor() lipgloss.AdaptiveColor {
	return GetRedColor()
}

func GetRedColor() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: "#ED567A", Dark: "#FF4672"}
}

func GetInactiveColor() lipgloss.AdaptiveColor {
	return subtleColor.get()
}

func GetSpinnerPoints() spinner.Model {
	spin := spinner.NewModel()
	switch profile {
	case te.ANSI, te.TrueColor:
		spin.Spinner = spinner.Points
	}
	spin.Style = MainStyle.Copy()
	return spin
}

func GetSpinnerDot() spinner.Model {
	spin := spinner.NewModel()
	switch profile {
	case te.ANSI:
		spin.Spinner = spinner.Points
	case te.TrueColor:
		spin.Spinner = spinner.Dot
	}
	spin.Style = MainStyle.Copy()
	return spin
}

func GetPrompt() string {
	return "> "
}

func GetFocusedPrompt() string {
	return MainStyle.Render("> ")
}

func GetCustomPrompt() string {
	return "  "
}

func GetFocusedCustomPrompt() string {
	return GetFocusedPrompt()
}

func GetFocusedText(text string) string {
	return MainStyle.Render(text)
}

// BUTTONS

// ButtonView renders something that resembles a button.
func ButtonView(text string, focused bool) string {
	return buttonStyling(text, false, focused)
}

// YesButtonView return a button reading "Yes".
func YesButtonView(focused bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("Y", true, focused) +
		buttonStyling("es  ", false, focused)
}

// NoButtonView returns a button reading "No.".
func NoButtonView(focused bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("N", true, focused) +
		buttonStyling("o  ", false, focused)
}

// OKButtonView returns a button reading "OK".
func OKButtonView(focused bool, defaultButton bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("OK", defaultButton, focused) +
		buttonStyling("  ", false, focused)
}

// CancelButtonView returns a button reading "Cancel.".
func CancelButtonView(focused bool, defaultButton bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("Cancel", defaultButton, focused) +
		buttonStyling("  ", false, focused)
}

func buttonStyling(str string, underline, focused bool) string {
	s := lipgloss.NewStyle().Foreground(creamColor.get())
	if focused {
		s = s.Background(MainStyle.GetForeground())
	} else {
		s = s.Background(lipgloss.AdaptiveColor{Light: "#827983", Dark: "#BDB0BE"})
	}
	if underline {
		s = s.Underline(true)
	}
	return s.Render(str)
}

// DisagreeButtonView returns a button reading "I disagree.".
func DisagreeButtonView(focused, defaultButton bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("I disagree", defaultButton, focused) +
		buttonStyling("  ", false, focused)
}

// AgreeButtonView returns a button reading "I agree.".
func AgreeButtonView(focused, defaultButton bool) string {
	return buttonStyling("  ", false, focused) +
		buttonStyling("I agree", defaultButton, focused) +
		buttonStyling("  ", false, focused)
}

func GetActiveCursor() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetMainColor())
}
