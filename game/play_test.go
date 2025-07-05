package game

import (
	"testing"
)

func TestPlayCard(t *testing.T) {
	t.Run("successfully plays a card", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		g.PlayerOrder = []string{p1.ID}
		g.CurrentPlayerIndex = 0
		g.State = StateInProgress

		// Give the player a card
		cardToPlay := &Card{ID: "c1", Name: "Test Card", Type: TypePropaganda}
		p1.Hand = []*Card{cardToPlay}

		err := g.PlayCard(p1.ID, cardToPlay.ID, "face_up")
		if err != nil {
			t.Fatalf("PlayCard failed unexpectedly: %v", err)
		}

		// Card should be removed from hand
		if len(p1.Hand) != 0 {
			t.Errorf("expected hand size to be 0, but got %d", len(p1.Hand))
		}

		// Card should be on the placemat
		if len(p1.Placemat.ActiveCards) != 1 || p1.Placemat.ActiveCards[0].ID != cardToPlay.ID {
			t.Errorf("card was not correctly moved to placemat")
		}
	})

	t.Run("fails if card not in hand", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		g.PlayerOrder = []string{p1.ID}
		g.CurrentPlayerIndex = 0
		g.State = StateInProgress

		err := g.PlayCard(p1.ID, "non-existent-card", "face_up")
		if err == nil {
			t.Error("expected an error when playing a card not in hand, but got nil")
		}
	})

	t.Run("fails if not player's turn", func(t *testing.T) {
		g := NewGame()
		p1, _ := g.AddPlayer("Player 1")
		p2, _ := g.AddPlayer("Player 2")
		g.PlayerOrder = []string{p1.ID, p2.ID}
		g.CurrentPlayerIndex = 0 // p1's turn
		g.State = StateInProgress

		// Give p2 a card
		cardToPlay := &Card{ID: "c1", Name: "Test Card", Type: TypePropaganda}
		p2.Hand = []*Card{cardToPlay}

		err := g.PlayCard(p2.ID, cardToPlay.ID, "face_up") // p2 tries to play on p1's turn
		if err == nil {
			t.Error("expected an error when playing out of turn, but got nil")
		}
	})
}
