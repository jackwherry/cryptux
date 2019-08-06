package main

import (
	"fmt"
	"io"
	"os"

	// "io/ioutil"
	// "net/http"
	"crypto/rand"
	"errors"
	"flag"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

// example command line: ./client -id 88226 -pass 88377671 [-server v.snazz.xyz]
var id = flag.Int("room", -1, "chat ID number (leave blank for new room)")
var passcode = flag.String("pass", "", "secret key for chat encryption")
var server = flag.String("server", "v.snazz.xyz", "hostname or IP address of cryptux server")

// using a single salt for all applications allows for known-plaintext attacks, right?
// TODO: Figure out how to mitigate this risk (does it matter?)
var salt = []byte(`js>ru/F7cug(3<-/,e~?"c#aUfq3Dqa!hX7G<Mf;4)~h<'U?bfW899gdE:3hxduK`)

// generateKey returns a deterministic key for the session
func generateKey(password string) [32]byte {
	var key [32]byte
	bigKey := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32) // 32 byte key
	copy(key[:], bigKey[:32])
	return key
}

func generateNonce() [24]byte {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}
	return nonce
}

func encryptMessage(message string, key [32]byte) []byte {
	nonce := generateNonce()
	// append message to end of nonce
	encrypted := secretbox.Seal(nonce[:], []byte(message), &nonce, &key)
	return encrypted
}

func decryptMessage(ciphertext []byte, key [32]byte) (string, error) {
	var decryptNonce [24]byte
	copy(decryptNonce[:], ciphertext[:24]) // nonce is first 24 bytes of ciphertext
	decrypted, ok := secretbox.Open(nil, ciphertext[24:], &decryptNonce, &key)
	if !ok {
		return "", errors.New("could not decrypt message")
	}
	return string(decrypted), nil
}

func main() {
	flag.Parse()
	if *passcode == "" {
		fmt.Println("Please provide a passcode (secret key) to encrypt your session.")
		os.Exit(1)
	}

	fmt.Println("Welcome to cryptux.")
	if *id == -1 {
		fmt.Println("You have not specified a chat ID number, so we'll open a new chat.")
	}
	key := generateKey(*passcode)
	encrypted := encryptMessage("Hey there!", key)
	decrypted, _ := decryptMessage(encrypted, key)
	fmt.Println(decrypted)
}
