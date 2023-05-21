package esox

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	log := zerolog.Ctx(r.Context())

	staticResources := GetStaticResources(r.Context())

	file, err := staticResources.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.NotFound(w, r)
		} else {
			log.Err(err).Msg("error opening static resource")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}
	defer file.Close()

	var buf [512]byte
	n, err := file.Read(buf[:])
	if err != nil && err != io.EOF {
		log.Err(err).Msg("error reading static resource")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	contentType := ""
	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript"
	} else {
		contentType = http.DetectContentType(buf[:n])
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	_, err = w.Write(buf[:n])
	if err != nil {
		log.Err(err).Msg("failed to write head of static resource to response writer")
		return
	}

	_, err = io.Copy(w, file)
	if err != nil {
		log.Err(err).Msg("failed to copy static resource to response writer")
		return
	}
}
