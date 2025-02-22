package logging

import (
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	ErrorLevel LogLevel = iota
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type ILogger interface {
	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Panicf(format string, args ...any)

	Traceln(args ...any)
	Debugln(args ...any)
	Infoln(args ...any)
	Warnln(args ...any)
	Errorln(args ...any)
	Panicln(args ...any)

	SetLevel(newLevel LogLevel)
}

func NewSimpleLogger(initialLevel LogLevel) (*SimpleLogger, error) {
	return &SimpleLogger{CurLevel: initialLevel, NormalLogger: log.New(os.Stdout, "", 0), ErrorLogger: log.New(os.Stderr, "", 0), Start: time.Now()}, nil
}

func NewStdErrSimpleLogger(initialLevel LogLevel) (*SimpleLogger, error) {
	return &SimpleLogger{CurLevel: initialLevel, NormalLogger: log.New(os.Stderr, "", 0), ErrorLogger: log.New(os.Stderr, "", 0), Start: time.Now()}, nil
}

func NewQuietLogger(initialLevel LogLevel) (*QuietLogger, error) {
	return &QuietLogger{CurLevel: initialLevel}, nil
}
