package scanner

import "go/token"

// compatible enums
const (
	eIllegal = int(token.ILLEGAL)

	tEOF        = int(token.EOF)
	tComment    = int(token.COMMENT)
	tIdentifier = int(token.IDENT)
	tInt        = int(token.INT)
	tFloat      = int(token.FLOAT)
	tImag       = int(token.IMAG)
	tRune       = int(token.CHAR)
	tString     = int(token.STRING)

	firstOp       = tAdd
	tAdd          = int(token.ADD)
	tSub          = int(token.SUB)
	tMul          = int(token.MUL)
	tQuo          = int(token.QUO)
	tRem          = int(token.REM)
	tAnd          = int(token.AND)
	tOr           = int(token.OR)
	tXor          = int(token.XOR)
	tShl          = int(token.SHL)
	tShr          = int(token.SHR)
	tAndNot       = int(token.AND_NOT)
	tAddAssign    = int(token.ADD_ASSIGN)
	tSubAssign    = int(token.SUB_ASSIGN)
	tMulAssign    = int(token.MUL_ASSIGN)
	tQuoAssign    = int(token.QUO_ASSIGN)
	tRemAssign    = int(token.REM_ASSIGN)
	tAndAssign    = int(token.AND_ASSIGN)
	tOrAssign     = int(token.OR_ASSIGN)
	tXorAssign    = int(token.XOR_ASSIGN)
	tShlAssign    = int(token.SHL_ASSIGN)
	tShrAssign    = int(token.SHR_ASSIGN)
	tAndNotAssign = int(token.AND_NOT_ASSIGN)
	tLogicAnd     = int(token.LAND)
	tLogicOr      = int(token.LOR)
	tArrow        = int(token.ARROW)
	tInc          = int(token.INC)
	tDec          = int(token.DEC)
	tEqual        = int(token.EQL)
	tLess         = int(token.LSS)
	tGreater      = int(token.GTR)
	tAssign       = int(token.ASSIGN)
	tNot          = int(token.NOT)
	tNotEqual     = int(token.NEQ)
	tLessEqual    = int(token.LEQ)
	tGreaterEqual = int(token.GEQ)
	tDefine       = int(token.DEFINE)
	tEllipsis     = int(token.ELLIPSIS)
	tLeftParen    = int(token.LPAREN)
	tLeftBrack    = int(token.LBRACK)
	tLeftBrace    = int(token.LBRACE)
	tComma        = int(token.COMMA)
	tPeriod       = int(token.PERIOD)
	tRightParen   = int(token.RPAREN)
	tRightBrack   = int(token.RBRACK)
	tRightBrace   = int(token.RBRACE)
	tSemiColon    = int(token.SEMICOLON)
	tColon        = int(token.COLON)
	lastOp        = tColon

	tBreak       = int(token.BREAK)
	tCase        = int(token.CASE)
	tChan        = int(token.CHAN)
	tConst       = int(token.CONST)
	tContinue    = int(token.CONTINUE)
	tDefault     = int(token.DEFAULT)
	tDefer       = int(token.DEFER)
	tElse        = int(token.ELSE)
	tFallthrough = int(token.FALLTHROUGH)
	tFor         = int(token.FOR)
	tFunc        = int(token.FUNC)
	tGo          = int(token.GO)
	tGoto        = int(token.GOTO)
	tIf          = int(token.IF)
	tImport      = int(token.IMPORT)
	tInterface   = int(token.INTERFACE)
	tMap         = int(token.MAP)
	tPackage     = int(token.PACKAGE)
	tRange       = int(token.RANGE)
	tReturn      = int(token.RETURN)
	tSelect      = int(token.SELECT)
	tStruct      = int(token.STRUCT)
	tSwitch      = int(token.SWITCH)
	tType        = int(token.TYPE)
	tVar         = int(token.VAR)
	lastGoToken  = tVar
)

// additional enums
const (
	// token start
	tNewline = (lastGoToken + 1) + iota
	tWhitespace
	tLineComment
	tLineCommentEOF
	tGeneralCommentSL
	tGeneralCommentML
	tRawStringLit
	tInterpretedStringLit
	// token end

	// error start
	eRune //94
	eRuneBOM
	eEscape
	eEscapeUnknown
	eEscapeBigU
	eRuneIncomplete
	eIncompleteEscape
	eStrIncomplete
	eRawStrIncomplete
	eCommentIncomplete
	eOctalLit
	eHexLit
	eStrWithNUL
	eStrWithBOM
	eStrWithWrongUTF8
	eCommentBOM
	eErrorEOF
	eErrorIllegal
	// error end
)
