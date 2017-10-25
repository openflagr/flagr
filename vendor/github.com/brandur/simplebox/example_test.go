package simplebox_test

import (
	"crypto/rand"
	"fmt"

	"github.com/brandur/simplebox"
)

func ExampleSimpleBox() {
	message := "hello"

	var secretKey [simplebox.KeySize]byte
	rand.Reader.Read(secretKey[:])
	box := simplebox.NewFromSecretKey(&secretKey)

	// Encrypt
	ciphertext := box.Encrypt([]byte(message))

	// Decrypt
	decrypted, err := box.Decrypt(ciphertext)
	if err != nil {
		panic(err)
	}

	// Prints:
	//
	// Decrypted: hello
	fmt.Printf("Decrypted: %v\n", decrypted)
}
