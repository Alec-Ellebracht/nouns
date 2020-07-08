package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"

	"github.com/gorilla/websocket"
)

//***********************************************************************************************
//
// Init
//
//***********************************************************************************************

var addr = flag.String("addr", ":80", "http service address")

var tpl *template.Template

// Parse all the html files in the templates folder
func init() {
	tpl = template.Must(template.ParseGlob("templates/*html"))
}

//***********************************************************************************************
//
// Start
//
//***********************************************************************************************

func main() {

	flag.Parse()

	mux := http.NewServeMux()

	// route handlers
	mux.HandleFunc("/", index)
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/ws/", socketHandler)
	mux.HandleFunc("/404", notfoundHandler)
	mux.HandleFunc("/join", joinHandler)
	mux.HandleFunc("/room/", roomHandler)

	// serves all the static resources for js and css
	mux.Handle("/resource/", http.StripPrefix("/resource/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(*addr, mux))
}

//***********************************************************************************************
//
// Internal Route Handlers
//
//***********************************************************************************************

// Handles the index page
func index(res http.ResponseWriter, req *http.Request) {

	tpl.ExecuteTemplate(res, "index.html", false)
}

// Handles the favicon
func faviconHandler(res http.ResponseWriter, req *http.Request) {

	http.ServeFile(res, req, "static/img/favicon.ico")
}

// Handles the 404 page
func notfoundHandler(res http.ResponseWriter, req *http.Request) {

	tpl.ExecuteTemplate(res, "notfound.html", nil)
}

// To upgrade a http connection to a websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// TO DO : check origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handles incoming websockets requests
func socketHandler(res http.ResponseWriter, req *http.Request) {

	// TO DO : add req parsing to get a specifc room
	// or a request to create a new room

	roomPath := path.Base(req.URL.String())
	roomID, _ := strconv.ParseInt(roomPath, 10, 64)
	log.Println("Socket attempt to connect to room", roomPath)

	room, ok := GetRoom(roomID)
	if !ok {

		log.Println("Socket error finding room", roomPath)
		http.Error(res, "We couldn't find the room you were looking for.", http.StatusNotFound)
		return
	}

	// upgrade the req to a websocket
	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {

		log.Println("Error upgrading conn to socket", err)
		http.Error(res, "Uh oh, there was an issue connecting to the host.", http.StatusBadRequest)
		return
	}
	log.Println("Client upgraded to websocket in room", roomID)

	// create the new client
	sid := CheckAndSetSession(res, req)
	NewClient(room, conn, sid)
}

// Handles the join page
func joinHandler(res http.ResponseWriter, req *http.Request) {

	// TO DO : create sessions in a better spot
	sid := CheckAndSetSession(res, req)
	log.Printf("Checked sid %T", sid)

	// send an error back if its not a post req
	if req.Method == http.MethodPost {

		err := req.ParseForm()
		if err != nil {
			log.Println("Error parsing Join form", err)
		}

		// guestName := req.Form["nickname"][0]
		roomID := req.Form["room"][0]

		if len(roomID) > 0 {

			route := fmt.Sprintf("/room/%v", roomID)
			http.Redirect(res, req, route, http.StatusSeeOther)

		} else {

			newRoom := CreateRoom()
			route := fmt.Sprintf("/room/%v", newRoom.ID)
			http.Redirect(res, req, route, http.StatusSeeOther)
		}

	} else if req.Method == http.MethodGet {

		tpl.ExecuteTemplate(res, "join.html", nil)

	} else {

		http.Redirect(res, req, "/404", http.StatusSeeOther)
	}

	return
}

// Handles the room page
func roomHandler(res http.ResponseWriter, req *http.Request) {

	if !ActiveSession(res, req) {

		http.Redirect(res, req, "/join", http.StatusSeeOther)

	} else if req.Method == http.MethodPost {

		log.Println(req.Method, "to ROOM handler")

	} else if req.Method == http.MethodGet {

		roomPath := path.Base(req.URL.String())
		roomID, _ := strconv.ParseInt(roomPath, 10, 64)

		log.Println(req.Method, "to join", roomPath)

		room, ok := GetRoom(roomID)
		if !ok {

			http.Redirect(res, req, "/404", http.StatusSeeOther)
			return
		}

		tpl.ExecuteTemplate(res, "room.html", room)

	} else {

		http.Redirect(res, req, "/404", http.StatusSeeOther)
	}

	return
}
