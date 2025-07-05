package game

import (
	"testing"
)

func TestAttack_SuccessfulHit(t *testing.T) {
	g := NewGame()
	p1, _ := g.AddPlayer("Player 1")
	p2, _ := g.AddPlayer("Player 2")
	p2.Population = 25000000 // Give target an initial population

	// Manually set player order and current turn
	g.PlayerOrder = []string{p1.ID, p2.ID}
	g.CurrentPlayerIndex = 0
	g.State = StateInProgress

	// Give Player 1 an attack card combo
	deliveryCard := &Card{ID: "d1", Type: TypeDeliverySystem, Name: "B-52"}
	warheadCard := &Card{ID: "w1", Type: TypeWarhead, Name: "10 Megaton", WarheadSize: 10}
	p1.Placemat.ActiveCards = []*Card{deliveryCard, warheadCard}

	initialPopulation := p2.Population

	err := g.Attack(p1.ID, p2.ID)
	if err != nil {
		t.Fatalf("Attack failed unexpectedly: %v", err)
	}

	expectedDamage := int64(warheadCard.WarheadSize) * 1000000
	expectedPopulation := initialPopulation - expectedDamage

	if p2.Population != expectedPopulation {
		t.Errorf("expected player 2 population to be %d, got %d", expectedPopulation, p2.Population)
	}

	// Check that the active cards were discarded
	if len(p1.Placemat.ActiveCards) != 0 {
		t.Errorf("expected active cards to be discarded, but found %d", len(p1.Placemat.ActiveCards))
	}
}

func TestAttack_DefendedByAntiMissile(t *testing.T) {
	g := NewGame()
	p1, _ := g.AddPlayer("Player 1")
	p2, _ := g.AddPlayer("Player 2")
	p2.Population = 25000000

	g.PlayerOrder = []string{p1.ID, p2.ID}
	g.CurrentPlayerIndex = 0
	g.State = StateInProgress

	// Attacker's cards
	deliveryCard := &Card{ID: "d1", Type: TypeDeliverySystem, Name: "B-52"}
	warheadCard := &Card{ID: "w1", Type: TypeWarhead, Name: "10 Megaton", WarheadSize: 10}
	p1.Placemat.ActiveCards = []*Card{deliveryCard, warheadCard}

	// Defender's card
	antiMissileCard := &Card{ID: "am1", Type: TypeAntiMissile, Name: "Anti-Missile"}
	p2.Placemat.ActiveCards = []*Card{antiMissileCard}

	initialPopulation := p2.Population

	err := g.Attack(p1.ID, p2.ID)
	if err != nil {
		t.Fatalf("Attack failed unexpectedly: %v", err)
	}

	// Population should be unchanged
	if p2.Population != initialPopulation {
		t.Errorf("expected population to be unchanged, but it changed from %d to %d", initialPopulation, p2.Population)
	}

	// Attacker's cards should be discarded
	if len(p1.Placemat.ActiveCards) != 0 {
		t.Errorf("expected attacker's cards to be discarded, but found %d", len(p1.Placemat.ActiveCards))
	}

	// Defender's anti-missile should be discarded
	if len(p2.Placemat.ActiveCards) != 0 {
		t.Errorf("expected defender's anti-missile to be discarded, but found %d", len(p2.Placemat.ActiveCards))
	}
}

func TestAttack_PlayerEliminationAndFinalStrike(t *testing.T) {
	g := NewGame()
	p1, _ := g.AddPlayer("Player 1")
	p2, _ := g.AddPlayer("Player 2")
	p2.Population = 5000000 // Low population to ensure elimination

	// Manually set player order for predictable testing
	g.PlayerOrder = []string{p1.ID, p2.ID}
	g.CurrentPlayerIndex = 0 // Player 1's turn
	g.State = StateInProgress

	// Give Player 1 a powerful attack
	deliveryCard := &Card{ID: "d1", Type: TypeDeliverySystem, Name: "Titan"}
	warheadCard := &Card{ID: "w1", Type: TypeWarhead, Name: "25 Megaton", WarheadSize: 25}
	p1.Placemat.ActiveCards = []*Card{deliveryCard, warheadCard}

	err := g.Attack(p1.ID, p2.ID)
	if err != nil {
		t.Fatalf("Attack failed unexpectedly: %v", err)
	}

	// 1. Check if target is eliminated
	if !p2.IsEliminated {
		t.Errorf("expected player 2 to be eliminated, but they were not")
	}
	if p2.Population != 0 {
		t.Errorf("expected player 2 population to be 0, but got %d", p2.Population)
	}

	// 2. Check for Final Strike state
	if g.State != StateFinalStrike {
		t.Errorf("expected game state to be '%s', but got '%s'", StateFinalStrike, g.State)
	}

	// 3. Check if it's the eliminated player's turn
	if g.CurrentPlayerIndex != 1 { // Player 2's index
		t.Errorf("expected CurrentPlayerIndex to be 1 for Final Strike, but got %d", g.CurrentPlayerIndex)
	}
}
