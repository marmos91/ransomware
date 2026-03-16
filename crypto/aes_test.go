package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"testing"
)

func randomBytes(t *testing.T, size int) []byte {
	t.Helper()
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("generate random bytes: %v", err)
	}
	return buf
}

func mustNewAesKey(t *testing.T) AesKey {
	t.Helper()
	key, err := NewRandomAesKey(AES_KEY_SIZE)
	if err != nil {
		t.Fatalf("generate AES key: %v", err)
	}
	return key
}

func TestNewRandomAesKey(t *testing.T) {
	t.Run("correct length", func(t *testing.T) {
		key := mustNewAesKey(t)
		if len(key) != AES_KEY_SIZE {
			t.Fatalf("expected key length %d, got %d", AES_KEY_SIZE, len(key))
		}
	})

	t.Run("uniqueness", func(t *testing.T) {
		key1 := mustNewAesKey(t)
		key2 := mustNewAesKey(t)
		if bytes.Equal(key1, key2) {
			t.Fatal("two random keys should not be equal")
		}
	})
}

func TestAesEncryptDecrypt(t *testing.T) {
	key := mustNewAesKey(t)

	tests := []struct {
		name string
		size int
	}{
		{"empty", 0},
		{"small 100B", 100},
		{"medium 1KB", 1024},
		{"large 2MB", 2 * 1024 * 1024},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plaintext := randomBytes(t, tc.size)

			ciphertext, err := AesEncrypt(plaintext, key)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}

			decrypted, err := AesDecrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Fatal("decrypted data does not match original")
			}
		})
	}

	t.Run("wrong key fails", func(t *testing.T) {
		ciphertext, err := AesEncrypt([]byte("secret data"), key)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		wrongKey := mustNewAesKey(t)
		if _, err = AesDecrypt(ciphertext, wrongKey); err == nil {
			t.Fatal("expected error decrypting with wrong key")
		}
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		if _, err := AesDecrypt([]byte{1, 2, 3}, key); err == nil {
			t.Fatal("expected error for short ciphertext")
		}
	})
}

func TestAesEncryptDecryptStream(t *testing.T) {
	key := mustNewAesKey(t)

	tests := []struct {
		name string
		size int
	}{
		{"empty", 0},
		{"sub-chunk 500KB", 500 * 1024},
		{"exact chunk 1MB", DefaultChunkSize},
		{"multi-chunk 2.5MB", int(2.5 * float64(DefaultChunkSize))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plaintext := randomBytes(t, tc.size)

			var encrypted bytes.Buffer
			if err := AesEncryptStream(bytes.NewReader(plaintext), &encrypted, key); err != nil {
				t.Fatalf("encrypt stream: %v", err)
			}

			var decrypted bytes.Buffer
			if err := AesDecryptStream(&encrypted, &decrypted, key); err != nil {
				t.Fatalf("decrypt stream: %v", err)
			}

			if !bytes.Equal(decrypted.Bytes(), plaintext) {
				t.Fatalf("decrypted data does not match original (got %d bytes, want %d)", decrypted.Len(), len(plaintext))
			}
		})
	}

	t.Run("wrong key fails", func(t *testing.T) {
		var encrypted bytes.Buffer
		if err := AesEncryptStream(bytes.NewReader([]byte("stream secret data")), &encrypted, key); err != nil {
			t.Fatalf("encrypt stream: %v", err)
		}

		wrongKey := mustNewAesKey(t)
		var decrypted bytes.Buffer
		if err := AesDecryptStream(bytes.NewReader(encrypted.Bytes()), &decrypted, wrongKey); err == nil {
			t.Fatal("expected error decrypting stream with wrong key")
		}
	})

	t.Run("corrupted chunk fails", func(t *testing.T) {
		plaintext := randomBytes(t, 1024)

		var encrypted bytes.Buffer
		if err := AesEncryptStream(bytes.NewReader(plaintext), &encrypted, key); err != nil {
			t.Fatalf("encrypt stream: %v", err)
		}

		data := encrypted.Bytes()
		if len(data) > 10 {
			data[10] ^= 0xff
		}

		var decrypted bytes.Buffer
		if err := AesDecryptStream(bytes.NewReader(data), &decrypted, key); err == nil {
			t.Fatal("expected error for corrupted chunk")
		}
	})

	t.Run("oversized chunk length fails", func(t *testing.T) {
		var buf bytes.Buffer
		oversize := uint32(DefaultChunkSize + 100 + 16 + 1)
		if err := binary.Write(&buf, binary.BigEndian, oversize); err != nil {
			t.Fatalf("write chunk length: %v", err)
		}

		var decrypted bytes.Buffer
		if err := AesDecryptStream(&buf, &decrypted, key); err == nil {
			t.Fatal("expected error for oversized chunk length")
		}
	})
}
