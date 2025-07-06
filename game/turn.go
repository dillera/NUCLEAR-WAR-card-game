package game

import "fmt"

// PlayCard handles a player's action to play a card.
func (g *Game) PlayCard(playerID, cardID, location string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. Find the player
	player, ok := g.Players[playerID]
	if !ok {
		return fmt.Errorf("player with ID %s not found", playerID)
	}

	// 2. Find the card in the player's hand
	var cardToPlay *Card
	cardIndex := -1
	for i, card := range player.Hand {
		if card.ID == cardID {
			cardToPlay = card
			cardIndex = i
			break
		}
	}

	if cardToPlay == nil {
		return fmt.Errorf("card with ID %s not found in player %s's hand", cardID, player.Name)
	}

	// 3. Enforce game state rules
	if g.State == "opening_round" {
		if cardToPlay.Type != "Secret" {
			return fmt.Errorf("only secret cards can be played during the opening round")
		}
		if location != "face_down_1" {
			return fmt.Errorf("secret cards must be played to the 'face_down_1' location during the opening round")
		}
	} else {
		// Check if it's the player's turn
		if g.PlayerOrder[g.CurrentPlayerIndex] != playerID {
			return fmt.Errorf("it is not player %s's turn", g.Players[g.PlayerOrder[g.CurrentPlayerIndex]].Name)
		}
	}

	// 4. Place the card on the placemat
	switch location {
	case "face_up":
		if g.State == "opening_round" {
			return fmt.Errorf("cannot play cards face up during the opening round")
		}
		player.Placemat.ActiveCards = append(player.Placemat.ActiveCards, cardToPlay)
	case "face_down_1":
		if player.Placemat.FaceDownCard1 != nil {
			return fmt.Errorf("face-down card slot 1 is already occupied")
		}
		player.Placemat.FaceDownCard1 = cardToPlay
	case "face_down_2":
		if player.Placemat.FaceDownCard2 != nil {
			return fmt.Errorf("face-down card slot 2 is already occupied")
		}
		player.Placemat.FaceDownCard2 = cardToPlay
	case "deterrent_1":
		if player.Placemat.Deterrent1 != nil {
			return fmt.Errorf("deterrent slot 1 is already occupied")
		}
		// TODO: Check if card is a valid deterrent
		player.Placemat.Deterrent1 = cardToPlay
	case "deterrent_2":
		if player.Placemat.Deterrent2 != nil {
			return fmt.Errorf("deterrent slot 2 is already occupied")
		}
		// TODO: Check if card is a valid deterrent
		player.Placemat.Deterrent2 = cardToPlay
	default:
		return fmt.Errorf("invalid card location: %s", location)
	}

	// 5. Remove the card from the player's hand
	player.Hand = append(player.Hand[:cardIndex], player.Hand[cardIndex+1:]...)

	fmt.Printf("Player %s played card '%s' to location %s\n", player.Name, cardToPlay.Name, location)

	// 6. Check if the game state should advance
	if g.State == "opening_round" {
		allPlayed := true
		for _, p := range g.Players {
			if p.Placemat.FaceDownCard1 == nil {
				allPlayed = false
				break
			}
		}
		if allPlayed {
			fmt.Println("All players have played their secret cards. Resolving opening secrets.")
			g.ResolveOpeningSecrets()
			g.State = StateInProgress
		}
	} else if g.State == StateInProgress {
		return nil
	}

	return nil
}

// PassTurn allows the current player to pass their turn.
func (g *Game) PassTurn(playerID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if it's the player's turn.
	if g.PlayerOrder[g.CurrentPlayerIndex] != playerID {
		return fmt.Errorf("it is not player %s's turn", playerID)
	}

	g.AdvanceTurn()
	return nil
}

// AdvanceTurn moves to the next active player, skipping those who are eliminated.
// This is an internal function and assumes a lock is already held.
func (g *Game) AdvanceTurn() {
	// Loop through players to find the next non-eliminated one.
	for i := 0; i < len(g.PlayerOrder); i++ {
		g.CurrentPlayerIndex = (g.CurrentPlayerIndex + 1) % len(g.PlayerOrder)
		nextPlayerID := g.PlayerOrder[g.CurrentPlayerIndex]
		if !g.Players[nextPlayerID].IsEliminated {
			fmt.Printf("--- Turn advanced. It is now %s's turn. ---\n", g.Players[nextPlayerID].Name)
			return
		}
	}
	// This case should ideally not be reached if a winner is declared correctly.
	fmt.Println("No active players left to advance turn to.")
}

// ResolveOpeningSecrets handles the simultaneous reveal of secret cards.
// This is an internal function and assumes a lock is already held.
func (g *Game) ResolveOpeningSecrets() {
	fmt.Println("--- Resolving Opening Secrets ---")
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		if player.Placemat.FaceDownCard1 != nil {
			// Reveal the card, log its effect, and then discard it.
			revealedCard := player.Placemat.FaceDownCard1
			player.Placemat.FaceDownCard1 = nil

			fmt.Printf("Player %s revealed secret: %s (%s)\n", player.Name, revealedCard.Name, revealedCard.Description)
			// TODO: Implement secret card effects based on revealedCard.ID or Name

			// By not placing the revealedCard on the FaceUpCard slot, we consider it resolved and discarded,
			// freeing the placemat for the main gameplay phase.
		}
	}
	fmt.Println("--- Finished Resolving Secrets ---")
}
