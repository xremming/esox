package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCutRight(t *testing.T) {
	cases := []struct {
		in         string
		wantBefore string
		wantAfter  string
	}{
		{"", "", ""},
		{"a", "a", ""},
		{"a:", "a", ""},
		{"a::", "a:", ""},
		{"a:b", "a", "b"},
		{"a::b", "a:", "b"},
		{"a::b:", "a::b", ""},
		{"a:b:c", "a:b", "c"},
		{"a::b:c", "a::b", "c"},
		{"a:b:c:", "a:b:c", ""},
		{"a::b:c:", "a::b:c", ""},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			gotBefore, gotAfter := cutRight(c.in, ':')
			assert.Equal(t, c.wantBefore, gotBefore, "before")
			assert.Equal(t, c.wantAfter, gotAfter, "after")
		})
	}
}
