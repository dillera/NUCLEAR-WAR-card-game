package game

import (
	"testing"
)

func TestAdvanceTurn(t *testing.T) {
	t.Run("advances to the next player in order", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		_, _ = g.AddPlayer("Player 2")

		g.CurrentPlayerIndex = 0

		g.AdvanceTurn()

		if g.CurrentPlayerIndex != 1 {
			t.Errorf("expected CurrentPlayerIndex to be 1, got %d", g.CurrentPlayerIndex)
		}

		activePlayerID := g.PlayerOrder[g.CurrentPlayerIndex]
		if g.Players[activePlayerID].ID != p1.ID {
			// This check is a bit tricky because player order is randomized.
			// A better test would be to check if the index advanced correctly.
		}
	})

	t.Run("wraps around to the first player after the last player", func(t *testing.T) {
		g := NewGame()
		_, _ = g.AddPlayer("Player 1")
		_, _ = g.AddPlayer("Player 2")

		g.CurrentPlayerIndex = len(g.PlayerOrder) - 1

		g.AdvanceTurn()

		if g.CurrentPlayerIndex != 0 {
			t.Errorf("expected CurrentPlayerIndex to be 0, got %d", g.CurrentPlayerIndex)
		}
	})
}

func TestAdvanceTurn_SkipsEliminatedPlayer(t *testing.T) {
	g := NewGame()
	p1, _ := g.AddPlayer("Player 1")
	p2, _ := g.AddPlayer("Player 2")
	p3, _ := g.AddPlayer("Player 3")

	// Manually set player order for predictable testing
	g.PlayerOrder = []string{p1.ID, p2.ID, p3.ID}

	// Eliminate the second player
	p2.IsEliminated = true

	// Set turn to Player 1 (index 0)
	g.CurrentPlayerIndex = 0

	// Advance the turn
	g.AdvanceTurn()

	// Expect the turn to skip Player 2 and go to Player 3 (index 2)
	if g.CurrentPlayerIndex != 2 {
		t.Errorf("expected CurrentPlayerIndex to be 2, got %d", g.CurrentPlayerIndex)
	}

	// Advance the turn again
	g.AdvanceTurn()

	// Expect the turn to wrap around to Player 1 (index 0)
	if g.CurrentPlayerIndex != 0 { // Should wrap back to Player 1
		t.Errorf("expected current player to be 0 after wrap-around, but got %d", g.CurrentPlayerIndex)
	}
}

func TestPassTurn(t *testing.T) {
	t.Run("advances to next player", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		p2, _ := g.AddPlayer("Player 2")
		g.PlayerOrder = []string{p1.ID, p2.ID}
		g.CurrentPlayerIndex = 0
		g.State = StateInProgress

		err := g.PassTurn(p1.ID)
		if err != nil {
			t.Fatalf("PassTurn failed unexpectedly: %v", err)
		}

		if g.CurrentPlayerIndex != 1 {
			t.Errorf("expected current player to be 1, but got %d", g.CurrentPlayerIndex)
		}
	})

	t.Run("wraps around to first player", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		p2, _ := g.AddPlayer("Player 2")
		g.PlayerOrder = []string{p1.ID, p2.ID}
		g.CurrentPlayerIndex = 1 // Last player's turn
		g.State = StateInProgress

		err := g.PassTurn(p2.ID)
		if err != nil {
			t.Fatalf("PassTurn failed unexpectedly: %v", err)
		}

		if g.CurrentPlayerIndex != 0 {
			t.Errorf("expected current player to be 0 after wrap-around, but got %d", g.CurrentPlayerIndex)
		}
	})

	t.Run("skips eliminated player", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		p2, _ := g.AddPlayer("Player 2")
		p3, _ := g.AddPlayer("Player 3")
		p2.IsEliminated = true
		g.PlayerOrder = []string{p1.ID, p2.ID, p3.ID}
		g.CurrentPlayerIndex = 0
		g.State = StateInProgress

		err := g.PassTurn(p1.ID)
		if err != nil {
			t.Fatalf("PassTurn failed unexpectedly: %v", err)
		}

		if g.CurrentPlayerIndex != 2 { // Should skip p2 and go to p3
			t.Errorf("expected current player to be 2, but got %d", g.CurrentPlayerIndex)
		}
	})

	t.Run("fails if not player's turn", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		p2, _ := g.AddPlayer("Player 2")
		g.PlayerOrder = []string{p1.ID, p2.ID}
		g.CurrentPlayerIndex = 0
		g.State = StateInProgress

		err := g.PassTurn(p2.ID) // p2 tries to pass on p1's turn
		if err == nil {
			t.Error("expected an error when a player passes out of turn, but got nil")
		}
	})
}
