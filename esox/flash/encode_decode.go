package flash

import (
	"encoding/base64"
	"strings"
)

func Encode(flashes []Data) string {
	var out []string
	var b strings.Builder
	for _, flash := range flashes {
		b.WriteByte(byte(flash.Level[0]))
		b.WriteByte(':')
		b.WriteString(base64.URLEncoding.EncodeToString([]byte(flash.Message)))

		out = append(out, b.String())
		b.Reset()
	}

	return strings.Join(out, ",")
}

func Decode(value string) []Data {
	flashes := strings.Split(value, ",")

	var out []Data
	for _, flash := range flashes {
		if len(flash) <= 2 {
			continue
		}

		level, ok := flashLevels[flash[0]]
		if !ok {
			level = LevelInfo
		}

		message, err := base64.URLEncoding.DecodeString(flash[2:])
		if err != nil {
			continue
		}

		out = append(out, Data{Level: level, Message: string(message)})
	}

	return out
}
