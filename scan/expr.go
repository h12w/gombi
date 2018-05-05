package scan

import (
	"h12.io/dfa"
)

var (
	Char        = dfa.Char
	Between     = dfa.Between
	BetweenByte = dfa.BetweenByte
	Str         = dfa.Str
	Con         = dfa.Con
	Or          = dfa.Or
	And         = dfa.And
	Optional    = dfa.Optional
	IfNot       = dfa.IfNot
	CharClass   = dfa.CharClass
)
