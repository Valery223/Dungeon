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

func TestEndToEndIntegration(t *testing.T) {
	// 1 Входные данные
	inputEvents := `[14:00:00] 1 1
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

	expectedOutput := `[14:00:00] Player [1] registered
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
