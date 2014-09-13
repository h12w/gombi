package scanner

import (
	"fmt"
	"unicode/utf8"

	"github.com/hailiang/dfa"
	"github.com/hailiang/gombi/scan"
)

func spec() (tokens, errors []scan.MID) {
	var (
		c     = scan.Char
		b     = scan.Between
		bb    = scan.BetweenByte
		s     = scan.Str
		con   = scan.Con
		or    = scan.Or
		ifNot = scan.IfNot
		class = scan.CharClass

		NUL           = s("\x00")
		BOM           = s("\uFEFF")
		anyByte       = bb(0, 255).Exclude(NUL)
		any           = b(0, utf8.MaxRune).Exclude(NUL, BOM)
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
		lineComment          = con(s(`//`), unicodeChar.ZeroOrMore(), newline)
		lineCommentEOF       = con(s(`//`), unicodeChar.ZeroOrMore())
		generalCommentSL     = con(s(`/*`), commentText(any.Exclude(s("\n"))), s(`/`))
		generalCommentML     = con(s(`/*`), commentText(any.Exclude(s("\n"))).ZeroOrOne(), s("\n"), commentText(any), s(`/`))
		ident                = con(letter, or(letter, unicodeDigit).ZeroOrMore())
		identifier           = ident.Exclude(keywords)
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
		bigUValue            = con(s(`\U00`), or(s(`10`), con(s(`0`), hexDigit)), hexDigit.Repeat(4))
		escapedChar          = con(s(`\`), c(`abfnrtv\'"`))
		unicodeValue         = or(unicodeChar.Exclude(c(`\`)), littleUValue, bigUValue, escapedChar)
		runeValue            = or(byteValue, unicodeValue.Exclude(s(`'`)))
		runeLit              = con(s(`'`), runeValue, s(`'`))
		rawStrValue          = or(unicodeChar.Exclude(c("`")), newline)
		rawStringLit         = con(s("`"), rawStrValue.ZeroOrMore(), s("`"))
		strValue             = or(unicodeValue.Exclude(s(`"`)), byteValue)
		interpretedStringLit = con(s(`"`), strValue.ZeroOrMore(), s(`"`))

		// errors //
		invalidBigU      = con(s(`\U`), hexDigit.Repeat(8)).Exclude(bigUValue)
		unknownEscape    = con(s(`\`), any.Exclude(octalDigit, c(`xUuabfnrtv\'"`)))
		incompleteEscape = con(s(`\`), or(
			empty,
			con(s(`x`), hexDigit.AtMost(1)),
			con(s(`u`), hexDigit.AtMost(3)),
			con(s(`U`), or(
				s(`0`).AtMost(2),
				s(`001`),
				con(s(`0010`), hexDigit.AtMost(3)),
				con(s(`000`), hexDigit.AtMost(4)),
			))))
		runeEscapeBigUErr    = con(s(`'`), invalidBigU, s(`'`))
		runeEscapeUnknownErr = con(s(`'`), unknownEscape, any.Exclude(s(`'`)).ZeroOrMore(), s(`'`))
		runeBOMErr           = s("'\uFEFF'")
		runeEscapeErr        = con(s(`'`), con(s(`\`), any.Exclude(s(`'`)).OneOrMore(), s(`'`))).Exclude(runeEscapeBigUErr, runeEscapeUnknownErr)
		runeErr              = or(
			con(s(`'`), runeValue.Complement(), s(`'`)),
			con(s(`'`), runeValue.AtLeast(2), s(`'`)),
		).Exclude(runeEscapeErr, runeEscapeUnknownErr, runeEscapeBigUErr, runeBOMErr)
		runeIncompleteEscapeErr = con(s(`'`), incompleteEscape)
		runeIncompleteErr       = con(s(`'`), or(empty, unicodeChar.Exclude(c(`\'`))))

		strIncompleteErr     = con(s(`"`), strValue.ZeroOrMore())
		rawStrIncompleteErr  = con(s("`"), rawStrValue.ZeroOrMore())
		commentIncompleteErr = con(s(`/*`), or(any.Exclude(c(`*`)).ZeroOrMore(), s(`*`)).Loop(ifNot('/')))
		octalLitErr          = con(s(`0`), octalDigit.ZeroOrMore(), c(`89`), decimalDigit.ZeroOrMore())
		hexLitErr            = con(s(`0`), c(`xX`))
		strWithNULErr        = con(s(`"`), strValue.ZeroOrMore(), NUL, or(NUL, strValue).ZeroOrMore(), s(`"`))
		strWithBOMErr        = con(s(`"`), or(strValue.ZeroOrOne(), BOM).OneOrMore(), s(`"`))
		strWithWrongUTF8Err  = con(s(`"`), or(anyByte.Exclude(s(`\`)), or(byteValue, escapedChar, littleUValue, bigUValue).ZeroOrOne()).Exclude(s(`"`)).OneOrMore(), s(`"`)).Exclude(strWithBOMErr)
		lineCommentBOMErr    = con(s(`//`), or(unicodeChar.ZeroOrOne(), BOM).OneOrMore(), or(newline, empty))
	)
	return []scan.MID{
			{whitespaces, tWhitespace},
			{s("\n"), tNewline},
			{lineComment, tLineComment},
			{lineCommentEOF, tLineCommentEOF},
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

			// error patterns that have to be ORed with token patterns
			{commentIncompleteErr, eCommentIncomplete},
			{octalLitErr, eOctalLit},
			{hexLitErr, eHexLit},
		},
		[]scan.MID{
			// error patterns that can be recognized by a second scan after an
			// eIllegal error
			{runeErr, eRune},
			{runeBOMErr, eRuneBOM},
			{runeEscapeErr, eEscape},
			{runeEscapeBigUErr, eEscapeBigU},
			{runeEscapeUnknownErr, eEscapeUnknown},
			{runeIncompleteEscapeErr, eIncompleteEscape},
			{runeIncompleteErr, eRuneIncomplete},
			{strIncompleteErr, eStrIncomplete},
			{rawStrIncompleteErr, eRawStrIncomplete},
			{strWithNULErr, eStrWithNUL},
			{strWithBOMErr, eStrWithBOM},
			{strWithWrongUTF8Err, eStrWithWrongUTF8},
			{lineCommentBOMErr, eCommentBOM},
		}
}

var (
	gTokenMatcher *scan.Matcher
	gErrorMatcher *scan.Matcher
)

func initMatcher() {
	tokenDefs, errorDefs := spec()
	gTokenMatcher = scan.NewMatcher(
		tEOF,
		eIllegal,
		tokenDefs,
	)
	gErrorMatcher = scan.NewMatcher(
		eErrorEOF,
		eErrorIllegal,
		errorDefs,
	)
	fmt.Println(gTokenMatcher.Count())
	fmt.Println(gErrorMatcher.Count())
	//fmt.Println(gErrorMatcher)
}

func getTokenMatcher() *scan.Matcher {
	if gTokenMatcher == nil {
		initMatcher()
	}
	return gTokenMatcher
}

func getErrorMatcher() *scan.Matcher {
	if gErrorMatcher == nil {
		initMatcher()
	}
	return gErrorMatcher
}

func init() {
	getTokenMatcher()
}
