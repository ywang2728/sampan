package log

type (
	Formatter interface {
		Format(*Entry) string
	}
	CLFormatter struct {
		template string
	}

	JsonFormatter struct {
		inline bool
	}
)
