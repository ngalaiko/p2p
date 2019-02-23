package logger

// Level is a logger level.
type Level uint

// Supported logger levels.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
	LevelPanic
)

var names = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"error": LevelError,
	"panic": LevelPanic,
}

// ParseLevel parses level from string.
func ParseLevel(l string) Level {
	if lvl, ok := names[l]; ok {
		return lvl
	}
	return LevelInfo
}
