package scan

import (
	"github.com/hailiang/dfa"
)

var (
	Char        = dfa.Char
	Between     = dfa.Between
	BetweenByte = dfa.BetweenByte
	Str         = dfa.Str
	Con         = dfa.Con
	Or          = dfa.Or
	And         = dfa.And
	ZeroOrOne   = dfa.ZeroOrOne
	ZeroOrMore  = dfa.ZeroOrMore
	OneOrMore   = dfa.OneOrMore
)
