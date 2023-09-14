package log

var _log *Log

func init() {
	_log = New().WithLevel(INFO).WithWriters(&ConsoleWriter{})
	go _log.logging()
}

func SetLevel(l Level) {
	_log.WithLevel(l)
}

func SetFormatter(f Formatter) {
	_log.WithFormatter(f)
}

func AddWriters(ws ...Writer) {
	_log.WithWriters(ws...)
}

func Fatal(s string) {
	_log.Fatal(s)
}

func Fatalf(s string, a any) {
	_log.Fatalf(s, a)
}

func Error(s string) {
	_log.Error(s)
}

func Errorf(s string, a any) {
	_log.Errorf(s, a)
}
func Warn(s string) {
	_log.Warn(s)
}

func Warnf(s string, a any) {
	_log.Warnf(s, a)
}
func Info(s string) {
	_log.Info(s)
}

func Infof(s string, a any) {
	_log.Infof(s, a)
}
func Debug(s string) {
	_log.Debug(s)
}

func Debugf(s string, a any) {
	_log.Debugf(s, a)
}
func Trace(s string) {
	_log.Trace(s)
}

func Tracef(s string, a any) {
	_log.Tracef(s, a)
}
