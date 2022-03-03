package entity

import (
	"log"
	"testing"
)

var deck *Deck

func GetDeck() *Deck {
	if deck == nil {
		deck = NewDeck()
	}
	return deck
}

func TestShuffle(t *testing.T) {
	deck := GetDeck()
	t.Logf("deck total %v", len(deck.Cards.GetCards()))
	t.Logf("deck detail")
	for _, card := range deck.Cards.GetCards() {
		t.Logf("card %v", card)
	}

	originDeck := *deck
	deck.Shuffle()

	for _, card := range deck.Cards.GetCards() {
		for _, oriCard := range originDeck.Cards.GetCards() {
			if card != oriCard {
				log.Fatalf("Detect difference %v with %v", card, oriCard)
				continue
			}
		}
	}
}

func TestDeal(t *testing.T) {
	deck := GetDeck()
	deck.Shuffle()
	cards1, err := deck.Deal(13)
	if err != nil {
		t.Errorf("deal1 error %v", err)
	}

	if deck.Dealt != 13 {
		t.Errorf("bad dealt")
	}

	t.Logf("deal1 result %v, dealt %v", cards1.GetCards(), deck.Dealt)

	cards2, err := deck.Deal(13)
	if err != nil {
		t.Errorf("deal2 error %v", err)
	}

	if deck.Dealt != 26 {
		t.Errorf("bad dealt")
	}

	t.Logf("deal2 result %v, dealt %v", cards2.GetCards(), deck.Dealt)

	cards3, err := deck.Deal(13)
	if err != nil {
		t.Errorf("deal3 error %v", err)
	}
	t.Logf("deal3 result %v, dealt %v", cards3.GetCards(), deck.Dealt)

	if deck.Dealt != 39 {
		t.Errorf("bad dealt")
	}

	cards4, err := deck.Deal(13)
	if err != nil {
		t.Errorf("deal4 error %v", err)
	}
	t.Logf("deal4 result %v, dealt %v", cards4.GetCards(), deck.Dealt)

	if deck.Dealt != 52 {
		t.Errorf("bad dealt")
	}
}
