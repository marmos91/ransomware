package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type AesKey []byte

const AES_KEY_SIZE = 32

// DefaultChunkSize is the plaintext chunk size for streaming encryption (1 MB).
const DefaultChunkSize = 1 << 20

// newGCM creates an AES-GCM cipher from the given key.
func newGCM(key AesKey) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// AesEncrypt encrypts a buffer with AES-GCM.
func AesEncrypt(plainText []byte, key AesKey) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plainText, nil), nil
}

// AesDecrypt decrypts a buffer with AES-GCM.
func AesDecrypt(cipherText []byte, key AesKey) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: got %d bytes, need at least %d", len(cipherText), nonceSize)
	}
	return gcm.Open(nil, cipherText[:nonceSize], cipherText[nonceSize:], nil)
}

// AesEncryptStream reads plaintext from r in chunks, encrypts each chunk with
// AES-GCM, and writes length-prefixed encrypted chunks to w. A zero-length
// marker signals end-of-stream.
//
// Wire format per chunk: [4-byte big-endian length][nonce + ciphertext + tag]
func AesEncryptStream(r io.Reader, w io.Writer, key AesKey) error {
	gcm, err := newGCM(key)
	if err != nil {
		return err
	}

	buf := make([]byte, DefaultChunkSize)
	nonce := make([]byte, gcm.NonceSize())

	for {
		n, readErr := io.ReadFull(r, buf)
		if n > 0 {
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
				return fmt.Errorf("generate nonce: %w", err)
			}

			encrypted := gcm.Seal(nonce, nonce, buf[:n], nil)

			if err := binary.Write(w, binary.BigEndian, uint32(len(encrypted))); err != nil {
				return fmt.Errorf("write chunk length: %w", err)
			}
			if _, err := w.Write(encrypted); err != nil {
				return fmt.Errorf("write chunk: %w", err)
			}
		}

		if errors.Is(readErr, io.EOF) || errors.Is(readErr, io.ErrUnexpectedEOF) {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read plaintext: %w", readErr)
		}
	}

	// End-of-stream marker.
	return binary.Write(w, binary.BigEndian, uint32(0))
}

// AesDecryptStream reads length-prefixed encrypted chunks from r, decrypts
// each with AES-GCM, and writes plaintext to w. A zero-length marker signals
// end-of-stream.
func AesDecryptStream(r io.Reader, w io.Writer, key AesKey) error {
	gcm, err := newGCM(key)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()

	for {
		var chunkLen uint32
		if err := binary.Read(r, binary.BigEndian, &chunkLen); err != nil {
			return fmt.Errorf("read chunk length: %w", err)
		}

		if chunkLen == 0 {
			return nil // end-of-stream
		}

		maxChunkLen := uint32(DefaultChunkSize + nonceSize + gcm.Overhead())
		if chunkLen > maxChunkLen {
			return fmt.Errorf("chunk length %d exceeds maximum %d", chunkLen, maxChunkLen)
		}
		if int(chunkLen) < nonceSize+gcm.Overhead() {
			return fmt.Errorf("chunk too small: %d bytes (minimum %d)", chunkLen, nonceSize+gcm.Overhead())
		}

		encrypted := make([]byte, chunkLen)
		if _, err := io.ReadFull(r, encrypted); err != nil {
			return fmt.Errorf("read chunk: %w", err)
		}

		plaintext, err := gcm.Open(nil, encrypted[:nonceSize], encrypted[nonceSize:], nil)
		if err != nil {
			return fmt.Errorf("decrypt chunk: %w", err)
		}

		if _, err := w.Write(plaintext); err != nil {
			return fmt.Errorf("write plaintext: %w", err)
		}
	}
}

// NewRandomAesKey creates a new random AES key.
func NewRandomAesKey(keySize int) (AesKey, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
