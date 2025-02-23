package logging

import (
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	ErrorLevel LogLevel = 0b00001
	WarnLevel  LogLevel = 0b00010
	InfoLevel  LogLevel = 0b00100
	DebugLevel LogLevel = 0b01000
	TraceLevel LogLevel = 0b10000
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

	EnableLevel(level LogLevel)
	DisableLevel(level LogLevel)
	Clone() ILogger
}

func initLogLevels(maxLevel LogLevel) LogLevel {
	var result LogLevel = 0
	if maxLevel > ErrorLevel {
		result |= ErrorLevel
	}
	if maxLevel > WarnLevel {
		result |= WarnLevel
	}
	if maxLevel > InfoLevel {
		result |= InfoLevel
	}
	if maxLevel > DebugLevel {
		result |= DebugLevel
	}
	result |= maxLevel
	return result
}

func NewSimpleLogger(initialLevel LogLevel) (*SimpleLogger, error) {
	initialLevel = initLogLevels(initialLevel)
	return &SimpleLogger{CurLevel: initialLevel, NormalLogger: log.New(os.Stdout, "", 0), ErrorLogger: log.New(os.Stderr, "", 0), Start: time.Now()}, nil
}

func NewStdErrSimpleLogger(initialLevel LogLevel) (*SimpleLogger, error) {
	initialLevel = initLogLevels(initialLevel)
	return &SimpleLogger{CurLevel: initialLevel, NormalLogger: log.New(os.Stderr, "", 0), ErrorLogger: log.New(os.Stderr, "", 0), Start: time.Now()}, nil
}
