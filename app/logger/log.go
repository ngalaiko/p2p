package logger

import (
	"fmt"
	"log"
)

// Logger is a multiple level logger.
type Logger struct {
	level  Level
	prefix string
}

// New is a logger constructor.
func New(l Level) *Logger {
	return &Logger{
		level: l,
	}
}

// Prefix adds prefix to every log message and returns a copy of the logger.
func (l *Logger) Prefix(p string) *Logger {
	lc := *l
	lc.prefix = fmt.Sprintf("%s: ", p)
	return &lc
}

// Debug prints debug level log.
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	log.Printf(fmt.Sprintf("[DEBUG] %s%s", l.prefix, format), v...)
}

// Info prints info level log.
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	log.Printf(fmt.Sprintf("[INFO] %s%s", l.prefix, format), v...)
}

// Error prints error level log.
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level > LevelError {
		return
	}
	log.Printf(fmt.Sprintf("[ERROR] %s%s", l.prefix, format), v...)
}

// Panic prints panic level log and panics.
func (l *Logger) Panic(format string, v ...interface{}) {
	panic(fmt.Sprintf(fmt.Sprintf("[PANIC] %s%s", l.prefix, format), v...))
}
