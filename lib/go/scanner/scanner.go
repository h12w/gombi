package scanner

import (
	"go/token"

	"github.com/hailiang/gombi/scan"
)

var (
	c     = scan.Char
	p     = scan.Pat
	con   = scan.Con
	or    = scan.Or
	orPat = scan.OrPat
	merge = scan.Merge

	any           = c(`\x00`).Negate()
	newline       = c(`\n`)
	unicodeChar   = any.Exclude(newline)
	unicodeLetter = c(`\pL`)
	unicodeDigit  = c(`\p{Nd}`)
	letter        = merge(unicodeLetter, c(`_`))
	decimalDigit  = c(`0-9`)
	octalDigit    = c(`0-7`)
	hexDigit      = c(`0-9A-Fa-f`)
	empty         = p(``)
	//whitespace    = c(` \t\n\r`)

	lineComments    = con(p(`//`), unicodeChar.ZeroOrMore(), or(newline, empty))
	generalComments = con(p(`/\*`), any.ZeroOrMore(), p(`\*/`))
	comment         = or(lineComments, generalComments)

	identifier = con(letter, or(letter, unicodeDigit).ZeroOrMore())
	//keyword    = orPat(
	//	`break`, `default`, `func`, `interface`, `select`,
	//	`case`, `defer`, `go`, `map`, `struct`,
	//	`chan`, `else`, `goto`, `package`, `switch`,
	//	`const`, `fallthrough`, `if`, `range`, `type`,
	//	`continue`, `for`, `import`, `return`, `var`,
	//)

	opDelim = orPat(
		`\+`, `&`, `\+=`, `&=`, `&&`, `==`, `!=`, `\(`, `\)`,
		`\-`, `\|`, `\-=`, `\|=`, `\|\|`, `<`, `<=`, `\[`, `\]`,
		`\*`, `^`, `\*=`, `^=`, `<\-`, `>`, `>=`, `\{`, `\}`,
		`/`, `<<`, `/=`, `<<=`, `\+\+`, `=`, `:=`, `,`, `;`,
		`%`, `>>`, `%=`, `>>=`, `\-\-`, `!`, `\.\.\.`, `\.`, `:`,
		`&^`, `&^=`,
	)

	intLit     = or(decimalLit, octalLit, hexLit)
	decimalLit = con(c(`1-9`), decimalDigit.ZeroOrMore())
	octalLit   = con(c(`0`), octalDigit.ZeroOrMore())
	hexLit     = con(c(`0`), c(`xX`), hexDigit.OneOrMore())

	floatLit = or(
		con(decimals, c(`\.`), decimals.ZeroOrOne(), exponent.ZeroOrOne()),
		con(decimals, exponent),
		con(c(`\.`), decimals, exponent.ZeroOrOne()))
	decimals = decimalDigit.OneOrMore()
	exponent = con(c(`eE`), c(`\+\-`).ZeroOrOne(), decimals)

	imaginaryLit = con(or(decimals, floatLit), c(`i`))

	runeLit        = con(c(`'`), or(unicodeValue, byteValue), c(`'`))
	unicodeValue   = or(unicodeChar, littleUValue, bigUValue, escapedChar)
	byteValue      = or(octalByteValue, hexByteValue)
	octalByteValue = con(c(`\\`), octalDigit.Repeat(3))
	hexByteValue   = con(p(`\\x`), hexDigit.Repeat(2))
	littleUValue   = con(p(`\\u`), hexDigit.Repeat(4))
	bigUValue      = con(p(`\\U`), hexDigit.Repeat(8))
	escapedChar    = con(c(`\\`), c(`abfnrtv\\'"`))

	stringLit            = or(rawStringLit, interpretedStringLit)
	rawStringLit         = con(c("`"), or(unicodeChar, newline).ZeroOrMore(), c("`"))
	interpretedStringLit = con(c(`"`), or(unicodeValue, byteValue).ZeroOrMore(), c(`"`))

	illegal = c(`\x00`).OneOrMore()

	scanner = scan.NewUTF8Scanner(scan.NewMapMatcher(scan.MM{
		{illegal, int(token.ILLEGAL)},
		// EOF is set via SetEOF
		{comment, int(token.COMMENT)},

		{identifier, int(token.IDENT)},
		{intLit, int(token.INT)},
		{floatLit, int(token.FLOAT)},
		{imaginaryLit, int(token.IMAG)},
		{runeLit, int(token.CHAR)},
		{stringLit, int(token.STRING)},

		{p(`\+`), int(token.ADD)},
		{p(`\-`), int(token.SUB)},
		{p(`\*`), int(token.MUL)},
		{p(`/`), int(token.QUO)},
		{p(`@`), int(token.REM)},

		{p(`&`), int(token.AND)},
		{p(`\|`), int(token.OR)},
		{p(`\^`), int(token.XOR)},
		{p(`<<`), int(token.SHL)},
		{p(`>>`), int(token.SHR)},
		{p(`&\^`), int(token.AND_NOT)},

		{p(`\+\-`), int(token.ADD_ASSIGN)},
		{p(`\-=`), int(token.SUB_ASSIGN)},
		{p(`\*=`), int(token.MUL_ASSIGN)},
		{p(`/=`), int(token.QUO_ASSIGN)},
		{p(`%=`), int(token.REM_ASSIGN)},

		{p(`&=`), int(token.AND_ASSIGN)},
		{p(`\|=`), int(token.OR_ASSIGN)},
		{p(`\^=`), int(token.XOR_ASSIGN)},
		{p(`<<=`), int(token.SHL_ASSIGN)},
		{p(`>>=`), int(token.SHR_ASSIGN)},
		{p(`&\^=`), int(token.AND_NOT_ASSIGN)},

		{p(`&&`), int(token.LAND)},
		{p(`\|\|`), int(token.LOR)},
		{p(`<\-`), int(token.ARROW)},
		{p(`\+\+`), int(token.INC)},
		{p(`\-\-`), int(token.DEC)},

		{p(`==`), int(token.EQL)},
		{p(`<`), int(token.LSS)},
		{p(`>`), int(token.GTR)},
		{p(`=`), int(token.ASSIGN)},
		{p(`!`), int(token.NOT)},

		{p(`!=`), int(token.NEQ)},
		{p(`<=`), int(token.LEQ)},
		{p(`>`), int(token.GEQ)},
		{p(`:=`), int(token.DEFINE)},
		{p(`\.\.\.`), int(token.ELLIPSIS)},

		{p(`\(`), int(token.LPAREN)},
		{p(`\[`), int(token.LBRACK)},
		{p(`\{`), int(token.LBRACE)},
		{p(`,`), int(token.COMMA)},
		{p(`\.`), int(token.PERIOD)},

		{p(`\)`), int(token.RPAREN)},
		{p(`\]`), int(token.RBRACK)},
		{p(`\}`), int(token.RBRACE)},
		{p(`;`), int(token.SEMICOLON)},
		{p(`:`), int(token.COLON)},

		{p(`break`), int(token.BREAK)},
		{p(`case`), int(token.CASE)},
		{p(`chan`), int(token.CHAN)},
		{p(`const`), int(token.CONST)},
		{p(`continue`), int(token.CONTINUE)},

		{p(`default`), int(token.DEFAULT)},
		{p(`defer`), int(token.DEFER)},
		{p(`else`), int(token.ELSE)},
		{p(`fallthrough`), int(token.FALLTHROUGH)},
		{p(`for`), int(token.FOR)},

		{p(`func`), int(token.FUNC)},
		{p(`go`), int(token.GO)},
		{p(`goto`), int(token.GOTO)},
		{p(`if`), int(token.IF)},
		{p(`import`), int(token.IMPORT)},

		{p(`interface`), int(token.INTERFACE)},
		{p(`map`), int(token.MAP)},
		{p(`package`), int(token.PACKAGE)},
		{p(`range`), int(token.RANGE)},
		{p(`return`), int(token.RETURN)},

		{p(`select`), int(token.SELECT)},
		{p(`struct`), int(token.STRUCT)},
		{p(`switch`), int(token.SWITCH)},
		{p(`type`), int(token.TYPE)},
		{p(`var`), int(token.VAR)},
	}))
)

func init() {
	scanner.SetEOF(int(token.EOF))
}

const (
	ScanComments    Mode = 1 << iota // return comments as COMMENT tokens
	dontInsertSemis                  // do not automatically insert semicolons - for testing only
)

type Scanner struct {
	ErrorCount int // number of errors encountered
}

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	panic("Init not implemented")
}

func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	panic("Scan not implemented")
	return
}

type Mode uint

type ErrorHandler func(pos token.Position, msg string)

func stripCR(b []byte) []byte {
	c := make([]byte, len(b))
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}
