package simplebox

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// Length in bytes of a secret key used for encryption and decryption.
	KeySize = 32

	// Length in bytes of a nonce value (which must be unique and may be
	// random) used for encryption and decryption.
	NonceSize = 24
)

// SimpleBox provides a simple wrapper around NaCl's secretbox with a
// self-contained random nonce strategy.
type SimpleBox struct {
	secretKey *[KeySize]byte
}

// Creates a SimpleBox from a secret key.
func NewFromSecretKey(secretKey *[KeySize]byte) *SimpleBox {
	return &SimpleBox{secretKey: secretKey}
}

// Decrypts the given ciphertext and returns plaintext. An appropriate error is
// included if decryption failed.
func (b *SimpleBox) Decrypt(cipher []byte) ([]byte, error) {
	if len(cipher) <= NonceSize {
		return nil, fmt.Errorf("Ciphertext is of invalid length.")
	}

	var nonce [NonceSize]byte
	copy(nonce[:], cipher[0:NonceSize])
	payload := cipher[NonceSize:]

	var opened []byte
	opened, ok := secretbox.Open(opened[:0], payload, &nonce, b.secretKey)

	if !ok {
		return nil, fmt.Errorf("Ciphertext could not be decrypted.")
	}

	return opened, nil
}

// Encrypts the given plaintext and returns ciphertext.
func (b *SimpleBox) Encrypt(plain []byte) []byte {
	var nonce [NonceSize]byte
	rand.Reader.Read(nonce[:])

	var box []byte
	box = secretbox.Seal(box[:0], plain, &nonce, b.secretKey)

	return append(nonce[:], box...)
}
