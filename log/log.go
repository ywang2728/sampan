package log

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	MsgChanCap = 1024
)

type (
	Logger interface {
		Fatal(string)
		Fatalf(string, any)
		Error(string)
		Errorf(string, any)
		Warn(string)
		Warnf(string, any)
		Info(string)
		Infof(string, any)
		Debug(string)
		Debugf(string, any)
		Trace(string)
		Tracef(string, any)
	}

	Entry struct {
		id     [16]byte
		level  Level
		now    time.Time
		msg    string
		source string
		line   int
		meta   map[string]any
	}

	Log struct {
		level     Level
		prefix    string
		msgChan   chan *Entry
		writers   []Writer
		formatter Formatter
	}
)

func newEntry(l Level, m string) (e *Entry) {
	e = &Entry{id: uuid.New(), level: l, now: time.Now(), msg: m, meta: make(map[string]any)}
	return
}

func (e *Entry) withSource(source string) *Entry {
	e.source = source
	return e
}

func (e *Entry) withLine(line int) *Entry {
	e.line = line
	return e
}

func (e *Entry) withMeta(k string, v any) *Entry {
	e.meta[k] = v
	return e
}

func New() (l *Log) {
	l = &Log{msgChan: make(chan *Entry, MsgChanCap), writers: []Writer{}}
	return l
}

func (l *Log) WithLevel(level Level) *Log {
	l.level = level
	return l
}

func (l *Log) WithPrefix(p string) *Log {
	l.prefix = p
	return l
}

func (l *Log) WithWriters(ws ...Writer) *Log {
	l.writers = append(l.writers, ws...)
	return l
}

func (l *Log) WithFormatter(f Formatter) *Log {
	l.formatter = f
	return l
}

func (l *Log) logging() {
	go func() {
		for e := range l.msgChan {
			if l.level >= e.level {
				for _, w := range l.writers {
					var s string
					if f, ok := w.GetFormatter(); ok {
						s = f.Format(e)
					} else {
						s = l.formatter.Format(e)
					}
					w.Write(s)
				}
			}
		}
	}()
}

func (l *Log) Fatal(s string) {
	l.msgChan <- newEntry(FATAL, s)
}

func (l *Log) Fatalf(s string, a any) {
	l.msgChan <- newEntry(FATAL, fmt.Sprintf(s, a))
}

func (l *Log) Error(s string) {
	l.msgChan <- newEntry(ERROR, s)
}

func (l *Log) Errorf(s string, a any) {
	l.msgChan <- newEntry(ERROR, fmt.Sprintf(s, a))
}

func (l *Log) Warn(s string) {
	l.msgChan <- newEntry(WARN, s)
}

func (l *Log) Warnf(s string, a any) {
	l.msgChan <- newEntry(WARN, fmt.Sprintf(s, a))
}

func (l *Log) Info(s string) {
	l.msgChan <- newEntry(INFO, s)
}

func (l *Log) Infof(s string, a any) {
	l.msgChan <- newEntry(INFO, fmt.Sprintf(s, a))
}

func (l *Log) Debug(s string) {
	l.msgChan <- newEntry(DEBUG, s)
}

func (l *Log) Debugf(s string, a any) {
	l.msgChan <- newEntry(DEBUG, fmt.Sprintf(s, a))
}

func (l *Log) Trace(s string) {
	l.msgChan <- newEntry(TRACE, s)
}

func (l *Log) Tracef(s string, a any) {
	l.msgChan <- newEntry(TRACE, fmt.Sprintf(s, a))
}
