package domain

import (
	"testing"
)

func defaultPlayerInDungeon() *Player {
	return &Player{
		ID:                    1,
		Status:                StatusInDungeon,
		HP:                    100,
		CurrentFloor:          1,
		CurrentFloorEnterTime: 1000,
		MonstersLeft:          []int{0, 2, 2},
		FloorCleared:          []bool{false, false, false},
		TimeSpentOnFloors:     []int{0, 0, 0},
	}
}

func defaultCfg() *DungeonConfig {
	return &DungeonConfig{
		Floors:      2,
		Monsters:    2,
		OpenAtSec:   1,
		CloseAtSec:  99999,
		DurationSec: 99999,
	}
}

// assertActionResult - вспомогательная функция для проверки результатов действий игрока в тестаx
func assertActionResult(t *testing.T, expectedStatus PlayerStatus, expectedRes ActionResult, actualPlayer *Player, actualRes ActionResult) {
	t.Helper()

	// 1. Проверяем изменение состояния
	if actualPlayer.Status != expectedStatus {
		t.Errorf("Статус: ожидалось %v, получено %v", expectedStatus, actualPlayer.Status)
	}

	// 2. Проверяем, принято ли действие
	if actualRes.IsAccepted != expectedRes.IsAccepted {
		t.Errorf("IsAccepted: ожидалось %v, получено %v", expectedRes.IsAccepted, actualRes.IsAccepted)
	}

	// 3. Проверяем ответное событие
	if expectedRes.OutgoingEvent == nil && actualRes.OutgoingEvent != nil {
		t.Errorf("OutgoingEvent: ожидалось nil, получено %+v", actualRes.OutgoingEvent)
		return
	}
	if expectedRes.OutgoingEvent != nil {
		if actualRes.OutgoingEvent == nil {
			t.Errorf("OutgoingEvent: ожидалось %+v, получено nil", expectedRes.OutgoingEvent)
			return
		}
		if expectedRes.OutgoingEvent.ID != actualRes.OutgoingEvent.ID {
			t.Errorf("OutgoingEvent ID: ожидалось %v, получено %v", expectedRes.OutgoingEvent.ID, actualRes.OutgoingEvent.ID)
		}
		if expectedRes.OutgoingEvent.IncomingEventID != actualRes.OutgoingEvent.IncomingEventID {
			t.Errorf("OutgoingEvent IncomingEventID: ожидалось %v, получено %v", expectedRes.OutgoingEvent.IncomingEventID, actualRes.OutgoingEvent.IncomingEventID)
		}
	}
}

// TestPlayer_StateNew проверяет все переходы из состояния New
// Таблица переходов:
//   - Registered - если регистрируется
//   - Disqual - если пытается войти в данж, не зарегистрировавшись
//   - Disqual - если получаем событие о том, что не может продолжать соревнование
//   - Fail - если получает событие, когда время вышло
//   - New - при всех остальных событиях, так же  отдаем событие  imposible move
func TestPlayer_StateNew(t *testing.T) {

	tests := []struct {
		name                 string
		inEvent              IncomingEvent
		expectedStatus       PlayerStatus
		expectedActionResult ActionResult
	}{
		{
			name:                 "Регистрация",
			inEvent:              IncomingEvent{ID: EventRegister, TimeSec: 100},
			expectedStatus:       StatusRegistered,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name:           "Вход без регистрации",
			inEvent:        IncomingEvent{ID: EventEnterDungeon, TimeSec: 100},
			expectedStatus: StatusDisqual,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDisqualified,
					IncomingEventID: EventEnterDungeon,
				}},
		},
		{
			name:                 "Не может продолжать соревнование",
			inEvent:              IncomingEvent{ID: EventCannotContinue, TimeSec: 100, Extra: "test"},
			expectedStatus:       StatusDisqual,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Время вышло",
			inEvent: IncomingEvent{
				ID:      EventRegister,
				TimeSec: 99999,
			},
			expectedStatus:       StatusFail,
			expectedActionResult: ActionResult{IsAccepted: false},
		},
		{
			name: "Невозможное действие",
			inEvent: IncomingEvent{
				ID:      EventKillMonster,
				TimeSec: 100,
			},
			expectedStatus: StatusNew,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutImpossible,
					IncomingEventID: EventKillMonster,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{ID: 1, Status: StatusNew}
			cfg := defaultCfg()
			result := player.ApplyEvent(tt.inEvent, cfg)

			assertActionResult(t, tt.expectedStatus, tt.expectedActionResult, player, result)
		})
	}
}

// TestPlayer_StateRegistered проверяет все переходы из состояния Registered
// Таблица переходов:
//   - InDungeon - если игрок входит в данж
//   - Disqual - если приходит  событие о том, что не может продолжать соревнование
//   - Disqual - если игрок пытается войти в данж до его открытия
//   - Fail - если получает событие, когда время вышло
//   - Registered - при всех остальных событиях, так же  отдаем событие  imposible move
func TestPlayer_StateRegistered(t *testing.T) {

	tests := []struct {
		name                 string
		inEvent              IncomingEvent
		expectedStatus       PlayerStatus
		expectedActionResult ActionResult
	}{
		{
			name: "Вход в данж",
			inEvent: IncomingEvent{
				ID:      EventEnterDungeon,
				TimeSec: 100,
			},
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Не может продолжать соревнование",
			inEvent: IncomingEvent{
				ID:      EventCannotContinue,
				TimeSec: 100,
				Extra:   "test",
			},
			expectedStatus:       StatusDisqual,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Пытается войти в данж до открытия",
			inEvent: IncomingEvent{
				ID:      EventEnterDungeon,
				TimeSec: 0,
			},
			expectedStatus: StatusDisqual,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDisqualified,
					IncomingEventID: EventEnterDungeon,
				}},
		},
		{
			name: "Время вышло",
			inEvent: IncomingEvent{
				ID:      EventEnterDungeon,
				TimeSec: 99999,
			},
			expectedStatus:       StatusFail,
			expectedActionResult: ActionResult{IsAccepted: false},
		},
		{
			name: "Невозможное действие",
			inEvent: IncomingEvent{
				ID:      EventKillMonster,
				TimeSec: 100,
			},
			expectedStatus: StatusRegistered,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutImpossible,
					IncomingEventID: EventKillMonster,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{ID: 1, Status: StatusRegistered}
			cfg := defaultCfg()
			result := player.ApplyEvent(tt.inEvent, cfg)

			assertActionResult(t, tt.expectedStatus, tt.expectedActionResult, player, result)
		})
	}
}

// TestPlayer_StateInDungeon проверяет все переходы из состояния InDungeon
// Таблица переходов:
//   - InDungeon - если игрок спускается на предыдущий этаж(если он не на первом этаже и не у босса)
//   - InDungeon - если игрок убивает монстра (если он не на последнем этаже и не у босса и в комнате с монстрами еще есть монстры)
//   - InDungeon - если игрок поднимается на следующий этаж(если он не на последнем этаже и не у босса и зачистил текущий этаж)
//   - InDungeon - если игрок входит в комнату с боссом (если он на последнем этаже и уже не у босса)
//   - InDungeon - если игрок убивает босса (если он на последнем этаже и у босса)
//     -InDungeon - если игрок восстанавливает HP
//     -InDungeon - если игрок получает урон, но не умирает
//     -Disqual - если приходит  событие о том, что не может продолжать соревнование
//     -Fail - если получаем событие, когда время вышло
//     -Fail - если получает урон, который убивает игрока(генерируем событие смерти)
//     -Fail - если выходит из данжа не зачистив подземелье
//     -Success - если выходит из данжа, зачистив подземелье
//     -InDungeon - если пытается подняться на следующий этаж, находясь на последнем этаже, событие  imposible move
//     -InDungeon - если пытается подняться на следующий этаж, не зачистив текущий этаж, событие  imposible move
//     -InDungeon - если пытается спуститься на предыдущий этаж, находясь на первом этаже, событие  imposible move
//     -InDungeon - при всех остальных событиях, так же  отдаем событие  imposible move
func TestPlayer_StateInDungeon(t *testing.T) {
	cfg := defaultCfg()
	tests := []struct {
		name                 string
		inEvent              IncomingEvent
		currentFloor         int
		monstersLeft         []int
		floorCleared         []bool
		bossDead             bool
		expectedStatus       PlayerStatus
		expectedActionResult ActionResult
	}{
		{
			name: "Спускается на предыдущий этаж",
			inEvent: IncomingEvent{
				ID:      EventPrevFloor,
				TimeSec: 100,
			},
			currentFloor:         2,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Убивает монстра",
			inEvent: IncomingEvent{
				ID:      EventKillMonster,
				TimeSec: 100,
			},
			currentFloor:         1,
			monstersLeft:         []int{0, 1, 2},
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Поднимается на следующий этаж",
			inEvent: IncomingEvent{
				ID:      EventNextFloor,
				TimeSec: 100,
			},
			currentFloor:         1,
			floorCleared:         []bool{true, true, true},
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Входит в комнату с боссом",
			inEvent: IncomingEvent{
				ID:      EventEnterBoss,
				TimeSec: 100,
			},
			currentFloor:         2,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Убивает босса",
			inEvent: IncomingEvent{
				ID:      EventKillBoss,
				TimeSec: 100,
			},
			currentFloor:         3,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Восстанавливает HP",
			inEvent: IncomingEvent{
				ID:      EventRestoreHP,
				TimeSec: 100,
				Value:   10,
			},
			currentFloor:         2,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Получает урон, но не умирает",
			inEvent: IncomingEvent{
				ID:      EventReceiveDamage,
				TimeSec: 100,
				Value:   10,
			},
			currentFloor:         1,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Не может продолжать соревнование",
			inEvent: IncomingEvent{
				ID:      EventCannotContinue,
				TimeSec: 100,
				Extra:   "test",
			},
			currentFloor:         1,
			expectedStatus:       StatusDisqual,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Время вышло",
			inEvent: IncomingEvent{
				ID:      EventRestoreHP,
				TimeSec: 99999,
				Value:   10,
			},
			expectedStatus:       StatusFail,
			expectedActionResult: ActionResult{IsAccepted: false},
		},
		{
			name: "Получает урон, который убивает игрока",
			inEvent: IncomingEvent{
				ID:      EventReceiveDamage,
				TimeSec: 100,
				Value:   100,
			},
			currentFloor:   1,
			expectedStatus: StatusFail,
			expectedActionResult: ActionResult{
				IsAccepted: true,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDead,
					IncomingEventID: EventReceiveDamage,
				},
			},
		},
		{
			name: "Выходит из данжа не зачистив подземелье",
			inEvent: IncomingEvent{
				ID:      EventLeaveDungeon,
				TimeSec: 100,
			},
			currentFloor:         1,
			floorCleared:         []bool{false, false, false},
			expectedStatus:       StatusFail,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Выходит из данжа, зачистив подземелье",
			inEvent: IncomingEvent{
				ID:      EventLeaveDungeon,
				TimeSec: 100,
			},
			currentFloor:         2,
			floorCleared:         []bool{false, true, true},
			bossDead:             true,
			expectedStatus:       StatusSuccess,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Пытается подняться на следующий этаж, находясь на последнем этаже",
			inEvent: IncomingEvent{
				ID:      EventNextFloor,
				TimeSec: 100,
			},
			currentFloor:   2,
			floorCleared:   []bool{true, true, true},
			bossDead:       true,
			expectedStatus: StatusInDungeon,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutImpossible,
					IncomingEventID: EventNextFloor,
				},
			},
		},
		{
			name: "Пытается подняться на следующий этаж, не зачистив текущий этаж",
			inEvent: IncomingEvent{
				ID:      EventNextFloor,
				TimeSec: 100,
			},
			currentFloor:   1,
			floorCleared:   []bool{false, false, false},
			bossDead:       false,
			expectedStatus: StatusInDungeon,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutImpossible,
					IncomingEventID: EventNextFloor,
				},
			},
		},
		{
			name: "Пытается спуститься на предыдущий этаж, находясь на первом этаже",
			inEvent: IncomingEvent{
				ID:      EventPrevFloor,
				TimeSec: 100,
			},
			currentFloor:   1,
			floorCleared:   []bool{true, true, true},
			bossDead:       true,
			expectedStatus: StatusInDungeon,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutImpossible,
					IncomingEventID: EventPrevFloor,
				},
			},
		},
		{
			name: "Невозможное действие",
			inEvent: IncomingEvent{
				ID:      EventRegister,
				TimeSec: 100,
			},
			currentFloor:   1,
			floorCleared:   []bool{true, true, true},
			bossDead:       true,
			expectedStatus: StatusInDungeon,
			expectedActionResult: ActionResult{
				IsAccepted: false,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutImpossible,
					IncomingEventID: EventRegister,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := defaultPlayerInDungeon()

			if tt.currentFloor != 0 {
				player.CurrentFloor = tt.currentFloor
			}
			if tt.monstersLeft != nil {
				player.MonstersLeft = tt.monstersLeft
			}
			if tt.floorCleared != nil {
				player.FloorCleared = tt.floorCleared
			}
			if tt.bossDead {
				player.BossDead = tt.bossDead
			}

			result := player.ApplyEvent(tt.inEvent, cfg)
			assertActionResult(t, tt.expectedStatus, tt.expectedActionResult, player, result)
		})
	}
}
