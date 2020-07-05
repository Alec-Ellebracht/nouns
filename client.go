package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Keeps track of all the open sessions
var sessions = make(map[string]*Client)

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
	if existing, ok := sessions[sid]; ok {

		client = existing
		client.room = room
		client.conn = conn

	} else {

		client = &Client{
			room,
			conn,
			make(chan []byte, 256),
		}

		sessions[sid] = client

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
