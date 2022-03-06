package heartfelt

import (
	"fmt"
	"log"
)

// Logger is a logger.
type Logger interface {
	Info(v ...interface{})
	Error(v ...interface{})
}

var _ Logger = new(defaultLogger)

type defaultLogger struct {
	logger *log.Logger
}

func newDefaultLogger() *defaultLogger {
	return &defaultLogger{
		logger: log.Default(),
	}
}

func (l defaultLogger) Info(v ...interface{}) {
	l.logger.Println("[INFO]", fmt.Sprint(v...))
}

func (l defaultLogger) Error(v ...interface{}) {
	l.logger.Println("[ERROR]", fmt.Sprint(v...))
}
