package scanner

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/hailiang/dfa"
	"github.com/hailiang/gombi/scan"
)

const (
	enableCache = false
)

func spec() (tokMatcher, errMatcher *scan.Matcher) {
	if enableCache {
		return tokMatcherCache.Init(), errMatcherCache.Init()
	}
	var (
		c     = scan.Char
		b     = scan.Between
		bb    = scan.BetweenByte
		s     = scan.Str
		con   = scan.Con
		and   = scan.And
		or    = scan.Or
		ifNot = scan.IfNot
		class = scan.CharClass

		NUL           = s("\x00")
		BOM           = s("\uFEFF")
		valid         = b(1, utf8.MaxRune).Exclude(BOM)
		newline       = s("\n")
		unicodeChar   = valid.Exclude(newline)
		unicodeLetter = class(`L`)
		unicodeDigit  = class(`Nd`)
		letter        = or(unicodeLetter, c(`_`))
		decimalDigit  = b('0', '9')
		octalDigit    = b('0', '7')
		hexDigit      = or(b('0', '9'), b('A', 'F'), b('a', 'f'))

		// http://www.cs.dartmouth.edu/~mckeeman/cs118/assignments/comment.html
		commentText = func(char *dfa.M) *dfa.M {
			return con(char.Exclude(`*`).Repeat(), `*`).Loop(ifNot('/'))
			//stars := s(`*`).AtLeast(1)
			//return con(or(char.Exclude(`*`), con(stars, char.Exclude(`/`))).Repeat(), stars)
		}
		keywords = or(`break`, `case`, `chan`, `const`, `continue`, `default`,
			`defer`, `else`, `fallthrough`, `for`, `func`, `go`, `goto`, `if`,
			`import`, `interface`, `map`, `package`, `range`, `return`, `select`,
			`struct`, `switch`, `type`, `var`)
		whitespaces          = c(" \t\r").AtLeast(1)
		lineCommentInfo      = con(`//line `, unicodeChar.AtLeast(1), newline)
		lineComment          = con(`//`, unicodeChar.Repeat(), newline).Exclude(lineCommentInfo)
		lineCommentEOF       = con(`//`, unicodeChar.Repeat())
		generalCommentSL     = con(`/*`, commentText(valid.Exclude("\n")), `/`)
		generalCommentML     = con(`/*`, commentText(valid.Exclude("\n")).Optional(), "\n", commentText(valid), `/`)
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
		escapedValue         = or(byteValue, littleUValue, bigUValue)
		unicodeValue         = or(unicodeChar.Exclude(`\`), escapedValue, escapedChar)
		runeValue            = unicodeValue.Exclude(`'`)
		runeLit              = con(`'`, runeValue, `'`)
		rawStrValue          = or(unicodeChar.Exclude("`"), newline)
		rawStrValues         = rawStrValue.Repeat()
		rawStringLit         = con("`", rawStrValue.Repeat(), "`")
		strValue             = unicodeValue.Exclude(`"`)
		strValues            = strValue.Repeat()
		interpretedStringLit = con(`"`, strValues, `"`)

		// errors //
		anyByte          = bb(0, 0xFF)
		anyRuneValue     = anyByte.Exclude(`'`, "\n")
		anyRuneValues    = anyByte.Exclude(`'`).Repeat()
		anyStrValues     = anyByte.Exclude(`"`, "\n").Repeat()
		anyRawStrValues  = anyByte.Exclude("`").Repeat()
		anyEscapedValue  = con(`\`, or(octalDigit, c("uUx")), anyByte.Exclude(`'`, `"`, "\n").AtMost(8))
		bigUNotInRange   = con(`\U`, hexDigit.Repeat(8)).Exclude(bigUValue)
		invalidUTF8      = b(0, 0x10ffff).InvalidPrefix()
		unknownEscape    = con(`\`, anyByte.Exclude(octalDigit, c(`xUuabfnrtv\'"`)))
		invalidEscape    = and(escapedValue.InvalidPrefix(), anyEscapedValue)
		incompleteEscape = and(escapedValue.Complement(), anyEscapedValue)
		wrongEscape      = or(invalidEscape, incompleteEscape)

		BOMErr                  = BOM
		BOMInRuneErr            = con(`'`, BOM, `'`)
		BOMInStrErr             = con(`"`, strValues, BOM, anyStrValues, `"`)
		BOMInRawStrErr          = con("`", rawStrValues, BOM, anyRawStrValues, "`")
		BOMInLineCommentErr     = con(`//`, unicodeChar.Repeat(), BOM, anyByte.Exclude(newline).Repeat(), or(newline, ""))
		NULErr                  = NUL
		NULInStrErr             = con(`"`, strValues, NUL, anyStrValues, `"`)
		bigURuneErr             = con(`'`, bigUNotInRange, `'`)
		bigUStrErr              = con(`"`, strValues, bigUNotInRange, anyStrValues, `"`)
		unknownEscapeInRuneErr  = con(`'`, unknownEscape, anyRuneValues, `'`)
		unknownEscapeInStrErr   = con(`"`, strValues, unknownEscape, anyStrValues, `"`)
		incompleteRuneEscapeErr = con(`'\`, anyRuneValues)
		incompleteRuneErr       = con(`'`, anyRuneValue.Exclude(`\`).Repeat())
		incompleteStrErr        = con(`"`, strValues)
		incompleteRawStrErr     = con("`", rawStrValue.Repeat())
		incompleteCommentErr    = con(`/*`, or(valid.Exclude(`*`).Repeat(), `*`).Loop(ifNot('/')))
		UTF8Err                 = invalidUTF8
		UTF8RuneErr             = con(`'`, invalidUTF8, `'`)
		UTF8StrErr              = con(`"`, strValues, invalidUTF8, anyStrValues, `"`)
		octalLitErr             = con(`0`, octalDigit.Repeat(), c(`89`), decimalDigit.Repeat())
		hexLitErr               = con(`0`, c(`xX`))
		strEscapeErr            = con(`"`, strValues, wrongEscape, anyStrValues, `"`).Exclude(bigUStrErr)
		runeEscapeErr           = con(`'`, wrongEscape, `'`)
		illegalRuneErr          = con(`'`, or(``, con(runeValue, anyRuneValue.AtLeast(1))), `'`)
	)
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
			{incompleteCommentErr, eIncompleteComment},
			{octalLitErr, eOctalLit},
			{hexLitErr, eHexLit},
		}),
		scan.NewMatcher(
			eErrorEOF,
			eErrorIllegal,
			[]scan.MID{
				// error patterns that can be recognized by a second scan after an
				// eIllegal error
				{BOMErr, eBOM},
				{BOMInRuneErr, eBOMInRune},
				{BOMInStrErr, eBOMInStr},
				{BOMInRawStrErr, eBOMInStr},
				{BOMInLineCommentErr, eBOMInComment},
				{NULErr, eNUL},
				{NULInStrErr, eNULInStr},
				{illegalRuneErr, eRune},
				{runeEscapeErr, eEscape},
				{strEscapeErr, eEscape},
				{bigURuneErr, eBigU},
				{bigUStrErr, eBigU},
				{unknownEscapeInRuneErr, eEscapeUnknown},
				{unknownEscapeInStrErr, eEscapeUnknown},
				{incompleteRuneEscapeErr, eIncompleteEscape},
				{incompleteRuneErr, eIncompleteRune},
				{incompleteStrErr, eIncompleteStr},
				{incompleteRawStrErr, eIncompleteRawStr},
				{UTF8Err, eUTF8},
				{UTF8RuneErr, eUTF8Rune},
				{UTF8StrErr, eUTF8Str},
			})
	if !enableCache {
		f, _ := os.Create("cache.go")
		fmt.Fprint(f, `
package scanner
import (
	"github.com/hailiang/dfa"
	"github.com/hailiang/gombi/scan"
)
`)
		fmt.Fprintln(f, "var tokMatcherCache = ")
		tokMatcher.WriteGo(f)
		fmt.Fprintln(f, "var errMatcherCache = ")
		errMatcher.WriteGo(f)
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
