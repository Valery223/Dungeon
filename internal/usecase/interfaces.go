package usecase

import "github.com/Valery223/Dungeon/internal/domain"

// EventReader интерфейс для получения входящих событий
type EventReader interface {
	// Read возвращает следующее событие, если поток завершен, возвращает io.EOF
	Read() (domain.IncomingEvent, error)
}

// EventWriter интерфейс для вывода результатов
type EventWriter interface {
	// WriteAccepted выводит подтверждение успешного действия
	// например: "Player [1] registered"
	WriteAccepted(e domain.IncomingEvent) error

	// WriteOutgoing выводит системные события
	// например  смерть, дисквалификация, невозможный ход
	WriteOutgoing(e domain.OutgoingEvent) error

	// WriteReport выводит финальный отчет по всем игрокам
	WriteReport(reports []ReportEntry) error
}
