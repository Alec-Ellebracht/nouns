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

	// check and track the clients session
	var client *Client
	if session, ok := sessions[sid]; ok {

		// session.lastActivity = time.Now()
		client = session.client
		client.room = room
		client.conn = conn

	} else {

		client = &Client{
			room,
			conn,
			make(chan []byte, 256),
		}

		sessions[sid] = &Session{client, time.Now()}

		// check into the room
		client.room.checkin <- client
	}

	// start listening for messages
	go reader(client)
	// TO DO : start the writer
}

// CheckAndSetSession gets the uuid session id if it exists in the cookies
// else it will add a new session id
func CheckAndSetSession(res http.ResponseWriter, req *http.Request) string {

	sid, err := req.Cookie("sid")

	if err == http.ErrNoCookie {

		sid = &http.Cookie{

			Name:  "sid",
			Value: uuid.New().String(),
			// Secure: true,
			HttpOnly: true,
		}

		http.SetCookie(res, sid)

	} else if err != nil {

		log.Println(err)
	}

	go cleanSessionStorage()

	return sid.Value
}

//***********************************************************************************************
//
// Internal
//
//***********************************************************************************************

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

// CleanSessionStorage periodically goes through all the stored session
// and removes any that have not been active for 3 hours or more
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
	send chan []byte
}

// Session tracks the client and the time they were last active
type Session struct {
	client       *Client
	lastActivity time.Time
}
