package parser

import (
	"go/token"

	"h12.me/gombi/parse"
)

var (
	builder = parse.NewBuilder()
	term    = builder.Term
	or      = builder.Or
	con     = builder.Con
	EOF     = parse.EOF
	newRule = parse.NewRule

	mList = func(item *parse.R, sep string) *parse.R {
		items := con(item, sep).AtLeast(1)
		return or(
			item,
			con(item, sep),
			con(items, item),
			con(items, item, sep),
		)
	}

	sList = func(item *parse.R, sep string) *parse.R {
		return or(
			item,
			con(item, con(sep, item).AtLeast(1)))
	}

	semiList = func(item *parse.R) *parse.R {
		return mList(item, ";")
	}

	commaList = func(item *parse.R) *parse.R {
		return sList(item, ",")
	}

	declList = func(item *parse.R) *parse.R {
		return or(
			item,
			con("(", ")"),
			con("(", semiList(item), ")"),
		)
	}

	identifier   = term("identifier")
	stringLit    = term("string")
	runeLit      = term("rune")
	intLit       = term("int")
	floatLit     = term("float")
	imaginaryLit = term("imag")

	// sourceExpr
	sourceExpr = con(or(expr, type_), ";", EOF).As("sourceExpr")

	// Packages

	sourceFile = con(packageClause, ";", or(
		EOF,
		con(importDecls, EOF),
		con(topLevelDecls, EOF),
		con(importDecls, topLevelDecls, EOF),
	)).As("sourceFile")
	importDecls   = con(importDecl, ";").AtLeast(1).As("importDecls")
	topLevelDecls = con(topLevelDecl, ";").AtLeast(1).As("topLevelDecls")

	packageClause = con("package", packageName).As("packageClause")
	packageName   = identifier // As??

	importDecl = con("import", declList(importSpec)).As("importDecl")
	importSpec = or(
		importPath,
		con(packageName, importPath),
		con("dot", importPath)).As("importSpec")
	importPath = stringLit

	// Declarations

	decl         = or(constDecl, typeDecl, varDecl).As("decl")
	topLevelDecl = or(decl, funcDecl, methodDecl).As("topLevelDecl")

	constDecl = con("const", declList(constSpec)).As("constDecl")
	valueSpec = or(
		identifierList,
		con(identifierList, "=", exprList),
		con(identifierList, type_, "=", exprList),
	).As("valueSpec")
	constSpec      = valueSpec
	identifierList = commaList(identifier).As("identifierList")
	exprList       = commaList(expr).As("exprList")

	typeDecl = con("type", declList(typeSpec)).As("typeDecl")
	typeSpec = con(identifier, type_).As("typeSpec")

	varDecl = con("var", declList(varSpec)).As("varDecl")
	varSpec = or(
		valueSpec,
		con(identifierList, type_),
	).As("varSpec")

	shortVarDecl = con(identifierList, ":=", exprList).As("shortVarDecl")

	funcDecl = con("func", funcName, or(func_, signature)).As("funcDecl")
	funcName = identifier
	func_    = con(signature, funcBody).As("func_")
	funcBody = block

	methodDecl = con("func", receiver, methodName, or(func_, signature)).As("methodDecl")
	receiver   = con("(", or(
		baseTypeName,
		con(identifier, baseTypeName),
		con("*", baseTypeName),
		con(identifier, "*", baseTypeName),
	), ")").As("receiver")
	baseTypeName = identifier

	// block

	block = or(
		con("{", "}"),
		con("{", semiList(stmt), "}"),
	).As("block")
	stmtList = con(stmt, ";").AtLeast(1).As("stmtList")

	// Statements
	stmt = newRule().As("stmt")
	_    = stmt.Define(or(
		decl, labeledStmt, simpleStmt,
		goStmt, returnStmt, breakStmt, continueStmt, gotoStmt,
		fallthroughStmt, block, ifStmt, switchStmt, selectStmt, forStmt,
		deferStmt))
	simpleStmt = or( /*emptyStmt,*/ exprStmt, sendStmt, incDecStmt, assignment, shortVarDecl).As("simpleStmt")

	//emptyStmt = null

	labeledStmt = con(label, ":", stmt).As("labeledStmt")
	label       = identifier

	exprStmt = expr

	sendStmt = con(channel, "<-", expr).As("sendStmt")
	channel  = expr

	incDecStmt = con(expr, or("++", "--")).As("incDecStmt")

	assignment = con(exprList, assignOp, exprList).As("assignment")
	assignOp   = or("=", "+=", "-=", "|=", "^=", "*=", "/=", "%=", "<<=", ">>=", "&=", "&^=").As("assignOp")

	condition = expr
	initCond  = or(condition, con(simpleStmt, ";", condition))
	ifStmt    = newRule().As("ifStmt")
	_         = ifStmt.Define(con("if",
		initCond,
		or(block, con(block, "else", or(ifStmt, block)))))

	switchStmt     = or(exprSwitchStmt, typeSwitchStmt).As("switchStmt")
	exprSwitchStmt = con("switch", or("{", con(initCond, "{")), or("}", con(exprCaseClause.AtLeast(1), "}"))).As("exprSwitchStmt")
	commaStmtList  = or(":", con(":", stmtList))
	exprCaseClause = con(exprSwitchCase, commaStmtList).As("exprCaseClause")
	exprSwitchCase = con("case", or(exprList, "default")).As("exprSwitchCase")

	typeSwitchStmt  = con("switch", or(typeSwitchGuard, con(simpleStmt, ";", typeSwitchGuard)), "{", or("}", con(typeCaseClause.AtLeast(1), "}"))).As("typeSwitchStmt")
	typeSwitchGuard = con(or(primaryExpr, con(identifier, ":=", primaryExpr)), ".", "(", "type", ")").As("typeSwitchGuard")
	typeCaseClause  = con(typeSwitchCase, commaStmtList).As("typeCaseClause")
	typeSwitchCase  = or(con("case", typeList), "default").As("typeSwitchCase")
	typeList        = commaList(type_).As("typeList")

	forStmt   = con("for", or(block, con(or(condition, forClause, rangeClause), block))).As("forStmt")
	forClause = con(
		or(";", con(initStmt, ";")),
		or(";", con(condition, ";")),
		or(";", con(postStmt, ";")),
	).As("forClause")
	initStmt    = simpleStmt
	postStmt    = simpleStmt
	rangeClause = con(or(con(exprList, "="), con(identifierList, ":=")), "range", expr).As("rangeClause")

	goStmt = con("go", expr).As("goStmt")

	selectStmt = con("select", or(
		con("{", "}"),
		con("{", commClause.AtLeast(1), "}"))).As("selectStmt")
	commClause = con(commCase, commaStmtList).As("commClause")
	commCase   = or(con("case", or(sendStmt, recvStmt)), "default").As("commCase")
	recvStmt   = or(
		recvExpr,
		con(or(con(exprList, "="), con(identifierList, ":=")), recvExpr)).As("recvStmt")
	recvExpr = expr

	returnStmt = or(
		"return",
		con("return", exprList)).As("returnStmt")

	breakStmt = or(
		"break",
		con("break", label)).As("breakStmt")

	continueStmt = or(
		"continue",
		con("continue", label)).As("continueStmt")

	gotoStmt = con("goto", label).As("gotoStmt")

	fallthroughStmt = term("fallthrough")

	//deferStmt = con("defer", expr).As("deferStmt")
	deferStmt = con("defer", callExpr).As("deferStmt") // restrict to callExpr

	// Expr

	expr      = newRule().As("expr")
	_         = expr.Define(or(unaryExpr, con(expr, binaryOp, unaryExpr)))
	unaryExpr = newRule().As("unaryExpr")
	_         = unaryExpr.Define(or(primaryExpr, con(unaryOp, unaryExpr)))

	operand     = or(literal, operandName /*methodExpr,*/, con("(", expr, ")")).As("operand") // remove ambiguous
	literal     = or(basicLit, compositeLit, funcLit).As("literal")
	basicLit    = or(intLit, floatLit, imaginaryLit, runeLit, stringLit).As("basicLit")
	operandName = identifier // remove ambiguous
	//operandName = or(identifier, qualifiedIdent).As("operandName")

	qualifiedIdent = con(packageName, ".", identifier).As("qualifiedIdent")

	compositeLit = con(literalType, literalValue).As("compositeLit")
	literalType  = or(arrayType, structType, con("[", "...", "]", elementType),
		sliceType, mapType, typeName).As("literalType")
	literalValue = newRule().As("literalValue")
	_            = literalValue.Define(or(con("{", "}"), con("{", elementList, "}")))
	elementList  = mList(element, ",").As("elementList")
	element      = or(
		value,
		con(key, ":", value)).As("element")
	key = expr
	//key          = or(fieldName, elementIndex).As("key")
	//fieldName    = identifier
	//elementIndex = expr
	value = or(expr, literalValue).As("value")

	funcLit = con("func", func_).As("funcLit")

	primaryExpr = newRule().As("primaryExpr")
	_           = primaryExpr.Define(or(
		operand,
		/*conversion, builtinCall,*/ // remove ambiguous
		con(primaryExpr, selector),
		con(primaryExpr, index),
		con(primaryExpr, slice),
		con(primaryExpr, typeAssertion),
		callExpr)) //callOrConversion
	callExpr = con(primaryExpr, call)
	selector = con(".", identifier).As("selector")
	index    = con("[", expr, "]").As("index")
	slice    = con("[", or(
		":",
		con(expr, ":"),
		con(":", expr),
		con(expr, ":", expr),
		con(":", expr, ":", expr),
		con(expr, ":", expr, ":", expr),
	), "]").As("slice")
	typeAssertion = con(".", "(", type_, ")").As("typeAssertion")
	call          = or(
		con("(", ")"),
		con("(", argumentList, ")"),
		con("(", argumentList, ",", ")"),
	).As("call")
	argumentList = or(exprList, con(exprList, "...")).As("argumentList")

	binaryOp = or("||", "&&", relOp, addOp, mulOp).As("binaryOp")
	relOp    = or("==", "!=", "<", "<=", ">", ">=")
	addOp    = or("+", "-", "|", "^")
	mulOp    = or("*", "/", "%", "<<", ">>", "&", "&^")
	unaryOp  = or("+", "-", "!", "^", "*", "&", "<-")

	methodExpr   = con(receiverType, ".", methodName).As("methodExpr")
	receiverType = newRule().As("receiverType")
	_            = receiverType.Define(or(typeName, con("(", "*", typeName, ")"), con("(", receiverType, ")")))

	// remove ambiguous
	//conversion = con(type_, "(", expr, opt(","), ")").As("conversion")

	// Built-in functions

	//builtinCall = con(identifier, "(", opt(builtinArgs, opt(",")), ")").As("builtinCall")
	//builtinArgs = or(con(type_, opt(",", argumentList)), argumentList).As("builtinArgs")

	// Types

	type_    = newRule().As("type_")
	_        = type_.Define(or(typeName, typeLit, con("(", type_, ")")))
	typeName = or(identifier, qualifiedIdent).As("typeName")
	typeLit  = or(arrayType, structType, pointerType, funcType, interfaceType,
		sliceType, mapType, channelType).As("typeLit")

	arrayType   = con("[", arrayLength, "]", elementType).As("arrayType")
	arrayLength = expr
	elementType = type_

	sliceType = con("[", "]", elementType).As("sliceType")

	structType     = con("struct", "{", semiList(fieldDecl), "}").As("structType")
	field          = or(con(identifierList, type_), anonymousField)
	fieldDecl      = or(field, con(field, tag)).As("fieldDecl")
	anonymousField = or(typeName, con("*", typeName)).As("anonymousField")
	tag            = stringLit

	pointerType = con("*", baseType).As("pointerType")
	baseType    = type_

	funcType   = con("func", signature).As("funcType")
	signature  = or(parameters, con(parameters, result)).As("signature")
	result     = or(parameters, type_).As("result")
	parameters = or(
		con("(", ")"),
		con("(", parameterList, ")")).As("parameters")
	parameterList = mList(parameterDecl, ",").As("parameterList")
	parameterDecl = or(
		type_,
		con(identifierList, type_),
		con(identifierList, "...", type_)).As("parameterDecl")

	interfaceType = con("interface", or(
		con("{", "}"),
		con("{", methodSpecs, "}"))).As("interfaceType")
	methodSpec        = or(con(methodName, signature), interfaceTypeName).As("methodSpec")
	methodSpecs       = semiList(methodSpec)
	methodName        = identifier
	interfaceTypeName = typeName

	mapType = con("map", "[", keyType, "]", elementType).As("mapType")
	keyType = type_

	channelType = con(
		or(
			"chan",
			con("chan", "<-"),
			con("<-", "chan"),
		), elementType).As("channelType")

	tokenTable = toTokenTable([]interface{}{
		//token.ILLEGAL:        ,
		token.EOF: parse.EOF,
		//token.COMMENT:        ,
		token.IDENT:          identifier,
		token.INT:            intLit,
		token.FLOAT:          floatLit,
		token.IMAG:           imaginaryLit,
		token.CHAR:           runeLit,
		token.STRING:         stringLit,
		token.ADD:            "+",
		token.SUB:            "-",
		token.MUL:            "*",
		token.QUO:            "/",
		token.REM:            "%",
		token.AND:            "&",
		token.OR:             "|",
		token.XOR:            "^",
		token.SHL:            "<<",
		token.SHR:            ">>",
		token.AND_NOT:        "&^",
		token.ADD_ASSIGN:     "+=",
		token.SUB_ASSIGN:     "-=",
		token.MUL_ASSIGN:     "*=",
		token.QUO_ASSIGN:     "/=",
		token.REM_ASSIGN:     "%=",
		token.AND_ASSIGN:     "&=",
		token.OR_ASSIGN:      "|=",
		token.XOR_ASSIGN:     "^=",
		token.SHL_ASSIGN:     "<<=",
		token.SHR_ASSIGN:     ">>=",
		token.AND_NOT_ASSIGN: "&^=",
		token.LAND:           "&&",
		token.LOR:            "||",
		token.ARROW:          "<-",
		token.INC:            "++",
		token.DEC:            "--",
		token.EQL:            "==",
		token.LSS:            "<",
		token.GTR:            ">",
		token.ASSIGN:         "=",
		token.NOT:            "!",
		token.NEQ:            "!=",
		token.LEQ:            "<=",
		token.GEQ:            ">=",
		token.DEFINE:         ":=",
		token.ELLIPSIS:       "...",
		token.LPAREN:         "(",
		token.LBRACK:         "[",
		token.LBRACE:         "{",
		token.COMMA:          ",",
		token.PERIOD:         ".",
		token.RPAREN:         ")",
		token.RBRACK:         "]",
		token.RBRACE:         "}",
		token.SEMICOLON:      ";",
		token.COLON:          ":",
		token.BREAK:          "break",
		token.CASE:           "case",
		token.CHAN:           "chan",
		token.CONST:          "const",
		token.CONTINUE:       "continue",
		token.DEFAULT:        "default",
		token.DEFER:          "defer",
		token.ELSE:           "else",
		token.FALLTHROUGH:    "fallthrough",
		token.FOR:            "for",
		token.FUNC:           "func",
		token.GO:             "go",
		token.GOTO:           "goto",
		token.IF:             "if",
		token.IMPORT:         "import",
		token.INTERFACE:      "interface",
		token.MAP:            "map",
		token.PACKAGE:        "package",
		token.RANGE:          "range",
		token.RETURN:         "return",
		token.SELECT:         "select",
		token.STRUCT:         "struct",
		token.SWITCH:         "switch",
		token.TYPE:           "type",
		token.VAR:            "var",
	})
)

func toTokenTable(a []interface{}) []*parse.R {
	rs := make([]*parse.R, len(a))
	for i := range a {
		switch o := a[i].(type) {
		case nil:
		case *parse.R:
			rs[i] = o
		case string:
			rs[i] = builder.Term(o)
		default:
			panic("element should be a string or a *parse.R")
		}
	}
	return rs
}

func init() {
	sourceExpr.InitTermSet()
	sourceFile.InitTermSet()
}
