package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	g := NewGame()

	if g.ID == "" {
		t.Error("expected a new game to have a non-empty ID")
	}

	if g.State != StateWaitingForPlayers {
		t.Errorf("expected game state to be '%s', but got '%s'", StateWaitingForPlayers, g.State)
	}

	if g.Players == nil {
		t.Error("expected players map to be initialized, but it was nil")
	}

	if len(g.Players) != 0 {
		t.Errorf("expected a new game to have 0 players, but got %d", len(g.Players))
	}

	if g.Deck == nil {
		t.Error("expected deck to be initialized, but it was nil")
	}

	// A full deck has a specific number of cards. Let's check against that.
	// This number might change if we add expansions, but it's a good check for now.
	expectedDeckSize := 100 // Based on the standard card list
	if len(g.Deck) != expectedDeckSize {
		t.Errorf("expected deck size to be %d, but got %d", expectedDeckSize, len(g.Deck))
	}
}

func TestAddPlayer(t *testing.T) {
	t.Run("successfully adds a player", func(t *testing.T) {
		g := NewGame()
		playerName := "Player 1"

		p, err := g.AddPlayer(playerName)
		if err != nil {
			t.Fatalf("AddPlayer failed unexpectedly: %v", err)
		}

		if len(g.Players) != 1 {
			t.Errorf("expected player count to be 1, got %d", len(g.Players))
		}

		if g.Players[p.ID] == nil {
			t.Errorf("player was not added to the Players map")
		}

		if len(g.PlayerOrder) != 1 || g.PlayerOrder[0] != p.ID {
			t.Errorf("player was not added to the PlayerOrder slice")
		}

		if p.Name != playerName {
			t.Errorf("expected player name to be '%s', got '%s'", playerName, p.Name)
		}
	})

	t.Run("returns an error for empty player name", func(t *testing.T) {
		g := NewGame()
		_, err := g.AddPlayer("")
		if err == nil {
			t.Error("expected an error when adding a player with an empty name, but got nil")
		}
	})
}

func TestStartGame(t *testing.T) {
	t.Run("successfully starts a game", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		p2, _ := g.AddPlayer("Player 2")

		err := g.StartGame()
		if err != nil {
			t.Fatalf("StartGame failed unexpectedly: %v", err)
		}

		if g.State != StateOpeningRound {
			t.Errorf("expected game state to be '%s', but got '%s'", StateOpeningRound, g.State)
		}

		// Check if players received population
		if p1.Population == 0 || p2.Population == 0 {
			t.Error("players should have received population, but they have 0")
		}

		// Check hand size and for one secret card
		for _, p := range g.Players {
			if len(p.Hand) != 9 { // 8 regular cards + 1 secret
				t.Errorf("expected hand size of 9 for player %s, but got %d", p.Name, len(p.Hand))
			}

			secretCardCount := 0
			for _, card := range p.Hand {
				if card.Type == TypeSecret {
					secretCardCount++
				}
			}
			if secretCardCount != 1 {
				t.Errorf("expected 1 secret card for player %s, but got %d", p.Name, secretCardCount)
			}
		}
	})

	t.Run("fails to start with less than 2 players", func(t *testing.T) {
		g := NewGame()
		_, _ = g.AddPlayer("Player 1")

		err := g.StartGame()
		if err == nil {
			t.Error("expected an error when starting a game with less than 2 players, but got nil")
		}
	})
}
