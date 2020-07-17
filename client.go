package main

import (
	"encoding/json"
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
func NewClient(uid string, name string, room *Room, conn *websocket.Conn) {

	client := &Client{
		room:   room,
		conn:   conn,
		UserID: uid,
		Name:   name,
		send:   make(chan interface{}),
	}

	sessions[uid] = &Session{client, time.Now()}

	// check into the room
	room.checkin <- client

	// start listening for messages
	go reader(client)
	go writer(client)

}

// AddCookies gets the uuid session id if it exists in the cookies
// else it will add a new session id along with the users chosen guestname
func AddCookies(res http.ResponseWriter, req *http.Request) {
	addGuestName(res, req)
	addUserID(res, req)
}

// ActiveSession checks to see if the vistor has a session id or not
func ActiveSession(res http.ResponseWriter, req *http.Request) (string, bool) {
	userID, err := req.Cookie("uid")
	if err != nil {
		log.Println("Error reading cookie for uid", err)
	}
	return userID.Value, err == nil
}

// GetGuestName gets the guestname for the guest
func GetGuestName(res http.ResponseWriter, req *http.Request) string {
	cookie, err := req.Cookie("guestname")
	if err != nil {
		log.Println("Error reading cookie for guestname", err)
		return "annonymous"
	}
	return cookie.Value
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

			submission := struct {
				Person string `json:"person"`
				Place  string `json:"place"`
				Thing  string `json:"thing"`
			}{}

			err := json.Unmarshal(body, &submission)
			if err != nil {
				log.Println("Error unmarshalling json for submission:", err)
				return
			}

			person := Noun{
				Type: Person,
				Text: submission.Person,
			}
			place := Noun{
				Type: Place,
				Text: submission.Place,
			}
			thing := Noun{
				Type: Thing,
				Text: submission.Thing,
			}

			client.room.CurrGame.Nouns.Add(person, place, thing)

		case "message":

			message := struct {
				Message string `json:"message"`
			}{}

			err := json.Unmarshal(body, &message)
			if err != nil {
				log.Panicln("Error unmarshalling json for message:", err)
				return
			}

			if client.room.CurrGame.Presenter.UserID == client.UserID {

				hint := message.Message
				client.room.publish <- &Hint{
					Text:   hint,
					Noun:   *client.room.CurrGame.CurrentNoun,
					client: client,
				}

			} else {

				guess := &Guess{
					Text:   message.Message,
					Player: client.Name,
					client: client,
				}

				client.room.CurrGame.DoGuess(guess)
			}

		case "start":

			client.room.CurrGame.Start()
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
			case *Noun:
				log.Println("Sending a Noun")
				env = Envelope{
					Type: "noun",
					Body: message,
				}
			case *Guess:
				log.Println("Sending a Guess")
				env = Envelope{
					Type: "guess",
					Body: message,
				}
			case *Hint:
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
			case PlayerAction:
				log.Println("Sending a player action message")
				env = Envelope{
					Type: "action",
					Body: message,
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

// addGuestName adds the guestname for the guest
func addGuestName(res http.ResponseWriter, req *http.Request) {

	err := req.ParseForm()
	if err != nil {
		log.Println("Error parsing Join form", err)
		http.Error(res, "Oh poop, something went wrong reading your request.", http.StatusBadRequest)
	}

	// get the guest name from the form
	// or use annonymous
	gn := "annonymous"
	xgn := req.Form["guestname"]
	if len(xgn) > 0 {
		gn = xgn[0]
	}

	gnc := &http.Cookie{
		Name:     "guestname",
		Value:    gn,
		HttpOnly: true,
		MaxAge:   0,
	}

	http.SetCookie(res, gnc)
}

// addUserId adds or bumps out the guests session
func addUserID(res http.ResponseWriter, req *http.Request) {

	uid, err := req.Cookie("uid")

	maxSession := int(time.Duration(time.Hour)/time.Second) * 2

	if err == http.ErrNoCookie {

		uid = &http.Cookie{

			Name:  "uid",
			Value: uuid.New().String(),
			// Secure: true,
			HttpOnly: true,
			MaxAge:   maxSession,
		}

	} else if err != nil {

		log.Println("Error checking uid cookie..", err)

	} else {

		// bump out the session
		uid.MaxAge = maxSession
	}

	http.SetCookie(res, uid)

	// TO DO : put this in a better spot
	go cleanSessionStorage()
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
	room   *Room
	conn   *websocket.Conn
	UserID string `json:"userID"`
	Name   string `json:"name"`
	send   chan interface{}
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
