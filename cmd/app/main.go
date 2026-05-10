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

// ConfigDTO - структура для загрузки конфигурации из json файла
type ConfigDTO struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

func main() {
	// 1 Читаем флаги
	configPath := flag.String("config", "config.json", "path to config file")
	eventsPath := flag.String("events", "events", "path to events log file")
	flag.Parse()

	// 2 Загружаем конфигурацию
	dungeonCfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 3 Открываем файл с событиями
	eventsFile, err := os.Open(*eventsPath)
	if err != nil {
		log.Fatalf("Failed to open events file: %v", err)
	}
	defer eventsFile.Close()

	// 4 Инициализация слоев инфраструктуры и приложения
	reader := io.NewStreamReader(eventsFile)
	writer := io.NewStreamWriter(os.Stdout) // Пишем в консоль

	// memory репозиторий для хранения состояния игроков в памяти
	repo := memory.NewInMemoryPlayerRepo()
	processor := usecase.NewEventProcessor(dungeonCfg, repo)
	// runner - это фасад, который связывает все компоненты и запускает игру
	runner := usecase.NewGameRunner(reader, writer, processor)

	// 5 Запуск
	if err := runner.Run(); err != nil {
		log.Fatalf("Game stopped with error: %v", err)
	}
}

// loadConfig загружает конфигурацию из JSON файла и преобразует ее в DungeonConfig
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

// parseTimeToSec переводит строку "HH:MM:SS" в секунды
func parseTimeToSec(t string) int {
	var h, m, s int
	_, _ = fmt.Sscanf(t, "%d:%d:%d", &h, &m, &s)
	return h*3600 + m*60 + s
}
