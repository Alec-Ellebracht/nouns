package main

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// Generates unique ints for room numbers
var lastRoomID int64

// Keeps track of all the rooms
var hotel = make(map[int64]*Room)

//***********************************************************************************************
//
// External
//
//***********************************************************************************************

// CreateRoom builds and tracks a new room
func CreateRoom() *Room {

	// increment the room identifier
	atomic.AddInt64(&lastRoomID, 1)
	roomID := atomic.LoadInt64(&lastRoomID)

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

// GetRoom checks for a specifc room by its id
func GetRoom(id int64) (*Room, bool) {

	room, ok := hotel[id]
	return room, ok
}

//***********************************************************************************************
//
// Internal
//
//***********************************************************************************************

// Run starts a room and sets up the front desk to check in and check out clients
func (room *Room) run() {
	for {

		select {

		case client := <-room.checkin:

			room.clients[client] = true

			log.Println("Client checked in to room..")
			log.Printf("Currently %v connected clients..\n", len(room.clients))

			go reader(client)
			// TO DO : start the writer

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

// Reader defines a reader which will listen for
// new messages being sent to this client
func reader(clt *Client) {

	// ack the client and send back the room number
	roomInfo := fmt.Sprintf("Room: %v", clt.room.id)
	clt.conn.WriteMessage(1, []byte(roomInfo))

	// if the reader returns then we checkout
	// the client since theyre no longer connected
	defer func() {
		clt.room.checkout <- clt
		clt.conn.Close() // kill the socket
	}()
	for {

		// read in all incoming messages
		_, msg, err := clt.conn.ReadMessage()
		if err != nil {
			// log and unregister the client if there is some error
			// such as close tab, nav away, etc...
			log.Println(err)
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				return
			}
		}

		// print out that message for clarity
		log.Println(string(msg))
	}
}

//***********************************************************************************************
//
// Structs
//
//***********************************************************************************************

// Client is a middleman connection and the room
type Client struct {
	room *Room
	conn *websocket.Conn
	send chan []byte
}

// Room is the game room for the clients to play in
type Room struct {
	id       int64
	clients  map[*Client]bool
	checkin  chan *Client
	checkout chan *Client
}
