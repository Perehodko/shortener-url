package cons_prod

import (
	"bufio"
	"encoding/json"
	"os"
)

type Producer interface {
	NewProducer(filename string) (*producer, error)
	WriteEvent(event *Event)
	Close() error
}

type Consumer interface {
	NewConsumer(filename string) (*consumer, error)
	ReadEvent() (*Event, error)
	Close() error
}

type Event struct {
	//ID       int    `json:"Id"`
	ShortURL string `json:"ShortURL"`
}

type producer struct {
	file *os.File
	// добавляем writer в Producerr
	writer *bufio.Writer
}

func NewProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	return &producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *producer) WriteEvent(event Event) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

func (p *producer) Close() error {
	return p.file.Close()
}

type consumer struct {
	file *os.File
	// добавляем reader в Consumer1
	reader *bufio.Reader
}

func NewConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &consumer{
		file: file,
		// создаём новый Reader
		reader: bufio.NewReader(file),
	}, nil
}

func (c *consumer) ReadEvent() (*Event, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	//data, err := c.reader.ReadBytes()

	if err != nil {
		return nil, err
	}

	// преобразуем данные из JSON-представления в структуру
	event := Event{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *consumer) Close() error {
	return c.file.Close()
}

//var events = Event{
//	ShortURL: "lalala19",
//}
//
//func main() {
//	fileName := "/Users/nperekhodko/Desktop/I/yandex_courses_go/coomon_go/ya_practicum/five_inc/cons_prod/events.log"
//
//	//defer os.Remove(fileName)
//	producer, err := NewProducer(fileName)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer producer.Close()
//
//	if err := producer.WriteEvent(events); err != nil {
//		log.Fatal(err)
//	}
//
//	consumer, err := NewConsumer(fileName)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer consumer.Close()
//
//	readEvent, err := consumer.ReadEvent()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(readEvent.ShortURL)
//
//}
