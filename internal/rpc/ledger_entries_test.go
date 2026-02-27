// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stellar/go-stellar-sdk/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeValidXDRKey generates a valid base64-encoded XDR LedgerKey (Account type)
// from a deterministic seed string. This produces keys that pass VerifyLedgerEntries.
func makeValidXDRKey(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	var pk xdr.Uint256
	copy(pk[:], hash[:])
	key := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeAccount,
		Account: &xdr.LedgerKeyAccount{
			AccountId: xdr.AccountId{
				Type:    xdr.PublicKeyTypePublicKeyTypeEd25519,
				Ed25519: &pk,
			},
		},
	}
	var buf bytes.Buffer
	if _, err := xdr.Marshal(&buf, key); err != nil {
		panic(fmt.Sprintf("failed to marshal test XDR key: %v", err))
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// makeValidXDRKeys generates n valid XDR keys with a common prefix.
func makeValidXDRKeys(prefix string, n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = makeValidXDRKey(fmt.Sprintf("%s_%d", prefix, i))
	}
	return keys
}

// TestGetLedgerEntries_EmptyKeys tests that empty key list returns empty map
func TestGetLedgerEntries_EmptyKeys(t *testing.T) {
	client := &Client{
		Network:      Testnet,
		CacheEnabled: false,
		AltURLs:      []string{"http://test.example.com"},
	}

	ctx := context.Background()
	result, err := client.GetLedgerEntries(ctx, []string{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

// TestGetLedgerEntries_FiveKeys tests fetching 5 related keys
func TestGetLedgerEntries_FiveKeys(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GetLedgerEntriesRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify request structure
		assert.Equal(t, "2.0", req.Jsonrpc)
		assert.Equal(t, "getLedgerEntries", req.Method)
		assert.Len(t, req.Params, 1)

		keys := req.Params[0].([]interface{})

		// Build response with entries for each key
		entries := make([]LedgerEntryResult, len(keys))

		for i, key := range keys {
			entries[i] = LedgerEntryResult{
				Key:                key.(string),
				Xdr:                "mock_xdr_data_" + key.(string),
				LastModifiedLedger: 12345,
				LiveUntilLedger:    12400,
			}
		}

		resp := GetLedgerEntriesResponse{
			Jsonrpc: "2.0",
			ID:      1,
		}
		resp.Result.Entries = entries
		resp.Result.LatestLedger = 12345

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Network:      Testnet,
		HorizonURL:   server.URL,
		SorobanURL:   server.URL,
		CacheEnabled: false,
		AltURLs:      []string{server.URL},
	}

	ctx := context.Background()
	keys := makeValidXDRKeys("five", 5)

	result, err := client.GetLedgerEntries(ctx, keys)

	require.NoError(t, err)
	assert.Len(t, result, 5)

	// Verify all keys are present
	for _, key := range keys {
		assert.Contains(t, result, key)
		assert.Equal(t, "mock_xdr_data_"+key, result[key])
	}
}

// TestGetLedgerEntries_LargeBatch tests batching with 100+ keys
func TestGetLedgerEntries_LargeBatch(t *testing.T) {
	requestCount := 0
	var mu sync.Mutex

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		var req GetLedgerEntriesRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		keys := req.Params[0].([]interface{})

		// Verify batch size is within limits (should be <= 50)
		assert.LessOrEqual(t, len(keys), 50, "Batch size should not exceed 50")

		// Build response
		entries := make([]LedgerEntryResult, len(keys))

		for i, key := range keys {
			entries[i] = LedgerEntryResult{
				Key:                key.(string),
				Xdr:                "xdr_" + key.(string),
				LastModifiedLedger: 12345,
				LiveUntilLedger:    12400,
			}
		}

		resp := GetLedgerEntriesResponse{
			Jsonrpc: "2.0",
			ID:      1,
		}
		resp.Result.Entries = entries
		resp.Result.LatestLedger = 12345

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Network:      Testnet,
		HorizonURL:   server.URL,
		SorobanURL:   server.URL,
		CacheEnabled: false,
		AltURLs:      []string{server.URL},
	}

	ctx := context.Background()

	// Generate 120 keys to test batching
	keys := makeValidXDRKeys("large", 120)

	result, err := client.GetLedgerEntries(ctx, keys)

	require.NoError(t, err)
	assert.Len(t, result, 120)

	// Verify all keys are present
	for _, key := range keys {
		assert.Contains(t, result, key)
		assert.Equal(t, "xdr_"+key, result[key])
	}

	// Verify that multiple requests were made (batching occurred)
	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, requestCount, 3, "Should have made at least 3 batched requests for 120 keys")
}

// TestGetLedgerEntries_ConcurrentBatches tests concurrent batch processing
func TestGetLedgerEntries_ConcurrentBatches(t *testing.T) {
	var requestTimes []time.Time
	var mu sync.Mutex

	// Create mock server with slight delay to test concurrency
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()

		// Small delay to simulate network latency
		time.Sleep(50 * time.Millisecond)

		var req GetLedgerEntriesRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		keys := req.Params[0].([]interface{})

		entries := make([]LedgerEntryResult, len(keys))

		for i, key := range keys {
			entries[i] = LedgerEntryResult{
				Key:                key.(string),
				Xdr:                "xdr_" + key.(string),
				LastModifiedLedger: 12345,
				LiveUntilLedger:    12400,
			}
		}

		resp := GetLedgerEntriesResponse{
			Jsonrpc: "2.0",
			ID:      1,
		}
		resp.Result.Entries = entries
		resp.Result.LatestLedger = 12345

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Network:      Testnet,
		HorizonURL:   server.URL,
		SorobanURL:   server.URL,
		CacheEnabled: false,
		AltURLs:      []string{server.URL},
	}

	ctx := context.Background()

	// Generate 150 keys to ensure multiple batches
	keys := makeValidXDRKeys("concurrent", 150)

	startTime := time.Now()
	result, err := client.GetLedgerEntries(ctx, keys)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.Len(t, result, 150)

	// Verify concurrent execution: with 3 batches at 50ms each,
	// sequential would take ~150ms, concurrent should be much faster
	// Allow some overhead but should be significantly less than sequential
	assert.Less(t, duration, 120*time.Millisecond,
		"Concurrent batching should complete faster than sequential")

	// Verify multiple requests were made concurrently
	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, len(requestTimes), 3, "Should have made at least 3 batched requests")
}

// TestGetLedgerEntries_TimeoutHandling tests timeout behavior
func TestGetLedgerEntries_TimeoutHandling(t *testing.T) {
	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than context timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{
		Network:      Testnet,
		HorizonURL:   server.URL,
		SorobanURL:   server.URL,
		CacheEnabled: false,
		AltURLs:      []string{server.URL},
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	keys := makeValidXDRKeys("timeout", 60) // Force batching

	_, err := client.GetLedgerEntries(ctx, keys)

	// Should get an error due to timeout
	require.Error(t, err)
}

// TestGetLedgerEntries_ErrorHandling tests error response handling
func TestGetLedgerEntries_ErrorHandling(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GetLedgerEntriesResponse{
			Jsonrpc: "2.0",
			ID:      1,
		}
		resp.Error = &struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    -32600,
			Message: "Invalid request",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Network:      Testnet,
		HorizonURL:   server.URL,
		SorobanURL:   server.URL,
		CacheEnabled: false,
		AltURLs:      []string{server.URL},
	}

	ctx := context.Background()
	keys := makeValidXDRKeys("error", 2)

	_, err := client.GetLedgerEntries(ctx, keys)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid request")
}

// TestChunkKeys tests the key chunking function
func TestChunkKeys(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		batchSize int
		expected  int // expected number of batches
	}{
		{
			name:      "exact multiple",
			keys:      make([]string, 100),
			batchSize: 50,
			expected:  2,
		},
		{
			name:      "with remainder",
			keys:      make([]string, 120),
			batchSize: 50,
			expected:  3,
		},
		{
			name:      "less than batch size",
			keys:      make([]string, 30),
			batchSize: 50,
			expected:  1,
		},
		{
			name:      "empty",
			keys:      []string{},
			batchSize: 50,
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := chunkKeys(tt.keys, tt.batchSize)
			assert.Len(t, batches, tt.expected)

			// Verify all keys are included
			totalKeys := 0
			for _, batch := range batches {
				totalKeys += len(batch)
				assert.LessOrEqual(t, len(batch), tt.batchSize)
			}
			assert.Equal(t, len(tt.keys), totalKeys)
		})
	}
}

// TestGetLedgerEntries_PartialFailure tests handling of partial batch failures
func TestGetLedgerEntries_PartialFailure(t *testing.T) {
	requestCount := 0
	var mu sync.Mutex

	// Create mock server that fails on second request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		currentCount := requestCount
		mu.Unlock()

		var req GetLedgerEntriesRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Fail on second batch
		if currentCount == 2 {
			resp := GetLedgerEntriesResponse{
				Jsonrpc: "2.0",
				ID:      1,
			}
			resp.Error = &struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    -32603,
				Message: "Internal error",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Success for other batches
		keys := req.Params[0].([]interface{})
		entries := make([]LedgerEntryResult, len(keys))

		for i, key := range keys {
			entries[i] = LedgerEntryResult{
				Key:                key.(string),
				Xdr:                "xdr_" + key.(string),
				LastModifiedLedger: 12345,
				LiveUntilLedger:    12400,
			}
		}

		resp := GetLedgerEntriesResponse{
			Jsonrpc: "2.0",
			ID:      1,
		}
		resp.Result.Entries = entries
		resp.Result.LatestLedger = 12345

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		Network:      Testnet,
		HorizonURL:   server.URL,
		SorobanURL:   server.URL,
		CacheEnabled: false,
		AltURLs:      []string{server.URL},
	}

	ctx := context.Background()

	// Generate enough keys to create multiple batches
	keys := makeValidXDRKeys("partial", 120)

	_, err := client.GetLedgerEntries(ctx, keys)

	// Should get an error due to partial failure
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch")
}
