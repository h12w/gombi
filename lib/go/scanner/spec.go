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
			return con(char.Exclude(c(`*`)).Repeat(), `*`).Loop(ifNot('/'))
		}
		keywords = or(`break`, `case`, `chan`, `const`, `continue`, `default`,
			`defer`, `else`, `fallthrough`, `for`, `func`, `go`, `goto`, `if`,
			`import`, `interface`, `map`, `package`, `range`, `return`, `select`,
			`struct`, `switch`, `type`, `var`)
		whitespaces          = c(" \t\r").AtLeast(1)
		lineComment          = con(`//`, unicodeChar.Repeat(), newline)
		lineCommentEOF       = con(`//`, unicodeChar.Repeat())
		generalCommentSL     = con(`/*`, commentText(any.Exclude("\n")), `/`)
		generalCommentML     = con(`/*`, commentText(any.Exclude("\n")).Optional(), "\n", commentText(any), `/`)
		ident                = con(letter, or(letter, unicodeDigit).Repeat())
		identifier           = ident.Exclude(keywords)
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
		interpretedStringLit = con(`"`, strValue.Repeat(), `"`)

		// errors //
		invalidBigU      = con(`\U`, hexDigit.Repeat(8)).Exclude(bigUValue)
		unknownEscape    = con(`\`, any.Exclude(octalDigit, c(`xUuabfnrtv\'"`)))
		incompleteEscape = con(`\`, or(
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
		runeEscapeUnknownErr = con(`'`, unknownEscape, any.Exclude(`'`).Repeat(), `'`)
		runeBOMErr           = con(`'`, BOM, `'`)
		runeEscapeErr        = con(`'`, con(`\`, any.Exclude(`'`).AtLeast(1), `'`)).Exclude(runeEscapeBigUErr, runeEscapeUnknownErr)
		runeErr              = or(
			con(`'`, runeValue.Complement(), `'`),
			con(`'`, runeValue.AtLeast(2), `'`),
		).Exclude(runeEscapeErr, runeEscapeUnknownErr, runeEscapeBigUErr, runeBOMErr)
		runeIncompleteEscapeErr = con(`'`, incompleteEscape)
		runeIncompleteErr       = con(`'`, or("", unicodeChar.Exclude(`\`, `'`)))

		strIncompleteErr     = con(`"`, strValue.Repeat())
		rawStrIncompleteErr  = con("`", rawStrValue.Repeat())
		commentIncompleteErr = con(`/*`, or(any.Exclude(`*`).Repeat(), `*`).Loop(ifNot('/')))
		octalLitErr          = con(`0`, octalDigit.Repeat(), c(`89`), decimalDigit.Repeat())
		hexLitErr            = con(`0`, c(`xX`))
		strWithNULErr        = con(`"`, strValue.Repeat(), NUL, or(NUL, strValue).Repeat(), `"`)
		strWithBOMErr        = con(`"`, or(strValue.Optional(), BOM).AtLeast(1), `"`)
		strWithWrongUTF8Err  = con(`"`, or(anyByte.Exclude(`\`), or(byteValue, escapedChar, littleUValue, bigUValue).Optional()).Exclude(`"`).AtLeast(1), `"`).Exclude(strWithBOMErr)
		lineCommentBOMErr    = con(`//`, or(unicodeChar.Optional(), BOM).AtLeast(1), or(newline, ""))
	)
	return []scan.MID{
			{whitespaces, tWhitespace},
			{"\n", tNewline},
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
