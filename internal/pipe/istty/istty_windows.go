//go:build windows
// +build windows

package istty

func isTTY() bool {
	return true
}
