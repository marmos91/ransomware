package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTree(t *testing.T, files []string) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range files {
		path := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	return root
}

func TestWalkAndCollect(t *testing.T) {
	t.Run("basic collection", func(t *testing.T) {
		root := setupTree(t, []string{
			"a.txt",
			"b.txt",
			"sub/c.txt",
		})

		files, err := WalkAndCollect(root, nil, nil, false, true)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		if len(files) != 3 {
			t.Fatalf("expected 3 files, got %d", len(files))
		}
	})

	t.Run("extension blacklist", func(t *testing.T) {
		root := setupTree(t, []string{
			"a.txt",
			"b.enc",
			"c.txt",
		})

		files, err := WalkAndCollect(root, []string{".enc"}, nil, false, true)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		if len(files) != 2 {
			t.Fatalf("expected 2 files, got %d", len(files))
		}
		for _, f := range files {
			if filepath.Ext(f) == ".enc" {
				t.Fatalf("blacklisted file should not be collected: %s", f)
			}
		}
	})

	t.Run("extension whitelist", func(t *testing.T) {
		root := setupTree(t, []string{
			"a.txt",
			"b.enc",
			"c.go",
		})

		files, err := WalkAndCollect(root, nil, []string{".txt"}, false, true)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		if filepath.Ext(files[0]) != ".txt" {
			t.Fatalf("expected .txt file, got %s", files[0])
		}
	})

	t.Run("non-recursive", func(t *testing.T) {
		root := setupTree(t, []string{
			"a.txt",
			"sub/b.txt",
			"sub/deep/c.txt",
		})

		files, err := WalkAndCollect(root, nil, nil, false, false)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file (non-recursive), got %d", len(files))
		}
	})

	t.Run("empty dir returns empty slice", func(t *testing.T) {
		root := t.TempDir()

		files, err := WalkAndCollect(root, nil, nil, false, true)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("whitelist takes precedence over blacklist", func(t *testing.T) {
		root := setupTree(t, []string{
			"a.txt",
			"b.enc",
			"c.go",
		})

		// Whitelist .txt, blacklist .txt — whitelist wins
		files, err := WalkAndCollect(root, []string{".txt"}, []string{".txt"}, false, true)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file (whitelist wins), got %d", len(files))
		}
	})
}

func TestWalkAndCollect_SkipHidden(t *testing.T) {
	root := setupTree(t, []string{
		"visible.txt",
		".hidden_file.txt",
	})
	hiddenDir := filepath.Join(root, ".hidden_dir")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "inside.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Run("skip hidden", func(t *testing.T) {
		files, err := WalkAndCollect(root, nil, nil, true, true)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}

		if len(files) != 1 {
			t.Fatalf("expected 1 visible file, got %d: %v", len(files), files)
		}
		if filepath.Base(files[0]) != "visible.txt" {
			t.Fatalf("expected visible.txt, got %s", files[0])
		}
	})

	t.Run("include hidden", func(t *testing.T) {
		files, err := WalkAndCollect(root, nil, nil, false, true)
		if err != nil {
			t.Fatalf("walk: %v", err)
		}
		if len(files) != 3 {
			t.Fatalf("expected 3 files (including hidden), got %d: %v", len(files), files)
		}
	})
}
