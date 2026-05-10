package domain

// DungeonConfig - dungeon configuration
type DungeonConfig struct {
	Floors      int
	Monsters    int
	OpenAtSec   int
	DurationSec int
	CloseAtSec  int
}
