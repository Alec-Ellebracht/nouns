package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gorilla/websocket"
)

var tpl *template.Template

// Parse all the html files in the templates folder
func init() {
	tpl = template.Must(template.ParseGlob("templates/*html"))
}

// Start
func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", index)
	mux.HandleFunc("/ws", socketHandler)
	mux.HandleFunc("/submit", submitHandler)
	mux.HandleFunc("/guess", guessHandler)

	// serves all the static resources for js and css
	mux.Handle("/resource/", http.StripPrefix("/resource/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// Handles the index page
func index(res http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(res, "index.html", false)
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

	convertedNounType, _ := strconv.Atoi(req.Form.Get("nounType"))

	n := Noun{
		nounType: convertedNounType,
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

		// fmt.Fprintf(res, "Your guess was "+outcome)
		res.Write([]byte("Your guess was " + outcome))

	} else if req.Method == http.MethodGet {

		n := submissions[0]
		tpl.ExecuteTemplate(res, "guess.html", n)
		return

	} else {

		res.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(res, "invalid_http_method")
		return
	}
}

// Handles incoming websockets requests
func socketHandler(res http.ResponseWriter, req *http.Request) {

	room, ok := getRoom(0)
	if !ok {

		room = createRoom()
		fmt.Println("Created room", room.id)
	}

	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Client upgraded to websocket..")

	client := &Client{
		room,
		conn,
		make(chan []byte, 256),
	}

	room.checkin <- client
}

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// func runGame() {

// 	fmt.Println("Initializing questions...")

// 	n := Noun{
// 		Person, "mr deez", []string{},
// 	}

// 	n2 := Noun{
// 		Thing, "deez", []string{},
// 	}

// 	n1 := Noun{
// 		Place, "deez street", []string{},
// 	}

// 	SubmitNoun(n)
// 	SubmitNoun(n1)
// 	SubmitNoun(n2)

// 	fmt.Println(submissions)

// 	for _, sub := range submissions {

// 		guess := "deez"
// 		fmt.Println("Is the", sub.printType(), `"`+guess+`"`, "?", sub.is(guess))
// 	}
// }
