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

// assertActionResult - helper function for checking the results of player actions in tests
func assertActionResult(
	t *testing.T,
	expectedStatus PlayerStatus,
	expectedRes ActionResult,
	actualPlayer *Player,
	actualRes ActionResult,
) {
	t.Helper()

	// 1. Check status change
	if actualPlayer.Status != expectedStatus {
		t.Errorf("Status: expected %v, got %v", expectedStatus, actualPlayer.Status)
	}

	// 2. Check if action was accepted
	if actualRes.IsAccepted != expectedRes.IsAccepted {
		t.Errorf("IsAccepted: expected %v, got %v", expectedRes.IsAccepted, actualRes.IsAccepted)
	}

	// 3. Check outgoing event
	if expectedRes.OutgoingEvent == nil && actualRes.OutgoingEvent != nil {
		t.Errorf("OutgoingEvent: expected nil, got %+v", actualRes.OutgoingEvent)
		return
	}
	if expectedRes.OutgoingEvent != nil {
		if actualRes.OutgoingEvent == nil {
			t.Errorf("OutgoingEvent: expected %+v, got nil", expectedRes.OutgoingEvent)
			return
		}
		if expectedRes.OutgoingEvent.ID != actualRes.OutgoingEvent.ID {
			t.Errorf(
				"OutgoingEvent ID: expected %v, got %v",
				expectedRes.OutgoingEvent.ID,
				actualRes.OutgoingEvent.ID,
			)
		}
		if expectedRes.OutgoingEvent.IncomingEventID != actualRes.OutgoingEvent.IncomingEventID {
			t.Errorf(
				"OutgoingEvent IncomingEventID: expected %v, got %v",
				expectedRes.OutgoingEvent.IncomingEventID,
				actualRes.OutgoingEvent.IncomingEventID,
			)
		}
	}
}

// TestPlayer_StateNew checks all transitions from New state
// Transition table:
//   - Registered - if registered
//   - Disqual - if tries to enter dungeon without registering
//   - Disqual - if receives event that cannot continue competition
//   - Fail - if receives event when time is up
//   - Fail - if receives lethal damage (generate death event)
//   - New - if receives damage but does not die
//   - New - if player restores HP
//   - New - for all other events, also send imposible move event
func TestPlayer_StateNew(t *testing.T) {

	tests := []struct {
		name                 string
		inEvent              IncomingEvent
		expectedStatus       PlayerStatus
		expectedActionResult ActionResult
	}{
		{
			name:                 "Registration",
			inEvent:              IncomingEvent{ID: EventRegister, TimeSec: 100},
			expectedStatus:       StatusRegistered,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name:           "Enter without registration",
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
			name:           "Cannot continue competition",
			inEvent:        IncomingEvent{ID: EventCannotContinue, TimeSec: 100, Extra: "test"},
			expectedStatus: StatusDisqual,
			expectedActionResult: ActionResult{
				IsAccepted: true,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDisqualified,
					IncomingEventID: EventCannotContinue,
				}},
		},
		{
			name: "Time is up",
			inEvent: IncomingEvent{
				ID:      EventRegister,
				TimeSec: 99999,
			},
			expectedStatus:       StatusFail,
			expectedActionResult: ActionResult{IsAccepted: false},
		},
		{
			name: "Impossible action",
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
		{
			name: "Receives damage but doesn't die",
			inEvent: IncomingEvent{
				ID:      EventReceiveDamage,
				TimeSec: 100,
				Value:   10,
				Extra:   "10",
			},
			expectedStatus: StatusNew,
			expectedActionResult: ActionResult{
				IsAccepted: true,
			},
		},
		{
			name: "Receives lethal damage",
			inEvent: IncomingEvent{
				ID:      EventReceiveDamage,
				TimeSec: 100,
				Value:   100,
				Extra:   "100",
			},
			expectedStatus: StatusFail,
			expectedActionResult: ActionResult{
				IsAccepted: true,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDead,
					IncomingEventID: EventReceiveDamage,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// player := &Player{ID: 1, Status: StatusNew}
			cfg := defaultCfg()
			player := NewPlayer(1, cfg, 100)
			result := player.ApplyEvent(tt.inEvent, cfg)

			assertActionResult(t, tt.expectedStatus, tt.expectedActionResult, player, result)
		})
	}
}

// TestPlayer_StateRegistered checks all transitions from Registered state
// Transition table:
//   - InDungeon - if player enters dungeon
//   - Disqual - if receives event that cannot continue competition
//   - Disqual - if player tries to enter dungeon before it opens
//   - Fail - if receives event when time is up
//   - Fail - if receives lethal damage (generate death event)
//   - Registered - if receives damage but doesn't die
//   - Registered - if player restores HP
//   - Registered - for all other events, also send imposible move event
func TestPlayer_StateRegistered(t *testing.T) {

	tests := []struct {
		name                 string
		inEvent              IncomingEvent
		expectedStatus       PlayerStatus
		expectedActionResult ActionResult
	}{
		{
			name: "Enter dungeon",
			inEvent: IncomingEvent{
				ID:      EventEnterDungeon,
				TimeSec: 100,
			},
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Cannot continue competition",
			inEvent: IncomingEvent{
				ID:      EventCannotContinue,
				TimeSec: 100,
				Extra:   "test",
			},
			expectedStatus: StatusDisqual,
			expectedActionResult: ActionResult{
				IsAccepted: true,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDisqualified,
					IncomingEventID: EventCannotContinue,
				}},
		},
		{
			name: "Tries to enter dungeon before opening",
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
			name: "Time is up",
			inEvent: IncomingEvent{
				ID:      EventEnterDungeon,
				TimeSec: 99999,
			},
			expectedStatus:       StatusFail,
			expectedActionResult: ActionResult{IsAccepted: false},
		},
		{
			name: "Impossible action",
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
		{
			name: "Receives damage but doesn't die",
			inEvent: IncomingEvent{
				ID:      EventReceiveDamage,
				TimeSec: 100,
				Value:   50,
				Extra:   "50",
			},
			expectedStatus:       StatusRegistered,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Receives lethal damage",
			inEvent: IncomingEvent{
				ID:      EventReceiveDamage,
				TimeSec: 100,
				Value:   100,
				Extra:   "100",
			},
			expectedStatus: StatusFail,
			expectedActionResult: ActionResult{
				IsAccepted: true,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDead,
					IncomingEventID: EventReceiveDamage,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{ID: 1, Status: StatusRegistered, HP: 100}
			cfg := defaultCfg()
			result := player.ApplyEvent(tt.inEvent, cfg)

			assertActionResult(t, tt.expectedStatus, tt.expectedActionResult, player, result)
		})
	}
}

// TestPlayer_StateInDungeon checks all transitions from InDungeon state
// Transition table:
//   - InDungeon - if player goes to previous floor (if not on first floor and not at boss)
//   - InDungeon - if player kills monster (if not on last floor and not at boss and there are more monsters in room)
//   - InDungeon - if player goes to next floor (if not on last floor and not at boss and cleared current floor)
//   - InDungeon - if player enters boss room (if on last floor and not at boss yet)
//   - InDungeon - if player kills boss (if on last floor and at boss)
//   - InDungeon - if player restores HP
//   - InDungeon - if player receives damage but doesn't die
//   - Disqual - if receives event that cannot continue competition
//   - Fail - if receives event when time is up
//   - Fail - if receives damage that kills player (generate death event)
//   - Fail - if exits dungeon without clearing it
//   - Success - if exits dungeon after clearing it
//   - InDungeon - if tries to go to next floor while on last floor, imposible move event
//   - InDungeon - if tries to go to next floor without clearing current floor, imposible move event
//   - InDungeon - if tries to go to previous floor while on first floor, imposible move event
//   - InDungeon - for all other events, also send imposible move event
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
			name: "Goes to previous floor",
			inEvent: IncomingEvent{
				ID:      EventPrevFloor,
				TimeSec: 100,
			},
			currentFloor:         2,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Kills monster",
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
			name: "Goes to next floor",
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
			name: "Enters boss room",
			inEvent: IncomingEvent{
				ID:      EventEnterBoss,
				TimeSec: 100,
			},
			currentFloor:         2,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Kills boss",
			inEvent: IncomingEvent{
				ID:      EventKillBoss,
				TimeSec: 100,
			},
			currentFloor:         3,
			expectedStatus:       StatusInDungeon,
			expectedActionResult: ActionResult{IsAccepted: true},
		},
		{
			name: "Restores HP",
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
			name: "Receives damage but doesn't die",
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
			name: "Cannot continue competition",
			inEvent: IncomingEvent{
				ID:      EventCannotContinue,
				TimeSec: 100,
				Extra:   "test",
			},
			currentFloor:   1,
			expectedStatus: StatusDisqual,
			expectedActionResult: ActionResult{
				IsAccepted: true,
				OutgoingEvent: &OutgoingEvent{
					ID:              EventOutDisqualified,
					IncomingEventID: EventCannotContinue,
				},
			},
		},
		{
			name: "Time is up",
			inEvent: IncomingEvent{
				ID:      EventRestoreHP,
				TimeSec: 99999,
				Value:   10,
			},
			expectedStatus:       StatusFail,
			expectedActionResult: ActionResult{IsAccepted: false},
		},
		{
			name: "Receives lethal damage",
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
			name: "Leaves dungeon without clearing it",
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
			name: "Leaves dungeon after clearing it",
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
			name: "Tries to go to next floor from last floor",
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
			name: "Tries to go to next floor without clearing current",
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
			name: "Tries to go to previous floor from first floor",
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
			name: "Impossible action",
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

// TestPlayer_TimeMetrics checks correctness of time metrics calculation (time on floors, time in dungeon, time with boss)
// Test scenarios:
// 1. Player cleared dungeon
// 2. Player died on floor
// 3. Player died with boss
// 4. Player left dungeon without clearing it
// 5. Time is up, player on floor
// 6. Time is up, player with boss
// 7. Time is up, player at registration
// 8. Time is up, player in New state
// 9. Event "Player cannot continue" received on floor
// 10. Event "Player cannot continue" received with boss
// 11. Event "Player cannot continue" received at registration
// 12. Event "Player cannot continue" received in New state
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
// 			name: "Player cleared dungeon",
// 			inEvents: IncomingEvent{
// 		},
// 	}
// }

// TestPlayer_TimeMetrics checks the correctness of time metrics calculation
// (time spent on floors, time with boss, save exit time).
func TestPlayer_TimeMetrics(t *testing.T) {
	// Config: 3 floors, 2 monsters each. Dungeon closes at second 1010.
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
		expectedFloorTimes   []int // Index matches floor number (0 - empty)
		expectedBossKillTime int
		expectedLeaveTime    int
	}{
		{
			name: "Player successfully cleared the entire dungeon",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20}, // Entered (CurrentFloorEnterTime = 20)

				// 1 Floor
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Killed last on floor 1. Time: 40-20 = 20
				{
					ID:      EventNextFloor,
					TimeSec: 50,
				}, // Moved to floor 2 (CurrentFloorEnterTime = 50)

				// 2 Floor
				{ID: EventKillMonster, TimeSec: 60},
				{ID: EventKillMonster, TimeSec: 80}, // Killed last on floor 2. Time: 80-50 = 30
				{
					ID:      EventNextFloor,
					TimeSec: 85,
				}, // Moved to floor 3 (CurrentFloorEnterTime = 85)
				{ID: EventEnterBoss, TimeSec: 90}, // Entered boss room (BossEnterTime = 90)

				// Boss
				{ID: EventKillBoss, TimeSec: 120},     // Killed boss. Time: 120-90 = 30
				{ID: EventLeaveDungeon, TimeSec: 130}, // Left
			},
			expectedStatus: StatusSuccess,
			expectedFloorTimes: []int{
				0,
				20,
				30,
			}, // 0-th index not used, floor 1 = 20s, floor 2 = 30s
			expectedBossKillTime: 30,
			expectedLeaveTime:    130,
		},
		{
			name: "Player died on floor 2",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Cleared floor 1. Time 20s
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventKillMonster, TimeSec: 60},
				{
					ID:      EventReceiveDamage,
					TimeSec: 70,
					Value:   100,
				}, // Died. Time on floor 2: 70-50 = 20
			},
			expectedStatus: StatusFail,
			// Floor 2
			expectedFloorTimes:   []int{0, 20, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    70, // Death time = exit time
		},
		{
			name: "Time is up while player was on a floor",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Cleared floor 1. Time 20s
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventKillMonster, TimeSec: 60},
				// Any event after CloseAtSec (1010)
				{ID: EventKillMonster, TimeSec: 1050},
			},
			expectedStatus:       StatusFail,
			expectedFloorTimes:   []int{0, 20, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    1010,
		},
		{
			name: "Player left dungeon without clearing it",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventLeaveDungeon, TimeSec: 40},
			},
			expectedStatus:       StatusFail,
			expectedFloorTimes:   []int{0, 0, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    40,
		},
		{
			name: "Player cannot continue (Cannot continue) on a floor",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventCannotContinue, TimeSec: 30, Extra: "Test test"},
			},
			expectedStatus:       StatusDisqual,
			expectedFloorTimes:   []int{0, 0, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    30,
		},
		{
			name: "Time is up while player was in New state",
			inEvents: []IncomingEvent{
				// Tries to register after dungeon closes
				{ID: EventRegister, TimeSec: 1050},
			},
			expectedStatus:       StatusFail,
			expectedFloorTimes:   []int{0, 0, 0},
			expectedBossKillTime: 0,
			expectedLeaveTime:    0,
		},
		{
			name: "Player moved from floor 2 to floor 1 and cannot continue (Cannot continue) on floor 1",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Cleared floor 1. Time 20s
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventPrevFloor, TimeSec: 60}, // Returned to floor 1. Time on floor 2 = 10s
				{ID: EventCannotContinue, TimeSec: 70, Extra: "Test test"},
			},
			expectedStatus:       StatusDisqual,
			expectedFloorTimes:   []int{0, 20, 10},
			expectedBossKillTime: 0,
			expectedLeaveTime:    70,
		},
		{
			name: "Player was with boss and received Cannot continue event",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{ID: EventEnterDungeon, TimeSec: 20},
				{ID: EventKillMonster, TimeSec: 30},
				{ID: EventKillMonster, TimeSec: 40}, // Cleared floor 1. Time 20s
				{ID: EventNextFloor, TimeSec: 50},
				{ID: EventKillMonster, TimeSec: 60},
				{ID: EventKillMonster, TimeSec: 70}, // Cleared floor 2. Time 20s
				{ID: EventNextFloor, TimeSec: 80},
				{
					ID:      EventEnterBoss,
					TimeSec: 90,
				}, // Time with boss started
				{
					ID:      EventCannotContinue,
					TimeSec: 100,
					Extra:   "Test test",
				}, // Receives "cannot continue" event with boss
			},
			expectedStatus:       StatusDisqual,
			expectedFloorTimes:   []int{0, 20, 20},
			expectedBossKillTime: 0, // Time with boss = 100 - 90 = 10
			expectedLeaveTime:    100,
		},
		{
			name: "Player was registered but received Cannot continue event before entering dungeon",
			inEvents: []IncomingEvent{
				{ID: EventRegister, TimeSec: 10},
				{
					ID:      EventCannotContinue,
					TimeSec: 20,
					Extra:   "Test test",
				}, // Receives "cannot continue" event before entering dungeon
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

			// Run all events through the state machine
			for _, event := range tt.inEvents {
				player.ApplyEvent(event, cfg)
			}

			// Check final status
			if player.Status != tt.expectedStatus {
				t.Errorf("Status: expected %v, got %v", tt.expectedStatus, player.Status)
			}

			// Check time arrays on floors
			// Use reflect.DeepEqual for fast slice comparison
			if !reflect.DeepEqual(player.TimeSpentOnFloors, tt.expectedFloorTimes) {
				// If arrays are empty (nil), DeepEqual may complain, check length
				if len(player.TimeSpentOnFloors) > 0 {
					t.Errorf(
						"Time on floors: expected %v, got %v",
						tt.expectedFloorTimes,
						player.TimeSpentOnFloors,
					)
				}
			}

			// Check boss kill time
			if player.BossKillOrExitTime != tt.expectedBossKillTime {
				t.Errorf(
					"Boss time: expected %d, got %d",
					tt.expectedBossKillTime,
					player.BossKillOrExitTime,
				)
			}

			// Check saved exit time
			if player.LeaveDungeonTime != tt.expectedLeaveTime {
				t.Errorf(
					"Exit time: expected %d, got %d",
					tt.expectedLeaveTime,
					player.LeaveDungeonTime,
				)
			}
		})
	}
}
