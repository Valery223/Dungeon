package usecase

import "github.com/Valery223/Dungeon/internal/domain"

// PlayerRepository интерфейс для работы с хранилищем игроков
type PlayerRepository interface {
	// Get возвращает игрока по ID или nil, если игрок не найден
	Get(id int) *domain.Player
	// Save сохраняет состояние игрока, если игрок новый, добавляем его ID в порядок
	Save(player *domain.Player)
	// GetAllOrdered возвращает всех игроков в порядке их первого появления
	GetAllOrdered() []*domain.Player
}
