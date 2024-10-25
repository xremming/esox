package esox

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

const StaticPrefix = "static"

const hashLength = 2 * sha256.Size

func integrityHash(r io.Reader) (pathHash string, integrity string, err error) {
	h := sha256.New()

	_, err = io.Copy(h, r)
	if err != nil {
		return
	}

	sum := h.Sum(nil)
	pathHash = hex.EncodeToString(sum)
	integrity = "sha256-" + base64.StdEncoding.EncodeToString(sum)
	return
}

func normalizeStaticPath(staticPath string) string {
	base := path.Base(staticPath)
	splitted := strings.Split(base, ".")
	if len(splitted) <= 2 {
		return staticPath
	}

	hash := splitted[len(splitted)-2]
	if len(hash) != hashLength {
		return staticPath
	}

	before := strings.Join(splitted[:len(splitted)-2], ".")
	after := splitted[len(splitted)-1]

	return path.Join(path.Dir(staticPath), fmt.Sprintf("%s.%s", before, after))
}

func staticPathWithHash(staticPath string, hash string) (string, error) {
	if len(staticPath) == 0 {
		return "", errors.New("invalid static path")
	}

	if len(hash) != hashLength {
		return "", errors.New("invalid hash length")
	}

	base := path.Base(staticPath)
	splitted := strings.Split(base, ".")
	if len(splitted) == 1 {
		return fmt.Sprintf("%s.%s", staticPath, hash), nil
	}

	before := strings.Join(splitted[:len(splitted)-1], ".")
	after := splitted[len(splitted)-1]

	return path.Join(path.Dir(staticPath), fmt.Sprintf("%s.%s.%s", before, hash, after)), nil
}

type StaticFile struct {
	io.ReadCloser
	Path         string
	PathWithHash string
	Integrity    string
}

func GetStaticFile(staticPath string) (StaticFile, error) {
	normalized := normalizeStaticPath(staticPath)
	file, err := os.Open(filepath.Join(StaticPrefix, normalized))
	if err != nil {
		return StaticFile{}, err
	}

	pathHash, integrity, err := integrityHash(file)
	if err != nil {
		return StaticFile{}, err
	}

	// after hash is calculated we need to reset the file pointer to the beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return StaticFile{}, err
	}

	pathWithHash, err := staticPathWithHash(staticPath, pathHash)
	if err != nil {
		return StaticFile{}, err
	}

	return StaticFile{
		ReadCloser:   file,
		Path:         normalized,
		PathWithHash: pathWithHash,
		Integrity:    integrity,
	}, err
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	normalizedPath := normalizeStaticPath(path)

	log := zerolog.Ctx(r.Context()).With().
		Str("path", path).
		Str("normalizedPath", normalizedPath).
		Logger()

	file, err := GetStaticFile(normalizedPath)
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

	if file.PathWithHash == path {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, no-cache")
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
