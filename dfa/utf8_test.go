package dfa

import (
	"testing"
	"unicode"
)

func TestUTF8(t *testing.T) {
	BetweenRune(0, unicode.MaxRune).saveSvg("utf8.svg")
}
