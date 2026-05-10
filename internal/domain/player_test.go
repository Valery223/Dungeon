package domain

import (
	"reflect"
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

// TestPlayer_TimeMetrics проверяет правильность расчета временных метрик (время на этажах, время в данже, время у босса)
// Сценарий теста:
// 1. Игрок прошел данж
// 2. Игрок умер на этаже
// 3. Игрок умер у босса
// 4. Игрок вышел из данжа, не зачистив его
// 5. Время вышло, игрок на этаже
// 6. Время вышло, игрок у босса
// 7. Время вышло, игрок у регистрации
// 8. Время вышло, игрок на этапе New
// 9. Пришло событие "Игрок не может продолжить" на этаже
// 10. Пришло событие "Игрок не может продолжить" у босса
// 11. Пришло событие "Игрок не может продолжить" на регистрации
// 12. Пришло событие "Игрок не может продолжить" на этапе New
// func TestPlayer_TimeMetrics(t *testing.T) {
// 	tests := []struct {
// 		name                      string
// 		inEvents                  []IncomingEvent
// 		bossDead                  bool
// 		expectedTimeSpentOnFloors []int
// 		expectedBossKillTime      int
// 		expectedTimeLeftInDungeon int
// 	}{
// 		{
// 			name: "Игрок прошел данж",
// 			inEvents: IncomingEvent{
// 		},
// 	}
// }

// TestPlayer_TimeMetrics проверяет правильность расчета временных метрик
// (время на этажах, время у босса, фиксация времени выхода).
func TestPlayer_TimeMetrics(t *testing.T) {
	// Конфиг: 2 этажа, по 2 монстра на каждом. Подземелье закрывается на 1000-й секунде.
	cfg := &DungeonConfig{
		Floors:      3,
		Monsters:    2,
		OpenAtSec:   10,
		CloseAtSec:  1010,
		DurationSec: 1000,
	}

	tests := []struct {
		name                 string
		inEvents             []IncomingEvent
		expectedStatus       PlayerStatus
		expectedFloorTimes   []int // Индекс совпадает с номером этажа (0 - пустой)
		expectedBossKillTime int
		expectedLeaveTime    int
	}{
		{
			name: "Игрок успешно прошел весь данж",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20}, // Вошел (CurrentFloorEnterTime = 20)

				// 1 Этаж
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Убил последнего на 1 этаже. Время: 40-20 = 20
				{ID: EventNextFloor, TimeSec: 50},   // Перешел на 2 этаж (CurrentFloorEnterTime = 50)

				// 2 Этаж
				{ID: EventKillMonster, TimeSec: 60},
				{ID: EventKillMonster, TimeSec: 80}, // Убил последнего на 2 этаже. Время: 80-50 = 30
				{ID: EventNextFloor, TimeSec: 85},   // Перешел на 3 этаж (CurrentFloorEnterTime = 85)
				{ID: EventEnterBoss, TimeSec: 90},   // Зашел к боссу (BossEnterTime = 90)

				// Босс
				{ID: EventKillBoss, TimeSec: 120},     // Убил босса. Время: 120-90 = 30
				{ID: EventLeaveDungeon, TimeSec: 130}, // Вышел
			},
			expectedStatus:       StatusSuccess,
			expectedFloorTimes:   []int{0, 20, 30}, // 0-й индекс не используется, 1 этаж = 20с, 2 этаж = 30с
			expectedBossKillTime: 30,
			expectedLeaveTime:    130,
		},
		{
			name: "Игрок умер на 2 этаже",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Зачистил 1 этаж. Время 20с
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventKillMonster, TimeSec: 60},
				{ID: EventReceiveDamage, TimeSec: 70, Value: 100}, // Умер. Время на 2 этаже: 70-50 = 20
			},
			expectedStatus: StatusFail,
			// 2 этаж
			expectedFloorTimes:   []int{0, 20, 20},
			expectedBossKillTime: 0,
			expectedLeaveTime:    70, // Время смерти = время выхода
		},
		{
			name: "Время вышло, пока игрок был на этаже",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Зачистил 1 этаж. Время 20с
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventKillMonster, TimeSec: 60},
				// Приходит любое событие после CloseAtSec (1010)
				{ID: EventKillMonster, TimeSec: 1050},
			},
			expectedStatus:       StatusFail,
			expectedFloorTimes:   []int{0, 20, 1010 - 50}, // Ничего не успел зачистить
			expectedBossKillTime: 0,
			expectedLeaveTime:    1010,
		},
		{
			name: "Игрок вышел из данжа, не зачистив его",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventLeaveDungeon, TimeSec: 40},
			},
			expectedStatus:       StatusFail,
			expectedFloorTimes:   []int{0, 20, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    40,
		},
		{
			name: "Игрок не может продолжить (Cannot continue) на этаже",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventCannotContinue, TimeSec: 30, Extra: "Test test"},
			},
			expectedStatus:       StatusDisqual,
			expectedFloorTimes:   []int{0, 10, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    30,
		},
		{
			name: "Время вышло, пока игрок был в статусе New",
			inEvents: []IncomingEvent{
				// Пытается зарегистрироваться после закрытия подземелья
				{ID: EventRegister, TimeSec: 1050},
			},
			expectedStatus:       StatusFail,
			expectedFloorTimes:   []int{0, 0, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    0,
		},
		{
			name: "Игрок перешл со 2 этажа на 1 этаж и не может продолжить (Cannot continue) на первом этаже",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Зачистил 1 этаж. Время 20с
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventPrevFloor, TimeSec: 60}, // Вернулся на 1 этаж. Время на 2 этаже = 10с
				{ID: EventCannotContinue, TimeSec: 70, Extra: "Test test"},
			},
			expectedStatus:       StatusDisqual,
			expectedFloorTimes:   []int{0, 20, 10},
			expectedBossKillTime: 0,
			expectedLeaveTime:    70,
		},
		{
			name: "Игрок был у боссаи получил событие Cannot continue",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Зачистил 1 этаж. Время 20с
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventKillMonster, TimeSec: 60},
				{ID: EventKillMonster, TimeSec: 70}, // Зачистил 2 этаж. Время 20с
				{ID: EventNextFloor, TimeSec: 80},
				{ID: EventEnterBoss, TimeSec: 90},                           // Время у босса началось
				{ID: EventCannotContinue, TimeSec: 100, Extra: "Test test"}, // Получает событие "не может продолжать" у босса
			},
			expectedStatus:       StatusDisqual,
			expectedFloorTimes:   []int{0, 20, 20},
			expectedBossKillTime: 10, // Время у босса = 100 - 90 = 10
			expectedLeaveTime:    100,
		},
		{
			name: "Игрок был зарегистрирован, но получил событие Cannot continue до входа в данж",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventCannotContinue, TimeSec: 20, Extra: "Test test"}, // Получает событие "не может продолжать" до входа в данж
			},
			expectedStatus:       StatusDisqual,
			expectedFloorTimes:   []int{0, 0, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{
				ID:                1,
				Status:            StatusNew,
				HP:                100,
				MonstersLeft:      make([]int, cfg.Floors+1),
				FloorCleared:      make([]bool, cfg.Floors+1),
				TimeSpentOnFloors: make([]int, cfg.Floors),
			}
			for i := 1; i <= cfg.Floors; i++ {
				player.MonstersLeft[i] = cfg.Monsters
			}

			// Прогоняем всю цепочку событий через автомат
			for _, event := range tt.inEvents {
				player.ApplyEvent(event, cfg)
			}

			// Проверяем финальный статус
			if player.Status != tt.expectedStatus {
				t.Errorf("Статус: ожидалось %v, получено %v", tt.expectedStatus, player.Status)
			}

			// Проверяем массивы времени на этажах
			// Используем reflect.DeepEqual для быстрого сравнения слайсов
			if !reflect.DeepEqual(player.TimeSpentOnFloors, tt.expectedFloorTimes) {
				// Если массивы пустые (nil), то DeepEqual может ругаться, сделаем проверку длины
				if len(player.TimeSpentOnFloors) > 0 {
					t.Errorf("Время на этажах: ожидалось %v, получено %v", tt.expectedFloorTimes, player.TimeSpentOnFloors)
				}
			}

			// Проверяем время убийства босса
			if player.BossKillOrExitTime != tt.expectedBossKillTime {
				t.Errorf("Время босса: ожидалось %d, получено %d", tt.expectedBossKillTime, player.BossKillOrExitTime)
			}

			// Проверяем зафиксированное время выхода
			if player.LeaveDungeonTime != tt.expectedLeaveTime {
				t.Errorf("Время выхода: ожидалось %d, получено %d", tt.expectedLeaveTime, player.LeaveDungeonTime)
			}
		})
	}
}
