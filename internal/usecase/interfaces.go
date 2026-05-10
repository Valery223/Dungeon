package usecase

import "github.com/Valery223/Dungeon/internal/domain"

// EventReader interface for receiving incoming events
type EventReader interface {
	// Read returns the next event, returns io.EOF if stream is finished
	Read() (domain.IncomingEvent, error)
}

// EventWriter interface for outputting results
type EventWriter interface {
	// WriteAccepted outputs confirmation of successful action
	// example: "Player [1] registered"
	WriteAccepted(e domain.IncomingEvent) error

	// WriteOutgoing outputs system events
	// example: death, disqualification, impossible move
	WriteOutgoing(e domain.OutgoingEvent) error

	// WriteReport outputs final report for all players
	WriteReport(reports []ReportEntry) error
}
