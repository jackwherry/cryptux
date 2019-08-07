package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

// example command line: ./client -id 88226 -pass 88377671 [-server v.snazz.xyz]
var room = flag.String("id", "", "unique ID for your chat room")
var passcode = flag.String("pass", "", "secret key for chat encryption (also must be the same as the other users)")
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
	return nonce // 24 cryptographically random bytes
}

func encryptMessage(message string, key [32]byte) []byte {
	nonce := generateNonce()
	// append message to end of nonce
	encrypted := secretbox.Seal(nonce[:], []byte(message), &nonce, &key)
	return encrypted
}

func decryptMessage(ciphertext []byte, key [32]byte) (string, error) {
	if len(ciphertext) <= 24 {
		return "", errors.New("ciphertext too short to contain nonce and data")
	}
	var decryptNonce [24]byte
	copy(decryptNonce[:], ciphertext[:24]) // nonce is first 24 bytes of ciphertext
	decrypted, ok := secretbox.Open(nil, ciphertext[24:], &decryptNonce, &key)
	if !ok {
		return "", errors.New("could not decrypt message")
	}
	return string(decrypted), nil
}

var lastMessage []byte

func getMessagesFromServer(key [32]byte) {
	for {
		response, err := http.Get("http://" + *server + ":8000/rooms/" + *room)
		if err != nil {
			fmt.Printf("The HTTP request to the server failed: %s \n ", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			decrypted, err := decryptMessage(data, key)
			if !(bytes.Equal(data, lastMessage)) {
				if err != nil {
					fmt.Println("A message has been sent, but you don't have the correct key to decrypt it.")
				} else {
					fmt.Println(decrypted)
				}
			}
			lastMessage = data
		}
		time.Sleep(75 * time.Millisecond)
	}
}

func sendMessageToServer(message string, key [32]byte) {
	ciphertext := encryptMessage(message, key)
	_, err := http.Post("http://"+*server+":8000/rooms/"+*room, "application/data", bytes.NewBuffer(ciphertext))
	if err != nil {
		fmt.Println("Could not send message.")
	}

}

func main() {
	flag.Parse()

	if *passcode == "" {
		fmt.Println("Usage: ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Welcome to cryptux.")
	if *room == "" {
		fmt.Println("Usage: ")
		flag.PrintDefaults()
		os.Exit(1)
	}
	key := generateKey(*passcode)
	encrypted := encryptMessage("Hey there!", key)
	decrypted, _ := decryptMessage(encrypted, key)
	fmt.Println(decrypted)

	go getMessagesFromServer(key)

	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString(byte(0x0A)) // break on newline

		sendMessageToServer(input, key)
	}
}
