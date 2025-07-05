package game

// GameState represents the overall state of the game (e.g., Peace, War).
const (
	StateWaitingForPlayers  = "waiting_for_players"
	StateOpeningRound       = "opening_round"
	StateMainPhasePlayCards = "main_phase_play_cards"
	StateInProgress        = "in_progress"
	StatePeace             = "peace"
	StateWar               = "war"
	StateFinalStrike       = "final_strike"
	StateFinished          = "finished"
)

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
	Hand       []*Card   `json:"-"` // Hide hand from other players in JSON response
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
type Game struct {
	ID                 string             `json:"id"`
	Players            map[string]*Player `json:"players"`
	PlayerOrder        []string           `json:"player_order"`
	CurrentPlayerIndex int                `json:"current_player_index"`
	Deck               []*Card            `json:"-"` // Deck is not sent to clients
	PopulationDeck     []*Card            `json:"-"` // Population deck is not sent to clients
	DiscardPile        []*Card            `json:"-"` // Discard pile is not sent
	PopulationBank     int64              `json:"-"` // Population bank is not sent
	State              string             `json:"state"`
	Winner             *Player            `json:"winner,omitempty"`
}
