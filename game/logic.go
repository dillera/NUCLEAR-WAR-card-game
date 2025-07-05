package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// NewGame creates and initializes a new game.
func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	popDeck := createPopulationDeck()
	nuclearDeck := createNuclearWarDeck()

	ShuffleCards(popDeck)
	ShuffleCards(nuclearDeck)

	gameID := uuid.New().String()

	// Calculate total population value for the bank
	var totalPopulation int64
	for _, card := range popDeck {
		totalPopulation += card.Value
	}

	return &Game{
		ID:                 gameID,
		Players:            make(map[string]*Player),
		PlayerOrder:        make([]string, 0, 6),
		CurrentPlayerIndex: 0,
		Deck:               nuclearDeck,
		PopulationDeck:     popDeck,
		DiscardPile:        make([]*Card, 0),
		PopulationBank:     totalPopulation,
		State:              StateWaitingForPlayers,
	}
}

// AddPlayer adds a new player to the game.
func (g *Game) AddPlayer(name string) (*Player, error) {
	if g.State != StateWaitingForPlayers {
		return nil, fmt.Errorf("cannot add players, game has already started")
	}
	if len(g.Players) >= 6 {
		return nil, fmt.Errorf("cannot add more than 6 players")
	}

	playerID := uuid.New().String()
	player := &Player{
		ID:         playerID,
		Name:       name,
		Population: 0,
		Hand:       make([]*Card, 0),
		Placemat:   Placemat{},
		IsActive:   true,
	}

	g.Players[playerID] = player
	g.PlayerOrder = append(g.PlayerOrder, playerID)

	return player, nil
}

// StartGame begins the game, dealing cards to players.
func (g *Game) StartGame() error {
	if g.State != StateWaitingForPlayers {
		return fmt.Errorf("game has already started")
	}
	if len(g.Players) < 2 {
		return fmt.Errorf("not enough players to start the game (minimum 2)")
	}

	// Determine number of population cards to deal
	numPopCards := 0
	switch len(g.Players) {
	case 2:
		numPopCards = 10
	case 3:
		numPopCards = 6
	case 4:
		numPopCards = 5
	case 5:
		numPopCards = 4
	case 6:
		numPopCards = 3
	}

	// Deal population cards
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		for i := 0; i < numPopCards; i++ {
			if len(g.PopulationDeck) > 0 {
				card := g.PopulationDeck[0]
				g.PopulationDeck = g.PopulationDeck[1:]
				player.Population += card.Value
				g.PopulationBank -= card.Value
			} else {
				return fmt.Errorf("not enough population cards in the deck to deal")
			}
		}
	}

	// Separate secret cards from the main deck.
	secretCards := make([]*Card, 0)
	nonSecretCards := make([]*Card, 0)
	for _, card := range g.Deck {
		if card.Type == "Secret" {
			secretCards = append(secretCards, card)
		} else {
			nonSecretCards = append(nonSecretCards, card)
		}
	}

	// Shuffle both decks
	rand.Shuffle(len(secretCards), func(i, j int) { secretCards[i], secretCards[j] = secretCards[j], secretCards[i] })
	rand.Shuffle(len(nonSecretCards), func(i, j int) { nonSecretCards[i], nonSecretCards[j] = nonSecretCards[j], nonSecretCards[i] })

	// Reassemble the main deck for drawing
	g.Deck = nonSecretCards

	// For a 2-player game, each player gets 9 cards total.
	numCardsToDeal := 9

	// Deal one secret card and the rest from the main deck to each player.
	for i, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		player.Hand = make([]*Card, 0)

		// Deal one secret card
		if i < len(secretCards) {
			player.Hand = append(player.Hand, secretCards[i])
		} else {
			// Handle case where there aren't enough secret cards (shouldn't happen with standard deck)
			fmt.Println("Warning: Not enough secret cards for all players.")
			player.Hand = append(player.Hand, g.drawCard()) // Draw a regular card instead
		}

		// Deal the remaining cards
		for j := 1; j < numCardsToDeal; j++ {
			player.Hand = append(player.Hand, g.drawCard())
		}
	}

	g.State = StateOpeningRound
	return nil
}

// drawCard removes and returns the top card from the deck.
func (g *Game) drawCard() *Card {
	if len(g.Deck) == 0 {
		// Reshuffle discard pile into deck if needed
		g.Deck = g.DiscardPile
		g.DiscardPile = make([]*Card, 0)
		ShuffleCards(g.Deck)
	}

	card := g.Deck[0]
	g.Deck = g.Deck[1:]
	return card
}
