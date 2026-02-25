// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package visualizer

import "github.com/dotandev/hintents/internal/terminal"

const (
	sgrReset   = "\033[0m"
	sgrBold    = "\033[1m"
	sgrDim     = "\033[2m"
	sgrRed     = "\033[31m"
	sgrGreen   = "\033[32m"
	sgrYellow  = "\033[33m"
	sgrBlue    = "\033[34m"
	sgrMagenta = "\033[35m"
	sgrCyan    = "\033[36m"
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

// Info returns an info indicator with theme-aware coloring.
func Info() string {
	if ColorEnabled() {
		return themeColors("info") + "[i]" + sgrReset
	}
	return "[i]"
}

// ContractBoundary returns a visual separator for cross-contract call transitions.
func ContractBoundary(fromContract, toContract string) string {
	if ColorEnabled() {
		return sgrMagenta + sgrBold + "--- contract boundary: " + fromContract + " -> " + toContract + " ---" + sgrReset
	}
	return "--- contract boundary: " + fromContract + " -> " + toContract + " ---"
}

// Symbol returns a symbol that may be styled; when colors are disabled, returns plain ASCII.
//
//nolint:gocyclo
func Symbol(name string) string {
	return defaultRenderer.Symbol(name)
}
