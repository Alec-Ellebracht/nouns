package main

import (
	"fmt"
	"log"
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

			log.Printf("Noun %v submitted to game..", noun.Noun)

		case guess := <-game.guess:

			guess.IsCorrect = game.currentNoun.is(guess.Guess)
			log.Printf("Is guess %v equal to %v? %v", guess.Guess, game.currentNoun.Noun, guess.IsCorrect)

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

		if &noun == game.currentNoun {
			isNext = true
		}

		if isNext {
			return noun
		}
	}

	return first
}

// func (game *Game) nextRound() Noun {
// 	return game.submissions[0]
// }

// Noun struct
type Noun struct {
	Type NounType
	Noun string
}

// PrintType prints out the type of Noun
// this is a Person, Place or Thing
func (n Noun) PrintType() string {
	return string(n.Type)
}

// Is compares the noun with the provided value
// and returns true if it is a match
func (n Noun) is(guess string) bool {

	lowerNoun := strings.ToLower(n.Noun)

	// exact match
	if lowerNoun == strings.ToLower(guess) {
		return true
	}

	sentence := strings.Fields(guess)

	match := false
	for _, word := range sentence {
		fmt.Println(word)
		if lowerNoun == strings.ToLower(word) {
			match = true
		}
	}

	return match
}

// Guess struct
type Guess struct {
	Guess     string
	IsCorrect bool
	Noun      string
	client    *Client
}

// Hint struct
type Hint struct {
	Hint   string
	Noun   Noun
	client *Client
}

// Start struct
type Start struct {
	IsStarted bool
}
