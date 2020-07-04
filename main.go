package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*html"))
}

func main() {

	http.HandleFunc("/", index)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/guess", guessHandler)

	http.Handle("/resource/", http.StripPrefix("/resource/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

func index(res http.ResponseWriter, req *http.Request) {

	tpl.ExecuteTemplate(res, "index.html", false)
}

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
