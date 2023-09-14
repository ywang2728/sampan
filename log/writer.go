package log

import "fmt"

type (
	Writer interface {
		Write(string)
		SetFormatter(Formatter)
		GetFormatter() (Formatter, bool)
	}

	ConsoleWriter struct {
		formatter Formatter
	}

	FileWriter struct {
		formatter Formatter
		path      string
	}
)

func (w *ConsoleWriter) SetFormatter(f Formatter) {
	w.formatter = f
}

func (w *ConsoleWriter) GetFormatter() (f Formatter, ok bool) {
	return w.formatter, w.formatter != nil
}

func (w *ConsoleWriter) Write(s string) {
	fmt.Println("console writer: " + s)
}
