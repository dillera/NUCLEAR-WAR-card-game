package game

import (
	"sync"
)

// MaxPlayers is the maximum number of players allowed in a game.
const MaxPlayers = 6

// CardType defines the type of a card.
const (
	TypePropaganda     = "Propaganda"
	TypeDeliverySystem = "Delivery System"
	TypeWarhead        = "Warhead"
	TypeAntiMissile    = "Anti-Missile"
	TypeSecret         = "Secret"
	TypeTopSecret      = "Top Secret"
)

// Player represents a player in the game.
type Player struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Population int64     `json:"population"`
	Hand       []*Card   `json:"hand"`
	Placemat     Placemat `json:"placemat"`
	IsActive     bool     `json:"is_active"`
	IsEliminated bool     `json:"is_eliminated"`
}

// Placemat holds the cards a player has in play.
type Placemat struct {
	ActiveCards   []*Card `json:"active_cards,omitempty"`
	FaceDownCard1 *Card `json:"-"` // Hidden from other players
	FaceDownCard2 *Card `json:"-"` // Hidden from other players
	Deterrent1    *Card `json:"deterrent_1,omitempty"`
	Deterrent2    *Card `json:"deterrent_2,omitempty"`
}

// PlayerView represents the game state from a single player's perspective,
// hiding information that should not be visible to them (e.g., other players' hands).
// This is used to provide a secure view of the game to each client.
type PlayerView struct {
	GameID             string        `json:"gameID"`
	PlayerName         string        `json:"playerName"`
	PlayerPopulation   int64         `json:"playerPopulation"`
	PlayerHand         []*Card       `json:"playerHand"`
	PlayerPlacemat     *Placemat     `json:"playerPlacemat"`
	Opponents          []Opponent    `json:"opponents"`
	CurrentTurnPlayer  string        `json:"currentTurnPlayer"`
	State              GameState     `json:"state"`
	Winner             *string       `json:"winner,omitempty"`
	TurnLog             []string      `json:"turnLog"`
	CurrentTurnPlayerId string        `json:"currentTurnPlayerId,omitempty"`
	AvailableCommands   []Command     `json:"availableCommands,omitempty"`
}

// Opponent represents a player as seen by another player.
// It redacts sensitive information like the player's hand.
type Opponent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Population  int64     `json:"population"`
	HandSize    int       `json:"handSize"`
	Placemat    *Placemat `json:"placemat"`
	IsEliminated bool      `json:"isEliminated"`
}

// Command represents a single action a player can take.
// This is used to dynamically inform the client about available options.
type Command struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// We can add fields for required parameters here later if needed
	// e.g., Params []string `json:"params"`
}

// Card represents a single card in the game.
type Card struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	Description      string   `json:"description,omitempty"`
	Value            int64    `json:"value,omitempty"`            // For Population cards
	WarheadSize      int      `json:"warhead_size,omitempty"`      // For Warheads
	CarryingCapacity int      `json:"carrying_capacity,omitempty"` // For Delivery Systems
	Intercepts       []string `json:"intercepts,omitempty"`      // For Anti-Missiles
}

// Game represents the state of a single game.
// GameState represents the different states a game can be in.
type GameState string

const (
	StateWaitingForPlayers GameState = "waiting_for_players"
	StateOpeningRound      GameState = "opening_round"
	StateInProgress        GameState = "in_progress"
	StateFinalStrike       GameState = "final_strike"
	StateGameOver          GameState = "game_over" // Replaces "StateFinished"
)

type Game struct {
	ID                 string             `json:"id"`
	Players            map[string]*Player `json:"players"`
	PlayerOrder        []string           `json:"player_order"`
	CurrentPlayerIndex int                `json:"current_player_index"`
	Deck               []*Card            `json:"-"` // Deck is not sent to clients
	PopulationDeck     []*Card            `json:"-"` // Population deck is not sent to clients
	DiscardPile        []*Card            `json:"-"` // Discard pile is not sent
	PopulationBank     int64              `json:"-"` // Population bank is not sent
	State              GameState          `json:"state"`
	Winner             *Player            `json:"winner,omitempty"`
	TurnLog            []string           `json:"turnLog"`
	mu                 sync.RWMutex       `json:"-"` // Mutex to protect game state
}
