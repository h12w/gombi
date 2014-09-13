package scanner

import (
	"fmt"
	"unicode/utf8"

	"github.com/hailiang/dfa"
	"github.com/hailiang/gombi/scan"
)

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

		// http://www.cs.dartmouth.edu/~mckeeman/cs118/assignments/comment.html
		commentText = func(char *dfa.M) *dfa.M {
			return con(char.Exclude(c(`*`)).ZeroOrMore(), s(`*`)).Loop(ifNot('/'))
		}
		keywords = or(s(`break`), s(`case`), s(`chan`), s(`const`),
			s(`continue`), s(`default`), s(`defer`), s(`else`), s(`fallthrough`),
			s(`for`), s(`func`), s(`go`), s(`goto`), s(`if`), s(`import`),
			s(`interface`), s(`map`), s(`package`), s(`range`), s(`return`),
			s(`select`), s(`struct`), s(`switch`), s(`type`), s(`var`))
		empty                = s(``)
		whitespaces          = c(" \t\r").OneOrMore()
		lineComment          = con(s(`//`), unicodeChar.ZeroOrMore(), or(newline, empty))
		generalCommentSL     = con(s(`/*`), commentText(any.Exclude(s("\n"))), s(`/`))
		generalCommentML     = con(s(`/*`), commentText(any.Exclude(s("\n"))).ZeroOrOne(), s("\n"), commentText(any), s(`/`))
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
		runeValue            = or(byteValue, unicodeValue.Exclude(s(`'`)))
		runeLit              = con(s(`'`), runeValue, s(`'`))
		rawStringLit         = con(s("`"), or(unicodeChar.Exclude(c("`")), newline).ZeroOrMore(), s("`"))
		interpretedStringLit = con(s(`"`), or(unicodeValue.Exclude(s(`"`)), byteValue).ZeroOrMore(), s(`"`))

		//		unknownEscape        = con(s(`\`), any.Exclude(octalDigit, c(`xUuabfnrtv\'"`)))
		//		runeUnknownEscapeErr = con(s(`'`), unknownEscape.Exclude(s(`'`)), s(`'`))
		//		runeEscapeCharErr    = con(s(`'`), con(s(`\`), any.Exclude(s(`'`)).OneOrMore()).Exclude(byteValue, littleUValue, bigUValue, escapedChar), s(`'`)).Exclude(runeUnknownEscapeErr)
		//		runeErr              = or(
		//			con(s(`'`), runeValue.Complement(), s(`'`)),
		//			con(s(`'`), runeValue.AtLeast(2), s(`'`)),
		//		).Exclude(runeEscapeCharErr, runeUnknownEscapeErr)
	)
	return []scan.MID{
		{whitespaces, tWhitespace},
		{s("\n"), tNewline},
		{lineComment, tLineComment},
		{generalCommentSL, tGeneralCommentSL},
		{generalCommentML, tGeneralCommentML},
		{identifier, tIdentifier},
		{intLit, tInt},
		{floatLit, tFloat},
		{imaginaryLit, tImag},
		{runeLit, tRune},
		{rawStringLit, tRawStringLit},
		{interpretedStringLit, tInterpretedStringLit},
		{s(`+`), tAdd},
		{s(`-`), tSub},
		{s(`*`), tMul},
		{s(`/`), tQuo},
		{s(`%`), tRem},
		{s(`&`), tAnd},
		{s(`|`), tOr},
		{s(`^`), tXor},
		{s(`<<`), tShl},
		{s(`>>`), tShr},
		{s(`&^`), tAndNot},
		{s(`+=`), tAddAssign},
		{s(`-=`), tSubAssign},
		{s(`*=`), tMulAssign},
		{s(`/=`), tQuoAssign},
		{s(`%=`), tRemAssign},
		{s(`&=`), tAndAssign},
		{s(`|=`), tOrAssign},
		{s(`^=`), tXorAssign},
		{s(`<<=`), tShlAssign},
		{s(`>>=`), tShrAssign},
		{s(`&^=`), tAndNotAssign},
		{s(`&&`), tLogicAnd},
		{s(`||`), tLogicOr},
		{s(`<-`), tArrow},
		{s(`++`), tInc},
		{s(`--`), tDec},
		{s(`==`), tEqual},
		{s(`<`), tLess},
		{s(`>`), tGreater},
		{s(`=`), tAssign},
		{s(`!`), tNot},
		{s(`!=`), tNotEqual},
		{s(`<=`), tLessEqual},
		{s(`>=`), tGreaterEqual},
		{s(`:=`), tDefine},
		{s(`...`), tEllipsis},
		{s(`(`), tLeftParen},
		{s(`[`), tLeftBrack},
		{s(`{`), tLeftBrace},
		{s(`,`), tComma},
		{s(`.`), tPeriod},
		{s(`)`), tRightParen},
		{s(`]`), tRightBrack},
		{s(`}`), tRightBrace},
		{s(`;`), tSemiColon},
		{s(`:`), tColon},
		{s(`break`), tBreak},
		{s(`case`), tCase},
		{s(`chan`), tChan},
		{s(`const`), tConst},
		{s(`continue`), tContinue},
		{s(`default`), tDefault},
		{s(`defer`), tDefer},
		{s(`else`), tElse},
		{s(`fallthrough`), tFallthrough},
		{s(`for`), tFor},
		{s(`func`), tFunc},
		{s(`go`), tGo},
		{s(`goto`), tGoto},
		{s(`if`), tIf},
		{s(`import`), tImport},
		{s(`interface`), tInterface},
		{s(`map`), tMap},
		{s(`package`), tPackage},
		{s(`range`), tRange},
		{s(`return`), tReturn},
		{s(`select`), tSelect},
		{s(`struct`), tStruct},
		{s(`switch`), tSwitch},
		{s(`type`), tType},
		{s(`var`), tVar},
		//		{runeErr, eRune},
		//		{runeEscapeCharErr, eRuneEscapeChar},
		//		{runeUnknownEscapeErr, eRuneUnknownEscape},
	}
}

var globalMatcher *scan.Matcher

func getMatcher() *scan.Matcher {
	if globalMatcher == nil {
		globalMatcher = scan.NewMatcher(
			tEOF,
			eIllegal,
			tokenDefs(),
		)
		fmt.Println(globalMatcher.Size())
	}
	return globalMatcher
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
