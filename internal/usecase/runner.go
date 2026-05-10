package usecase

import (
	"io"
)

type GameRunner struct {
	reader    EventReader
	writer    EventWriter
	processor *EventProcessor
}

func NewGameRunner(r EventReader, w EventWriter, p *EventProcessor) *GameRunner {
	return &GameRunner{
		reader:    r,
		writer:    w,
		processor: p,
	}
}

// Run запускает основной цикл обработки
func (r *GameRunner) Run() error {
	for {
		// 1 Читаем событие
		inEvent, err := r.reader.Read()
		if err != nil {
			if err == io.EOF {
				break // Поток завершен
			}
			return err // Ошибка чтения
		}

		// 2 Бизнес логика обрабатывает событие
		result := r.processor.ProcessEvent(inEvent)

		// 3 Выводим результаты
		// Если действие было принято, то выводим его
		if result.IsAccepted {
			r.writer.WriteAccepted(inEvent)
		}
		// Если есть исходящее событие, то выводим его
		if result.OutgoingEvent != nil {
			r.writer.WriteOutgoing(*result.OutgoingEvent)
		}
	}

	// 4 Поток завершен. Генерируем и пишем финальный отчет
	reports := r.processor.GenerateFinalReport()
	return r.writer.WriteReport(reports)
}
