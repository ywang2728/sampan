package log

import (
	"github.com/google/uuid"
	"strings"
	"sync"
	"time"
)

const (
	OFF Level = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG
	TRACE
)

type (
	Level int

	Writer interface {
		write(string)
	}

	Formatter interface {
		format(*Entry) string
	}

	Logger interface {
		Trace(*Entry)
		Debug(*Entry)
		Info(*Entry)
		Warn(*Entry)
		Error(*Entry)
		Fatal(*Entry)
		SetFormatter(*Formatter)
		AddWriters(...*Writer)
	}

	Entry struct {
		id    [16]byte
		level Level
		now   time.Time
		msg   string
		file  string
		line  int
		data  any
	}

	Log struct {
		mtx       sync.Mutex
		wg        sync.WaitGroup
		level     Level
		prefix    string
		msgBuff   []*Entry
		msgChan   chan *Entry
		writers   []*Writer
		formatter *Formatter
	}

	ConsoleWriter struct {
		formatter *Formatter
	}

	FileWriter struct {
		formatter *Formatter
		path      string
	}

	CLFormatter struct {
		template string
	}

	JsonFormatter struct {
		inline bool
	}
)

var _levels = [7]string{"OFF", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}

func (l Level) String() string {
	return _levels[l]
}

func (l Level) Index() int {
	return int(l)
}

func parseLevel(l string) (level Level) {
	switch strings.ToUpper(l) {
	case "FATAL":
		level = FATAL
	case "ERROR":
		level = ERROR
	case "WARN":
		level = WARN
	case "INFO":
		level = INFO
	case "DEBUG":
		level = DEBUG
	case "TRACE":
		level = TRACE
	default:
		level = OFF
	}
	return
}

func newEntry(l, m string) (e *Entry) {
	e = &Entry{id: uuid.New(), level: parseLevel(l), now: time.Now(), msg: m}
	return
}

var _log *Log

func (l Log) SetLevel(level Level) {
	l.level = level
}

func (l Log) AddWriters(ws ...*Writer) {
	l.writers = append(l.writers, ws...)
}

func (l Log) SetFormatter(f *Formatter) {
	l.formatter = f
}

func (l Log) Trace(e *Entry) {

}

func SetLevel(level string) {
	_log.SetLevel(parseLevel(level))
}

func SetFormatter(f *Formatter) {
	_log.SetFormatter(f)
}

func AddWriters(ws ...*Writer) {
	_log.AddWriters(ws...)
}

func Trace(msg string) {
	newEntry("trace", msg)
}

func init() {
	_log = &Log{level: INFO}
}
