package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteToFileWithMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    os.FileMode
		content []byte
	}{
		{"mode 0600", 0600, []byte("private content")},
		{"mode 0644", 0644, []byte("public content")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.txt")

			if err := WriteToFileWithMode(path, tc.content, tc.mode); err != nil {
				t.Fatalf("write: %v", err)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read back: %v", err)
			}
			if string(data) != string(tc.content) {
				t.Fatalf("content mismatch: got %q, want %q", data, tc.content)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat: %v", err)
			}
			if info.Mode().Perm() != tc.mode {
				t.Fatalf("mode mismatch: got %o, want %o", info.Mode().Perm(), tc.mode)
			}
		})
	}
}

func TestWriteStringToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "hello world"

	if err := WriteStringToFile(path, content); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(data) != content {
		t.Fatalf("content mismatch: got %q, want %q", data, content)
	}
}

func TestReadStringFileContent(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		want := "file content"

		if err := os.WriteFile(path, []byte(want), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		got, err := ReadStringFileContent(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := ReadStringFileContent("/nonexistent/path/file.txt")
		if err == nil {
			t.Fatal("expected error for non-existent file")
		}
	})
}

func TestDeleteFileIfExists(t *testing.T) {
	t.Run("delete existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")

		if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		if err := DeleteFileIfExists(path); err != nil {
			t.Fatalf("delete: %v", err)
		}

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("file should not exist after deletion")
		}
	})

	t.Run("delete non-existent is no-op", func(t *testing.T) {
		err := DeleteFileIfExists("/nonexistent/path/file.txt")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
}
