package parser

import "github.com/hailiang/gombi/parse"

var (
	term = parse.Term
	rule = parse.Rule
	or   = parse.Or
	con  = parse.Con
	self = parse.Self

	comma      = term(";")
	leftParen  = term("(")
	rightParen = term(")")
	importTok  = term("import")

	sourceFile = con(packageClause, comma, con(importDecl, comma).ZeroOrMore(), con(topLevelDecl, comma).ZeroOrMore())

//	importDecl = con(importTok, or(importSpec, con(leftParen, )
)
