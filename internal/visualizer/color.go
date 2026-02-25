// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package visualizer

import (
	"github.com/dotandev/hintents/internal/terminal"
)

// ANSI SGR (Select Graphic Rendition) escape codes for terminal colors.
const (
	sgrRed     = "\033[31m"
	sgrGreen   = "\033[32m"
	sgrYellow  = "\033[33m"
	sgrBlue    = "\033[34m"
	sgrMagenta = "\033[35m"
	sgrCyan    = "\033[36m"
	sgrBold    = "\033[1m"
	sgrDim     = "\033[2m"
)

var defaultRenderer terminal.Renderer = terminal.NewANSIRenderer()

// ColorEnabled reports whether ANSI color output should be used.
func ColorEnabled() bool {
	return defaultRenderer.IsTTY()
}

// Colorize returns text with ANSI color if enabled, otherwise plain text.
func Colorize(text string, color string) string {
	return defaultRenderer.Colorize(text, color)
}

// Success returns a success indicator.
func Success() string {
	return defaultRenderer.Success()
}

// Warning returns a warning indicator.
func Warning() string {
	return defaultRenderer.Warning()
}

// Error returns an error indicator.
func Error() string {
	return defaultRenderer.Error()
}

// Info returns an info indicator.
func Info() string {
	return Colorize("[i]", "cyan")
}

// ContractBoundary returns a visual separator indicating a cross-contract
// transition from fromContract to toContract.
func ContractBoundary(fromContract, toContract string) string {
	text := "--- contract boundary: " + fromContract + " -> " + toContract + " ---"
	return Colorize(text, sgrMagenta+sgrBold)
}

// Symbol returns a symbol that may be styled; when colors disabled, returns plain ASCII equivalent.
//
//nolint:gocyclo
func Symbol(name string) string {
	return defaultRenderer.Symbol(name)
}
