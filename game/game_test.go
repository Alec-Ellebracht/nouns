package game

import (
	"testing"
)

func TestStart(t *testing.T) {

	game := Game{}

	p1 := Player{"one", 0}
	n1 := Noun{Person, "dumbledore"}

	game.Players.Add(&p1)
	game.Nouns.Add(n1)

	game.Start()

	if game.Presenter != &p1 {
		t.Error("Expected", "player 1", "got", game.Presenter)
	}

	if game.CurrentNoun.Text != n1.Text {
		t.Error("Expected", "noun 1", "got", game.CurrentNoun)
	}

}

func TestPlayerAdd(t *testing.T) {

	game := Game{}

	p1 := Player{Name: "one"}
	p2 := Player{Name: "two"}
	p3 := Player{Name: "three"}

	game.Players.Add(&p1)
	game.Players.Add(&p2, &p3)

	if len(game.Players.All) != 3 {
		t.Error("Expected", "3 players", "got", len(game.Players.All))
	}
}

func TestPlayerFirst(t *testing.T) {

	game := Game{}

	p1 := Player{Name: "one"}
	p2 := Player{Name: "two"}
	p3 := Player{Name: "three"}

	game.Players.Add(&p1)
	game.Players.Add(&p2, &p3)

	if game.Players.First().Name != p1.Name {
		t.Error("Expected", "one", "got", game.Players.First().Name)
	}
}

func TestNounAdd(t *testing.T) {

	game := Game{}

	n1 := Noun{Person, "dumbledore"}
	n2 := Noun{Place, "hogwarts"}
	n3 := Noun{Thing, "wand"}

	game.Nouns.Add(n1)
	game.Nouns.Add(n2, n3)

	if len(game.Nouns.All) != 3 {
		t.Error("Expected", "3 Nouns", "got", len(game.Nouns.All))
	}
}

func TestNounFirst(t *testing.T) {

	game := Game{}

	n1 := Noun{Person, "dumbledore"}
	n2 := Noun{Place, "hogwarts"}
	n3 := Noun{Thing, "wand"}

	game.Nouns.Add(n1, n2, n3)

	if game.Nouns.First().Text != n1.Text {
		t.Error("Expected", "dumbledore", "got", game.Nouns.First())
	}
}
