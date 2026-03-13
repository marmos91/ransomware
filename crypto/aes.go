package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type AesKey []byte

const AES_KEY_SIZE = 32

// Encrypts a buffer with AES
func AesEncrypt(plainText []byte, key AesKey) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	cypherText := gcm.Seal(nonce, nonce, plainText, nil)

	return cypherText, nil
}

// Decrypts a buffer with AES
func AesDecrypt(cypherText []byte, key AesKey) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	if len(cypherText) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short: %d bytes, need at least %d", len(cypherText), gcm.NonceSize())
	}

	plainText, err := gcm.Open(nil, cypherText[:gcm.NonceSize()], cypherText[gcm.NonceSize():], nil)

	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// Creates a new random AES key
func NewRandomAesKey(keySize int) (AesKey, error) {
	key := make([]byte, keySize)

	_, err := io.ReadFull(rand.Reader, key)

	if err != nil {
		return nil, err
	}

	return key, nil
}
