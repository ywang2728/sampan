package log

const (
	OFF Level = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG
	TRACE
)

type Level int

var _levels = [7]string{"OFF", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}

func (l Level) Name() string {
	return _levels[l]
}

func (l Level) Index() int {
	return int(l)
}

func (l Level) MarshalJSON() ([]byte, error) {
	return []byte(`"` + l.Name() + `"`), nil
}
