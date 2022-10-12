package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const RSA_KEY_SIZE = 2048

type Keypair struct {
	Public  *rsa.PublicKey
	Private *rsa.PrivateKey
}

func NewRandomRsaKeypair(keySize int) (*Keypair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)

	if err != nil {
		return nil, err
	}

	return &Keypair{
		Public:  &privateKey.PublicKey,
		Private: privateKey,
	}, nil
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) (string, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return "", err
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem), nil
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("key type is not RSA")
}

func RsaEncrypt(plainText []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	cipherText, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		plainText,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return cipherText, nil
}

func RsaDecrypt(cipherText []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	plainText, err := privateKey.Decrypt(
		nil,
		cipherText,
		&rsa.OAEPOptions{
			Hash: crypto.SHA256,
		},
	)

	if err != nil {
		return nil, err
	}

	return plainText, err
}
