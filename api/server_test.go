package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"testing"

	"nuclear-war-game-server/game"
)

// setupTestServer initializes a new server and a test recorder.
func setupTestServer() (*Server, *httptest.ResponseRecorder) {
	s := NewServer()
	rr := httptest.NewRecorder()
	return s, rr
}

// createGame is a helper function to create a new game via API call for testing.
func createGame(t *testing.T, s *Server) *game.Game {
	req, err := http.NewRequest("POST", "/games", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("createGame handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var g game.Game
	if err := json.NewDecoder(rr.Body).Decode(&g); err != nil {
		t.Fatalf("could not parse response JSON from createGame: %v", err)
	}

	if g.ID == "" {
		t.Fatal("createGame returned a game with an empty ID")
	}

	// The handler adds the game to s.games. We need to retrieve the pointer to the
	// actual game instance from the server's map to check its state directly.
	s.mu.Lock()
	gameInstance, ok := s.games[g.ID]
	s.mu.Unlock()
	if !ok {
		t.Fatalf("game %s not found in server map after creation", g.ID)
	}

	return gameInstance
}

func TestCreateGameHandler(t *testing.T) {
	s, rr := setupTestServer()

	req, err := http.NewRequest("POST", "/games", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.Handler(s.router)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resp game.Game
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response JSON: %v", err)
	}

	if resp.ID == "" {
		t.Errorf("expected a game_id in response, but it was empty")
	}

	if len(s.games) != 1 {
		t.Errorf("expected server to have 1 game, but it has %d", len(s.games))
	}
}

func TestJoinGameHandler(t *testing.T) {
	s, rr := setupTestServer()

	// First, create a game to join
	g := game.NewGame()
	s.games[g.ID] = g

	// Prepare the join request
	joinReq := JoinGameRequest{PlayerName: "Test Player"}
	body, _ := json.Marshal(joinReq)

	// The game ID is part of the URL, not the body
	url := fmt.Sprintf("/games/%s/join", g.ID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	handler := http.Handler(s.router)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp game.Player
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response JSON: %v", err)
	}

	if resp.ID == "" {
		t.Errorf("expected a player_id in response, but it was empty")
	}

	if resp.Name != "Test Player" {
		t.Errorf("expected player name to be 'Test Player', but got '%s'", resp.Name)
	}

	if len(g.Players) != 1 {
		t.Errorf("expected game to have 1 player, but it has %d", len(g.Players))
	}
}

func TestStartGameHandler(t *testing.T) {
	s, rr := setupTestServer()

	// First, create a game and add players
	g := game.NewGame()
	g.AddPlayer("Player 1")
	g.AddPlayer("Player 2")
	s.games[g.ID] = g

	// Prepare the start game request
	url := fmt.Sprintf("/games/%s/start", g.ID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.Handler(s.router)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp game.Game
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response JSON: %v", err)
	}

	if resp.State != game.StateOpeningRound {
		t.Errorf("expected game state to be '%s', but got '%s'", game.StateOpeningRound, resp.State)
	}
}

func TestPlayCardHandler(t *testing.T) {
	s, rr := setupTestServer()

	// Setup game with two players and start it
	g := game.NewGame()
	p1, _ := g.AddPlayer("Player 1")
	g.AddPlayer("Player 2")
	s.games[g.ID] = g
	if err := g.StartGame(); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// Find the secret card in player 1's hand to play
	var secretCard *game.Card
	for _, c := range p1.Hand {
		if c.Type == "Secret" {
			secretCard = c
			break
		}
	}
	if secretCard == nil {
		t.Fatal("player 1 has no secret card to play")
	}

	// Prepare the play card request for the opening round
	playReq := PlayCardRequest{
		PlayerID: p1.ID,
		CardID:   secretCard.ID,
		Location: "face_down_1",
	}
	body, _ := json.Marshal(playReq)

	url := fmt.Sprintf("/games/%s/play", g.ID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	handler := http.Handler(s.router)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		// Log the response body for debugging
		t.Logf("Response body: %s", rr.Body.String())
	}

	// Verify the card was moved from hand to placemat
	updatedPlayer, ok := g.Players[p1.ID]
	if !ok {
		t.Fatal("player not found in game state after playing card")
	}

	cardFoundInHand := false
	for _, c := range updatedPlayer.Hand {
		if c.ID == secretCard.ID {
			cardFoundInHand = true
			break
		}
	}
	if cardFoundInHand {
		t.Errorf("played card is still in player's hand")
	}

	if updatedPlayer.Placemat.FaceDownCard1 == nil || updatedPlayer.Placemat.FaceDownCard1.ID != secretCard.ID {
		t.Errorf("played secret card is not in the correct face_down_1 location on the placemat")
	}
}

func TestPassHandler(t *testing.T) {
	s, rr := setupTestServer()

	// Setup game with two players and start it
	g := game.NewGame()
	p1, _ := g.AddPlayer("Player 1")
	g.AddPlayer("Player 2")
	s.games[g.ID] = g
	if err := g.StartGame(); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// It's P1's turn. Let's have them pass.
	passReq := struct {
		PlayerID string `json:"playerID"`
	}{
		PlayerID: p1.ID,
	}
	body, _ := json.Marshal(passReq)

	url := fmt.Sprintf("/games/%s/pass", g.ID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	handler := http.Handler(s.router)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Response body: %s", rr.Body.String())
	}

	// Verify turn has advanced to player 2
	if g.CurrentPlayerIndex != 1 {
		t.Errorf("expected current player index to be 1, but got %d", g.CurrentPlayerIndex)
	}
}

func TestAttackHandler(t *testing.T) {
	s, rr := setupTestServer()

	// Setup game with two players and start it
	g := game.NewGame()
	p1, _ := g.AddPlayer("Player 1")
	p2, _ := g.AddPlayer("Player 2")
	s.games[g.ID] = g
	if err := g.StartGame(); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// Manually set up the attacker's placemat for a valid attack
	// In a real game, these would be played on previous turns.
	deliveryCard := &game.Card{ID: "card-b52", Name: "B-52 Bomber", Type: "Delivery System"}
	warheadCard := &game.Card{ID: "card-10mt", Name: "10 Megaton Warhead", Type: "Warhead", WarheadSize: 10}
	p1.Placemat.ActiveCards = append(p1.Placemat.ActiveCards, deliveryCard, warheadCard)

	initialTargetPopulation := p2.Population

	// Prepare the attack request
	attackReq := AttackRequest{
		AttackerID: p1.ID,
		TargetID:   p2.ID,
	}
	body, _ := json.Marshal(attackReq)

	url := fmt.Sprintf("/games/%s/attack", g.ID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	handler := http.Handler(s.router)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Response body: %s", rr.Body.String())
	}

	// Verify the target's population was reduced
	updatedTarget, ok := g.Players[p2.ID]
	if !ok {
		t.Fatal("target player not found in game state after attack")
	}

	if updatedTarget.Population >= initialTargetPopulation {
		t.Errorf("expected target population to decrease, but it did not. Initial: %d, Current: %d", initialTargetPopulation, updatedTarget.Population)
	}
}

func TestJoinGameHandler_Errors(t *testing.T) {
	// Helper to perform a join request
	joinGame := func(s *Server, gameID, playerName string) *httptest.ResponseRecorder {
		joinReq := JoinGameRequest{PlayerName: playerName}
		body, _ := json.Marshal(joinReq)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/games/%s/join", gameID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		s.router.ServeHTTP(rr, req)
		return rr
	}

	t.Run("join full game", func(t *testing.T) {
		s, _ := setupTestServer() // Create a fresh server for this test case
		g := createGame(t, s)

		// Fill the game with max players
		for i := 0; i < game.MaxPlayers; i++ {
			playerName := fmt.Sprintf("Player %d", i+1)
			res := joinGame(s, g.ID, playerName)
			if res.Code != http.StatusOK {
				t.Fatalf("expected status 200 for player %d, got %d", i+1, res.Code)
			}
		}

		// Try to join the full game
		res := joinGame(s, g.ID, "Late Player")
		if res.Code != http.StatusConflict {
			t.Errorf("expected status %d for full game, got %d", http.StatusConflict, res.Code)
		}
	})

	t.Run("join started game", func(t *testing.T) {
		s, _ := setupTestServer() // Create a fresh server for this test case
		g := createGame(t, s)

		// Add two players
		joinGame(s, g.ID, "Player 1")
		joinGame(s, g.ID, "Player 2")

		// Start the game
		startReq, _ := http.NewRequest("POST", fmt.Sprintf("/games/%s/start", g.ID), nil)
		startRR := httptest.NewRecorder()
		s.router.ServeHTTP(startRR, startReq)
		if startRR.Code != http.StatusOK {
			t.Fatalf("Failed to start game, status: %d", startRR.Code)
		}

		// Try to join the started game
		res := joinGame(s, g.ID, "Late Player")
		if res.Code != http.StatusConflict {
			t.Errorf("expected status %d for started game, got %d", http.StatusConflict, res.Code)
		}
	})
}

func TestGetGameHandler(t *testing.T) {
	s, _ := setupTestServer()

	// Setup game with two players and start it
	g := game.NewGame()
	p1, _ := g.AddPlayer("Player 1")
	p2, _ := g.AddPlayer("Player 2")
	s.games[g.ID] = g
	if err := g.StartGame(); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	t.Run("full game view", func(t *testing.T) {
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/games/%s", g.ID)
		req, _ := http.NewRequest("GET", url, nil)

		handler := http.Handler(s.router)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var respGame game.Game
		if err := json.Unmarshal(rr.Body.Bytes(), &respGame); err != nil {
			t.Fatalf("could not parse response JSON: %v", err)
		}

		if len(respGame.Players[p2.ID].Hand) == 0 {
			t.Errorf("expected to see other player's hand in full view, but it was empty")
		}
	})

	t.Run("player-specific view", func(t *testing.T) {
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/games/%s?playerID=%s", g.ID, p1.ID)
		req, _ := http.NewRequest("GET", url, nil)

		handler := http.Handler(s.router)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var playerView game.PlayerView
		if err := json.Unmarshal(rr.Body.Bytes(), &playerView); err != nil {
			t.Fatalf("could not parse player view JSON: %v", err)
		}

		if len(playerView.PlayerHand) == 0 {
			t.Errorf("expected to see own hand, but it was empty")
		}

		// Check that other players' hands are hidden but their hand size is visible
		for _, otherPlayer := range playerView.Opponents {
			if otherPlayer.HandSize == 0 {
				t.Errorf("expected opponent's hand size to be visible, but it was 0 for player %s", otherPlayer.Name)
			}
		}
	})
}
