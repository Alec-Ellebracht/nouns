package main

import (
	"encoding/json"
	"fmt"
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
		room: room,
		conn: conn,
		send: make(chan interface{}),
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
func ActiveSession(res http.ResponseWriter, req *http.Request) (string, bool) {
	sid, err := req.Cookie("sid")
	return sid.Value, err == nil
}

// SetName finds the client object and sets the name field
// once the player tells us what it is
func SetName(sid string, name string) {
	session, ok := sessions[sid]
	if !ok {
		log.Println("name not set for", sid)
		return
	}
	session.client.name = name
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
		var body json.RawMessage
		env := Envelope{Body: &body}

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

			err := json.Unmarshal(body, &nouns)
			if err != nil {
				log.Println("Error unmarshalling json for submission:", err)
				return
			}

			client.room.CurrGame.submit <- Noun{
				Type: Person,
				Text: nouns.Person,
			}
			client.room.CurrGame.submit <- Noun{
				Type: Place,
				Text: nouns.Place,
			}
			client.room.CurrGame.submit <- Noun{
				Type: Thing,
				Text: nouns.Thing,
			}

		case "message":

			message := struct {
				Message string `json:"message"`
			}{}

			err := json.Unmarshal(body, &message)
			if err != nil {
				log.Panicln("Error unmarshalling json for guess:", err)
				return
			}

			if client.room.CurrGame.currentPlayer == client {

				hint := message.Message
				client.room.publish <- Hint{
					Text:   hint,
					Noun:   *client.room.CurrGame.currentNoun,
					client: client,
				}

			} else {

				guess := message.Message
				isCorrect := client.room.CurrGame.currentNoun.is(guess)

				var noun string
				if isCorrect {
					noun = client.room.CurrGame.currentNoun.Text

					go func() {
						next := client.room.CurrGame.nextNoun()
						client.room.CurrGame.currentNoun = &next
						client.room.CurrGame.currentPlayer.send <- next
					}()
				}

				client.room.publish <- Guess{
					Text:      guess,
					IsCorrect: isCorrect,
					Noun:      noun,
					Player:    client.name,
					client:    client,
				}

			}

		case "start":

			// TO DO : move to game file
			client.room.publish <- Start{true}

			shuffled := client.room.CurrGame.submissions

			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

			client.room.CurrGame.submissions = shuffled
			client.room.CurrGame.currentNoun = &shuffled[0]
			client.room.CurrGame.currentPlayer = client

			time.Sleep(time.Second * 2)
			fmt.Println(&shuffled[0])
			client.send <- shuffled[0]

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
					Body: message,
				}
			case Guess:
				log.Println("Sending a Guess")
				env = Envelope{
					Type: "guess",
					Body: message,
				}
			case Hint:
				log.Println("Sending a Hint")
				env = Envelope{
					Type: "hint",
					Body: message,
				}
			case Start:
				log.Println("Sending a start message")
				env = Envelope{
					Type: "start",
					Body: nil,
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
	name string
	send chan interface{}
}

// Session tracks the client and the time they were last active
type Session struct {
	client       *Client
	lastActivity time.Time
}

// Envelope allows for better json comms on the websocket
type Envelope struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}
