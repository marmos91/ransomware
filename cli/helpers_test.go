package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marmos91/ransomware/crypto"
)

type testKeys struct {
	Keypair         *crypto.Keypair
	AesKey          crypto.AesKey
	EncryptedAesKey []byte
	KeySizeBits     uint16
}

func setupTestKeys(t *testing.T) testKeys {
	t.Helper()

	kp, err := crypto.NewRandomRsaKeypair(2048)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	aesKey, err := crypto.NewRandomAesKey(crypto.AES_KEY_SIZE)
	if err != nil {
		t.Fatalf("generate AES key: %v", err)
	}

	encryptedAesKey, err := crypto.RsaEncrypt(aesKey, kp.Public)
	if err != nil {
		t.Fatalf("encrypt AES key: %v", err)
	}

	return testKeys{
		Keypair:         kp,
		AesKey:          aesKey,
		EncryptedAesKey: encryptedAesKey,
		KeySizeBits:     uint16(kp.Public.Size() * 8),
	}
}

func createTestFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	return path
}
