package scanner

import (
	"go/token"
	"unicode/utf8"

	"github.com/hailiang/dfa"
	"github.com/hailiang/gombi/scan"
)

var globalMatcher *scan.Matcher

func getMatcher() *scan.Matcher {
	if globalMatcher == nil {
		globalMatcher = scan.NewMatcher(
			tEOF,
			tIllegal,
			tokenDefs(),
		)
	}
	return globalMatcher
}

func tokenDefs() []scan.MID {
	var (
		c     = scan.Char
		b     = scan.Between
		s     = scan.Str
		con   = scan.Con
		or    = scan.Or
		ifNot = scan.IfNot
		class = scan.CharClass

		any           = b(1, utf8.MaxRune)
		newline       = c("\n")
		unicodeChar   = any.Exclude(newline)
		unicodeLetter = class(`L`)
		unicodeDigit  = class(`Nd`)
		letter        = or(unicodeLetter, c(`_`))
		decimalDigit  = b('0', '9')
		octalDigit    = b('0', '7')
		hexDigit      = or(b('0', '9'), b('A', 'F'), b('a', 'f'))

		empty = s(``)

		whitespaces = c(" \t\r").OneOrMore()
		lineComment = con(s(`//`), unicodeChar.ZeroOrMore(), or(newline, empty))
		// http://www.cs.dartmouth.edu/~mckeeman/cs118/assignments/comment.html
		commentText = func(char *dfa.M) *dfa.M {
			return con(char.Exclude(c(`*`)).ZeroOrMore(), s(`*`)).Loop(ifNot('/'))
		}
		generalCommentSL = con(s(`/*`), commentText(any.Exclude(s("\n"))), s(`/`))
		generalCommentML = con(s(`/*`), commentText(any.Exclude(s("\n"))).ZeroOrOne(), s("\n"), commentText(any), s(`/`))
		keywords         = or(s(`break`), s(`case`), s(`chan`), s(`const`),
			s(`continue`), s(`default`), s(`defer`), s(`else`), s(`fallthrough`),
			s(`for`), s(`func`), s(`go`), s(`goto`), s(`if`), s(`import`),
			s(`interface`), s(`map`), s(`package`), s(`range`), s(`return`),
			s(`select`), s(`struct`), s(`switch`), s(`type`), s(`var`))
		identifier           = con(letter, or(letter, unicodeDigit).ZeroOrMore()).Exclude(keywords)
		hexLit               = con(s(`0`), c("xX"), hexDigit.OneOrMore())
		decimalLit           = con(b('1', '9'), decimalDigit.ZeroOrMore())
		octalLit             = con(s(`0`), octalDigit.ZeroOrMore())
		intLit               = or(hexLit, decimalLit, octalLit)
		decimals             = decimalDigit.OneOrMore()
		exponent             = con(c("eE"), c("+-").ZeroOrOne(), decimals)
		floatLit1            = con(decimals, s(`.`), decimals.ZeroOrOne(), exponent.ZeroOrOne())
		floatLit2            = con(decimals, exponent)
		floatLit3            = con(s(`.`), decimals, exponent.ZeroOrOne())
		floatLit             = or(floatLit1, floatLit2, floatLit3)
		imaginaryLit         = con(or(floatLit, decimals), s(`i`))
		hexByteValue         = con(s(`\x`), hexDigit.Repeat(2))
		octalByteValue       = con(s(`\`), octalDigit.Repeat(3))
		byteValue            = or(hexByteValue, octalByteValue)
		littleUValue         = con(s(`\u`), hexDigit.Repeat(4))
		bigUValue            = con(s(`\U`), hexDigit.Repeat(8))
		escapedChar          = con(s(`\`), c(`abfnrtv\'"`))
		unicodeValue         = or(unicodeChar.Exclude(c(`\`)), littleUValue, bigUValue, escapedChar)
		runeLit              = con(s(`'`), or(byteValue, unicodeValue.Exclude(s(`'`))), s(`'`))
		rawStringLit         = con(s("`"), or(unicodeChar.Exclude(c("`")), newline).ZeroOrMore(), s("`"))
		interpretedStringLit = con(s(`"`), or(unicodeValue.Exclude(s(`"`)), byteValue).ZeroOrMore(), s(`"`))
	)
	return []scan.MID{
		{whitespaces, tWhitespace},
		{s("\n"), tNewline},
		{s(`if`), int(token.IF)},
		{s(`break`), int(token.BREAK)},
		{s(`case`), int(token.CASE)},
		{s(`chan`), int(token.CHAN)},
		{s(`const`), int(token.CONST)},
		{s(`continue`), int(token.CONTINUE)},
		{s(`default`), int(token.DEFAULT)},
		{s(`defer`), int(token.DEFER)},
		{s(`else`), int(token.ELSE)},
		{s(`fallthrough`), int(token.FALLTHROUGH)},
		{s(`for`), int(token.FOR)},
		{s(`func`), int(token.FUNC)},
		{s(`goto`), int(token.GOTO)},
		{s(`go`), int(token.GO)},
		{s(`import`), int(token.IMPORT)},
		{s(`interface`), int(token.INTERFACE)},
		{s(`map`), int(token.MAP)},
		{s(`package`), int(token.PACKAGE)},
		{s(`range`), int(token.RANGE)},
		{s(`return`), int(token.RETURN)},
		{s(`select`), int(token.SELECT)},
		{s(`struct`), int(token.STRUCT)},
		{s(`switch`), int(token.SWITCH)},
		{s(`type`), int(token.TYPE)},
		{s(`var`), int(token.VAR)},
		{identifier, int(token.IDENT)},
		{lineComment, tLineComment},
		{generalCommentSL, tGeneralCommentSL},
		{generalCommentML, tGeneralCommentML},
		{runeLit, int(token.CHAR)},
		{imaginaryLit, int(token.IMAG)},
		{floatLit, int(token.FLOAT)},
		{intLit, int(token.INT)},
		{rawStringLit, tRawStringLit},
		{interpretedStringLit, tInterpretedStringLit},
		{s(`...`), int(token.ELLIPSIS)},
		{s(`.`), int(token.PERIOD)},
		{s(`(`), int(token.LPAREN)},
		{s(`)`), int(token.RPAREN)},
		{s(`{`), int(token.LBRACE)},
		{s(`}`), int(token.RBRACE)},
		{s(`,`), int(token.COMMA)},
		{s(`==`), int(token.EQL)},
		{s(`=`), int(token.ASSIGN)},
		{s(`:=`), int(token.DEFINE)},
		{s(`:`), int(token.COLON)},
		{s(`[`), int(token.LBRACK)},
		{s(`]`), int(token.RBRACK)},
		{s(`*=`), int(token.MUL_ASSIGN)},
		{s(`*`), int(token.MUL)},
		{s(`+=`), int(token.ADD_ASSIGN)},
		{s(`++`), int(token.INC)},
		{s(`+`), int(token.ADD)},
		{s(`-=`), int(token.SUB_ASSIGN)},
		{s(`--`), int(token.DEC)},
		{s(`-`), int(token.SUB)},
		{s(`/=`), int(token.QUO_ASSIGN)},
		{s(`/`), int(token.QUO)},
		{s(`%=`), int(token.REM_ASSIGN)},
		{s(`%`), int(token.REM)},
		{s(`|=`), int(token.OR_ASSIGN)},
		{s(`||`), int(token.LOR)},
		{s(`|`), int(token.OR)},
		{s(`^=`), int(token.XOR_ASSIGN)},
		{s(`^`), int(token.XOR)},
		{s(`<<=`), int(token.SHL_ASSIGN)},
		{s(`<<`), int(token.SHL)},
		{s(`>>=`), int(token.SHR_ASSIGN)},
		{s(`>>`), int(token.SHR)},
		{s(`&^=`), int(token.AND_NOT_ASSIGN)},
		{s(`&^`), int(token.AND_NOT)},
		{s(`&=`), int(token.AND_ASSIGN)},
		{s(`&&`), int(token.LAND)},
		{s(`&`), int(token.AND)},
		{s(`!=`), int(token.NEQ)},
		{s(`!`), int(token.NOT)},
		{s(`<=`), int(token.LEQ)},
		{s(`<-`), int(token.ARROW)},
		{s(`<`), int(token.LSS)},
		{s(`>=`), int(token.GEQ)},
		{s(`>`), int(token.GTR)},
		{s(`;`), int(token.SEMICOLON)},
	}
}

func init() {
	getMatcher()
}

type gombiScanner struct {
	scan.Scanner
}

func newGombiScanner() gombiScanner {
	return gombiScanner{scan.Scanner{Matcher: getMatcher()}}
}

//type gombiScanner struct {
//	scan.GotoScanner
//}
//
//func newGombiScanner() gombiScanner {
//	return gombiScanner{scan.GotoScanner{
//		Match:   match,
//		EOF:     int(token.EOF),
//		Illegal: int(token.ILLEGAL)}}
//}
