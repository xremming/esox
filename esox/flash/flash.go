package flash

import "net/http"

func Flash(r *http.Request, level Level, message string) {
	flashes := FromContext(r.Context())
	flashes = append(flashes, Data{Level: level, Message: message})
	*r = *r.WithContext(NewContext(r.Context(), flashes))
}

func Info(r *http.Request, message string) {
	Flash(r, LevelInfo, message)
}

func Success(r *http.Request, message string) {
	Flash(r, LevelSuccess, message)
}

func Warning(r *http.Request, message string) {
	Flash(r, LevelWarning, message)
}

func Error(r *http.Request, message string) {
	Flash(r, LevelError, message)
}
