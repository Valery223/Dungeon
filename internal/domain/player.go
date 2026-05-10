// Здесь описаны сущности игрока, правила обработки событий и конечный автомат
package domain

import "fmt"

// PlayerStatus - определяет фазу игрока в процессе прохождения данжа
type PlayerStatus int

// Возможные состояния игрока (FSM States)
const (
	StatusNew        PlayerStatus = iota // Игрок еще не зарегистрирован
	StatusRegistered                     // Зарегистрирован
	StatusInDungeon                      // В данже
	StatusSuccess                        // Успешно прошел данж
	StatusFail                           // Провалился (умер, вышел раньше времени, закончилось время и т.д.)
	StatusDisqual                        // Дисквалифицирован
)

// ActionResult описывает ответ "конечного автомата" на входящее событие
type ActionResult struct {
	IsAccepted    bool           // Указывает, валидно ли действие с точки зрения правил
	OutgoingEvent *OutgoingEvent // Содержит исходящии события  (31, 32, 33) если они произошли, иначе nil
}

// Player храннит данные о текущем прохождении игрока
type Player struct {
	ID     int
	Status PlayerStatus
	HP     int

	CurrentFloor int
	BossDead     bool

	// Для каждого этажа храним, сколько монстров осталось и зачищен ли этаж
	MonstersLeft []int
	FloorCleared []bool

	// Временные метки для расчета времени прохождения
	EnterDungeonTime int
	LeaveDungeonTime int

	// Для каждого этажа храним, сколько времени игрок провел на нем
	CurrentFloorEnterTime int
	TimeSpentOnFloors     []int

	BossEnterTime      int
	BossKillOrExitTime int
}

// NewPlayer - создние  доменной сущности
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

// ApplyEvent применяет событие "e" к состоянию игрока
// Использует настройки "cfg" для проверки ограничений (количество этажей, время закрытия)
// Возвращает ActionResult с решением о том, как "автомат" выполнил событие
func (p *Player) ApplyEvent(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status == StatusDisqual || p.Status == StatusSuccess || p.Status == StatusFail {
		// Игнорируем
		return ActionResult{IsAccepted: false}
	}

	// Если игрок пытается что-то сделать после закрытия данжа, то  мы его принудительно закрываем
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

// ForceClose принудительно закрывает данж для игрока, переводя его в статус Fail, если он еще не завершил испытание
func (p *Player) ForceClose(closeTimeSec int, cfg *DungeonConfig) {
	// Если игрок был в данже, то фиксируем время, проведенное на последнем этаже
	if p.Status == StatusInDungeon {
		p.accumulateUnclearedTime(closeTimeSec, cfg)
		p.LeaveDungeonTime = closeTimeSec
	}

	// Если игрок еще не завершил испытание, то он проваливается
	if p.Status == StatusInDungeon || p.Status == StatusRegistered || p.Status == StatusNew {
		p.Status = StatusFail
	}

}

// Обработчики конкретных состояний

// register переводит игрока в статус StatusRegistered
// Возвращает событие (Impossible Move), если игрок не  в статусе StatusNew.
func (p *Player) register(e IncomingEvent) ActionResult {
	// Игрок может зарегистрироваться только если он новый
	if p.Status != StatusNew {
		return p.impossibleMove(e)
	}

	p.Status = StatusRegistered

	return ActionResult{IsAccepted: true}
}

// enterDungeon переводит игрока в статус StatusInDungeon, фиксирует время входа и текущий этаж
// Использует "cfg" для проверки ограничений (время открытия)
func (p *Player) enterDungeon(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Если игрок уже в данже, то он не может войти снова
	if p.Status == StatusInDungeon {
		return p.impossibleMove(e)
	}

	// Игрок может войти в данж только если он зарегистрирован
	if p.Status != StatusRegistered {
		p.Status = StatusDisqual
		return p.disqualify(e)
	}

	// Если игрок пытается войти в данж до его открытия, то он дисквалифицируется
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

// killMonster обрабатывает событие убийства монстра
func (p *Player) killMonster(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Проверяем,  что игрок в данже, что это не последний этаж с боссом и что на этаже есть монстры
	if p.Status != StatusInDungeon || p.CurrentFloor >= cfg.Floors || p.MonstersLeft[p.CurrentFloor] == 0 {
		return p.impossibleMove(e)
	}

	// Убовляем монстра
	p.MonstersLeft[p.CurrentFloor]--

	// Если монстров не осталось, то фиксируем время прохождения этажа
	if p.MonstersLeft[p.CurrentFloor] == 0 {
		p.completeCurrentFloor(e.TimeSec, cfg)
	}

	return ActionResult{IsAccepted: true}
}

// nextFloor обрабатывает событие перехода на следующий этаж
func (p *Player) nextFloor(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Проверяем, что игрок в данже, что это не последний этаж и что комната зачищена
	if p.Status != StatusInDungeon || p.CurrentFloor >= cfg.Floors || !p.FloorCleared[p.CurrentFloor] {
		return p.impossibleMove(e)
	}

	// Если игрок пытается подняться на следующий этаж, но комната не зачишена, то фиксируем время, проведенное на этаже
	// Если в будущем можно будет переходить на следующий этаж, не убив всех монстров
	p.accumulateUnclearedTime(e.TimeSec, cfg)

	p.CurrentFloor++
	if p.CurrentFloor == cfg.Floors {
		p.FloorCleared[p.CurrentFloor] = true // На последнем этаже нет монстров, он сразу считается зачищенным
	}
	p.CurrentFloorEnterTime = e.TimeSec

	return ActionResult{IsAccepted: true}
}

// prevFloor обрабатывает событие перехода на предыдущий этаж
func (p *Player) prevFloor(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Проверяем, что игрок в данже, что это не первый этаж и не комната с боссом
	if p.Status != StatusInDungeon || p.CurrentFloor <= 1 || p.CurrentFloor > cfg.Floors {
		return p.impossibleMove(e)
	}

	// Если игрок пытается спуститься на предыдущий этаж, но комната не зачишена, то фиксируем время, проведенное на этаже
	p.accumulateUnclearedTime(e.TimeSec, cfg)

	p.CurrentFloor--
	p.CurrentFloorEnterTime = e.TimeSec

	return ActionResult{IsAccepted: true}
}

// enterBoss обрабатывает событие входа в комнату с боссом
func (p *Player) enterBoss(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status != StatusInDungeon || p.CurrentFloor != cfg.Floors {
		return p.impossibleMove(e)
	}

	p.CurrentFloor = cfg.Floors + 1 // Этаж босса
	p.BossEnterTime = e.TimeSec
	return ActionResult{IsAccepted: true}
}

// killBoss обрабатывает событие убийства босса
func (p *Player) killBoss(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status != StatusInDungeon || p.CurrentFloor != cfg.Floors+1 || p.BossDead {
		return p.impossibleMove(e)
	}

	p.BossDead = true

	p.completeCurrentFloor(e.TimeSec, cfg)

	return ActionResult{IsAccepted: true}
}

// leaveDungeon обрабатывает событие выхода из данжа
// переводит игрока в статус Success или Fail в зависимости от того, зачистил ли он все этажи и убил босса
func (p *Player) leaveDungeon(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status != StatusInDungeon {
		return p.impossibleMove(e)
	}

	// Если игрок выходит и что-то не зачишено, то фиксируем время незачищенных этажей
	p.accumulateUnclearedTime(e.TimeSec, cfg)
	// Фиксируем время выхода
	p.LeaveDungeonTime = e.TimeSec

	// Проверяем, все ли этажи зачишены и убит ли босс
	allCleared := true
	for i := 1; i < len(p.FloorCleared); i++ {
		if !p.FloorCleared[i] {
			allCleared = false
			break
		}
	}

	if allCleared && p.BossDead {
		// Если все этажи зачишены и босс убит, то игрок успешно прошел данж
		p.Status = StatusSuccess
	} else {
		// Иначе он провалился
		p.Status = StatusFail
	}

	return ActionResult{IsAccepted: true}
}

// cannotContinue обрабатывает событие, когда игрок не может продолжать
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
		IsAccepted: true,
	}
}

// restoreHP обрабатывает событие восстановления HP
func (p *Player) restoreHP(e IncomingEvent) ActionResult {
	p.HP += e.Value
	if p.HP > 100 {
		p.HP = 100
	}

	return ActionResult{
		IsAccepted: true,
	}
}

// receiveDamage обрабатывает событие получения урона
// При смерти переводит игрока в статус Fail и генерирует событие смерти
func (p *Player) receiveDamage(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	p.HP -= e.Value
	if p.HP <= 0 {
		p.HP = 0

		p.Status = StatusFail

		// time
		if p.CurrentFloor < cfg.Floors {
			// Если игрок умер на обычном этаже, то фиксируем время, проведенное на нем
			p.TimeSpentOnFloors[p.CurrentFloor] += e.TimeSec - p.CurrentFloorEnterTime
		} else if p.CurrentFloor == cfg.Floors+1 {
			// Если игрок умер в комнате с боссом, то фиксируем время, проведенное на ней
			p.BossKillOrExitTime = e.TimeSec - p.BossEnterTime
		}
		p.LeaveDungeonTime = e.TimeSec
		return p.dead(e)
	}

	return ActionResult{
		IsAccepted: true,
	}
}

// Вспомогательные методы

// disqualify - событие дисквалификации игрока
// outgoing event 31
func (p *Player) disqualify(e IncomingEvent) ActionResult {
	return ActionResult{
		IsAccepted:    false,
		OutgoingEvent: p.buildEvent(e, EventOutDisqualified, "")}
}

// dead - событие смерти игрока
// outgoing event 32
func (p *Player) dead(e IncomingEvent) ActionResult {
	return ActionResult{
		IsAccepted:    true,
		OutgoingEvent: p.buildEvent(e, EventOutDead, ""),
	}
}

// impossibleMove - событие невозможного действия
// outgoing event 33
func (p *Player) impossibleMove(e IncomingEvent) ActionResult {
	return ActionResult{
		IsAccepted:    false,
		OutgoingEvent: p.buildEvent(e, EventOutImpossible, fmt.Sprintf("%d", e.ID)),
	}
}

// buildEvent - вспомогательный метод для генерации исходящего события
// Создает OutgoingEvent с типом "outID" и опциональным параметром "extra"
func (p *Player) buildEvent(e IncomingEvent, outID EventID, extra string) *OutgoingEvent {
	return &OutgoingEvent{
		TimeSec:         e.TimeSec,
		PlayerID:        p.ID,
		ID:              outID,
		IncomingEventID: e.ID,
		ExtraParam:      extra,
	}
}

// completeCurrentFloor вызывается когда убит последний монстр или босс
// Он фиксирует итоговое время и помечает этаж как пройденный
func (p *Player) completeCurrentFloor(currentTime int, cfg *DungeonConfig) {
	if p.CurrentFloor == cfg.Floors+1 {
		// Логика для босса
		p.BossDead = true
		p.BossKillOrExitTime += currentTime - p.BossEnterTime
	} else {
		// Логика для обычного этажа
		p.FloorCleared[p.CurrentFloor] = true
		p.TimeSpentOnFloors[p.CurrentFloor] += currentTime - p.CurrentFloorEnterTime
	}
}

// accumulateUnclearedTime вызывается при прерывании: смена этажа, выход, смерть,
// Добавляет прошедшее время в копилку этажа, только если он еще не зачищен
func (p *Player) accumulateUnclearedTime(currentTime int, cfg *DungeonConfig) {
	if p.Status != StatusInDungeon {
		return
	}

	if p.CurrentFloor == cfg.Floors+1 {
		if !p.BossDead {
			p.BossKillOrExitTime += currentTime - p.BossEnterTime
		}
	} else {
		// Добавляем время только если на этаже еще остались монстры
		if !p.FloorCleared[p.CurrentFloor] {
			p.TimeSpentOnFloors[p.CurrentFloor] += currentTime - p.CurrentFloorEnterTime
		}
	}
}

// Публичные методы для расчета итогового отчета
func (p *Player) TotalDungeonTime() int {
	if p.EnterDungeonTime == 0 || p.LeaveDungeonTime == 0 {
		return 0
	}
	return p.LeaveDungeonTime - p.EnterDungeonTime
}

func (p *Player) AvgFloorTime(cfg *DungeonConfig) int {
	sum := 0
	clearedCount := 0

	// Считаем только для пройденных этажей
	for i := 1; i <= cfg.Floors; i++ {
		if p.FloorCleared[i] {
			sum += p.TimeSpentOnFloors[i]
			clearedCount++
		}
	}

	// По заданию среднее время можно трактовать по разному
	// Буду считать, что если игрок не зачистил все этажи, то среднее время 0, так как он не выполнил условие для получения результата
	if clearedCount != cfg.Floors {
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
