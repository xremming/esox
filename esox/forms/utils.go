package forms

import (
	"encoding/base64"
	"hash/fnv"
	"strings"
	"unicode"
)

const FormatDatetimeLocal = "2006-01-02T15:04"

func removeWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, c := range s {
		if unicode.IsSpace(c) {
			continue
		}

		b.WriteRune(c)
	}

	return b.String()
}

func idFromName(names ...string) string {
	if len(names) == 0 {
		panic("esox: idFromName: no names provided")
	}

	hash := fnv.New32()
	for _, name := range names {
		hash.Write([]byte(name))
	}

	name := removeWhitespace(strings.Join(names, "_"))

	return name + "_" + base64.URLEncoding.EncodeToString(hash.Sum(nil))
}
