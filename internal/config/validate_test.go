// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"testing"
)

func TestTimeoutValidator_Zero(t *testing.T) {
	cfg := &Config{RequestTimeout: 0}
	err := TimeoutValidator{}.Validate(cfg)
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
	if !strings.Contains(err.Error(), "greater than 0") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTimeoutValidator_Negative(t *testing.T) {
	cfg := &Config{RequestTimeout: -5}
	err := TimeoutValidator{}.Validate(cfg)
	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
}

func TestTimeoutValidator_TooLarge(t *testing.T) {
	cfg := &Config{RequestTimeout: 999}
	err := TimeoutValidator{}.Validate(cfg)
	if err == nil {
		t.Fatal("expected error for timeout > 300")
	}
	if !strings.Contains(err.Error(), "at most") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTimeoutValidator_ValidBounds(t *testing.T) {
	for _, v := range []int{1, 15, 150, 300} {
		cfg := &Config{RequestTimeout: v}
		if err := (TimeoutValidator{}).Validate(cfg); err != nil {
			t.Errorf("timeout %d should be valid: %v", v, err)
		}
	}
}

func TestCrashReportingValidator_DisabledOK(t *testing.T) {
	cfg := &Config{CrashReporting: false}
	if err := (CrashReportingValidator{}).Validate(cfg); err != nil {
		t.Errorf("disabled crash reporting should pass: %v", err)
	}
}

func TestCrashReportingValidator_EnabledNoEndpoint(t *testing.T) {
	cfg := &Config{CrashReporting: true}
	err := CrashReportingValidator{}.Validate(cfg)
	if err == nil {
		t.Fatal("expected error when crash reporting enabled with no endpoint")
	}
	if !strings.Contains(err.Error(), "neither crash_endpoint nor crash_sentry_dsn") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCrashReportingValidator_EnabledWithEndpoint(t *testing.T) {
	cfg := &Config{CrashReporting: true, CrashEndpoint: "https://crash.example.com"}
	if err := (CrashReportingValidator{}).Validate(cfg); err != nil {
		t.Errorf("should pass with crash_endpoint set: %v", err)
	}
}

func TestCrashReportingValidator_EnabledWithDSN(t *testing.T) {
	cfg := &Config{CrashReporting: true, CrashSentryDSN: "https://key@sentry.io/1"}
	if err := (CrashReportingValidator{}).Validate(cfg); err != nil {
		t.Errorf("should pass with valid sentry dsn: %v", err)
	}
}

func TestCrashReportingValidator_BadDSNScheme(t *testing.T) {
	cfg := &Config{CrashReporting: true, CrashSentryDSN: "http://key@sentry.io/1"}
	err := CrashReportingValidator{}.Validate(cfg)
	if err == nil {
		t.Fatal("expected error for http sentry dsn")
	}
	if !strings.Contains(err.Error(), "https scheme") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCompositeValidate_FirstFailure(t *testing.T) {
	cfg := &Config{
		RpcUrl:         "",
		Network:        NetworkTestnet,
		LogLevel:       "info",
		RequestTimeout: 15,
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation to fail for empty rpc_url")
	}
}

func TestCompositeValidate_AllPass(t *testing.T) {
	cfg := &Config{
		RpcUrl:         "https://soroban-testnet.stellar.org",
		Network:        NetworkTestnet,
		LogLevel:       "info",
		RequestTimeout: 15,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected all validators to pass, got %v", err)
	}
}

func TestCompositeValidate_MultipleIssues(t *testing.T) {
	cfg := &Config{
		RpcUrl:         "ftp://bad",
		Network:        Network("bogus"),
		LogLevel:       "verbose",
		RequestTimeout: -1,
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation to fail")
	}
}

func BenchmarkValidators(b *testing.B) {
	cfg := &Config{
		RpcUrl:         "https://soroban-testnet.stellar.org",
		Network:        NetworkTestnet,
		LogLevel:       "info",
		RequestTimeout: 15,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}
