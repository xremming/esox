package esox

import (
	"encoding/base64"
	"strings"
)

type FlashLevel string

const (
	FlashLevelInfo    FlashLevel = "info"
	FlashLevelSuccess FlashLevel = "success"
	FlashLevelWarning FlashLevel = "warning"
	FlashLevelError   FlashLevel = "error"
)

var flashLevels = map[byte]FlashLevel{
	'i': FlashLevelInfo,
	's': FlashLevelSuccess,
	'w': FlashLevelWarning,
	'e': FlashLevelError,
}

type FlashData struct {
	Level   FlashLevel
	Message string
}

func encodeFlashCookie(flashData []FlashData) string {
	var out []string
	var b strings.Builder
	for _, data := range flashData {
		b.WriteByte(byte(data.Level[0]))
		b.WriteByte(':')
		b.WriteString(base64.URLEncoding.EncodeToString([]byte(data.Message)))

		out = append(out, b.String())
		b.Reset()
	}

	return strings.Join(out, ",")
}

func decodeFlashCookie(cookie string) []FlashData {
	flashes := strings.Split(cookie, ",")

	var out []FlashData
	for _, flash := range flashes {
		if len(flash) <= 2 {
			continue
		}

		level, ok := flashLevels[flash[0]]
		if !ok {
			level = FlashLevelInfo
		}

		message, err := base64.URLEncoding.DecodeString(flash[2:])
		if err != nil {
			continue
		}

		out = append(out, FlashData{Level: level, Message: string(message)})
	}

	return out
}
