package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Valery223/Dungeon/internal/domain"
	"github.com/Valery223/Dungeon/internal/infrastructure/io"
	"github.com/Valery223/Dungeon/internal/infrastructure/memory"
	"github.com/Valery223/Dungeon/internal/usecase"
)

// ConfigDTO - structure for loading configuration from json file
type ConfigDTO struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

func main() {
	// 1 Read flags
	configPath := flag.String("config", "config.json", "path to config file")
	eventsPath := flag.String("events", "events", "path to events log file")
	flag.Parse()

	// 2 Load configuration
	dungeonCfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 3 Open events file
	eventsFile, err := os.Open(*eventsPath)
	if err != nil {
		log.Fatalf("Failed to open events file: %v", err)
	}
	defer func() {
		if err := eventsFile.Close(); err != nil {
			log.Printf("Failed to close events file: %v", err)
		}
	}()

	// 4 Initialize infrastructure and application layers
	reader := io.NewStreamReader(eventsFile)
	writer := io.NewStreamWriter(os.Stdout) // Write to console

	// memory repository for storing player state in memory
	repo := memory.NewInMemoryPlayerRepo()
	processor := usecase.NewEventProcessor(dungeonCfg, repo)
	// runner - this is a facade that connects all components and starts the game
	runner := usecase.NewGameRunner(reader, writer, processor)

	// 5 Start
	if err := runner.Run(); err != nil {
		log.Fatalf("Game stopped with error: %v", err)
	}
}

// loadConfig loads configuration from JSON file and converts it to DungeonConfig
func loadConfig(path string) (*domain.DungeonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}

	var dto ConfigDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	openAtSec := parseTimeToSec(dto.OpenAt)
	durationSec := dto.Duration * 3600

	return &domain.DungeonConfig{
		Floors:      dto.Floors,
		Monsters:    dto.Monsters,
		OpenAtSec:   openAtSec,
		DurationSec: durationSec,
		CloseAtSec:  openAtSec + durationSec,
	}, nil
}

// parseTimeToSec converts string 'HH:MM:SS' to seconds
func parseTimeToSec(t string) int {
	var h, m, s int
	_, _ = fmt.Sscanf(t, "%d:%d:%d", &h, &m, &s)
	return h*3600 + m*60 + s
}
