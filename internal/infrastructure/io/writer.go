package io

import (
	"fmt"
	"io"

	"github.com/Valery223/Dungeon/internal/domain"
	"github.com/Valery223/Dungeon/internal/usecase"
)

// StreamWriter outputs results to stream
type StreamWriter struct {
	out io.Writer
}

// NewStreamWriter creates a new StreamWriter for the given io.Writer
func NewStreamWriter(w io.Writer) *StreamWriter {
	return &StreamWriter{out: w}
}

// WriteAccepted outputs successful action confirmation
// for example: 'Player [1] registered'
func (w *StreamWriter) WriteAccepted(e domain.IncomingEvent) error {
	timeStr := formatTime(e.TimeSec)
	var action string

	switch e.ID {
	case domain.EventRegister:
		action = "registered"
	case domain.EventEnterDungeon:
		action = "entered the dungeon"
	case domain.EventKillMonster:
		action = "killed the monster"
	case domain.EventNextFloor:
		action = "went to the next floor"
	case domain.EventPrevFloor:
		action = "went to the previous floor"
	case domain.EventEnterBoss:
		action = "entered the boss's floor"
	case domain.EventKillBoss:
		action = "killed the boss"
	case domain.EventLeaveDungeon:
		action = "left the dungeon"
	case domain.EventCannotContinue:
		action = fmt.Sprintf("cannot continue due to [%s]", e.Extra)
	case domain.EventRestoreHP:
		action = fmt.Sprintf("has restored [%d] of health", e.Value)
	case domain.EventReceiveDamage:
		action = fmt.Sprintf("recieved [%d] of damage", e.Value)
	}

	_, err := fmt.Fprintf(w.out, "[%s] Player [%d] %s\n", timeStr, e.PlayerID, action)
	return err
}

// WriteOutgoing outputs system events
// for example death, disqualification, impossible move
func (w *StreamWriter) WriteOutgoing(e domain.OutgoingEvent) error {
	timeStr := formatTime(e.TimeSec)
	var action string

	switch e.ID {
	case domain.EventOutDisqualified:
		action = "is disqualified"
	case domain.EventOutDead:
		action = "is dead"
	case domain.EventOutImpossible:
		action = fmt.Sprintf("makes imposible move [%s]", e.ExtraParam)
	}

	_, err := fmt.Fprintf(w.out, "[%s] Player [%d] %s\n", timeStr, e.PlayerID, action)
	return err
}

// WriteReport outputs final report for all players
func (w *StreamWriter) WriteReport(reports []usecase.ReportEntry) error {
	_, err := fmt.Fprintln(w.out, "\nFinal report:")
	if err != nil {
		return err
	}

	for _, r := range reports {
		var statusStr string
		switch r.Status {
		case domain.StatusSuccess:
			statusStr = "SUCCESS"
		case domain.StatusFail:
			statusStr = "FAIL"
		case domain.StatusDisqual:
			statusStr = "DISQUAL"
		}

		tTotal := formatTime(r.TotalTime)
		tAvg := formatTime(r.AvgFloorTime)
		tBoss := formatTime(r.BossKillTime)

		_, err = fmt.Fprintf(w.out, "[%s] %d [%s, %s, %s] HP:%d\n",
			statusStr, r.PlayerID, tTotal, tAvg, tBoss, r.HP)
		if err != nil {
			return err
		}
	}
	return nil
}

// formatTime converts seconds to 'HH:MM:SS'
func formatTime(sec int) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
