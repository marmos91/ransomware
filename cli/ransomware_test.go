package cli

import (
	"bytes"
	"os"
	"slices"
	"testing"
	"time"
)

func TestSplitCommaSeparated(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"single", ".enc", []string{".enc"}},
		{"multiple", ".enc,.txt,.go", []string{".enc", ".txt", ".go"}},
		{"with spaces", " .enc , .txt , .go ", []string{".enc", ".txt", ".go"}},
		{"trailing comma", ".enc,.txt,", []string{".enc", ".txt"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitCommaSeparated(tc.input)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidateEncSuffix(t *testing.T) {
	tests := []struct {
		name    string
		suffix  string
		wantErr bool
	}{
		{".enc is ok", ".enc", false},
		{"enc without dot", "enc", true},
		{"empty", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEncSuffix(tc.suffix)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateEncSuffix(%q) error = %v, wantErr %v", tc.suffix, err, tc.wantErr)
			}
		})
	}
}

func TestFileHeader_WriteRead(t *testing.T) {
	hdr := fileHeader{
		KeySizeBits:  2048,
		FileMode:     0644,
		ModTime:      time.Now().Unix(),
		PartialBytes: 512,
	}

	var buf bytes.Buffer
	if err := hdr.writeTo(&buf); err != nil {
		t.Fatalf("write header: %v", err)
	}

	var got fileHeader
	if err := got.readFrom(&buf); err != nil {
		t.Fatalf("read header: %v", err)
	}

	if got != hdr {
		t.Fatalf("header mismatch: got %+v, want %+v", got, hdr)
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	k := setupTestKeys(t)
	dir := t.TempDir()
	original := []byte("original file content")
	srcPath := createTestFile(t, dir, "test.txt", original)

	mtime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(srcPath, mtime, mtime); err != nil {
		t.Fatalf("set mtime: %v", err)
	}

	if err := encryptFile(srcPath, k.AesKey, k.EncryptedAesKey, k.KeySizeBits, 0, ".enc"); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	encPath := srcPath + ".enc"
	if _, err := os.Stat(encPath); err != nil {
		t.Fatalf("encrypted file should exist: %v", err)
	}

	if err := decryptFile(encPath, k.Keypair.Private, ".enc"); err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	decrypted, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read decrypted: %v", err)
	}
	if !bytes.Equal(decrypted, original) {
		t.Fatal("decrypted content does not match original")
	}

	info, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("stat decrypted: %v", err)
	}
	if info.ModTime().Unix() != mtime.Unix() {
		t.Fatalf("mtime not restored: got %v, want %v", info.ModTime(), mtime)
	}
}

func TestEncryptDecryptFile_Partial(t *testing.T) {
	k := setupTestKeys(t)
	dir := t.TempDir()

	original := make([]byte, 2048)
	for i := range original {
		original[i] = byte(i % 256)
	}
	srcPath := createTestFile(t, dir, "test.txt", original)

	if err := encryptFile(srcPath, k.AesKey, k.EncryptedAesKey, k.KeySizeBits, 512, ".enc"); err != nil {
		t.Fatalf("encrypt partial: %v", err)
	}

	encPath := srcPath + ".enc"
	if err := decryptFile(encPath, k.Keypair.Private, ".enc"); err != nil {
		t.Fatalf("decrypt partial: %v", err)
	}

	decrypted, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read decrypted: %v", err)
	}
	if !bytes.Equal(decrypted, original) {
		t.Fatalf("decrypted content does not match original (got %d bytes, want %d)", len(decrypted), len(original))
	}
}

func TestVerifyFile(t *testing.T) {
	k := setupTestKeys(t)
	dir := t.TempDir()
	srcPath := createTestFile(t, dir, "test.txt", []byte("content to verify"))

	if err := encryptFile(srcPath, k.AesKey, k.EncryptedAesKey, k.KeySizeBits, 0, ".enc"); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	encPath := srcPath + ".enc"
	if err := verifyFile(encPath, k.Keypair.Private); err != nil {
		t.Fatalf("verify should succeed: %v", err)
	}
}

func TestReadEncryptedHeader_WrongKey(t *testing.T) {
	kA := setupTestKeys(t)
	kB := setupTestKeys(t)

	dir := t.TempDir()
	srcPath := createTestFile(t, dir, "test.txt", []byte("secret"))

	if err := encryptFile(srcPath, kA.AesKey, kA.EncryptedAesKey, kA.KeySizeBits, 0, ".enc"); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	encPath := srcPath + ".enc"
	f, err := os.Open(encPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	if _, _, err = readEncryptedHeader(f, kB.Keypair.Private); err == nil {
		t.Fatal("expected error when reading header with wrong key")
	}
}
