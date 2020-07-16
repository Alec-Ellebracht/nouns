package main

import (
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

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
	Room         *Room
	Nouns        Bowl
	Players      Group
	Presenter    *Player
	Host         *Player
	CurrentNoun  *Noun
	StartingTime time.Duration
	Rounds       int
	Broadcast    chan interface{}
}

// NewGame constructor for a game
func NewGame(host *Player) Game {

	game := Game{
		Nouns:        Bowl{},
		Players:      Group{},
		Presenter:    nil,
		Host:         host,
		CurrentNoun:  nil,
		StartingTime: 3,
		Rounds:       3,
		Broadcast:    make(chan interface{}),
	}
	return game
}

// Start begins the game
func (g *Game) Start() {

	g.Players.Shuffle()
	g.Nouns.Shuffle()

	g.Presenter = g.Players.First()
	g.CurrentNoun = g.Nouns.First()

}

// DoGuess checks the guess against the current noun
func (g *Game) DoGuess(guess *Guess) bool {

	if g.CurrentNoun.Is(guess.Text) {
		guess.IsCorrect = true
		g.CurrentNoun = g.Nouns.Next()
	}
	return guess.IsCorrect
}

// DoPass moves to the next noun in the bowl
func (g *Game) DoPass() *Noun {

	g.CurrentNoun = g.Nouns.Next()
	return g.CurrentNoun
}

// Bowl struct
type Bowl struct {
	Current int
	All     []Noun
	Guessed []Noun
}

// First gets the starting item in the slice
func (b *Bowl) First() *Noun {
	if len(b.All) == 0 {
		log.Fatalln("Cannot call first without adding nouns to the bowl.")
	}
	return &b.All[0]
}

// Next gets the next item in the slice
func (b *Bowl) Next() *Noun {
	if b.Current == len(b.All)-1 {
		b.Current = 0
	} else {
		b.Current++
	}
	return &b.All[b.Current]
}

// Shuffle randomizes the slice
func (b *Bowl) Shuffle() {

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(b.All),
		func(i, j int) { b.All[i], b.All[j] = b.All[j], b.All[i] })
}

// Add appends an item to the list
func (b *Bowl) Add(nouns ...Noun) {
	b.All = append(b.All, nouns...)
}

// Noun struct
type Noun struct {
	Type NounType `json:"type"`
	Text string   `json:"text"`
}

// Is compares the noun with the provided value
// and returns true if it is a match
func (n Noun) Is(s string) bool {

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

// Player struct
type Player struct {
	Name  string
	Score int
}

// IncrementScore adds to the players current score
func (p *Player) IncrementScore(i int) {
	p.Score += i
}

// SetName adds to the players current score
func (p *Player) SetName(s string) {
	p.Name = s
}

// Group struct
type Group struct {
	Current int
	All     []*Player
}

// First gets the starting item in the slice
func (g *Group) First() *Player {
	if len(g.All) == 0 {
		log.Fatalln("Cannot call first without adding players to the group.")
	}
	return g.All[0]
}

// Next gets the next item in the slice
func (g *Group) Next() *Player {
	if g.Current == len(g.All)-1 {
		g.Current = 0
	} else {
		g.Current++
	}
	return g.All[g.Current]
}

// Shuffle randomizes the slice
func (g *Group) Shuffle() {

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(g.All),
		func(i, j int) { g.All[i], g.All[j] = g.All[j], g.All[i] })
}

// Add appends an item to the list
func (g *Group) Add(others ...*Player) {
	g.All = append(g.All, others...)
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
