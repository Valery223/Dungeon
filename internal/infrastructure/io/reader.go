package io

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Valery223/Dungeon/internal/domain"
)

// StreamReader читает входящие события из потока
type StreamReader struct {
	scanner *bufio.Scanner
}

// NewStreamReader создает новый StreamReader для данного io.Reader
func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{
		scanner: bufio.NewScanner(r),
	}
}

// Read читает следующее событие из потока, возвращает io.EOF при завершении
func (r *StreamReader) Read() (domain.IncomingEvent, error) {
	if r.scanner.Scan() {
		line := strings.TrimSpace(r.scanner.Text())
		// if line == "" {
		// 	continue // Пропускаем пустые строки
		// }

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

// parseLineSimple - парсинг строки в IncomingEvent, ожидая формат: [HH:MM:SS] PlayerID EventID [ExtraParam...]
func parseLineSimple(line string) (domain.IncomingEvent, error) {
	var event domain.IncomingEvent

	// 1 Разбиваем строку по пробелам
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return event, fmt.Errorf("not enough fields")
	}

	// 2 Парсим время "[14:49:02]"
	timeStr := strings.Trim(parts[0], "[]")
	event.TimeSec = parseTime(timeStr)

	// 3 Парсим ID игрока и ID события
	playerID, err1 := strconv.Atoi(parts[1])
	eventID, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return event, fmt.Errorf("invalid numeric fields")
	}

	event.PlayerID = playerID
	event.ID = domain.EventID(eventID)

	// 4 Парсим Extra параметр, если он есть
	if len(parts) > 3 {
		if event.ID == domain.EventReceiveDamage || event.ID == domain.EventRestoreHP {
			// Для урона и лечения это число
			val, _ := strconv.Atoi(parts[3])
			event.Value = val
		} else {
			// Для причины "CannotContinue" это строка
			event.Extra = strings.Join(parts[3:], " ")
		}
	}

	return event, nil
}

// parseTime переводит время в секунды
func parseTime(t string) int {
	var h, m, s int
	fmt.Sscanf(t, "%d:%d:%d", &h, &m, &s)
	return h*3600 + m*60 + s
}
