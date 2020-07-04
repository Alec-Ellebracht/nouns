package main

import "fmt"

// noun types
const (
	Person = iota
	Place
	Thing
)

// noun struct
type Noun struct {
	nounType int
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

func (n Noun) is(s string) string {
	if n.noun == s {
		return "Correct!"
	} else {
		return "incorrect..."
	}
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
