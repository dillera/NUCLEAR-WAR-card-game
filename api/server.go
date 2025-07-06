package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nuclear-war-game-server/game"
	"sync"

	"github.com/gorilla/mux"
)

// Server holds the state of the API server, including all active games.
type Server struct {
	games  map[string]*game.Game
	mu     sync.Mutex
	router *mux.Router
}

// NewServer creates a new API server instance.
func NewServer() *Server {
	s := &Server{
		games:  make(map[string]*game.Game),
		router: mux.NewRouter(),
	}
	s.routes()
	return s
}

// getGameFromRequest retrieves the game from the request URL.
func (s *Server) getGameFromRequest(r *http.Request) (*game.Game, error) {
	vars := mux.Vars(r)
	gameID := vars["gameID"]

	s.mu.Lock()
	g, ok := s.games[gameID]
	s.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("game not found")
	}

	return g, nil
}

func (s *Server) createGameHandler(w http.ResponseWriter, r *http.Request) {

	newGame := game.NewGame()

	s.mu.Lock()
	s.games[newGame.ID] = newGame
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGame)
}

// JoinGameRequest is the expected body for a join game request.
type JoinGameRequest struct {
	PlayerName string `json:"playerName"`
}

// PlayCardRequest is the expected body for a play card request.
type PlayCardRequest struct {
	PlayerID string `json:"playerID"`
	CardID   string `json:"cardID"`
	Location string `json:"location"` // e.g., "face_up", "face_down_1", "deterrent_1"
}

// GameViewForPlayer is a custom view of the game state for a specific player,
// including their hand.
type GameViewForPlayer struct {
	*game.Game
	PlayerHand []*game.Card `json:"player_hand,omitempty"`
}

func (s *Server) gameHandler(w http.ResponseWriter, r *http.Request) {
	g, err := s.getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	s.getGameStateHandler(w, r, g)
}

func (s *Server) getGameStateHandler(w http.ResponseWriter, r *http.Request, g *game.Game) {
	playerID := r.URL.Query().Get("playerID")

	w.Header().Set("Content-Type", "application/json")

	// If a playerID is provided, return a view specific to that player
	if playerID != "" {
		if _, ok := g.Players[playerID]; !ok {
			http.Error(w, "Player not found in this game", http.StatusNotFound)
			return
		}
		playerView := g.NewPlayerView(playerID)

		json.NewEncoder(w).Encode(playerView)
		return
	}

	// Otherwise, return the general game state
	json.NewEncoder(w).Encode(g)
}

func (s *Server) joinGameHandler(w http.ResponseWriter, r *http.Request) {
	// Log the request body for debugging
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	r.Body.Close() //  must close
	// And now set a new body, which will simulate the same data we read:
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var joinReq JoinGameRequest
	if err := json.Unmarshal(bodyBytes, &joinReq); err == nil {
		log.Printf("Received join request from player '%s'", joinReq.PlayerName)
	}

	g, err := s.getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req JoinGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerName == "" {
		http.Error(w, "Player name must be provided", http.StatusBadRequest)
		return
	}

	if g.IsFull() {
		http.Error(w, "Game is full", http.StatusConflict)
		return
	}

	if g.HasStarted() {
		http.Error(w, "Game has already started", http.StatusConflict)
		return
	}

	player, err := g.AddPlayer(req.PlayerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Player '%s' successfully joined game %s", req.PlayerName, g.ID)

	log.Printf("Game %s: Player '%s' joined (total players: %d)", g.ID, req.PlayerName, len(g.Players))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(player)
}

func (s *Server) startGameHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("START GAME: Handler called for URL: %s", r.URL.Path)
	
	g, err := s.getGameFromRequest(r)
	if err != nil {
		log.Printf("START GAME ERROR: Failed to get game: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("START GAME: Found game %s with %d players in state: %s", g.ID, len(g.Players), g.State)
	
	if r.Method != http.MethodPost {
		log.Printf("START GAME ERROR: Invalid method: %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("START GAME: Calling g.StartGame() for game %s", g.ID)
	if err := g.StartGame(); err != nil {
		log.Printf("START GAME ERROR: Failed to start game: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("START GAME: Game %s successfully started! New state: %s", g.ID, g.State)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

// AttackRequest defines the expected body for an attack request.
type AttackRequest struct {
	AttackerID string `json:"attackerID"`
	TargetID   string `json:"targetID"`
}

func (s *Server) attackHandler(w http.ResponseWriter, r *http.Request) {
	g, err := s.getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req AttackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := g.Attack(req.AttackerID, req.TargetID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(g)
}

func (s *Server) passHandler(w http.ResponseWriter, r *http.Request) {
	game, err := s.getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var reqBody struct {
		PlayerID string `json:"playerID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := game.PassTurn(reqBody.PlayerID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(game)
}

func (s *Server) playCardHandler(w http.ResponseWriter, r *http.Request) {
	g, err := s.getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req PlayCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := g.PlayCard(req.PlayerID, req.CardID, req.Location); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

// Start runs the HTTP server.
func (s *Server) routes() {
	s.router.HandleFunc("/games", s.createGameHandler).Methods("POST")
	s.router.HandleFunc("/games/{gameID}", s.gameHandler).Methods("GET")
	s.router.HandleFunc("/games/{gameID}/join", s.joinGameHandler).Methods("POST")
	s.router.HandleFunc("/games/{gameID}/start", s.startGameHandler).Methods("POST")
	s.router.HandleFunc("/games/{gameID}/play", s.playCardHandler).Methods("POST")
	s.router.HandleFunc("/games/{gameID}/attack", s.attackHandler).Methods("POST")
	s.router.HandleFunc("/games/{gameID}/pass", s.passHandler).Methods("POST")
}

// Start runs the HTTP server.
func (s *Server) Start() {
	fmt.Println("Nuclear War server listening on port 8080...")
	if err := http.ListenAndServe(":8080", s.router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
