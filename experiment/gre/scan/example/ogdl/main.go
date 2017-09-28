package main

import (
	"fmt"

	"h12.me/gombi/scan"
)

func main() {
	var (
		char   = scan.Char
		pat    = scan.Pat
		merge  = scan.Merge
		or     = scan.Or
		con    = scan.Con
		Tokens = scan.Tokens

		nonctrl = char(`[:cntrl:]`).Negate()
		indent  = char(`\t `)
		lbreak  = char(`\n\r`)
		space   = merge(indent, lbreak)
		inline  = merge(nonctrl, indent)
		any     = merge(nonctrl, space)
		invalid = any.Negate()
		delim   = char(`,{}`)
		empty   = pat(``)

		newline        = or(lbreak, pat(`\r\n`))
		inlineComment  = con(pat(`//`), inline.ZeroOrMore(), or(newline, empty))
		quoted         = or(inline.Exclude(char(`"`)), pat(`\\"`))
		quotedString   = con(pat(`"`), quoted.ZeroOrMore(), pat(`"`))
		unquoted       = nonctrl.Exclude(merge(delim, char(` `)))
		unquotedString = unquoted.OneOrMore()

		tokens = Tokens(
			invalid,
			inlineComment,
			char(`{`),
			char(`}`),
			char(`,`),
			quotedString,
			unquotedString,
			space.OneOrMore(),
		)
	)

	pp("nonctrl", nonctrl)
	pp("indent", indent)
	pp("lbreak", lbreak)
	pp("space", space)
	pp("inline", inline)
	pp("any", any)
	pp("invalid", invalid)
	pp("delim", delim)
	pp("newline", newline)
	pp("inlineComment", inlineComment)
	pp("quoted", quoted)
	pp("quotedString", quotedString)
	pp("unquoted", unquoted)
	pp("unquotedString", unquotedString)
	pp("tokens", tokens)
}

func pp(v ...interface{}) {
	fmt.Println(v...)
}
