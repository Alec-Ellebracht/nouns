package main

import (
	"fmt"
	"strings"
)

// noun types
type NounType int

const (
	Person NounType = iota
	Place
	Thing
)

// noun struct
type Noun struct {
	nounType NounType
	noun     string
	hints    []string
}

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

// guess struct
type Guess struct {
	guess string
}

// for state

// a list of submitted nouns
var submissions []Noun

// func to submit a new noun to the pool
func SubmitNoun(n Noun) {
	submissions = append(submissions, n)

	fmt.Println("current submissions", submissions)
}
