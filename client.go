package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Keeps track of all the open sessions
var sessions = make(map[string]*Session)
var lastClean = time.Now()

//***********************************************************************************************
//
// External
//
//***********************************************************************************************

// NewClient checks the session id cookie to see if the client
// already has an open session and either connects them back to
// their open session or creates a new one
func NewClient(room *Room, conn *websocket.Conn, sid string) {

	client := &Client{
		room,
		conn,
		make(chan interface{}),
	}

	sessions[sid] = &Session{client, time.Now()}

	// check into the room
	room.checkin <- client

	// start listening for messages
	go reader(client)
	go writer(client)

}

// CheckAndSetSession gets the uuid session id if it exists in the cookies
// else it will add a new session id
func CheckAndSetSession(res http.ResponseWriter, req *http.Request) string {

	sid, err := req.Cookie("sid")

	maxSession := int(time.Duration(time.Hour)/time.Second) * 2

	if err == http.ErrNoCookie {

		sid = &http.Cookie{

			Name:  "sid",
			Value: uuid.New().String(),
			// Secure: true,
			HttpOnly: true,
			MaxAge:   maxSession,
		}

	} else if err != nil {

		log.Println("Error checking cookie..", err)

	} else {

		// bump out the session
		sid.MaxAge = maxSession
	}

	http.SetCookie(res, sid)

	go cleanSessionStorage()

	return sid.Value
}

// ActiveSession checks to see if the vistor has a session id or not
func ActiveSession(res http.ResponseWriter, req *http.Request) bool {
	_, err := req.Cookie("sid")
	return err == nil
}

//***********************************************************************************************
//
// Internal
//
//***********************************************************************************************

// Reader defines a reader which will listen for
// new messages being sent to this client
func reader(client *Client) {

	// ack the client and send back the room number
	roomInfo := fmt.Sprintf("Successfully joined room %v", client.room.ID)
	client.conn.WriteMessage(1, []byte(roomInfo))

	// if the reader returns then we checkout
	// the client since theyre no longer connected
	defer func() {
		client.conn.Close() // kill the socket
		client.room.checkout <- client
	}()
	for {

		msg := struct {
			Event  string `json:"event"`
			Person string `json:"person"`
			Place  string `json:"place"`
			Thing  string `json:"thing"`
		}{}

		err := client.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Received error reading json from client:", err)
			return
		}
		log.Printf("Event: %v, person: %v, place: %v, thing: %v", msg.Event, msg.Person, msg.Place, msg.Thing)

		switch msg.Event {

		case "SUBMIT":
			person := Noun{
				Type: Person,
				Noun: msg.Person,
			}
			place := Noun{
				Type: Place,
				Noun: msg.Place,
			}
			thing := Noun{
				Type: Thing,
				Noun: msg.Thing,
			}
			client.room.CurrGame.submit <- person
			client.room.CurrGame.submit <- place
			client.room.CurrGame.submit <- thing
			client.room.publish <- thing

		case "GUESS":
			client.room.CurrGame.guess <- Guess{
				Guess:  msg.Thing,
				client: client,
			}

		case "HINT":
			client.room.CurrGame.guess <- Guess{
				Guess:  msg.Thing,
				client: client,
			}
		}

	}
}

// Writer will listen for messages from other clients
// and relay them to this client
func writer(client *Client) {

	// if the reader returns then we checkout
	// the client since theyre no longer connected
	defer func() {
		client.conn.Close() // kill the socket
		client.room.checkout <- client
	}()
	for {

		select {

		case message, ok := <-client.send:

			log.Printf("Message type %T:", message)

			if !ok {
				log.Println("Client channel closed:", ok)
				return
			}

			err := client.conn.WriteJSON(message)
			if err != nil {
				log.Println("Received error writing json to client:", err)
				return
			}
		}
	}
}

// CleanSessionStorage periodically goes through all the stored sessions
// and removes any that have not been active for 3 hours or more
// This is the best we can do for now but this could be better
func cleanSessionStorage() {

	if time.Now().Sub(lastClean) > (time.Second * 30) {

		log.Println("Running session cleanup..")
		i := 0

		for key, session := range sessions {
			if time.Now().Sub(session.lastActivity) > (time.Hour * 3) {
				delete(sessions, key)
				i++
			}
		}
		log.Println("Removed", i, "old sessions..")

		lastClean = time.Now()
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
	send chan interface{}
}

// Session tracks the client and the time they were last active
type Session struct {
	client       *Client
	lastActivity time.Time
}
