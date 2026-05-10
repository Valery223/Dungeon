package memory

import "github.com/Valery223/Dungeon/internal/domain"

// InMemoryPlayerRepo implementation of PlayerRepository
type InMemoryPlayerRepo struct {
	players map[int]*domain.Player
	order   []int
}

func NewInMemoryPlayerRepo() *InMemoryPlayerRepo {
	return &InMemoryPlayerRepo{
		players: make(map[int]*domain.Player),
		order:   make([]int, 0),
	}
}

// Get returns a player by ID or nil if player not found
func (r *InMemoryPlayerRepo) Get(id int) *domain.Player {
	return r.players[id]
}

// Save saves player state
// If player is new, add their ID to the order
func (r *InMemoryPlayerRepo) Save(player *domain.Player) {
	if _, exists := r.players[player.ID]; !exists {
		r.order = append(r.order, player.ID)
	}
	r.players[player.ID] = player
}

// GetAllOrdered returns all players in order of their first appearance
func (r *InMemoryPlayerRepo) GetAllOrdered() []*domain.Player {
	result := make([]*domain.Player, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.players[id])
	}
	return result
}
