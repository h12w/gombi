package dfa

import (
	"testing"
	"unicode"
)

func TestUTF8(t *testing.T) {
	Between(0, unicode.MaxRune).saveSvg("utf8.svg")
	Between(0, unicode.MaxRune).minimize().saveSvg("utf8min.svg")
}
