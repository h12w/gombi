package scanner

import (
	"fmt"
	"go/token"
	"path/filepath"
	"unicode/utf8"

	"github.com/hailiang/gombi/scan"
)

const (
	tNewline = 1000 + iota
	tWhitespace
	tLineComment
	tGeneralComment
	tRawStringLit
	tInterpretedStringLit
)

const (
	firstOp = token.ADD
	lastOp  = token.COLON
)

var (
	c     = scan.Char
	s     = scan.Str
	p     = scan.Pat
	con   = scan.Con
	or    = scan.Or
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
	empty         = s(``)
	whitespaces   = c(` \t\r`).OneOrMore()

	lineComment    = con(s(`//`), unicodeChar.ZeroOrMore(), or(newline, empty))
	generalComment = con(s(`/*`), any.ZeroOrMore().Ungreedy(), s(`*/`))
	//comment        = or(lineComment, generalComment)

	identifier = con(letter, or(letter, unicodeDigit).ZeroOrMore())

	intLit     = or(hexLit, decimalLit, octalLit)
	decimalLit = con(c(`1-9`), decimalDigit.ZeroOrMore())
	octalLit   = con(s(`0`), octalDigit.ZeroOrMore())
	hexLit     = con(s(`0`), c(`xX`), hexDigit.OneOrMore())

	floatLit = or(
		con(decimals, s(`.`), decimals.ZeroOrOne(), exponent.ZeroOrOne()),
		con(decimals, exponent),
		con(s(`.`), decimals, exponent.ZeroOrOne()))
	decimals = decimalDigit.OneOrMore()
	exponent = con(c(`eE`), c(`\+\-`).ZeroOrOne(), decimals)

	imaginaryLit = con(or(decimals, floatLit), c(`i`))

	runeLit        = con(c(`'`), or(unicodeValue, byteValue), c(`'`))
	unicodeValue   = or(unicodeChar, littleUValue, bigUValue, escapedChar)
	byteValue      = or(octalByteValue, hexByteValue)
	octalByteValue = con(s(`\`), octalDigit.Repeat(3))
	hexByteValue   = con(s(`\x`), hexDigit.Repeat(2))
	littleUValue   = con(s(`\u`), hexDigit.Repeat(4))
	bigUValue      = con(s(`\U`), hexDigit.Repeat(8))
	escapedChar    = con(s(`\`), c(`abfnrtv\\'"`))

	//stringLit            = or(rawStringLit, interpretedStringLit)
	rawStringLit         = con(s("`"), or(unicodeChar, newline).ZeroOrMore().Ungreedy(), s("`"))
	interpretedStringLit = con(s(`"`), or(unicodeValue, byteValue).ZeroOrMore().Ungreedy(), s(`"`))

	illegal = c(`\x00`).OneOrMore()

	scanner = scan.NewScanner(scan.NewMapMatcher(scan.MM{
		{newline, tNewline},
		{whitespaces, tWhitespace},

		{illegal, int(token.ILLEGAL)},
		// EOF is set via SetEOF
		// COMMENT
		{lineComment, tLineComment},
		{generalComment, tGeneralComment},

		{imaginaryLit, int(token.IMAG)},
		{floatLit, int(token.FLOAT)},
		{intLit, int(token.INT)},
		{runeLit, int(token.CHAR)},
		{rawStringLit, tRawStringLit},
		{interpretedStringLit, tInterpretedStringLit},

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

		{s(`==`), int(token.EQL)},
		{s(`=`), int(token.ASSIGN)},

		{s(`!=`), int(token.NEQ)},
		{s(`!`), int(token.NOT)},
		{s(`<=`), int(token.LEQ)},
		{s(`<-`), int(token.ARROW)},
		{s(`<`), int(token.LSS)},
		{s(`>=`), int(token.GEQ)},
		{s(`>`), int(token.GTR)},
		{s(`:=`), int(token.DEFINE)},
		{s(`:`), int(token.COLON)},
		{s(`...`), int(token.ELLIPSIS)},
		{s(`.`), int(token.PERIOD)},

		{s(`(`), int(token.LPAREN)},
		{s(`[`), int(token.LBRACK)},
		{s(`{`), int(token.LBRACE)},
		{s(`,`), int(token.COMMA)},

		{s(`)`), int(token.RPAREN)},
		{s(`]`), int(token.RBRACK)},
		{s(`}`), int(token.RBRACE)},
		{s(`;`), int(token.SEMICOLON)},

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
		{s(`if`), int(token.IF)},
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
	s          *scan.Scanner
	ErrorCount int // number of errors encountered

	pos     int
	tok     token.Token
	lastTok token.Token
	val     []byte

	file *token.File  // source file handle
	dir  string       // directory portion of file.Name()
	err  ErrorHandler // error reporting; or nil
	mode Mode         // scanning mode
}

func skipBOM(buf []byte) []byte {
	r, size := utf8.DecodeRune(buf)
	if size > 0 && r == 0xFEFF {
		buf = buf[size:]
	}
	return buf
}

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	s.s = scanner
	s.s.SetBuffer(skipBOM(src))

	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.err = err
	s.mode = mode
}

func (s *Scanner) scan() {
	s.pos = s.s.Pos()
	if !s.s.Scan() {
		s.tok = token.ILLEGAL
		s.val = s.s.Token().Value
		fmt.Println(s.s.Error()) // DEBUG
	} else {
		s.lastTok = s.tok
		t := s.s.Token()
		s.tok, s.val = token.Token(t.ID), t.Value
	}
	switch s.tok {
	case tNewline:
		s.file.AddLine(s.s.Pos())
	case tLineComment:
		s.addLineFromValue()
		s.tok = token.COMMENT
		if s.val[len(s.val)-1] == '\n' {
			s.val = s.val[:len(s.val)-1]
		}
	case tGeneralComment:
		s.addLineFromValue()
		s.tok = token.COMMENT
	case tRawStringLit:
		s.addLineFromValue()
		s.tok = token.STRING
		s.val = stripCR(s.val)
	case tInterpretedStringLit:
		s.tok = token.STRING
	}
	if s.tok != token.SEMICOLON && firstOp <= s.tok && s.tok <= lastOp {
		s.val = nil
	}
	if s.tok == token.COMMENT {
		s.val = stripCR(s.val)
	}
}

func (s *Scanner) addLineFromValue() {
	for i, c := range s.val {
		if c == '\n' {
			s.file.AddLine(s.pos + i)
		}
	}
}

func (s *Scanner) semiNeeded() bool {
	if s.tok == tNewline {
		switch s.lastTok {
		case token.IDENT, token.INT, token.FLOAT, token.IMAG, token.CHAR,
			token.STRING, token.BREAK, token.CONTINUE, token.FALLTHROUGH,
			token.RETURN, token.INC, token.DEC, token.RPAREN, token.RBRACK,
			token.RBRACE:
			return true
		}
	}
	return false
}

func (s *Scanner) Scan() (token.Pos, token.Token, string) {
	if s.semiNeeded() { // second time
		s.lastTok = token.SEMICOLON
		return s.file.Pos(s.pos), s.tok, string(s.val)
	}
	for {
		s.scan()
		if s.semiNeeded() { // first time
			return s.file.Pos(s.pos), token.SEMICOLON, ""
		}
		switch s.tok {
		case tNewline, tWhitespace:
			continue
		}
		break
	}
	return s.file.Pos(s.pos), s.tok, string(s.val)
}

type Mode uint

type ErrorHandler func(pos token.Position, msg string)

func stripCR(b []byte) []byte {
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			b[i] = ch
			i++
		}
	}
	return b[:i]
}
