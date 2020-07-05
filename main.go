package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gorilla/websocket"
)

//***********************************************************************************************
//
// Init
//
//***********************************************************************************************

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

	mux := http.NewServeMux()

	// route handlers
	mux.HandleFunc("/", index)
	mux.HandleFunc("/ws", socketHandler)
	mux.HandleFunc("/404", notfoundHandler)
	mux.HandleFunc("/join", joinHandler)
	mux.HandleFunc("/submit", submitHandler)
	mux.HandleFunc("/guess", guessHandler)

	// serves all the static resources for js and css
	mux.Handle("/resource/", http.StripPrefix("/resource/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", mux))
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

	room, ok := GetRoom(1)
	if !ok {

		room = CreateRoom()
		log.Println("Created room", room.id)
	}

	// upgrade the req to a websocket
	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		log.Println("Error upgrading conn to socket", err)
		return
	}
	log.Println("Client upgraded to websocket..")

	// create the new client
	sid := CheckAndSetSession(res, req)
	NewClient(room, conn, sid)
}

// Handles the 404 page
func notfoundHandler(res http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(res, "notfound.html", nil)
}

// Handles the join page
func joinHandler(res http.ResponseWriter, req *http.Request) {

	// TO DO : create sessions in a better spot
	sid := CheckAndSetSession(res, req)
	log.Printf("Checked sid %T", sid)

	// send an error back if its not a post req
	if req.Method == http.MethodPost {

		req.ParseForm()

		fmt.Println(req.Form)

	} else if req.Method == http.MethodGet {

		tpl.ExecuteTemplate(res, "join.html", nil)

	} else {

		http.Redirect(res, req, "/404", http.StatusSeeOther)
	}

	return
}

// Handles the Noun submission page
func submitHandler(res http.ResponseWriter, req *http.Request) {

	// send an error back if its not a post req
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(res, "invalid_http_method")
		return
	}

	// Must call ParseForm() before working with data
	req.ParseForm()
	log.Println(req.Form)

	intNounType, _ := strconv.Atoi(req.Form.Get("nounType"))

	n := Noun{
		nounType: NounType(intNounType),
		noun:     req.Form.Get("noun"),
		hints:    []string{req.Form.Get("hint")},
	}

	SubmitNoun(n)

	http.Redirect(res, req, "/guess", http.StatusFound)
}

// Handles the Noun guessing page
func guessHandler(res http.ResponseWriter, req *http.Request) {

	// send an error back if its not a post req
	if req.Method == http.MethodPost {

		req.ParseForm()

		n := submissions[0]
		outcome := n.is(req.Form.Get("guess"))

		fmt.Fprintf(res, fmt.Sprintf("Your guess was %v", outcome))

	} else if req.Method == http.MethodGet {

		n := submissions[0]
		tpl.ExecuteTemplate(res, "guess.html", n)

	} else {

		res.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(res, "invalid_http_method")
	}

	return
}
