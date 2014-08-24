package scanner

import (
	"go/token"

	"github.com/hailiang/gombi/gre/scan"
)

const (
	tNewline = 1000 + iota
	tWhitespace
	tLineComment
	tGeneralCommentSL
	tGeneralCommentML
	tRawStringLit
	tInterpretedStringLit
	tSkip
)

var (
	c     = scan.Char
	s     = scan.Str
	con   = scan.Con
	or    = scan.Or
	merge = scan.Merge

	illegal       = c(`\x00`)
	any           = illegal.Negate()
	newline       = c(`\n`)
	unicodeChar   = any.Exclude(newline)
	unicodeLetter = c(`\p{L}`)
	unicodeDigit  = c(`\p{Nd}`)
	//unicodeLetter = c(`A-Za-z0-9۰۱۸६४ŝ`)
	//unicodeDigit = c(`0-9９８７６`)
	letter       = merge(unicodeLetter, c(`_`))
	decimalDigit = c(`0-9`)
	octalDigit   = c(`0-7`)
	hexDigit     = c(`0-9A-Fa-f`)
	empty        = s(``)
	whitespaces  = c(` \t\r`).OneOrMore()

	lineComment          = con(s(`//`), unicodeChar.ZeroOrMore(), or(newline, empty))
	generalCommentSL     = con(s(`/*`), any.Exclude(newline).ZeroOrMore().Ungreedy(), s(`*/`))
	generalCommentML     = con(s(`/*`), any.ZeroOrMore().Ungreedy(), s(`*/`))
	identifier           = con(letter, or(letter, unicodeDigit).ZeroOrMore())
	intLit               = or(hexLit, decimalLit, octalLit)
	decimalLit           = con(c(`1-9`), decimalDigit.ZeroOrMore())
	octalLit             = con(s(`0`), octalDigit.ZeroOrMore())
	hexLit               = con(s(`0`), c(`xX`), hexDigit.OneOrMore())
	floatLit             = or(floatLit1, floatLit2, floatLit3)
	floatLit1            = con(decimals, s(`.`), decimals.ZeroOrOne(), exponent.ZeroOrOne())
	floatLit2            = con(decimals, exponent)
	floatLit3            = con(s(`.`), decimals, exponent.ZeroOrOne())
	decimals             = decimalDigit.OneOrMore()
	exponent             = con(c(`eE`), c(`\+\-`).ZeroOrOne(), decimals)
	imaginaryLit         = con(or(decimals, floatLit), c(`i`))
	runeLit              = con(c(`'`), or(unicodeValue, byteValue), c(`'`))
	unicodeValue         = or(unicodeChar, littleUValue, bigUValue, escapedChar)
	unicodeStrValue      = or(unicodeChar.Exclude(c(`"`)), littleUValue, bigUValue, escapedChar)
	byteValue            = or(octalByteValue, hexByteValue)
	octalByteValue       = con(s(`\`), octalDigit.Repeat(3))
	hexByteValue         = con(s(`\x`), hexDigit.Repeat(2))
	littleUValue         = con(s(`\u`), hexDigit.Repeat(4))
	bigUValue            = con(s(`\U`), hexDigit.Repeat(8))
	escapedChar          = con(s(`\`), c(`abfnrtv\\'"`))
	rawStringLit         = con(s("`"), or(unicodeChar.Exclude(c("`")), newline).ZeroOrMore(), s("`"))
	interpretedStringLit = con(s(`"`), or(unicodeStrValue, byteValue).ZeroOrMore(), s(`"`))

	matcher = scan.NewMapMatcher(scan.MM{
		{newline, tNewline},
		{whitespaces, tWhitespace},
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
		{s(`+=`), int(token.ADD_ASSIGN)},
		{s(`++`), int(token.INC)},
		{s(`+`), int(token.ADD)},
		{s(`-=`), int(token.SUB_ASSIGN)},
		{s(`--`), int(token.DEC)},
		{s(`-`), int(token.SUB)},
		{s(`*=`), int(token.MUL_ASSIGN)},
		{s(`*`), int(token.MUL)},
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
		{s(`[`), int(token.LBRACK)},
		{s(`]`), int(token.RBRACK)},
		{s(`;`), int(token.SEMICOLON)},
		{scan.Pat(`.`), int(token.ILLEGAL)},
	})
)

func init() {
	matcher.EOF = int(token.EOF)
	matcher.Illegal = int(token.ILLEGAL)
}

type gombiScanner struct {
	scan.Scanner
}

func newGombiScanner() gombiScanner {
	s := gombiScanner{scan.Scanner{
		Matcher: matcher,
	}}
	return s
}
