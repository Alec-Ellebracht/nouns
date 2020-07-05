package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

const (
	roomID = iota
)

// Keeps track of all the rooms
var hotel = make(map[int]*Room)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	room *Room

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

type Room struct {
	// Room id
	id int

	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	checkin chan *Client

	// Unregister requests from clients.
	checkout chan *Client
}

func (room *Room) run() {
	for {

		select {

		case client := <-room.checkin:

			room.clients[client] = true

			log.Println("Client checked in to room..")
			log.Printf("Currently %v connected clients..\n", len(room.clients))

			go reader(client)

		case client := <-room.checkout:

			if _, ok := room.clients[client]; ok {

				delete(room.clients, client)
				close(client.send)
			}
			log.Println("Client checked out of room..")
			log.Printf("Currently %v connected clients..\n", len(room.clients))
		}
	}
}

// Creates and tracks a new room
func createRoom() *Room {

	newRoom := &Room{
		id:       roomID,
		checkin:  make(chan *Client),
		checkout: make(chan *Client),
		clients:  make(map[*Client]bool),
	}

	// add room to the list of active rooms
	hotel[newRoom.id] = newRoom

	// start the room in a routine
	go newRoom.run()

	return newRoom
}

// Checks for a specifc room
func getRoom(id int) (*Room, bool) {

	if _, ok := hotel[id]; ok {

		return hotel[id], true
	}

	return nil, false
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(clt *Client) {

	welcome := fmt.Sprintf("Room: %v", clt.room.id)
	clt.conn.WriteMessage(1, []byte(welcome))

	for {

		// read in a message
		_, msg, err := clt.conn.ReadMessage()
		if err != nil {
			log.Println("return")
			return
		}

		// print out that message for clarity
		log.Println(string(msg))
	}
}
