package main

import (
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
				game.currentNoun = noun
			}

			log.Printf("Noun %v submitted to game..", noun.noun)

		case guess := <-game.guess:

			outcome := game.currentNoun.is(guess.guess)
			log.Printf("Is guess %v equal to %v? %v", guess.guess, game.currentNoun.noun, outcome)
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
}

//***********************************************************************************************
//
// Enums
//
//***********************************************************************************************

// NounType is the type of noun
// this is a Person, Place or Thing
type NounType int

const (
	Person NounType = iota
	Place
	Thing
)

//***********************************************************************************************
//
// Structs
//
//***********************************************************************************************

// Game struct
type Game struct {
	submissions []*Noun
	currentNoun *Noun
	submit      chan *Noun
	guess       chan *Guess
}

// Noun struct
type Noun struct {
	nounType NounType
	noun     string
	hints    []string
}

// PrintType prints out the type of Noun
// this is a Person, Place or Thing
func (n Noun) PrintType() string {

	switch n.nounType {

	case Person:
		return "Person"
	case Place:
		return "Place"
	case Thing:
		return "Thing"
	}
	return ""
}

func (n Noun) is(s string) bool {
	return strings.ToLower(n.noun) == strings.ToLower(s)
}

// Guess struct
type Guess struct {
	guess string
}
