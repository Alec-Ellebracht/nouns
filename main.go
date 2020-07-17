package main

import (
	"flag"
	"log"
	"net/http"
	"path"
	"runtime"
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
	mux.HandleFunc("/admin", adminHandler)
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

// Handles the admin page
func adminHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {

		http.Redirect(res, req, "/404", http.StatusSeeOther)
	}

	adminData := struct {
		Routines int
		Cpus     int
		Rooms    int
		Sessions int
	}{
		runtime.NumGoroutine(),
		runtime.NumCPU(),
		len(hotel),
		len(sessions),
	}

	tpl.ExecuteTemplate(res, "admin.html", adminData)
}

// Handles the join page
func joinHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		err := req.ParseForm()
		if err != nil {
			log.Println("Error parsing Join form", err)
			http.Error(res, "Oh poop, something went wrong reading your request.", http.StatusBadRequest)
		}

		AddCookies(res, req)
		roomPath := GenerateRoomPath(req.Form)

		http.Redirect(res, req, roomPath, http.StatusSeeOther)

	} else if req.Method == http.MethodGet {

		tpl.ExecuteTemplate(res, "join.html", nil)

	} else {

		http.Redirect(res, req, "/404", http.StatusSeeOther)
	}

	return
}

// Handles the room page
func roomHandler(res http.ResponseWriter, req *http.Request) {

	_, isActive := ActiveSession(res, req)

	if !isActive {

		http.Redirect(res, req, "/join", http.StatusSeeOther)

	} else if req.Method == http.MethodGet {

		roomPath := path.Base(req.URL.Path)
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

// To upgrade a http connection to a websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// TO DO : check origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handles incoming websockets requests
func socketHandler(res http.ResponseWriter, req *http.Request) {

	roomPath := path.Base(req.URL.String())
	roomID, _ := strconv.ParseInt(roomPath, 10, 64)
	log.Println("Socket attempt to connect to room", roomPath)

	// upgrade the req to a websocket
	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {

		log.Println("Error upgrading conn to socket", err)
		http.Error(res, "Uh oh, there was an issue connecting to the host.", http.StatusBadRequest)
		return
	}
	log.Println("Client upgraded to websocket in room", roomID)

	room, ok := GetRoom(roomID)
	if !ok {

		log.Println("Socket error finding room", roomPath)

		conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "We couldn't find the room you were looking for."))

		conn.Close()

		return
	}

	// create the new client
	sid, _ := ActiveSession(res, req)
	NewClient(sid, room, conn)
}
