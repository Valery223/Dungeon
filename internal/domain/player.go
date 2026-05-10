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

	BossEnterTime int
	BossKillTime  int
}

// ApplyEvent применяет событие "e" к состоянию игрока
// Использует настройки "cfg" для проверки ограничений (количество этажей, время закрытия)
// Возвращает ActionResult с решением о том, как "автомат" выполнил событие
func (p *Player) ApplyEvent(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	if p.Status == StatusDisqual || p.Status == StatusSuccess || p.Status == StatusFail {
		// Игнорируем
		return ActionResult{IsAccepted: false}
	}

	// Если данж закрылся, то игрок провалился(хотя не должно сюда попасть, так как мы не должны принимать события после закрытия данжа)
	// Должно раньше обработаться, но на всякий случай
	if e.TimeSec >= cfg.CloseAtSec {
		p.Status = StatusFail
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
		return p.leaveDungeon(e)
	case EventCannotContinue:
		return p.cannotContinue(e)
	case EventRestoreHP:
		return p.restoreHP(e)
	case EventReceiveDamage:
		return p.receiveDamage(e)
	default:
		return p.impossibleMove(e)
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
		p.FloorCleared[p.CurrentFloor] = true
		p.TimeSpentOnFloors[p.CurrentFloor] += e.TimeSec - p.CurrentFloorEnterTime
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
	// В зависимости от задания, пока убрал(не может уйти, пока не убьет всех монстров)
	// if !p.FloorCleared[p.CurrentFloor] {
	// 	p.TimeSpentOnFloors[p.CurrentFloor] += e.TimeSec - p.CurrentFloorEnterTime
	// }

	p.CurrentFloor++
	if p.CurrentFloor == cfg.Floors {
		p.FloorCleared[p.CurrentFloor] = true // На последнем этаже нет монстров, он сразу считается зачищенным
	} else {
		p.CurrentFloorEnterTime = e.TimeSec
	}
	return ActionResult{IsAccepted: true}
}

// prevFloor обрабатывает событие перехода на предыдущий этаж
func (p *Player) prevFloor(e IncomingEvent, cfg *DungeonConfig) ActionResult {
	// Проверяем, что игрок в данже, что это не первый этаж и не комната с боссом
	if p.Status != StatusInDungeon || p.CurrentFloor <= 1 || p.CurrentFloor > cfg.Floors {
		return p.impossibleMove(e)
	}

	// Если игрок пытается спуститься на предыдущий этаж, но комната не зачишена, то фиксируем время, проведенное на этаже
	if !p.FloorCleared[p.CurrentFloor] {
		p.TimeSpentOnFloors[p.CurrentFloor] += e.TimeSec - p.CurrentFloorEnterTime
	}

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
	p.BossKillTime = e.TimeSec - p.BossEnterTime
	return ActionResult{IsAccepted: true}
}

// leaveDungeon обрабатывает событие выхода из данжа
// переводит игрока в статус Success или Fail в зависимости от того, зачистил ли он все этажи и убил босса
func (p *Player) leaveDungeon(e IncomingEvent) ActionResult {
	if p.Status != StatusInDungeon {
		return p.impossibleMove(e)
	}

	p.LeaveDungeonTime = e.TimeSec
	allCleared := true
	for i := 1; i < len(p.FloorCleared); i++ {
		if !p.FloorCleared[i] {
			allCleared = false
			break
		}
	}

	if allCleared && p.BossDead {
		p.Status = StatusSuccess
	} else {
		p.Status = StatusFail
	}

	return ActionResult{IsAccepted: true}
}

// cannotContinue обрабатывает событие, когда игрок не может продолжать
func (p *Player) cannotContinue(e IncomingEvent) ActionResult {
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
func (p *Player) receiveDamage(e IncomingEvent) ActionResult {
	p.HP -= e.Value
	if p.HP <= 0 {
		p.HP = 0
		p.Status = StatusFail
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
