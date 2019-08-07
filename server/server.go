package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Room stores the last message of the room and its ID
type Room struct {
	id          string
	lastMessage []byte // contents unknown to the server
	// possibly generate a salt per room, instead of a single hardcoded one?
}

var rooms []Room

// GetLastMessageEndpoint sends the last message for the specified room
func GetLastMessageEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range rooms {
		if item.id == params["id"] {
			fmt.Fprint(w, string(item.lastMessage))
			return
		}
	}
	fmt.Fprint(w, "")
}

// PostMessageOrCreateRoom posts a message and/or creates a room
func PostMessageOrCreateRoom(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	for idx, item := range rooms {
		if item.id == params["id"] {
			lastMessage, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Println(err, "room existed")
			}
			if len(lastMessage) >= 250 {
				return
			}
			var room Room
			room.id = item.id
			room.lastMessage = lastMessage
			rooms[idx] = room
			return
		}
	}

	// otherwise, the room doesn't exist
	var room Room
	room.id = params["id"]
	lastMessage, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err, "room did not exist")
	}
	room.lastMessage = lastMessage
	rooms = append(rooms, room)
	return
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/rooms/{id}", GetLastMessageEndpoint).Methods("GET")
	router.HandleFunc("/rooms/{id}", PostMessageOrCreateRoom).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}
