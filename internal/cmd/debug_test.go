// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOverrideState(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantEntries int
		wantErr     bool
	}{
		{
			name: "valid override with entries",
			content: `{
				"ledger_entries": {
					"key1": "value1",
					"key2": "value2"
				}
			}`,
			wantEntries: 2,
			wantErr:     false,
		},
		{
			name: "empty ledger entries",
			content: `{
				"ledger_entries": {}
			}`,
			wantEntries: 0,
			wantErr:     false,
		},
		{
			name: "null ledger entries",
			content: `{
				"ledger_entries": null
			}`,
			wantEntries: 0,
			wantErr:     false,
		},
		{
			name:        "invalid json",
			content:     `{invalid json}`,
			wantEntries: 0,
			wantErr:     true,
		},
		{
			name: "missing ledger_entries field",
			content: `{
				"other_field": "value"
			}`,
			wantEntries: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "override.json")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			entries, err := loadOverrideState(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadOverrideState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(entries) != tt.wantEntries {
				t.Errorf("loadOverrideState() got %d entries, want %d", len(entries), tt.wantEntries)
			}
		})
	}
}

func TestLoadOverrideState_FileNotFound(t *testing.T) {
	_, err := loadOverrideState("/nonexistent/path/to/file.json")
	if err == nil {
		t.Error("loadOverrideState() expected error for nonexistent file, got nil")
	}
}

func TestLoadOverrideState_RealWorldExample(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "override.json")
	
	override := OverrideData{
		LedgerEntries: map[string]string{
			"AAAAAAAAAAC6hsKutUTv8P4rkKBTPJIKJvhqEMH3L9sEqKnG9nT/bQ==": "AAAABgAAAAFv8F+E0D/BE04jR47s+JhGi1Q/T/yxfC8UgG88j68rAAAAAAAAAAB+SCAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=",
			"test_account_balance": "base64_encoded_balance_data",
		},
	}

	data, err := json.MarshalIndent(override, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	entries, err := loadOverrideState(tmpFile)
	if err != nil {
		t.Fatalf("loadOverrideState() unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("loadOverrideState() got %d entries, want 2", len(entries))
	}

	expectedKey := "AAAAAAAAAAC6hsKutUTv8P4rkKBTPJIKJvhqEMH3L9sEqKnG9nT/bQ=="
	if val, ok := entries[expectedKey]; !ok {
		t.Errorf("loadOverrideState() missing expected key %s", expectedKey)
	} else if val != "AAAABgAAAAFv8F+E0D/BE04jR47s+JhGi1Q/T/yxfC8UgG88j68rAAAAAAAAAAB+SCAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=" {
		t.Errorf("loadOverrideState() wrong value for key %s", expectedKey)
	}
}

func TestOverrideData_JSONMarshaling(t *testing.T) {
	original := OverrideData{
		LedgerEntries: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded OverrideData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(decoded.LedgerEntries) != len(original.LedgerEntries) {
		t.Errorf("decoded entries count = %d, want %d", len(decoded.LedgerEntries), len(original.LedgerEntries))
	}

	for key, val := range original.LedgerEntries {
		if decoded.LedgerEntries[key] != val {
			t.Errorf("decoded[%s] = %s, want %s", key, decoded.LedgerEntries[key], val)
		}
	}
}
