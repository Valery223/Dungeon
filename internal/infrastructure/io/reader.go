package io

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Valery223/Dungeon/internal/domain"
)

// StreamReader reads incoming events from a stream
type StreamReader struct {
	scanner *bufio.Scanner
}

// NewStreamReader creates a new StreamReader for the given io.Reader
func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{
		scanner: bufio.NewScanner(r),
	}
}

// Read reads the next event from the stream, returns io.EOF on completion
func (r *StreamReader) Read() (domain.IncomingEvent, error) {
	for r.scanner.Scan() {
		line := strings.TrimSpace(r.scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		event, err := parseLineSimple(line)
		if err != nil {
			return domain.IncomingEvent{}, fmt.Errorf("failed to parse line '%s': %v", line, err)
		}
		return event, nil
	}

	if err := r.scanner.Err(); err != nil {
		return domain.IncomingEvent{}, err
	}
	return domain.IncomingEvent{}, io.EOF
}

// parseLineSimple - parse a line into IncomingEvent, expecting format: [HH:MM:SS] PlayerID EventID [ExtraParam...]
func parseLineSimple(line string) (domain.IncomingEvent, error) {
	var event domain.IncomingEvent

	// 1 Split line by spaces
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return event, fmt.Errorf("not enough fields")
	}

	// 2 Parse time '[14:49:02]'
	if !strings.HasPrefix(parts[0], "[") || !strings.HasSuffix(parts[0], "]") {
		return event, fmt.Errorf("invalid time format")
	}
	timeStr := strings.Trim(parts[0], "[]")
	var err error
	event.TimeSec, err = parseTime(timeStr)
	if err != nil {
		return event, fmt.Errorf("invalid time format")
	}

	// 3 Parse player ID and event ID
	playerID, err1 := strconv.Atoi(parts[1])
	eventID, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return event, fmt.Errorf("invalid numeric fields")
	}

	event.PlayerID = playerID
	event.ID = domain.EventID(eventID)

	// 4 Parse Extra parameter if present
	if len(parts) > 3 {
		if event.ID == domain.EventReceiveDamage || event.ID == domain.EventRestoreHP {
			// For damage and healing this is a number
			val, _ := strconv.Atoi(parts[3])
			event.Value = val
		} else {
			// For 'CannotContinue' reason this is a string
			event.Extra = strings.Join(parts[3:], " ")
		}
	}

	return event, nil
}

// parseTime converts time to seconds
func parseTime(t string) (int, error) {
	var h, m, s int
	_, err := fmt.Sscanf(t, "%d:%d:%d", &h, &m, &s)
	if err != nil {
		return 0, fmt.Errorf("invalid time format")
	}
	return h*3600 + m*60 + s, nil
}
