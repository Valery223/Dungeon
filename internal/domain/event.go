package domain

// Здесь описаны сущности событий (входящие и исходящие)
type EventID int

// Константы для идентификаторов событий
const (
	// Incoming events
	EventRegister       EventID = 1  // Регистрация игрока
	EventEnterDungeon   EventID = 2  // Вход в данж
	EventKillMonster    EventID = 3  // Убийство монстра
	EventNextFloor      EventID = 4  // Переход на следующий этаж
	EventPrevFloor      EventID = 5  // Переход на предыдущий этаж
	EventEnterBoss      EventID = 6  // Вход в комнату с боссом
	EventKillBoss       EventID = 7  // Убийство босса
	EventLeaveDungeon   EventID = 8  // Выход из данжа
	EventCannotContinue EventID = 9  // Невозможность продолжить
	EventRestoreHP      EventID = 10 // Восстановление HP
	EventReceiveDamage  EventID = 11 // Получение урона

	// Outgoing events
	EventOutDisqualified EventID = 31 // Игрок дисквалифицирован
	EventOutDead         EventID = 32 // Игрок умер
	EventOutImpossible   EventID = 33 // Невозможное действие
)

// IncomingEvent представляет событие, которое приходит от игрока
type IncomingEvent struct {
	ID       EventID
	TimeSec  int
	PlayerID int
	Value    int    // распарсенное из ExtraParam (урон, хп)
	Extra    string // оригинальный ExtraParam
}

// OutgoingEvent представляет исходящее событие
type OutgoingEvent struct {
	ID              EventID
	IncomingEventID EventID
	TimeSec         int
	PlayerID        int
	ExtraParam      string
}
