package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
)

type EventType byte

const (
	_                     = iota // 0
	EventDelete EventType = iota // 1
	EventPut                     // 2
)

// Event - описывает конкретное событие
type Event struct {
	Sequence  uint64    // Уникальный порядковый номер записи
	EventType EventType //Выполненое действие
	Key       string    // Ключ, затронутый транзакцией
	Value     string    // Значение для транзакции PUT
}

// TransactionLogger - методы для добавления в журнал транзакций
type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error

	ReadEvents() (<-chan Event, <-chan error)
	Run()
}

// Реализуем интерфейс

type FileTransactionLogger struct {
	events       chan<- Event // Канал только для записи; для передачи событий
	errors       <-chan error // Калан только для чтения; для приема ошибок
	lastSequence uint64       // Последний используемый порядковый номер
	file         *os.File     // Местоположение файла журнала
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

// NewFileTransactionLogger создает экземпляр TransactionLogger
// возвращаем интерфейс, т.к. в данном случае возможена  фабрика
func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction file: %w", err)
	}
	return &FileTransactionLogger{file: file}, nil
}

// Run добавляет записи в конец журнала
func (l *FileTransactionLogger) Run() {
	// Создаем буфиризированный канал для контурентной записис в файл транзакций
	// Буферизированный канал позволит службе потребителю обработать короткие всплески событий без замедления из за дискового io
	// Если буфер заполнится, то методы записи будут блокироватся до момента, когда горутина записи освободит в нем место
	events := make(chan Event, 16)
	l.events = events

	// Канал ошибок
	// Буфет 1 позволит отправлять ошибки без блокировок
	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		// Извлекаем слудующее событие
		for e := range events {
			// Увеличиваем порядковый номер
			l.lastSequence++
			// Записываем событие в журнал
			_, err := fmt.Fprintf(
				l.file,
				"%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

// ReadEvents прочитать события из файла
func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event

		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			// Читаем построчно файл
			line := scanner.Text()

			fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value)
			// Если последнего элемента нет в строке, то не пажаем с EOF, а заменяем пустой срокой
			uv, err := url.QueryUnescape(e.Value)
			if err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}
			e.Value = uv

			// Проверка целостности, порядковый номер последовательно увеличивается?
			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction number out of sequence")
				return
			}
			// Запоминаем порядковый номер
			l.lastSequence = e.Sequence
			outEvent <- e
		}
		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()
	return outEvent, outError
}
