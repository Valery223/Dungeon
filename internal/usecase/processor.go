package usecase

import (
	"github.com/Valery223/Dungeon/internal/domain"
)

// ReportEntry - structure for final report
type ReportEntry struct {
	Status       domain.PlayerStatus
	PlayerID     int
	TotalTime    int
	AvgFloorTime int
	BossKillTime int
	HP           int
}

type EventProcessor struct {
	cfg  *domain.DungeonConfig
	repo PlayerRepository
}

func NewEventProcessor(cfg *domain.DungeonConfig, repo PlayerRepository) *EventProcessor {
	return &EventProcessor{
		cfg:  cfg,
		repo: repo,
	}
}

func (p *EventProcessor) ProcessEvent(e domain.IncomingEvent) domain.ActionResult {
	player := p.repo.Get(e.PlayerID)

	if player == nil {
		player = domain.NewPlayer(e.PlayerID, p.cfg, 100)
		p.repo.Save(player)
	}
	return player.ApplyEvent(e, p.cfg)
}

func (p *EventProcessor) GenerateFinalReport() []ReportEntry {
	players := p.repo.GetAllOrdered()
	reports := make([]ReportEntry, 0, len(players))

	// Force close dungeon for all players who are still in it
	// Players receive Fail status
	for _, player := range players {
		player.ForceClose(p.cfg.CloseAtSec, p.cfg)
	}

	// Form report for all players in order of their first appearance
	for _, player := range players {
		reports = append(reports, ReportEntry{
			Status:       player.Status,
			PlayerID:     player.ID,
			TotalTime:    player.TotalDungeonTime(),
			AvgFloorTime: player.AvgFloorTime(p.cfg),
			BossKillTime: player.FinalBossTime(),
			HP:           player.HP,
		})
	}

	return reports
}
