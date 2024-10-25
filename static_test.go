package esox

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testHash string

func init() {
	var (
		buf bytes.Buffer
		err error
	)
	buf.WriteString("test")
	testHash, _, err = integrityHash(&buf)
	if err != nil {
		panic(err)
	}
}

func TestNormalizeStaticPath(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"plain", "plain"},
		{"plain file", "plain file"},
		{"styles.css", "styles.css"},
		{"styles.min.css", "styles.min.css"},

		{
			fmt.Sprintf("styles.%s.css", testHash),
			"styles.css",
		}, {
			fmt.Sprintf("styles.min.%s.css", testHash),
			"styles.min.css",
		}, {
			"folder/styles.css",
			"folder/styles.css",
		}, {
			"folder/styles.min.css",
			"folder/styles.min.css",
		}, {
			fmt.Sprintf("folder/styles.%s.css", testHash),
			"folder/styles.css",
		}, {
			fmt.Sprintf("folder/styles.min.%s.css", testHash),
			"folder/styles.min.css",
		}, {
			"folder.min/styles.css",
			"folder.min/styles.css",
		}, {
			"folder.min/styles.min.css",
			"folder.min/styles.min.css",
		}, {
			fmt.Sprintf("folder.min/styles.%s.css", testHash),
			"folder.min/styles.css",
		}, {
			fmt.Sprintf("folder.min/styles.min.%s.css", testHash),
			"folder.min/styles.min.css",
		}, {
			fmt.Sprintf("folder.min.%s/styles.%s.css", testHash, testHash),
			fmt.Sprintf("folder.min.%s/styles.css", testHash),
		}, {
			fmt.Sprintf("folder.min.%s/styles.min.%s.css", testHash, testHash),
			fmt.Sprintf("folder.min.%s/styles.min.css", testHash),
		}, {
			fmt.Sprintf("folder.min.%s.css/styles.%s.css", testHash, testHash),
			fmt.Sprintf("folder.min.%s.css/styles.css", testHash),
		}, {
			fmt.Sprintf("folder.min.%s.css/styles.min.%s.css", testHash, testHash),
			fmt.Sprintf("folder.min.%s.css/styles.min.css", testHash),
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			out := normalizeStaticPath(c.in)
			assert.Equal(t, c.out, out)
		})
	}
}

func TestStaticPathWithHash(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"plain", "plain.%s"},
		{"plain file", "plain file.%s"},
		{"styles.css", "styles.%s.css"},
		{"styles.min.css", "styles.min.%s.css"},
		{"folder/styles.min.css", "folder/styles.min.%s.css"},
		{"folder.min/styles.min.css", "folder.min/styles.min.%s.css"},
		{"folder.min.css/styles.min.css", "folder.min.css/styles.min.%s.css"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			out, err := staticPathWithHash(c.in, testHash)
			assert.NoError(t, err)

			expected := fmt.Sprintf(c.out, testHash)
			assert.Equal(t, expected, out)
		})
	}
}

func TestStaticPathWithHashError(t *testing.T) {
	cases := []struct {
		in   string
		hash string
	}{
		{"", testHash},
		{"plain", ""},
		{"styles.css", "invalid"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			_, err := staticPathWithHash(c.in, c.hash)
			assert.Error(t, err)
		})
	}
}
