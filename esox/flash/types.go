package flash

type Level string

const (
	LevelInfo    Level = "info"
	LevelSuccess Level = "success"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

var flashLevels = map[byte]Level{
	'i': LevelInfo,
	's': LevelSuccess,
	'w': LevelWarning,
	'e': LevelError,
}

type Data struct {
	Level   Level
	Message string
}
