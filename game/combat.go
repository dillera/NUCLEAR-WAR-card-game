package game

import (
	"fmt"
)

// Attack handles a player's action to attack another player.
func (g *Game) Attack(attackerID, targetID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. Find attacker and target players.
	attacker, ok := g.Players[attackerID]
	if !ok {
		return fmt.Errorf("attacker with ID %s not found", attackerID)
	}
	target, ok := g.Players[targetID]
	if !ok {
		return fmt.Errorf("target with ID %s not found", targetID)
	}

	isFinalStrike := g.State == StateFinalStrike

	// If it's a final strike, only the current (eliminated) player can attack.
	if isFinalStrike {
		if attacker.ID != g.PlayerOrder[g.CurrentPlayerIndex] {
			return fmt.Errorf("it is player %s's Final Strike, not player %s's", g.Players[g.PlayerOrder[g.CurrentPlayerIndex]].Name, attacker.Name)
		}
	} else {
		// In a normal turn, an eliminated player cannot attack.
		if attacker.IsEliminated {
			return fmt.Errorf("player %s is eliminated and cannot attack", attacker.Name)
		}
		// TODO: Add check to ensure it's the attacker's actual turn.
	}

	// 2. Validate the attack combination on the attacker's placemat.
	var deliverySystem *Card
	var warhead *Card

	for _, card := range attacker.Placemat.ActiveCards {
		if card.Type == TypeDeliverySystem {
			deliverySystem = card
		} else if card.Type == TypeWarhead {
			warhead = card
		}
	}

	if deliverySystem == nil {
		return fmt.Errorf("attacker %s has no delivery system in play", attacker.Name)
	}

	if warhead == nil {
		return fmt.Errorf("attacker %s has no warhead in play", attacker.Name)
	}

	// TODO: Check if the warhead size is compatible with the delivery system's payload.

	fmt.Printf("Player %s is attacking Player %s with a %d megaton warhead on a %s!\n",
		attacker.Name, target.Name, warhead.WarheadSize, deliverySystem.Name)

	// 3. Check for defense.
	var antiMissile *Card
	for _, card := range target.Placemat.ActiveCards {
		if card.Type == TypeAntiMissile {
			antiMissile = card
			break
		}
	}

	// Remove attacker's cards regardless of outcome.
	newAttackerCards := []*Card{}
	for _, card := range attacker.Placemat.ActiveCards {
		if card.ID != deliverySystem.ID && card.ID != warhead.ID {
			newAttackerCards = append(newAttackerCards, card)
		}
	}
	attacker.Placemat.ActiveCards = newAttackerCards

	if antiMissile != nil {
		fmt.Printf("Player %s's attack was intercepted by an Anti-Missile!\n", attacker.Name)
		// Remove the used anti-missile card from the target's placemat.
		newTargetCards := []*Card{}
		for _, card := range target.Placemat.ActiveCards {
			if card.ID != antiMissile.ID {
				newTargetCards = append(newTargetCards, card)
			}
		}
		target.Placemat.ActiveCards = newTargetCards
	} else {
		// 1 megaton = 1 million population
		damage := int64(warhead.WarheadSize) * 1000000
		target.Population -= damage
		fmt.Printf("Attack successful! Player %s loses %d population.\n", target.Name, damage)

		if target.Population <= 0 {
			target.Population = 0
			if !target.IsEliminated {
				target.IsEliminated = true
				fmt.Printf("Player %s has been eliminated!\n", target.Name)

				// Set up the Final Strike
				g.State = StateFinalStrike
				for i, playerID := range g.PlayerOrder {
					if playerID == target.ID {
						g.CurrentPlayerIndex = i
						break
					}
				}
				fmt.Printf("Player %s gets a Final Strike! It is now their turn.\n", target.Name)
				// End the current attack function here. The next action must be an attack from the eliminated player.
				return nil
			}
		}
	}

	// If this was a Final Strike, reset state before checking for winner and advancing turn.
	if isFinalStrike {
		g.State = StateInProgress
		fmt.Printf("Player %s has completed their Final Strike.\n", attacker.Name)
	}

	// 5. Check for a winner.
	g.checkForWinner()

	// 6. An attack is a turn-ending action, but only if the game isn't over.
	if g.State != StateGameOver {
		g.AdvanceTurn()
	}

	return nil
}

// checkForWinner checks if there is only one player left and declares them the winner.
// This is an internal function and assumes a lock is already held.
func (g *Game) checkForWinner() {
	activePlayers := []*Player{}
	for _, player := range g.Players {
		if !player.IsEliminated {
			activePlayers = append(activePlayers, player)
		}
	}

	if len(activePlayers) == 1 {
		g.Winner = activePlayers[0]
		g.State = StateGameOver
		fmt.Printf("Player %s has won the game!\n", g.Winner.Name)
	}
}
