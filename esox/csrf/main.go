package csrf

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func getCSRF(r *http.Request) string {
	log := zerolog.Ctx(r.Context())

	if t := r.FormValue("_csrf"); len(t) > 0 {
		log.Debug().Msg("CSRF token found in form value _csrf.")
		return t
	} else if t := r.URL.Query().Get("_csrf"); len(t) > 0 {
		log.Debug().Msg("CSRF token found in query parameter _csrf.")
		return t
	} else if t := r.Header.Get("X-CSRF-TOKEN"); len(t) > 0 {
		log.Debug().Msg("CSRF token found in header X-CSRF-TOKEN.")
		return t
	} else if t := r.Header.Get("X-XSRF-TOKEN"); len(t) > 0 {
		log.Debug().Msg("CSRF token found in header X-XSRF-TOKEN.")
		return t
	}

	return ""
}

func sign(secret string, value string) []byte {
	signer := hmac.New(sha256.New, []byte(secret))
	signer.Write([]byte(value))
	return signer.Sum(nil)
}

const (
	DefaultMaxAge time.Duration = 30 * time.Minute
)

type CSRF struct {
	// Secrets is a list of secrets used to sign the CSRF token. The first
	// secret will be used to sign the token, and the rest will be used to
	// verify the token. If the token is signed with a secret that is not
	// in this list, it will be considered invalid.
	Secrets []string

	// MaxAge is the maximum age of a CSRF token. If the token is older than
	// MaxAge, it will be considered invalid. If MaxAge is 0, DefaultMaxAge
	// will be used.
	MaxAge time.Duration
}

func (csrf CSRF) Generate() string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature := sign(csrf.Secrets[0], timestamp)
	signatureBase64URL := base64.URLEncoding.EncodeToString(signature)
	return timestamp + "." + signatureBase64URL
}

var (
	ErrTokenInvalid   = errors.New("CSRF token is invalid")
	ErrTokenSignature = errors.New("CSRF token has an invalid signature")
	ErrTokenExpired   = errors.New("CSRF token has expired")
)

func (csrf CSRF) Validate(ctx context.Context, token string) error {
	log := zerolog.Ctx(ctx).With().Str("csrf", token).Logger()

	splittedValue := strings.SplitN(token, ".", 2)
	if len(splittedValue) != 2 {
		log.Error().Msg("CSRF token is invalid.")
		return ErrTokenInvalid
	}

	timestamp, signatureBase64URL := splittedValue[0], splittedValue[1]

	parsedTimestamp, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		log.Err(err).Msg("CSRF token's timestamp is invalid.")
		return ErrTokenInvalid
	}

	signature, err := base64.URLEncoding.DecodeString(signatureBase64URL)
	if err != nil {
		log.Err(err).Msg("CSRF token's signature is invalid.")
		return ErrTokenInvalid
	}

	maxAge := csrf.MaxAge
	if maxAge == 0 {
		maxAge = 30 * time.Minute
	}

	ok := false
	for _, secret := range csrf.Secrets {
		possibleSignature := sign(secret, timestamp)
		if hmac.Equal(signature, possibleSignature) {
			ok = true
		}
	}

	if !ok {
		log.Debug().Msg("CSRF token's signature is invalid.")
		return ErrTokenSignature
	}

	if time.Now().UTC().Sub(parsedTimestamp) > maxAge {
		log.Debug().Msg("CSRF token has expired.")
		return ErrTokenExpired
	}

	return nil
}
