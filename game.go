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

			outcome := game.currentNoun.is(guess.Guess)
			log.Printf("Is guess %v equal to %v? %v", guess.Guess, game.currentNoun.Noun, outcome)

		case hint := <-game.hint:
			fmt.Println(hint)
		}
	}
}

//***********************************************************************************************
//
// Enums
//
//***********************************************************************************************

func cleanupGame(game *Game) {
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
	submissions []Noun
	currentNoun *Noun
	submit      chan Noun
	guess       chan Guess
	hint        chan Guess
}

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

func (n Noun) is(s string) bool {
	return strings.ToLower(n.Noun) == strings.ToLower(s)
}

// Guess struct
type Guess struct {
	Guess  string
	client *Client
}

// Hint struct
type Hint struct {
	Hint   string
	client *Client
}
