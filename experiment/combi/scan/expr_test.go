package scan

import (
	"fmt"
	"testing"

	"h12.io/gspec"
)

var (
	c          = Char
	b          = Between
	merge      = Merge
	s          = Str
	con        = Con
	or         = Or
	zeroOrOne  = ZeroOrOne
	zeroOrMore = ZeroOrMore
	oneOrMore  = OneOrMore
	repeat     = Repeat
)

func TestExpr(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	for _, tc := range []struct {
		sg fmt.Stringer
		s  string
	}{
		{c(`a`), `[a]`},
		{c(`abc`), `[a-c]`},
		{c(`[`), `[\[]`},
		{c(`]`), `[\]]`},
		{c(`-`), `[\-]`},
		{b('a', 'c'), `[a-c]`},
		{s(`xy`), `xy`},

		{c(`b`).Negate(), `[\x00-ac-\U0010ffff]`},
		{c(`b`).Negate().Negate(), `[b]`},
		{c(`bf`).Negate(), `[\x00-ac-eg-\U0010ffff]`},
		{c(`bf`).Negate().Negate(), `[bf]`},
		{c("\x00").Negate(), `[\x01-\U0010ffff]`},
		{c("\x00").Negate().Negate(), `[\x00]`},
		{c("\U0010ffff").Negate(), `[\x00-\U0010fffe]`},
		{c("\U0010ffff").Negate().Negate(), `[\U0010ffff]`},

		{c(`abcde`).Exclude(c(`bd`)), `[ace]`},
	} {
		expect(tc.sg.String()).Equal(tc.s)
	}
}
