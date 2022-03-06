package heartfelt

import (
	"log"
	"os"
)

type Logger interface {
	Info(v ...interface{})
	Err(v ...interface{})
}

var _ Logger = new(defaultLogger)

type defaultLogger struct {
	infoLogger *log.Logger
	errLogger  *log.Logger
}

func newDefaultLogger() *defaultLogger {
	flag := log.LstdFlags | log.Lmsgprefix
	return &defaultLogger{
		infoLogger: log.New(os.Stderr, "[INFO]", flag),
		errLogger:  log.New(os.Stderr, "[ERROR]", flag),
	}
}

func (l defaultLogger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

func (l defaultLogger) Err(v ...interface{}) {
	l.errLogger.Println(v...)
}
