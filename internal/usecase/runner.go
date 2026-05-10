package usecase

import (
	"io"
)

type GameRunner struct {
	reader    EventReader
	writer    EventWriter
	processor *EventProcessor
}

func NewGameRunner(r EventReader, w EventWriter, p *EventProcessor) *GameRunner {
	return &GameRunner{
		reader:    r,
		writer:    w,
		processor: p,
	}
}

// Run starts the main processing loop
func (r *GameRunner) Run() error {
	for {
		// 1 Read event
		inEvent, err := r.reader.Read()
		if err != nil {
			if err == io.EOF {
				break // Stream completed
			}
			return err // Read error
		}

		// 2 Business logic processes the event
		result := r.processor.ProcessEvent(inEvent)

		// 3 Output results
		// If action was accepted, output it
		if result.IsAccepted {
			err := r.writer.WriteAccepted(inEvent)
			if err != nil {
				return err
			}
		}
		// If there is an outgoing event, output it
		if result.OutgoingEvent != nil {
			err := r.writer.WriteOutgoing(*result.OutgoingEvent)
			if err != nil {
				return err
			}
		}
	}

	// 4 Stream completed. Generate and write final report
	reports := r.processor.GenerateFinalReport()
	return r.writer.WriteReport(reports)
}
