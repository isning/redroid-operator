package main

import (
	"testing"
)

func TestParseWatchNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantNil  bool
	}{
		{
			name:    "empty string returns nil (cluster-wide)",
			input:   "",
			wantNil: true,
		},
		{
			name:    "whitespace-only returns nil (cluster-wide)",
			input:   "  ",
			wantNil: true,
		},
		{
			name:    "comma-only returns nil (cluster-wide)",
			input:   ",,,",
			wantNil: true,
		},
		{
			name:    "mixed blank entries returns nil (cluster-wide)",
			input:   " , , ",
			wantNil: true,
		},
		{
			name:     "single namespace",
			input:    "ns-a",
			wantKeys: []string{"ns-a"},
		},
		{
			name:     "single namespace with surrounding spaces",
			input:    "  ns-a  ",
			wantKeys: []string{"ns-a"},
		},
		{
			name:     "multiple namespaces",
			input:    "ns-a,ns-b,ns-c",
			wantKeys: []string{"ns-a", "ns-b", "ns-c"},
		},
		{
			name:     "multiple namespaces with spaces",
			input:    " ns-a , ns-b , ns-c ",
			wantKeys: []string{"ns-a", "ns-b", "ns-c"},
		},
		{
			name:     "duplicates are deduplicated by map",
			input:    "ns-a,ns-a,ns-b",
			wantKeys: []string{"ns-a", "ns-b"},
		},
		{
			name:     "blank entries mixed with valid are skipped",
			input:    "ns-a,,ns-b, ,ns-c",
			wantKeys: []string{"ns-a", "ns-b", "ns-c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseWatchNamespaces(tc.input)

			if tc.wantNil {
				if got != nil {
					t.Errorf("expected nil map, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("expected non-nil map with keys %v, got nil", tc.wantKeys)
			}
			if len(got) != len(tc.wantKeys) {
				t.Errorf("expected %d keys, got %d: %v", len(tc.wantKeys), len(got), got)
			}
			for _, k := range tc.wantKeys {
				if _, ok := got[k]; !ok {
					t.Errorf("expected key %q in map, got %v", k, got)
				}
			}
		})
	}
}
