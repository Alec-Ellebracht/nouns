package main

import (
	"log"
	"sync/atomic"
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

	game := &Game{
		submissions: []Noun{},
		currentNoun: nil,
		submit:      make(chan Noun),
		guess:       make(chan Guess),
	}

	go game.Run()

	newRoom := &Room{
		ID:       roomID,
		CurrGame: game,
		checkin:  make(chan *Client),
		checkout: make(chan *Client),
		publish:  make(chan interface{}),
		clients:  make(map[*Client]bool),
	}

	// add room to the list of active rooms
	hotel[newRoom.ID] = newRoom

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

	defer func() {

		log.Printf("All guests have left room %v, sending in the cleanup crew..\n", room.ID)

		close(room.checkin)
		close(room.checkout)
		delete(hotel, room.ID)
	}()
	for {

		select {

		case client := <-room.checkin:

			room.clients[client] = true

			log.Println("Client checked in to room..")
			log.Printf("Currently %v connected clients..\n", len(room.clients))

			// room.publish <- "Someone else joined the room!"

		case client := <-room.checkout:

			if _, ok := room.clients[client]; ok {

				delete(room.clients, client)
				close(client.send)
			}
			log.Println("Client checked out of room..")
			log.Printf("Currently %v connected clients..\n", len(room.clients))

		case message := <-room.publish:

			for client := range room.clients {

				select {

				case client.send <- message:

				default:
					close(client.send)
					delete(room.clients, client)
				}
			}
		}
	}
}

//***********************************************************************************************
//
// Structs
//
//***********************************************************************************************

// Room is the game room for the clients to play in
type Room struct {
	ID       int64
	CurrGame *Game
	clients  map[*Client]bool
	checkin  chan *Client
	checkout chan *Client
	publish  chan interface{}
}

func (room *Room) empty() bool {
	return len(room.clients) == 0
}
