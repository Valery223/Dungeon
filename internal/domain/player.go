// This file describes player entities, event handling rules and the finite state machine
package domain

import "fmt"

// PlayerStatus - determines the phase of a player progressing through the dungeon
type PlayerStatus int

// Possible player states (FSM States)
const (
	StatusNew        PlayerStatus = iota // Player not yet registered
	StatusRegistered                     // Registered
	StatusInDungeon                      // In dungeon
	StatusSuccess                        // Successfully cleared dungeon
	StatusFail                           // Failed (died, left early, time ran out, etc.)
	StatusDisqual                        // Disqualified
)

// ActionResult describes the response of the finite state machine to an incoming event
type ActionResult struct {
	IsAccepted    bool           // Indicates whether the action is valid according to rules
	OutgoingEvent *OutgoingEvent // Contains outgoing events (31, 32, 33) if they occurred, otherwise nil
}

// Player stores data about a player's current dungeon run
type Player struct {
	ID     int
	Status PlayerStatus
	HP     int

	CurrentFloor int
	BossDead     bool

	// For each floor, we store how many monsters remain and if the floor is cleared
	MonstersLeft []int
	FloorCleared []bool

	// Time stamps for calculating play time
	EnterDungeonTime int
	LeaveDungeonTime int

	// For each floor, we store how much time the player spent on it
	CurrentFloorEnterTime int
	TimeSpentOnFloors     []int

	BossEnterTime      int
	BossKillOrExitTime int
}

// NewPlayer - creates a domain entity
func NewPlayer(id int, cfg *DungeonConfig, hp int) *Player {
	p := &Player{
		ID:                id,
		Status:            StatusNew,
		HP:                hp,
		MonstersLeft:      make([]int, cfg.Floors+2),
		FloorCleared:      make([]bool, cfg.Floors+1),
		TimeSpentOnFloors: make([]int, cfg.Floors+2),
	}

	for i := 1; i <= cfg.Floors; i++ {
		p.MonstersLeft[i] = cfg.Monsters
	}
	return p
}

// ApplyEvent applies event "e" to the player state
// Uses settings "cfg" to verify constraints (number of floors, close time)
// Returns ActionResult with decision on how the "state machine" executed the event
func (p *Player) ApplyEvent(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status == StatusDisqual || p.Status == StatusSuccess || p.Status == StatusFail {
		// Ignore
		return ActionResult{IsAccepted: false}
	}

	// If player tries to do something after dungeon closes, we forcefully close it
	if e.TimeSec >= cfg.CloseAtSec {
		p.ForceClose(cfg.CloseAtSec, cfg)
		return ActionResult{IsAccepted: false}
	}

	switch e.ID {
	case EventRegister:
		return p.register(e)
	case EventEnterDungeon:
		return p.enterDungeon(e, cfg)
	case EventKillMonster:
		return p.killMonster(e, cfg)
	case EventNextFloor:
		return p.nextFloor(e, cfg)
	case EventPrevFloor:
		return p.prevFloor(e, cfg)
	case EventEnterBoss:
		return p.enterBoss(e, cfg)
	case EventKillBoss:
		return p.killBoss(e, cfg)
	case EventLeaveDungeon:
		return p.leaveDungeon(e, cfg)
	case EventCannotContinue:
		return p.cannotContinue(e, cfg)
	case EventRestoreHP:
		return p.restoreHP(e)
	case EventReceiveDamage:
		return p.receiveDamage(e, cfg)
	default:
		return p.impossibleMove(e)
	}
}

// ForceClose forcefully closes the dungeon for a player, changing status to Fail if not already finished
func (p *Player) ForceClose(closeTimeSec int, cfg *DungeonConfig) {
	// If player was in dungeon, fix time spent on last floor
	if p.Status == StatusInDungeon {
		p.accumulateUnclearedTime(closeTimeSec, cfg)
		p.LeaveDungeonTime = closeTimeSec
	}

	// If player hasn't finished the challenge, they fail
	if p.Status == StatusInDungeon || p.Status == StatusRegistered || p.Status == StatusNew {
		p.Status = StatusFail
	}

}

// Handlers for specific states

// register transitions player to StatusRegistered status
// Returns an event (Impossible Move) if player is not in StatusNew status.
func (p *Player) register(e IncomingEvent) ActionResult {
	// Player can only register if they are new
	if p.Status != StatusNew {
		return p.impossibleMove(e)
	}

	p.Status = StatusRegistered

	return ActionResult{IsAccepted: true}
}

// enterDungeon transitions player to StatusInDungeon status, fixes entry time and current floor
// Uses "cfg" to verify constraints (open time)
func (p *Player) enterDungeon(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// If player is already in dungeon, they cannot enter again
	if p.Status == StatusInDungeon {
		return p.impossibleMove(e)
	}

	// Player can enter dungeon only if they are registered
	if p.Status != StatusRegistered {
		p.Status = StatusDisqual
		return p.disqualify(e)
	}

	// If player tries to enter dungeon before it opens, they are disqualified
	if e.TimeSec < cfg.OpenAtSec {
		p.Status = StatusDisqual
		return p.disqualify(e)
	}

	p.Status = StatusInDungeon
	p.CurrentFloor = 1
	p.EnterDungeonTime = e.TimeSec
	p.CurrentFloorEnterTime = e.TimeSec

	return ActionResult{IsAccepted: true}
}

// killMonster handles the event of killing a monster
func (p *Player) killMonster(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Check that player is in dungeon, not on last floor with boss, and there are monsters on the floor
	if p.Status != StatusInDungeon || p.CurrentFloor >= cfg.Floors || p.MonstersLeft[p.CurrentFloor] == 0 {
		return p.impossibleMove(e)
	}

	// Remove a monster
	p.MonstersLeft[p.CurrentFloor]--

	// If no monsters left, fix time spent on floor
	if p.MonstersLeft[p.CurrentFloor] == 0 {
		p.completeCurrentFloor(e.TimeSec, cfg)
	}

	return ActionResult{IsAccepted: true}
}

// nextFloor handles the event of moving to the next floor
func (p *Player) nextFloor(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Check that player is in dungeon, that it's not the last floor, and that the room is cleared
	if p.Status != StatusInDungeon || p.CurrentFloor >= cfg.Floors || !p.FloorCleared[p.CurrentFloor] {
		return p.impossibleMove(e)
	}

	// If player tries to move to the next floor, fix time spent on current floor if not yet cleared
	// In the future, we may allow moving to the next floor without clearing all monsters
	p.accumulateUnclearedTime(e.TimeSec, cfg)

	p.CurrentFloor++
	if p.CurrentFloor == cfg.Floors {
		p.FloorCleared[p.CurrentFloor] = true // The last floor has no monsters, it is immediately considered cleared
	}
	p.CurrentFloorEnterTime = e.TimeSec

	return ActionResult{IsAccepted: true}
}

// prevFloor handles the event of moving to the previous floor
func (p *Player) prevFloor(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Check that player is in dungeon, that it's not the first floor and not the boss room
	if p.Status != StatusInDungeon || p.CurrentFloor <= 1 || p.CurrentFloor > cfg.Floors {
		return p.impossibleMove(e)
	}

	// If player tries to move to the previous floor, fix time spent on current floor if not yet cleared
	p.accumulateUnclearedTime(e.TimeSec, cfg)

	p.CurrentFloor--
	p.CurrentFloorEnterTime = e.TimeSec

	return ActionResult{IsAccepted: true}
}

// enterBoss handles the event of entering the boss room
func (p *Player) enterBoss(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status != StatusInDungeon || p.CurrentFloor != cfg.Floors {
		return p.impossibleMove(e)
	}

	p.CurrentFloor = cfg.Floors + 1 // Boss floor
	p.BossEnterTime = e.TimeSec
	return ActionResult{IsAccepted: true}
}

// killBoss handles the event of killing the boss
func (p *Player) killBoss(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status != StatusInDungeon || p.CurrentFloor != cfg.Floors+1 || p.BossDead {
		return p.impossibleMove(e)
	}

	p.BossDead = true

	p.completeCurrentFloor(e.TimeSec, cfg)

	return ActionResult{IsAccepted: true}
}

// leaveDungeon handles the event of leaving the dungeon
// transitions player to Success or Fail status depending on whether they cleared all floors and killed the boss
func (p *Player) leaveDungeon(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status != StatusInDungeon {
		return p.impossibleMove(e)
	}

	// If player exits and something is not cleared, fix time for uncleared floors
	p.accumulateUnclearedTime(e.TimeSec, cfg)
	// Fix exit time
	p.LeaveDungeonTime = e.TimeSec

	// Check if all floors are cleared and boss is killed
	allCleared := true
	for i := 1; i < len(p.FloorCleared); i++ {
		if !p.FloorCleared[i] {
			allCleared = false
			break
		}
	}

	if allCleared && p.BossDead {
		// If all floors are cleared and boss is killed, player successfully completed the dungeon
		p.Status = StatusSuccess
	} else {
		// Otherwise they failed
		p.Status = StatusFail
	}

	return ActionResult{IsAccepted: true}
}

// cannotContinue handles the event when player cannot continue
func (p *Player) cannotContinue(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// time
	if p.Status == StatusInDungeon {
		if p.CurrentFloor < cfg.Floors && !p.FloorCleared[p.CurrentFloor] {
			p.TimeSpentOnFloors[p.CurrentFloor] += e.TimeSec - p.CurrentFloorEnterTime
		} else if p.CurrentFloor == cfg.Floors+1 && !p.BossDead {
			p.BossKillOrExitTime = e.TimeSec - p.BossEnterTime
		}
		p.LeaveDungeonTime = e.TimeSec
	}
	p.Status = StatusDisqual
	return ActionResult{
		IsAccepted:    true,
		OutgoingEvent: p.buildEvent(e, EventOutDisqualified, e.Extra),
	}
}

// restoreHP handles the HP restoration event
func (p *Player) restoreHP(e IncomingEvent) ActionResult {
	p.HP += e.Value
	if p.HP > 100 {
		p.HP = 100
	}

	return ActionResult{
		IsAccepted: true,
	}
}

// receiveDamage handles the damage reception event
// On death, transitions player to Fail status and generates death event
func (p *Player) receiveDamage(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	p.HP -= e.Value
	if p.HP <= 0 {
		p.HP = 0

		if p.Status == StatusInDungeon {
			// time
			if p.CurrentFloor < cfg.Floors {
				// If player died on a regular floor, fix time spent on it
				p.TimeSpentOnFloors[p.CurrentFloor] += e.TimeSec - p.CurrentFloorEnterTime
			} else if p.CurrentFloor == cfg.Floors+1 {
				// If player died in the boss room, fix time spent on it
				p.BossKillOrExitTime = e.TimeSec - p.BossEnterTime
			}
		}
		p.Status = StatusFail
		p.LeaveDungeonTime = e.TimeSec
		return p.dead(e)
	}

	return ActionResult{
		IsAccepted: true,
	}
}

// Helper methods

// disqualify - player disqualification event
// outgoing event 31
func (p *Player) disqualify(e IncomingEvent) ActionResult {
	return ActionResult{
		IsAccepted:    false,
		OutgoingEvent: p.buildEvent(e, EventOutDisqualified, "")}
}

// dead - player death event
// outgoing event 32
func (p *Player) dead(e IncomingEvent) ActionResult {
	return ActionResult{
		IsAccepted:    true,
		OutgoingEvent: p.buildEvent(e, EventOutDead, ""),
	}
}

// impossibleMove - impossible action event
// outgoing event 33
func (p *Player) impossibleMove(e IncomingEvent) ActionResult {
	return ActionResult{
		IsAccepted:    false,
		OutgoingEvent: p.buildEvent(e, EventOutImpossible, fmt.Sprintf("%d", e.ID)),
	}
}

// buildEvent - helper method for generating outgoing event
// Creates OutgoingEvent with type "outID" and optional parameter "extra"
func (p *Player) buildEvent(e IncomingEvent, outID EventID, extra string) *OutgoingEvent {
	return &OutgoingEvent{
		TimeSec:         e.TimeSec,
		PlayerID:        p.ID,
		ID:              outID,
		IncomingEventID: e.ID,
		ExtraParam:      extra,
	}
}

// completeCurrentFloor is called when the last monster or boss is killed
// It fixes the final time and marks the floor as cleared
func (p *Player) completeCurrentFloor(currentTime int, cfg *DungeonConfig) {
	if p.CurrentFloor == cfg.Floors+1 {
		// Boss logic
		p.BossDead = true
		p.BossKillOrExitTime += currentTime - p.BossEnterTime
	} else {
		// Regular floor logic
		p.FloorCleared[p.CurrentFloor] = true
		p.TimeSpentOnFloors[p.CurrentFloor] += currentTime - p.CurrentFloorEnterTime
	}
}

// accumulateUnclearedTime is called when interrupted: floor change, exit, death
// Adds elapsed time to the floor's time bank, only if it hasn't been cleared yet
func (p *Player) accumulateUnclearedTime(currentTime int, cfg *DungeonConfig) {
	if p.Status != StatusInDungeon {
		return
	}

	if p.CurrentFloor == cfg.Floors+1 {
		if !p.BossDead {
			p.BossKillOrExitTime += currentTime - p.BossEnterTime
		}
	} else {
		// Add time only if monsters still remain on the floor
		if !p.FloorCleared[p.CurrentFloor] {
			p.TimeSpentOnFloors[p.CurrentFloor] += currentTime - p.CurrentFloorEnterTime
		}
	}
}

// Public methods for calculating final report
func (p *Player) TotalDungeonTime() int {
	if p.EnterDungeonTime == 0 || p.LeaveDungeonTime == 0 {
		return 0
	}
	return p.LeaveDungeonTime - p.EnterDungeonTime
}

func (p *Player) AvgFloorTime(cfg *DungeonConfig) int {
	sum := 0
	clearedCount := 0

	// Count only for cleared floors
	// Start from 1, as floor 0 is the lobby, and floor cfg.Floors is an empty room, not included in the calculation
	for i := 1; i < cfg.Floors; i++ {
		if p.FloorCleared[i] {
			sum += p.TimeSpentOnFloors[i]
			clearedCount++
		}
	}

	// According to the task, average time can be interpreted in different ways
	// I count that if player didn't clear all floors, average time is 0, as they didn't meet the condition for getting a result
	if clearedCount != cfg.Floors-1 {
		return 0
	}

	if clearedCount > 0 {
		return sum / clearedCount
	}
	return 0
}

func (p *Player) FinalBossTime() int {
	if p.BossDead {
		return p.BossKillOrExitTime
	}
	return 0
}
