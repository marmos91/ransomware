//go:build !windows

package fs

import "testing"

func TestIsHidden(t *testing.T) {
	tests := []struct {
		path   string
		hidden bool
	}{
		{".hidden", true},
		{".gitignore", true},
		{"visible.txt", false},
		{"path/to/normal.go", false},
		{"path/to/.secret", true},
		{".", false},
		{"..", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			got, err := IsHidden(tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.hidden {
				t.Fatalf("IsHidden(%q) = %v, want %v", tc.path, got, tc.hidden)
			}
		})
	}
}
