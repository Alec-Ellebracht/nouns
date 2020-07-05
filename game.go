package main

import (
	"fmt"
	"strings"
)

// for state

// a list of submitted nouns
var submissions []Noun

//***********************************************************************************************
//
// External Funcs
//
//***********************************************************************************************

// SubmitNoun handles a new noun submission and adds it to the pool
func SubmitNoun(n Noun) {
	submissions = append(submissions, n)

	fmt.Println("current submissions", submissions)
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
