package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

//***********************************************************************************************
//
// External Funcs
//
//***********************************************************************************************

// Run starts the game
func (game *Game) Run() {

	defer cleanupGame(game)
	for {

		select {

		case noun := <-game.submit:

			game.submissions = append(game.submissions, noun)
			if game.currentNoun == nil {
				game.currentNoun = &noun
			}

			log.Printf("Noun %v submitted to game..", noun.Text)

		case guess := <-game.guess:

			guess.IsCorrect = game.currentNoun.is(guess.Text)
			log.Printf("Is guess %v equal to %v? %v", guess.Text, game.currentNoun.Text, guess.IsCorrect)

			// go func() {
			// 	game.room.publish <- guess
			// }()

		case hint := <-game.hint:
			fmt.Println(hint)

			// go func() {
			// 	game.room.publish <- hint
			// }()
		}
	}
}

//***********************************************************************************************
//
// Enums
//
//***********************************************************************************************

func cleanupGame(game *Game) {

	log.Println("Cleaning up the game...")

	close(game.submit)
	close(game.guess)
	close(game.hint)
}

//***********************************************************************************************
//
// Enums
//
//***********************************************************************************************

// NounType is the type of noun
// this is a Person, Place or Thing
type NounType string

const (
	Person NounType = "person"
	Place  NounType = "place"
	Thing  NounType = "thing"
)

//***********************************************************************************************
//
// Structs
//
//***********************************************************************************************

// Game struct
type Game struct {
	room          *Room
	submissions   []Noun
	currentNoun   *Noun
	currentPlayer *Client
	submit        chan Noun
	guess         chan Guess
	hint          chan Hint
}

func (game *Game) nextPlayer() *Client {

	var first *Client
	for client := range game.room.clients {
		first = client
		break
	}

	isNext := false
	for client := range game.room.clients {

		if client == game.currentPlayer {
			isNext = true
		}

		if isNext {
			return client
		}
	}

	return first
}

func (game *Game) nextNoun() Noun {

	first := game.submissions[0]

	isNext := false
	for _, noun := range game.submissions {

		if isNext {
			return noun
		}

		if noun.Text == game.currentNoun.Text {
			isNext = true
		}
	}

	return first
}

// func (game *Game) nextRound() Noun {
// 	return game.submissions[0]
// }

// Noun struct
type Noun struct {
	Type NounType `json:"type"`
	Text string   `json:"text"`
}

// PrintType prints out the type of Noun
// this is a Person, Place or Thing
func (n Noun) PrintType() string {
	return string(n.Type)
}

// Is compares the noun with the provided value
// and returns true if it is a match
func (n Noun) is(s string) bool {

	nonLetter, err := regexp.Compile(`[^\w]`)
	if err != nil {
		log.Fatal(err)
	}

	noun := nonLetter.ReplaceAllString(strings.ToLower(n.Text), " ")
	guess := nonLetter.ReplaceAllString(strings.ToLower(s), " ")

	// exact match
	if noun == guess {
		return true
	}

	sentence := strings.Fields(guess)

	match := false
	for _, word := range sentence {

		if noun == word {
			match = true
		}
	}

	return match
}

// Guess struct
type Guess struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
	Noun      string `json:"noun"`
	Player    string `json:"player"`
	client    *Client
}

// Hint struct
type Hint struct {
	Text   string `json:"text"`
	Noun   Noun   `json:"noun"`
	client *Client
}

// Start struct
type Start struct {
	IsStarted bool `json:"isStarted"`
}
