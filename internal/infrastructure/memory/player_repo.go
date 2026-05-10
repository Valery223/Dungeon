package memory

import "github.com/Valery223/Dungeon/internal/domain"

// InMemoryPlayerRepo реализация PlayerRepository
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

// Get возвращает игрока по ID или nil, если игрок не найден
func (r *InMemoryPlayerRepo) Get(id int) *domain.Player {
	return r.players[id]
}

// Save сохраняет состояние игрока
// Если игрок новый, добавляем его ID в порядок
func (r *InMemoryPlayerRepo) Save(player *domain.Player) {
	if _, exists := r.players[player.ID]; !exists {
		r.order = append(r.order, player.ID)
	}
	r.players[player.ID] = player
}

// GetAllOrdered возвращает всех игроков в порядке их первого появления
func (r *InMemoryPlayerRepo) GetAllOrdered() []*domain.Player {
	result := make([]*domain.Player, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.players[id])
	}
	return result
}
