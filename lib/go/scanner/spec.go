package scanner

import (
	"fmt"
	"unicode/utf8"

	"github.com/hailiang/dfa"
	"github.com/hailiang/gombi/scan"
)

const (
	tokMatcherCache = "tok.cache"
	errMatcherCache = "err.cache"
	enableCache     = false
)

func spec() (tokMatcher, errMatcher *scan.Matcher) {
	if enableCache {
		tokMatcher, _ = scan.LoadMatcher(tokMatcherCache)
		errMatcher, _ = scan.LoadMatcher(errMatcherCache)
		if tokMatcher != nil && errMatcher != nil {
			return
		}
	}
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
		anyByte       = bb(0x01, 0xff)
		any           = b(1, utf8.MaxRune).Exclude(BOM)
		newline       = s("\n")
		unicodeChar   = any.Exclude(newline)
		unicodeLetter = class(`L`)
		unicodeDigit  = class(`Nd`)
		letter        = or(unicodeLetter, c(`_`))
		decimalDigit  = b('0', '9')
		octalDigit    = b('0', '7')
		hexDigit      = or(b('0', '9'), b('A', 'F'), b('a', 'f'))

		// http://www.cs.dartmouth.edu/~mckeeman/cs118/assignments/comment.html
		commentText = func(char *dfa.M) *dfa.M {
			return con(char.Exclude(c(`*`)).Repeat(), `*`).Loop(ifNot('/'))
		}
		keywords = or(`break`, `case`, `chan`, `const`, `continue`, `default`,
			`defer`, `else`, `fallthrough`, `for`, `func`, `go`, `goto`, `if`,
			`import`, `interface`, `map`, `package`, `range`, `return`, `select`,
			`struct`, `switch`, `type`, `var`)
		whitespaces          = c(" \t\r").AtLeast(1)
		lineCommentInfo      = con(`//line `, unicodeChar.AtLeast(1), newline)
		lineComment          = con(`//`, unicodeChar.Repeat(), newline).Exclude(lineCommentInfo)
		lineCommentEOF       = con(`//`, unicodeChar.Repeat())
		generalCommentSL     = con(`/*`, commentText(any.Exclude("\n")), `/`)
		generalCommentML     = con(`/*`, commentText(any.Exclude("\n")).Optional(), "\n", commentText(any), `/`)
		identifier           = con(letter, or(letter, unicodeDigit).Repeat()).Exclude(keywords)
		hexLit               = con(`0`, c("xX"), hexDigit.AtLeast(1))
		decimalLit           = con(b('1', '9'), decimalDigit.Repeat())
		octalLit             = con(`0`, octalDigit.Repeat())
		intLit               = or(hexLit, decimalLit, octalLit)
		decimals             = decimalDigit.AtLeast(1)
		exponent             = con(c("eE"), c("+-").Optional(), decimals)
		floatLit1            = con(decimals, `.`, decimals.Optional(), exponent.Optional())
		floatLit2            = con(decimals, exponent)
		floatLit3            = con(`.`, decimals, exponent.Optional())
		floatLit             = or(floatLit1, floatLit2, floatLit3)
		imaginaryLit         = con(or(floatLit, decimals), `i`)
		hexByteValue         = con(`\x`, hexDigit.Repeat(2))
		octalByteValue       = con(`\`, octalDigit.Repeat(3))
		byteValue            = or(hexByteValue, octalByteValue)
		littleUValue         = con(`\u`, hexDigit.Repeat(4))
		bigUValue            = con(`\U00`, or(`10`, con(`0`, hexDigit)), hexDigit.Repeat(4))
		escapedChar          = con(`\`, c(`abfnrtv\'"`))
		unicodeValue         = or(unicodeChar.Exclude(`\`), littleUValue, bigUValue, escapedChar)
		runeValue            = or(byteValue, unicodeValue.Exclude(`'`))
		runeLit              = con(`'`, runeValue, `'`)
		rawStrValue          = or(unicodeChar.Exclude("`"), newline)
		rawStringLit         = con("`", rawStrValue.Repeat(), "`")
		strValue             = or(unicodeValue.Exclude(`"`), byteValue)
		strValues            = strValue.Repeat()
		interpretedStringLit = con(`"`, strValues, `"`)

		// errors //
		invalidBigU       = con(`\U`, hexDigit.Repeat(8)).Exclude(bigUValue)
		unknownRuneEscape = con(`\`, any.Exclude(octalDigit, c(`xUuabfnrtv\'`)))
		incompleteEscape  = con(`\`, or(
			"",
			con(`x`, hexDigit.AtMost(1)),
			con(`u`, hexDigit.AtMost(3)),
			con(`U`, or(
				``,
				`0`,
				`00`,
				con(`000`, hexDigit.AtMost(4)),
				`001`,
				con(`0010`, hexDigit.AtMost(3)),
			))))

		runeEscapeBigUErr    = con(`'`, invalidBigU, `'`)
		runeEscapeUnknownErr = con(`'`, unknownRuneEscape, any.Exclude(`'`).Repeat(), `'`)
		runeEscapeErr        = con(`'\`, any.Exclude(`'`).AtLeast(1), `'`).Exclude(runeEscapeBigUErr, runeEscapeUnknownErr)
		runeBOMErr           = con(`'`, BOM, `'`)
		runeErr              = or(
			con(`'`, runeValue.Complement().Exclude(`'`), `'`),
			con(`'`, runeValue.AtLeast(2), `'`),
		).Exclude(runeEscapeBigUErr, runeEscapeUnknownErr, runeEscapeErr, runeBOMErr)
		runeIncompleteEscapeErr = con(`'`, incompleteEscape)
		runeIncompleteErr       = con(`'`, or(``, unicodeChar.Exclude(`\`, `'`)))

		anyStrValues     = bb(0, 0xFF).Exclude(`"`).Repeat()
		strIncompleteErr = con(`"`, strValues)
		strNULErr        = con(`"`, strValues, NUL, anyStrValues, `"`)
		strBOMErr        = con(`"`, strValues, BOM, anyStrValues, `"`)
		wrongUTF8        = bb(1, 0xff).Repeat(1, 4).Exclude(b(0, 0x10ffff)).Minimize()
		strWrongUTF8Err  = con(`"`, strValues, wrongUTF8, anyStrValues, `"`).Exclude(strBOMErr)

		rawStrIncompleteErr  = con("`", rawStrValue.Repeat())
		commentIncompleteErr = con(`/*`, or(any.Exclude(`*`).Repeat(), `*`).Loop(ifNot('/')))
		lineCommentBOMErr    = con(`//`, or(unicodeChar.Optional(), BOM).AtLeast(1), or(newline, ""))
		octalLitErr          = con(`0`, octalDigit.Repeat(), c(`89`), decimalDigit.Repeat())
		hexLitErr            = con(`0`, c(`xX`))
	)
	_ = anyByte
	//wrongUTF8.SaveSVG("w.svg")
	//sss.SaveSVG("sss.svg")
	tokMatcher, errMatcher = scan.NewMatcher(
		tEOF,
		eIllegal,
		[]scan.MID{
			{whitespaces, tWhitespace},
			{"\n", tNewline},
			{lineComment, tLineComment},
			{lineCommentEOF, tLineCommentEOF},
			{lineCommentInfo, tLineCommentInfo},
			{generalCommentSL, tGeneralCommentSL},
			{generalCommentML, tGeneralCommentML},
			{identifier, tIdentifier},
			{intLit, tInt},
			{floatLit, tFloat},
			{imaginaryLit, tImag},
			{runeLit, tRune},
			{rawStringLit, tRawStringLit},
			{interpretedStringLit, tInterpretedStringLit},
			{`+`, tAdd},
			{`-`, tSub},
			{`*`, tMul},
			{`/`, tQuo},
			{`%`, tRem},
			{`&`, tAnd},
			{`|`, tOr},
			{`^`, tXor},
			{`<<`, tShl},
			{`>>`, tShr},
			{`&^`, tAndNot},
			{`+=`, tAddAssign},
			{`-=`, tSubAssign},
			{`*=`, tMulAssign},
			{`/=`, tQuoAssign},
			{`%=`, tRemAssign},
			{`&=`, tAndAssign},
			{`|=`, tOrAssign},
			{`^=`, tXorAssign},
			{`<<=`, tShlAssign},
			{`>>=`, tShrAssign},
			{`&^=`, tAndNotAssign},
			{`&&`, tLogicAnd},
			{`||`, tLogicOr},
			{`<-`, tArrow},
			{`++`, tInc},
			{`--`, tDec},
			{`==`, tEqual},
			{`<`, tLess},
			{`>`, tGreater},
			{`=`, tAssign},
			{`!`, tNot},
			{`!=`, tNotEqual},
			{`<=`, tLessEqual},
			{`>=`, tGreaterEqual},
			{`:=`, tDefine},
			{`...`, tEllipsis},
			{`(`, tLeftParen},
			{`[`, tLeftBrack},
			{`{`, tLeftBrace},
			{`,`, tComma},
			{`.`, tPeriod},
			{`)`, tRightParen},
			{`]`, tRightBrack},
			{`}`, tRightBrace},
			{`;`, tSemiColon},
			{`:`, tColon},
			{`break`, tBreak},
			{`case`, tCase},
			{`chan`, tChan},
			{`const`, tConst},
			{`continue`, tContinue},
			{`default`, tDefault},
			{`defer`, tDefer},
			{`else`, tElse},
			{`fallthrough`, tFallthrough},
			{`for`, tFor},
			{`func`, tFunc},
			{`go`, tGo},
			{`goto`, tGoto},
			{`if`, tIf},
			{`import`, tImport},
			{`interface`, tInterface},
			{`map`, tMap},
			{`package`, tPackage},
			{`range`, tRange},
			{`return`, tReturn},
			{`select`, tSelect},
			{`struct`, tStruct},
			{`switch`, tSwitch},
			{`type`, tType},
			{`var`, tVar},

			// error patterns that have to be ORed with token patterns
			{commentIncompleteErr, eCommentIncomplete},
			{octalLitErr, eOctalLit},
			{hexLitErr, eHexLit},
		}),
		scan.NewMatcher(
			eErrorEOF,
			eErrorIllegal,
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
				{strNULErr, eStrWithNUL},
				{strBOMErr, eStrWithBOM},
				{strWrongUTF8Err, eStrWithWrongUTF8},
				{lineCommentBOMErr, eCommentBOM},
			})
	if enableCache {
		tokMatcher.SaveCache(tokMatcherCache)
		errMatcher.SaveCache(errMatcherCache)
	}
	return
}

var (
	gTokenMatcher *scan.Matcher
	gErrorMatcher *scan.Matcher
)

func initMatcher() {
	gTokenMatcher, gErrorMatcher = spec()
	fmt.Println(gTokenMatcher.Count())
	fmt.Println(gErrorMatcher.Count())
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
