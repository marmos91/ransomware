package fs

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestProcessFilesParallel(t *testing.T) {
	t.Run("sequential workers=1", func(t *testing.T) {
		files := []string{"a", "b", "c"}
		var processed []string
		err := ProcessFilesParallel(files, 1, func(path string) error {
			processed = append(processed, path)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(processed) != 3 {
			t.Fatalf("expected 3 processed, got %d", len(processed))
		}
	})

	t.Run("concurrent workers=4", func(t *testing.T) {
		files := make([]string, 20)
		for i := range files {
			files[i] = fmt.Sprintf("file_%d", i)
		}

		var mu sync.Mutex
		var processed []string
		err := ProcessFilesParallel(files, 4, func(path string) error {
			mu.Lock()
			processed = append(processed, path)
			mu.Unlock()
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(processed) != 20 {
			t.Fatalf("expected 20 processed, got %d", len(processed))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		err := ProcessFilesParallel(nil, 4, func(path string) error {
			t.Fatal("should not be called")
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error aggregation", func(t *testing.T) {
		files := []string{"a", "b", "c"}
		err := ProcessFilesParallel(files, 1, func(path string) error {
			return fmt.Errorf("failed: %s", path)
		})
		if err == nil {
			t.Fatal("expected error")
		}
		// Sequential mode returns on first error
		if !strings.Contains(err.Error(), "failed: a") {
			t.Fatalf("expected first error, got: %v", err)
		}
	})

	t.Run("concurrent error aggregation", func(t *testing.T) {
		files := []string{"a", "b", "c", "d"}
		err := ProcessFilesParallel(files, 4, func(path string) error {
			return fmt.Errorf("fail: %s", path)
		})
		if err == nil {
			t.Fatal("expected error")
		}
		// All errors should be aggregated
		errStr := err.Error()
		for _, f := range files {
			if !strings.Contains(errStr, "fail: "+f) {
				t.Fatalf("expected error for %s in: %v", f, errStr)
			}
		}
	})

	t.Run("workers clamped to file count", func(t *testing.T) {
		files := []string{"a", "b"}
		var mu sync.Mutex
		var processed []string
		err := ProcessFilesParallel(files, 100, func(path string) error {
			mu.Lock()
			processed = append(processed, path)
			mu.Unlock()
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(processed) != 2 {
			t.Fatalf("expected 2 processed, got %d", len(processed))
		}
	})
}
