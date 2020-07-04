package main

import "github.com/gorilla/websocket"

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	room *Room

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

type Room struct {
	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func (room *Room) run() {
	for {

		select {

		case client := <-room.register:
			room.clients[client] = true

		case client := <-room.unregister:
			if _, ok := room.clients[client]; ok {

				delete(room.clients, client)
				close(client.send)
			}
		}
	}
}

func createRoom() *Room {
	return &Room{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}
