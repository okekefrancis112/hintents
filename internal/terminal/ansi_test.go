// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package terminal

import (
	"os"
	"strings"
	"testing"
)

func TestANSIRenderer_IsTTY(t *testing.T) {
	// Test NO_COLOR — need a fresh renderer for each scenario since sync.Once caches
	t.Run("NO_COLOR", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")
		r := NewANSIRenderer()
		if r.IsTTY() {
			t.Error("IsTTY() should be false when NO_COLOR is set")
		}
	})

	// Test FORCE_COLOR
	t.Run("FORCE_COLOR", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		os.Setenv("FORCE_COLOR", "1")
		defer os.Unsetenv("FORCE_COLOR")
		r := NewANSIRenderer()
		if !r.IsTTY() {
			t.Error("IsTTY() should be true when FORCE_COLOR is set")
		}
	})

	// Test TERM=dumb
	t.Run("TERM_dumb", func(t *testing.T) {
		os.Unsetenv("FORCE_COLOR")
		os.Unsetenv("NO_COLOR")
		os.Setenv("TERM", "dumb")
		defer os.Unsetenv("TERM")
		r := NewANSIRenderer()
		if r.IsTTY() {
			t.Error("IsTTY() should be false when TERM=dumb")
		}
	})
}

func TestANSIRenderer_Colorize(t *testing.T) {
	t.Run("with_color", func(t *testing.T) {
		os.Setenv("FORCE_COLOR", "1")
		os.Unsetenv("NO_COLOR")
		defer os.Unsetenv("FORCE_COLOR")

		r := NewANSIRenderer()
		text := "hello"
		colored := r.Colorize(text, "red")
		if !strings.Contains(colored, "\033[31m") {
			t.Errorf("Expected red color code, got %q", colored)
		}
	})

	t.Run("no_color", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		os.Unsetenv("FORCE_COLOR")
		defer os.Unsetenv("NO_COLOR")

		r := NewANSIRenderer()
		text := "hello"
		plain := r.Colorize(text, "red")
		if strings.Contains(plain, "\033") {
			t.Errorf("Expected plain text when NO_COLOR is set, got %q", plain)
		}
	})
}

func TestANSIRenderer_Symbols(t *testing.T) {
	r := NewANSIRenderer()
	os.Setenv("FORCE_COLOR", "1")
	defer os.Unsetenv("FORCE_COLOR")

	if r.Symbol("check") != "[OK]" {
		t.Errorf("Expected [OK] for check symbol, got %q", r.Symbol("check"))
	}

	os.Setenv("NO_COLOR", "1")
	if r.Symbol("check") != "[OK]" {
		t.Errorf("Expected [OK] for check symbol when NO_COLOR, got %q", r.Symbol("check"))
	}
}
