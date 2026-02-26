// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
)

// New is a proxy to the standard errors.New
func New(text string) error {
	return errors.New(text)
}

// Is is a proxy to the standard errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As is a proxy to the standard errors.As
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Sentinel errors for comparison with errors.Is
var (
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrRPCConnectionFailed  = errors.New("RPC connection failed")
	ErrRPCTimeout           = errors.New("RPC request timed out")
	ErrAllRPCFailed         = errors.New("all RPC endpoints failed")
	ErrSimulatorNotFound    = errors.New("simulator binary not found")
	ErrSimulationFailed     = errors.New("simulation execution failed")
	ErrSimCrash             = errors.New("simulator process crashed")
	ErrInvalidNetwork       = errors.New("invalid network")
	ErrMarshalFailed        = errors.New("failed to marshal request")
	ErrUnmarshalFailed      = errors.New("failed to unmarshal response")
	ErrSimulationLogicError = errors.New("simulation logic error")
	ErrRPCError             = errors.New("RPC server returned an error")
	ErrValidationFailed     = errors.New("validation failed")
	ErrProtocolUnsupported  = errors.New("unsupported protocol version")
	ErrArgumentRequired     = errors.New("required argument missing")
	ErrAuditLogInvalid      = errors.New("audit log verification failed")
	ErrSessionNotFound      = errors.New("session not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrLedgerNotFound       = errors.New("ledger not found")
	ErrLedgerArchived       = errors.New("ledger has been archived")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrConfigFailed         = errors.New("configuration error")
	ErrNetworkNotFound      = errors.New("network not found")
)

type LedgerNotFoundError struct {
	Sequence uint32
	Message  string
}

func (e *LedgerNotFoundError) Error() string {
	return e.Message
}

func (e *LedgerNotFoundError) Is(target error) bool {
	return target == ErrLedgerNotFound
}

type LedgerArchivedError struct {
	Sequence uint32
	Message  string
}

func (e *LedgerArchivedError) Error() string {
	return e.Message
}

func (e *LedgerArchivedError) Is(target error) bool {
	return target == ErrLedgerArchived
}

type RateLimitError struct {
	Message string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimitExceeded
}

// Wrap functions for consistent error wrapping
func WrapTransactionNotFound(err error) error {
	return &ErstError{
		 Code:    ErstLedgerNotFound,
		 Message: "transaction not found",
		 OrigErr: err,
	}
}

func WrapRPCConnectionFailed(err error) error {
	return &ErstError{
		 Code:    ErstRPCConnectionFailed,
		 Message: "RPC connection failed",
		 OrigErr: err,
	}
}

func WrapSimulatorNotFound(msg string) error {
	return &ErstError{
		 Code:    ErstSimulatorNotFound,
		 Message: msg,
	}
}

func WrapSimulationFailed(err error, stderr string) error {
	return &ErstError{
		 Code:    ErstSimulationFailed,
		 Message: stderr,
		 OrigErr: err,
	}
}

func WrapInvalidNetwork(network string) error {
	return &ErstError{
		 Code:    ErstInvalidNetwork,
		 Message: network + ". Must be one of: testnet, mainnet, futurenet",
	}
}

func WrapMarshalFailed(err error) error {
	return &ErstError{
		 Code:    ErstValidationFailed,
		 Message: "failed to marshal request",
		 OrigErr: err,
	}
}

func WrapUnmarshalFailed(err error, output string) error {
	return &ErstError{
		 Code:    ErstValidationFailed,
		 Message: output,
		 OrigErr: err,
	}
}

func WrapSimulationLogicError(msg string) error {
	return &ErstError{
		 Code:    ErstSimulationLogicError,
		 Message: msg,
	}
}

func WrapRPCTimeout(err error) error {
	return &ErstError{
		 Code:    ErstRPCTimeout,
		 Message: "RPC request timed out",
		 OrigErr: err,
	}
}

func WrapAllRPCFailed() error {
	return &ErstError{
		 Code:    ErstAllRPCFailed,
		 Message: "all RPC endpoints failed",
	}
}

func WrapRPCError(url string, msg string, code int) error {
	return &ErstError{
		 Code:    ErstRPCError,
		 Message: fmt.Sprintf("from %s: %s (code %d)", url, msg, code),
	}
}

func WrapSimCrash(err error, stderr string) error {
	msg := stderr
	if msg == "" && err != nil {
		 msg = err.Error()
	}
	return &ErstError{
		 Code:    ErstSimCrash,
		 Message: msg,
		 OrigErr: err,
	}
}

func WrapValidationError(msg string) error {
	return &ErstError{
		 Code:    ErstValidationFailed,
		 Message: msg,
	}
}

func WrapProtocolUnsupported(version uint32) error {
	return &ErstError{
		 Code:    ErstValidationFailed,
		 Message: fmt.Sprintf("unsupported protocol version: %d", version),
	}
}

func WrapCliArgumentRequired(arg string) error {
	return &ErstError{
		 Code:    ErstValidationFailed,
		 Message: "--" + arg,
	}
}

func WrapAuditLogInvalid(msg string) error {
	return &ErstError{
		 Code:    ErstValidationFailed,
		 Message: msg,
	}
}

func WrapSessionNotFound(sessionID string) error {
	return &ErstError{
		 Code:    ErstValidationFailed,
		 Message: sessionID,
	}
}

func WrapUnauthorized(msg string) error {
	if msg != "" {
		 return &ErstError{
			  Code:    ErstUnauthorized,
			  Message: msg,
		 }
	}
	return &ErstError{
		 Code:    ErstUnauthorized,
		 Message: "unauthorized",
	}
}

func WrapLedgerNotFound(sequence uint32) error {
	return &ErstError{
		 Code:    ErstLedgerNotFound,
		 Message: fmt.Sprintf("ledger %d not found (may be archived or not yet created)", sequence),
	}
}

func WrapLedgerArchived(sequence uint32) error {
	return &ErstError{
		 Code:    ErstLedgerArchived,
		 Message: fmt.Sprintf("ledger %d has been archived and is no longer available", sequence),
	}
}

func WrapRateLimitExceeded() error {
	return &ErstError{
		 Code:    ErstRateLimitExceeded,
		 Message: "rate limit exceeded, please try again later",
	}
}

func WrapConfigError(msg string, err error) error {
	if err != nil {
		 return &ErstError{
			  Code:    ErstConfigFailed,
			  Message: msg + ": " + err.Error(),
			  OrigErr: err,
		 }
	}
	return &ErstError{
		 Code:    ErstConfigFailed,
		 Message: msg,
	}
}

func WrapNetworkNotFound(network string) error {
	return &ErstError{
		 Code:    ErstNetworkNotFound,
		 Message: network,
	}
}
