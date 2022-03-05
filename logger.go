package heartfelt

import "log"

type Logger interface {
	Info(string)
	Err(string)
}

var _ Logger = new(defaultLogger)

type defaultLogger struct {
	logger *log.Logger
}

func newDefaultLogger() *defaultLogger {
	return &defaultLogger{log.Default()}
}

func (l defaultLogger) Info(msg string) {
	l.logger.Println(msg)
}

func (l defaultLogger) Err(msg string) {
	l.logger.Println(msg)
}
