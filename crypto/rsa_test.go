package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func mustNewRsaKeypair(t *testing.T, bits int) *Keypair {
	t.Helper()
	kp, err := NewRandomRsaKeypair(bits)
	if err != nil {
		t.Fatalf("generate %d-bit keypair: %v", bits, err)
	}
	return kp
}

func TestNewRandomRsaKeypair(t *testing.T) {
	t.Run("2048 bit", func(t *testing.T) {
		kp := mustNewRsaKeypair(t, 2048)
		if kp.Private == nil || kp.Public == nil {
			t.Fatal("keypair fields must not be nil")
		}
		if kp.Private.Size()*8 != 2048 {
			t.Fatalf("expected 2048-bit key, got %d", kp.Private.Size()*8)
		}
		if &kp.Private.PublicKey != kp.Public {
			t.Fatal("public key should be derived from private key")
		}
	})

	t.Run("4096 bit", func(t *testing.T) {
		kp := mustNewRsaKeypair(t, 4096)
		if kp.Private.Size()*8 != 4096 {
			t.Fatalf("expected 4096-bit key, got %d", kp.Private.Size()*8)
		}
	})
}

func TestRsaEncryptDecrypt(t *testing.T) {
	kp := mustNewRsaKeypair(t, 2048)
	plaintext := randomBytes(t, 32)

	t.Run("round-trip 32-byte payload", func(t *testing.T) {
		ciphertext, err := RsaEncrypt(plaintext, kp.Public)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		decrypted, err := RsaDecrypt(ciphertext, kp.Private)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Fatal("decrypted data does not match original")
		}
	})

	t.Run("wrong key fails", func(t *testing.T) {
		ciphertext, err := RsaEncrypt(plaintext, kp.Public)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		wrongKp := mustNewRsaKeypair(t, 2048)
		if _, err = RsaDecrypt(ciphertext, wrongKp.Private); err == nil {
			t.Fatal("expected error decrypting with wrong key")
		}
	})
}

func TestExportParsePrivateKey(t *testing.T) {
	kp := mustNewRsaKeypair(t, 2048)

	t.Run("PEM round-trip", func(t *testing.T) {
		pemStr := ExportRsaPrivateKeyAsPemStr(kp.Private)
		if pemStr == "" {
			t.Fatal("exported PEM should not be empty")
		}

		parsed, err := ParseRsaPrivateKeyFromPemStr(pemStr)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}

		if !kp.Private.Equal(parsed) {
			t.Fatal("parsed key does not match original")
		}
	})

	t.Run("invalid PEM fails", func(t *testing.T) {
		if _, err := ParseRsaPrivateKeyFromPemStr("not a pem"); err == nil {
			t.Fatal("expected error for invalid PEM")
		}
	})
}

func TestExportParsePublicKey(t *testing.T) {
	kp := mustNewRsaKeypair(t, 2048)

	t.Run("PEM round-trip", func(t *testing.T) {
		pemStr, err := ExportRsaPublicKeyAsPemStr(kp.Public)
		if err != nil {
			t.Fatalf("export: %v", err)
		}

		parsed, err := ParseRsaPublicKeyFromPemStr(pemStr)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}

		if !kp.Public.Equal(parsed) {
			t.Fatal("parsed key does not match original")
		}
	})

	t.Run("invalid PEM fails", func(t *testing.T) {
		if _, err := ParseRsaPublicKeyFromPemStr("not a pem"); err == nil {
			t.Fatal("expected error for invalid PEM")
		}
	})

	t.Run("non-RSA key type fails", func(t *testing.T) {
		ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("generate EC key: %v", err)
		}
		der, err := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
		if err != nil {
			t.Fatalf("marshal EC public key: %v", err)
		}
		ecPEM := string(pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: der,
		}))

		if _, err = ParseRsaPublicKeyFromPemStr(ecPEM); err == nil {
			t.Fatal("expected error for non-RSA key")
		}
	})
}
