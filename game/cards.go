package game

import (
	"fmt"
	"math/rand"
)

// createPopulationDeck creates the initial 20 population cards.
func createPopulationDeck() []*Card {
	return []*Card{
		{ID: "pop-1m-1", Name: "1 Million", Type: "Population", Value: 1000000},
		{ID: "pop-1m-2", Name: "1 Million", Type: "Population", Value: 1000000},
		{ID: "pop-1m-3", Name: "1 Million", Type: "Population", Value: 1000000},
		{ID: "pop-1m-4", Name: "1 Million", Type: "Population", Value: 1000000},
		{ID: "pop-1m-5", Name: "1 Million", Type: "Population", Value: 1000000},
		{ID: "pop-2m-1", Name: "2 Million", Type: "Population", Value: 2000000},
		{ID: "pop-2m-2", Name: "2 Million", Type: "Population", Value: 2000000},
		{ID: "pop-2m-3", Name: "2 Million", Type: "Population", Value: 2000000},
		{ID: "pop-2m-4", Name: "2 Million", Type: "Population", Value: 2000000},
		{ID: "pop-5m-1", Name: "5 Million", Type: "Population", Value: 5000000},
		{ID: "pop-5m-2", Name: "5 Million", Type: "Population", Value: 5000000},
		{ID: "pop-5m-3", Name: "5 Million", Type: "Population", Value: 5000000},
		{ID: "pop-10m-1", Name: "10 Million", Type: "Population", Value: 10000000},
		{ID: "pop-10m-2", Name: "10 Million", Type: "Population", Value: 10000000},
		{ID: "pop-10m-3", Name: "10 Million", Type: "Population", Value: 10000000},
		{ID: "pop-25m-1", Name: "25 Million", Type: "Population", Value: 25000000},
		{ID: "pop-25m-2", Name: "25 Million", Type: "Population", Value: 25000000},
		{ID: "pop-25m-3", Name: "25 Million", Type: "Population", Value: 25000000},
		{ID: "pop-25m-4", Name: "25 Million", Type: "Population", Value: 25000000},
		{ID: "pop-25m-5", Name: "25 Million", Type: "Population", Value: 25000000},
	}
}

// createNuclearWarDeck creates the 100-card Nuclear War deck.
// Note: This is a sample deck. We can adjust the card distribution later.
func createNuclearWarDeck() []*Card {
	deck := make([]*Card, 0, 100)

	// Propaganda (20 cards)
	for i := 0; i < 20; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("prop-%d", i), Name: "Propaganda", Type: TypePropaganda, Description: "Steal 1 million population from a player."})
	}

	// Delivery Systems (30 cards)
	for i := 0; i < 15; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("missile-%d", i), Name: "ICBM", Type: TypeDeliverySystem, CarryingCapacity: 100, Description: "Carries a warhead up to 100 megatons."})
	}
	for i := 0; i < 15; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("bomber-%d", i), Name: "B-52 Bomber", Type: TypeDeliverySystem, CarryingCapacity: 200, Description: "Carries multiple warheads up to a total of 200 megatons."})
	}

	// Warheads (30 cards)
	for i := 0; i < 10; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("wh-10m-%d", i), Name: "10 Megaton Warhead", Type: TypeWarhead, WarheadSize: 10, Description: "A 10-megaton warhead."})
	}
	for i := 0; i < 10; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("wh-25m-%d", i), Name: "25 Megaton Warhead", Type: TypeWarhead, WarheadSize: 25, Description: "A 25-megaton warhead."})
	}
	for i := 0; i < 10; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("wh-100m-%d", i), Name: "100 Megaton Warhead", Type: TypeWarhead, WarheadSize: 100, Description: "A 100-megaton warhead."})
	}

	// Anti-Missiles (10 cards)
	for i := 0; i < 10; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("anti-missile-%d", i), Name: "Anti-Missile System", Type: TypeAntiMissile, Intercepts: []string{"ICBM"}, Description: "Intercepts ICBMs."})
	}

	// Secrets (10 cards)
	for i := 0; i < 10; i++ {
		deck = append(deck, &Card{ID: fmt.Sprintf("secret-%d", i), Name: "Secret: Spy Network", Type: TypeSecret, Description: "Look at another player's hand."})
	}

	return deck
}

// ShuffleCards shuffles a slice of cards.
func ShuffleCards(cards []*Card) {
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}
