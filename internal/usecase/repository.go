package usecase

import "github.com/Valery223/Dungeon/internal/domain"

// PlayerRepository interface for working with player storage
type PlayerRepository interface {
	// Get returns a player by ID or nil if player not found
	Get(playerID int) *domain.Player
	// Save saves player state, if player is new, add their ID to the order
	Save(player *domain.Player)
	// GetAllOrdered returns all players in the order of their first appearance
	GetAllOrdered() []*domain.Player
}
