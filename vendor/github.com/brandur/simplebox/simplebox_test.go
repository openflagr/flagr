package simplebox

import (
	"crypto/rand"
	"fmt"
	"testing"
)

var (
	simpleBox *SimpleBox
)

func init() {
	var secretKey [KeySize]byte
	rand.Reader.Read(secretKey[:])
	simpleBox = NewFromSecretKey(&secretKey)
}

func TestDecryptBadCipher(t *testing.T) {
	garbage := "not-actually-ciphertext-but-longer-than-a-nonce-at-least"
	plain, err := simpleBox.Decrypt([]byte(garbage))
	if plain != nil {
		t.Errorf("Expected plaintext to be nil")
	}
	expected := fmt.Errorf("Ciphertext could not be decrypted.")
	if err.Error() != expected.Error() {
		t.Errorf("Expected error to be '%v', but got '%v'", expected, err)
	}
}

func TestDecryptBadLength(t *testing.T) {
	garbage := "not-actually-ciphertext"
	plain, err := simpleBox.Decrypt([]byte(garbage))
	if plain != nil {
		t.Errorf("Expected plaintext to be nil")
	}
	expected := fmt.Errorf("Ciphertext is of invalid length.")
	if err.Error() != expected.Error() {
		t.Errorf("Expected error to be '%v', but got '%v'", expected, err)
	}
}

func TestEncryptionSymmetry(t *testing.T) {
	expected := "hello"
	cipher := simpleBox.Encrypt([]byte(expected))
	actual, err := simpleBox.Decrypt(cipher)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != expected {
		t.Errorf("Expected plaintext to be '%v', but got '%v'",
			expected, string(actual))
	}
}

func TestNewFromSecretKey(t *testing.T) {
	var secretKey [KeySize]byte
	for i := 0; i < KeySize; i++ {
		secretKey[i] = byte(i)
	}

	simpleBox := NewFromSecretKey(&secretKey)
	if simpleBox.secretKey != &secretKey {
		t.Errorf("Expected secretKey to be '%v', but got '%v'",
			secretKey, simpleBox.secretKey)
	}
}
