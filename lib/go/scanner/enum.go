package scanner

import "go/token"

const (
	tSemi        = int(token.SEMICOLON)
	tComment     = int(token.COMMENT)
	tString      = int(token.STRING)
	tEOF         = int(token.EOF)
	tIllegal     = int(token.ILLEGAL)
	tIdent       = int(token.IDENT)
	tRParen      = int(token.RPAREN)
	tRBrack      = int(token.RBRACK)
	tRBrace      = int(token.RBRACE)
	tInt         = int(token.INT)
	tFloat       = int(token.FLOAT)
	tImag        = int(token.IMAG)
	tChar        = int(token.CHAR)
	tReturn      = int(token.RETURN)
	tInc         = int(token.INC)
	tDec         = int(token.DEC)
	tBreak       = int(token.BREAK)
	tContinue    = int(token.CONTINUE)
	tFallthrough = int(token.FALLTHROUGH)
	lastGoToken  = int(token.VAR)
	firstOp      = int(token.ADD)
	lastOp       = int(token.COLON)
)

const (
	tNewline = (lastGoToken + 1) + iota
	tWhitespace
	tLineComment
	tGeneralCommentSL
	tGeneralCommentML
	tRawStringLit
	tInterpretedStringLit
	tokenCount
)
