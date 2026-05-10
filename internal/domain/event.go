package domain

// Event entities are described here (incoming and outgoing)
type EventID int

// Constants for event identifiers
const (
	// Incoming events
	EventRegister       EventID = 1  // Player registration
	EventEnterDungeon   EventID = 2  // Entering the dungeon
	EventKillMonster    EventID = 3  // Killing a monster
	EventNextFloor      EventID = 4  // Moving to next floor
	EventPrevFloor      EventID = 5  // Moving to previous floor
	EventEnterBoss      EventID = 6  // Entering boss room
	EventKillBoss       EventID = 7  // Killing the boss
	EventLeaveDungeon   EventID = 8  // Leaving dungeon
	EventCannotContinue EventID = 9  // Cannot continue
	EventRestoreHP      EventID = 10 // HP restoration
	EventReceiveDamage  EventID = 11 // Receiving damage

	// Outgoing events
	EventOutDisqualified EventID = 31 // Player disqualified
	EventOutDead         EventID = 32 // Player died
	EventOutImpossible   EventID = 33 // Impossible action
)

// IncomingEvent represents an event coming from the player
type IncomingEvent struct {
	ID       EventID
	TimeSec  int
	PlayerID int
	Value    int    // parsed from ExtraParam (damage, hp)
	Extra    string // original ExtraParam
}

// OutgoingEvent represents an outgoing event
type OutgoingEvent struct {
	ID              EventID
	IncomingEventID EventID
	TimeSec         int
	PlayerID        int
	ExtraParam      string
}
