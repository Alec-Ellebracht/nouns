package main

import (
	"encoding/json"
	"log"
	"math/rand"
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

	// if the reader returns then we checkout
	// the client since theyre no longer connected
	defer func() {
		client.conn.Close() // kill the socket
		client.room.checkout <- client
	}()
	for {

		// unmarshal to the envelope first
		// then based on the type we further back
		// into the specific type of message
		var msg json.RawMessage
		env := Envelope{Msg: &msg}

		err := client.conn.ReadJSON(&env)
		if err != nil {
			log.Println("Received error reading json from client:", err)
			return
		}
		log.Printf("Received %v type payload..", env.Type)

		switch env.Type {

		case "submit":

			nouns := struct {
				Person string `json:"person"`
				Place  string `json:"place"`
				Thing  string `json:"thing"`
			}{}

			err := json.Unmarshal(msg, &nouns)
			if err != nil {
				log.Println("Error unmarshalling json for submission:", err)
				return
			}

			client.room.CurrGame.submit <- Noun{
				Type: Person,
				Noun: nouns.Person,
			}
			client.room.CurrGame.submit <- Noun{
				Type: Place,
				Noun: nouns.Place,
			}
			client.room.CurrGame.submit <- Noun{
				Type: Thing,
				Noun: nouns.Thing,
			}

		case "guess":

			guess := struct {
				Guess string `json:"guess"`
			}{}

			err := json.Unmarshal(msg, &guess)
			if err != nil {
				log.Panicln("Error unmarshalling json for guess:", err)
				return
			}

			client.room.publish <- Guess{
				Guess:  guess.Guess,
				client: client,
			}

		case "hint":

			hint := struct {
				Hint string `json:"guess"`
			}{}

			err := json.Unmarshal(msg, &hint)
			if err != nil {
				log.Panicln("Error unmarshalling json for hint:", err)
				return
			}

			client.room.CurrGame.hint <- Hint{
				Hint:   hint.Hint,
				client: client,
			}

		case "start":

			nouns := client.room.CurrGame.submissions

			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(nouns), func(i, j int) { nouns[i], nouns[j] = nouns[j], nouns[i] })

			client.room.publish <- Hint{
				Hint:   "A wizarding school",
				Noun:   nouns[0],
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

			env := Envelope{}

			switch message.(type) {
			case Noun:
				log.Println("Sending a Noun")
				env = Envelope{
					Type: "noun",
					Msg:  message,
				}
			case Guess:
				log.Println("Sending a Guess")
				env = Envelope{
					Type: "guess",
					Msg:  message,
				}
			case Hint:
				log.Println("Sending a Hint")
				env = Envelope{
					Type: "hint",
					Msg:  message,
				}
			}

			err := client.conn.WriteJSON(env)
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

// Envelope allows for better json comms on the websocket
type Envelope struct {
	Type string
	Msg  interface{}
}
