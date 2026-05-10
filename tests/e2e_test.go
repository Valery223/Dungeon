package tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Valery223/Dungeon/internal/domain"
	"github.com/Valery223/Dungeon/internal/infrastructure/io"
	"github.com/Valery223/Dungeon/internal/infrastructure/memory"
	"github.com/Valery223/Dungeon/internal/usecase"
)

// Тесты end-to-end для проверки всей системы на примере из задания

// TestEndToEndIntegration_from_example - тестирует систему на примере данном в задании
func TestEndToEndIntegration_from_example(t *testing.T) {
	// 1 Входные данные
	inputEvents := `
[14:00:00] 1 1
[14:00:00] 2 1
[14:10:00] 2 2
[14:10:00] 3 2
[14:11:00] 2 5
[14:12:00] 3 3
[14:14:00] 2 3
[14:27:00] 2 11 60
[14:29:00] 2 11 50
[14:40:00] 1 2
[14:41:00] 1 3
[14:44:00] 1 11 50
[14:45:00] 1 3
[14:48:00] 1 4
[14:48:00] 1 6
[14:49:00] 1 11 25
[14:49:02] 1 10 80
[14:50:00] 1 11 65
[14:59:00] 1 7
[15:04:00] 1 8`

	expectedOutput := `
[14:00:00] Player [1] registered
[14:00:00] Player [2] registered
[14:10:00] Player [2] entered the dungeon
[14:10:00] Player [3] is disqualified
[14:11:00] Player [2] makes imposible move [5]
[14:14:00] Player [2] killed the monster
[14:27:00] Player [2] recieved [60] of damage
[14:29:00] Player [2] recieved [50] of damage
[14:29:00] Player [2] is dead
[14:40:00] Player [1] entered the dungeon
[14:41:00] Player [1] killed the monster
[14:44:00] Player [1] recieved [50] of damage
[14:45:00] Player [1] killed the monster
[14:48:00] Player [1] went to the next floor
[14:48:00] Player [1] entered the boss's floor
[14:49:00] Player [1] recieved [25] of damage
[14:49:02] Player [1] has restored [80] of health
[14:50:00] Player [1] recieved [65] of damage
[14:59:00] Player [1] killed the boss
[15:04:00] Player [1] left the dungeon

Final report:
[SUCCESS] 1 [00:24:00, 00:05:00, 00:11:00] HP:35
[FAIL] 2 [00:19:00, 00:00:00, 00:00:00] HP:0
[DISQUAL] 3 [00:00:00, 00:00:00, 00:00:00] HP:100
`

	// 2 Настраиваем Конфиг
	cfg := &domain.DungeonConfig{
		Floors:      2,
		Monsters:    2,
		OpenAtSec:   14*3600 + 5*60, // 14:05:00
		DurationSec: 2 * 3600,
		CloseAtSec:  (14+2)*3600 + 5*60,
	}

	// 3 Подменяем ввод и вывод
	inBuf := strings.NewReader(inputEvents)
	outBuf := new(bytes.Buffer)

	reader := io.NewStreamReader(inBuf)
	writer := io.NewStreamWriter(outBuf)
	repo := memory.NewInMemoryPlayerRepo()
	processor := usecase.NewEventProcessor(cfg, repo)
	runner := usecase.NewGameRunner(reader, writer, processor)

	// 4Запускаем систему
	err := runner.Run()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 5 Проверяем результат
	actualOutput := strings.TrimSpace(outBuf.String())
	expectedClean := strings.TrimSpace(expectedOutput)

	if actualOutput != expectedClean {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expectedClean, actualOutput)
	}
}

// TestEndToEndIntegration_one_player - тестирует систему на примере одного игрока,
// который успешно проходит игру
func TestEndToEndIntegration_one_player(t *testing.T) {
	// 1 Входные данные
	inputEvents := `
[14:00:00] 1 1
[14:05:10] 1 2
[14:05:20] 1 3
[14:05:30] 1 11 40
[14:05:40] 1 11 40
[14:05:50] 1 3
[14:06:00] 1 3
[14:06:10] 1 10 40
[14:06:20] 1 4
[14:06:30] 1 3
[14:06:40] 1 3
[14:06:50] 1 3
[14:07:00] 1 4
[14:07:10] 1 6
[14:07:20] 1 11 10
[14:07:30] 1 7
[14:07:40] 1 8`

	expectedOutput := `
[14:00:00] Player [1] registered
[14:05:10] Player [1] entered the dungeon
[14:05:20] Player [1] killed the monster
[14:05:30] Player [1] recieved [40] of damage
[14:05:40] Player [1] recieved [40] of damage
[14:05:50] Player [1] killed the monster
[14:06:00] Player [1] killed the monster
[14:06:10] Player [1] has restored [40] of health
[14:06:20] Player [1] went to the next floor
[14:06:30] Player [1] killed the monster
[14:06:40] Player [1] killed the monster
[14:06:50] Player [1] killed the monster
[14:07:00] Player [1] went to the next floor
[14:07:10] Player [1] entered the boss's floor
[14:07:20] Player [1] recieved [10] of damage
[14:07:30] Player [1] killed the boss
[14:07:40] Player [1] left the dungeon

Final report:
[SUCCESS] 1 [00:02:30, 00:00:40, 00:00:20] HP:50
`

	// 2 Настраиваем Конфиг
	cfg := &domain.DungeonConfig{
		Floors:      3,
		Monsters:    3,
		OpenAtSec:   14*3600 + 5*60, // 14:05:00
		DurationSec: 2 * 3600,
		CloseAtSec:  (14+2)*3600 + 5*60,
	}

	// 3 Подменяем ввод и вывод
	inBuf := strings.NewReader(inputEvents)
	outBuf := new(bytes.Buffer)

	reader := io.NewStreamReader(inBuf)
	writer := io.NewStreamWriter(outBuf)
	repo := memory.NewInMemoryPlayerRepo()
	processor := usecase.NewEventProcessor(cfg, repo)
	runner := usecase.NewGameRunner(reader, writer, processor)

	// 4Запускаем систему
	err := runner.Run()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 5 Проверяем результат
	actualOutput := strings.TrimSpace(outBuf.String())
	expectedClean := strings.TrimSpace(expectedOutput)

	if actualOutput != expectedClean {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expectedClean, actualOutput)
	}
}

// TestEndToEndIntegration_one_player - тестирует систему на примере одного игрока,
// который успешно проходит игру
func TestEndToEndIntegration_one_player_2_with_imposibe_move(t *testing.T) {
	// 1 Входные данные
	inputEvents := `
[14:00:00] 1 3
[14:00:10] 1 4
[14:00:20] 1 1
[14:00:30] 1 3
[14:00:40] 1 5
[14:00:50] 1 6
[14:01:00] 1 7
[14:00:10] 1 8
[14:00:20] 1 10 12
[14:05:10] 1 2
[14:05:20] 1 3
[14:05:30] 1 3
[14:05:40] 1 4
[14:05:50] 1 6
[14:06:00] 1 7
[14:07:10] 1 8
`

	expectedOutput := `
[14:00:00] Player [1] makes imposible move [3]
[14:00:10] Player [1] makes imposible move [4]
[14:00:20] Player [1] registered
[14:00:30] Player [1] makes imposible move [3]
[14:00:40] Player [1] makes imposible move [5]
[14:00:50] Player [1] makes imposible move [6]
[14:01:00] Player [1] makes imposible move [7]
[14:00:10] Player [1] makes imposible move [8]
[14:00:20] Player [1] has restored [12] of health
[14:05:10] Player [1] entered the dungeon
[14:05:20] Player [1] killed the monster
[14:05:30] Player [1] killed the monster
[14:05:40] Player [1] went to the next floor
[14:05:50] Player [1] entered the boss's floor
[14:06:00] Player [1] killed the boss
[14:07:10] Player [1] left the dungeon

Final report:
[SUCCESS] 1 [00:02:00, 00:00:20, 00:00:10] HP:100
`

	// 2 Настраиваем Конфиг
	cfg := &domain.DungeonConfig{
		Floors:      2,
		Monsters:    2,
		OpenAtSec:   14*3600 + 5*60, // 14:05:00
		DurationSec: 2 * 3600,
		CloseAtSec:  (14+2)*3600 + 5*60,
	}

	// 3 Подменяем ввод и вывод
	inBuf := strings.NewReader(inputEvents)
	outBuf := new(bytes.Buffer)

	reader := io.NewStreamReader(inBuf)
	writer := io.NewStreamWriter(outBuf)
	repo := memory.NewInMemoryPlayerRepo()
	processor := usecase.NewEventProcessor(cfg, repo)
	runner := usecase.NewGameRunner(reader, writer, processor)

	// 4Запускаем систему
	err := runner.Run()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 5 Проверяем результат
	actualOutput := strings.TrimSpace(outBuf.String())
	expectedClean := strings.TrimSpace(expectedOutput)

	if actualOutput != expectedClean {
		t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expectedClean, actualOutput)
	}
}

// TestEndToEndIntegration_state_new - тестирует переходы из состояния NEW
// на примере одного игрока,
// Возможные переходы:
// NEW -> REGISTERED (при регистрации)
// NEW -> DISQUALIFIED (при входе в данж до регистрации)
// NEW -> DISQUALIFIED (не может продолжать)
// NEW -> FAIL (время вышло, данж закрылся)
// NEW -> FAIL (умирает)
// NEW -> NEW (невалидное действие)
func TestEndToEndIntegration_state_new(t *testing.T) {
	startInput := `
	`
	startOutput := `
	`

	cfg := &domain.DungeonConfig{
		Floors:      2,
		Monsters:    2,
		OpenAtSec:   14*3600 + 5*60, // 14:05:00
		DurationSec: 2 * 3600,
		CloseAtSec:  (14+2)*3600 + 5*60,
	}

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name: "Registered after registration",
			input: `
[14:00:00] 1 1
			`,
			output: `
[14:00:00] Player [1] registered

Final report:
[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Disqualified for entering dungeon before registration",
			input: `
[14:00:00] 1 2
			`,
			output: `
[14:00:00] Player [1] is disqualified

Final report:
[DISQUAL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Disqualified for not being able to continue",
			input: `
[14:00:00] 1 9 reson1 reson2
			`,
			output: `
[14:00:00] Player [1] cannot continue due to [reson1 reson2]
[14:00:00] Player [1] is disqualified

Final report:
[DISQUAL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Failed for time out",
			input: `
[23:00:00] 1 1
			`,
			output: `

Final report:
[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Failed for dying",
			input: `
[14:00:00] 1 11 150
			`,
			output: `
[14:00:00] Player [1] recieved [150] of damage
[14:00:00] Player [1] is dead

Final report:
[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:0
			`,
		},
		{
			name: "Invalid action",
			input: `
[14:00:00] 1 8
			`,
			output: `
[14:00:00] Player [1] makes imposible move [8]

Final report:
[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
	}

	for _, tt := range tests {
		input := startInput + tt.input
		expectedOutput := startOutput + tt.output
		inBuf := strings.NewReader(input)
		outBuf := new(bytes.Buffer)

		reader := io.NewStreamReader(inBuf)
		writer := io.NewStreamWriter(outBuf)
		repo := memory.NewInMemoryPlayerRepo()
		processor := usecase.NewEventProcessor(cfg, repo)
		runner := usecase.NewGameRunner(reader, writer, processor)

		err := runner.Run()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualOutput := strings.TrimSpace(outBuf.String())
		expectedClean := strings.TrimSpace(expectedOutput)

		if actualOutput != expectedClean {
			t.Errorf("Test '%s' failed. Output mismatch.\nExpected:\n%s\nGot:\n%s", tt.name, expectedClean, actualOutput)
		}
	}
}

// TestEndToEndIntegration_state_registered - тестирует переходы из состояния REGISTERED
// на примере одного игрока,
// Возможные переходы:
// REGISTERED -> DUNGEON (при входе в данж после открытия)
// REGISTERED -> DISQUALIFIED (при входе в данж до открытия)
// REGISTERED -> DISQUALIFIED (не может продолжать)
// REGISTERED -> FAIL (время вышло, данж закрылся)
// REGISTERED -> FAIL (умирает)
// REGISTERED -> REGISTERED (невалидное действие)
func TestEndToEndIntegration_state_registered(t *testing.T) {
	startInput := `
[14:00:00] 1 1`
	startOutput := `
[14:00:00] Player [1] registered`

	cfg := &domain.DungeonConfig{
		Floors:      2,
		Monsters:    2,
		OpenAtSec:   14*3600 + 5*60, // 14:05:00
		DurationSec: 2 * 3600,
		CloseAtSec:  (14+2)*3600 + 5*60,
	}

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name: "Dungeon after opening",
			input: `
[14:05:10] 1 2
			`,
			output: `
[14:05:10] Player [1] entered the dungeon

Final report:
[FAIL] 1 [01:59:50, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Disqualified for entering dungeon before opening",
			input: `
[14:00:00] 1 2
			`,
			output: `
[14:00:00] Player [1] is disqualified

Final report:
[DISQUAL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Disqualified for not being able to continue",
			input: `
[14:00:00] 1 9 reson1 reson2
			`,
			output: `
[14:00:00] Player [1] cannot continue due to [reson1 reson2]
[14:00:00] Player [1] is disqualified

Final report:
[DISQUAL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Failed for time out",
			input: `
[23:00:00] 1 2
			`,
			output: `

Final report:
[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Failed for dying",
			input: `
[14:00:00] 1 11 150
			`,
			output: `
[14:00:00] Player [1] recieved [150] of damage
[14:00:00] Player [1] is dead

Final report:
[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:0
			`,
		},
		{
			name: "Invalid action",
			input: `
[14:00:00] 1 8
			`,
			output: `
[14:00:00] Player [1] makes imposible move [8]

Final report:
[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100
			`,
		},
	}

	for _, tt := range tests {
		input := startInput + tt.input
		expectedOutput := startOutput + tt.output
		inBuf := strings.NewReader(input)
		outBuf := new(bytes.Buffer)

		reader := io.NewStreamReader(inBuf)
		writer := io.NewStreamWriter(outBuf)
		repo := memory.NewInMemoryPlayerRepo()
		processor := usecase.NewEventProcessor(cfg, repo)
		runner := usecase.NewGameRunner(reader, writer, processor)

		err := runner.Run()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualOutput := strings.TrimSpace(outBuf.String())
		expectedClean := strings.TrimSpace(expectedOutput)

		if actualOutput != expectedClean {
			t.Errorf("Test '%s' failed. Output mismatch.\nExpected:\n%s\nGot:\n%s", tt.name, expectedClean, actualOutput)
		}
	}
}

// TestEndToEndIntegration_state_dungeon - тестирует переходы из состояния DUNGEON
// на примере одного игрока,
// Больше вариантов на невалидные действия,
// есть в тестах player_test.go,
// Возможные переходы:
// DUNGEON -> SUCCESS (при зачистки данжа)
// DUNGEON -> DISQUALIFIED (не может продолжать)
// DUNGEON -> FAIL (время вышло, данж закрылся)
// DUNGEON -> FAIL (умирает)
// DUNGEON -> FAIL (выходит из данжа до его зачистки)
// DUNGEON -> DUNGEON (невалидное действие)
func TestEndToEndIntegration_state_dungeon(t *testing.T) {
	startInput := `
[14:00:00] 1 1
[14:05:10] 1 2`
	startOutput := `
[14:00:00] Player [1] registered
[14:05:10] Player [1] entered the dungeon`

	cfg := &domain.DungeonConfig{
		Floors:      2,
		Monsters:    2,
		OpenAtSec:   14*3600 + 5*60, // 14:05:00
		DurationSec: 2 * 3600,
		CloseAtSec:  (14+2)*3600 + 5*60,
	}

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name: "Success after clearing the dungeon",
			input: `
[14:05:20] 1 3
[14:05:30] 1 3
[14:05:40] 1 4
[14:05:50] 1 6
[14:06:00] 1 7
[14:06:10] 1 8`,
			output: `
[14:05:20] Player [1] killed the monster
[14:05:30] Player [1] killed the monster
[14:05:40] Player [1] went to the next floor
[14:05:50] Player [1] entered the boss's floor
[14:06:00] Player [1] killed the boss
[14:06:10] Player [1] left the dungeon

Final report:
[SUCCESS] 1 [00:01:00, 00:00:20, 00:00:10] HP:100`,
		},
		{
			name: "Disqualified for not being able to continue",
			input: `
[14:10:00] 1 9 reson1 reson2
			`,
			output: `
[14:10:00] Player [1] cannot continue due to [reson1 reson2]
[14:10:00] Player [1] is disqualified

Final report:
[DISQUAL] 1 [00:04:50, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Failed for time out",
			input: `
			`,
			output: `

Final report:
[FAIL] 1 [01:59:50, 00:00:00, 00:00:00] HP:100
			`,
		},
		{
			name: "Failed for dying",
			input: `
[14:10:00] 1 11 150
			`,
			output: `
[14:10:00] Player [1] recieved [150] of damage
[14:10:00] Player [1] is dead

Final report:
[FAIL] 1 [00:04:50, 00:00:00, 00:00:00] HP:0
			`,
		},
		{
			name: "Invalid action",
			input: `
[14:10:00] 1 5
			`,
			output: `
[14:10:00] Player [1] makes imposible move [5]

Final report:
[FAIL] 1 [01:59:50, 00:00:00, 00:00:00] HP:100
			`,
		},
	}

	for _, tt := range tests {
		input := startInput + tt.input
		expectedOutput := startOutput + tt.output
		inBuf := strings.NewReader(input)
		outBuf := new(bytes.Buffer)

		reader := io.NewStreamReader(inBuf)
		writer := io.NewStreamWriter(outBuf)
		repo := memory.NewInMemoryPlayerRepo()
		processor := usecase.NewEventProcessor(cfg, repo)
		runner := usecase.NewGameRunner(reader, writer, processor)

		err := runner.Run()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		actualOutput := strings.TrimSpace(outBuf.String())
		expectedClean := strings.TrimSpace(expectedOutput)

		if actualOutput != expectedClean {
			t.Errorf("Test '%s' failed. Output mismatch.\nExpected:\n%s\nGot:\n%s", tt.name, expectedClean, actualOutput)
		}
	}
}
