package parser

import (
	"go/token"

	"github.com/hailiang/gombi/parse"
)

var (
	builder = parse.NewBuilder()
	term    = builder.Term
	rule    = builder.Rule
	recur   = builder.Recur
	or      = builder.Or
	con     = builder.Con
	null    = parse.Null
	EOF     = parse.EOF
	newRule = parse.NewRule
	opt     = func(rules ...interface{}) *parse.R {
		return con(rules...).Optional()
	}
	repeat = func(rules ...interface{}) *parse.R {
		return con(rules...).Repeat()
	}

	declList = func(item *parse.R) *parse.R {
		return or(item, con("(", repeat(item, ";"), opt(item), ")"))
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

	sourceFile = con(packageClause, ";", repeat(importDecl, ";"), repeat(topLevelDecl, ";")).As("sourceFile")

	packageClause = con("package", packageName).As("packageClause")
	packageName   = identifier // As??

	importDecl = con("import", declList(importSpec)).As("importDecl")
	importSpec = con(or("dot", packageName).Optional(), importPath).As("importSpec")
	importPath = stringLit

	// Declarations

	decl         = or(constDecl, typeDecl, varDecl).As("decl")
	topLevelDecl = or(decl, funcDecl, methodDecl).As("topLevelDecl")

	constDecl      = con("const", declList(constSpec)).As("constDecl")
	constSpec      = con(identifierList, opt(opt("type"), "=", exprList)).As("constSpec")
	identifierList = con(identifier, repeat(",", identifier)).As("identifierList")
	exprList       = con(expr, repeat(",", expr)).As("exprList")

	typeDecl = con("type", declList(typeSpec)).As("typeDecl")
	typeSpec = con(identifier, type_).As("typeSpec")

	varDecl = con("var", declList(varSpec)).As("varDecl")
	varSpec = con(identifierList, or(type_, con(opt(type_), "=", exprList))).As("varSpec")

	shortVarDecl = con(identifierList, ":=", exprList).As("shortVarDecl")

	funcDecl = con("func", funcName, or(func_, signature)).As("funcDecl")
	funcName = identifier
	func_    = con(signature, funcBody).As("func_")
	funcBody = block

	methodDecl   = con("func", receiver, methodName, or(func_, signature)).As("methodDecl")
	receiver     = con("(", opt(identifier), opt("*"), baseTypeName, ")").As("receiver")
	baseTypeName = identifier

	// block

	block    = con("{", stmtList, opt(stmt), "}").As("block")
	stmtList = repeat(stmt, ";").As("stmtList")

	// Statements
	stmt = newRule().As("stmt")
	_    = stmt.Define(or(
		decl, labeledStmt, simpleStmt,
		goStmt, returnStmt, breakStmt, continueStmt, gotoStmt,
		fallthroughStmt, block, ifStmt, switchStmt, selectStmt, forStmt,
		deferStmt))
	simpleStmt = or(emptyStmt, exprStmt, sendStmt, incDecStmt, assignment, shortVarDecl).As("simpleStmt")

	emptyStmt = null

	labeledStmt = con(label, ":", stmt).As("labeledStmt")
	label       = identifier

	exprStmt = expr

	sendStmt = con(channel, "<-", expr).As("sendStmt")
	channel  = expr

	incDecStmt = con(expr, or("++", "--")).As("incDecStmt")

	assignment = con(exprList, assignOp, exprList).As("assignment")
	assignOp   = or("=", "+=", "-=", "|=", "^=", "*=", "/=", "%=", "<<=", ">>=", "&=", "&^=").As("assignOp")

	ifStmt = newRule().As("ifStmt")
	_      = ifStmt.Define(con("if", opt(simpleStmt, ";"), expr, block, opt("else", or(ifStmt, block))))

	switchStmt     = or(exprSwitchStmt, typeSwitchStmt).As("switchStmt")
	exprSwitchStmt = con("switch", opt(simpleStmt, ";"), opt(expr), "{", exprCaseClause.Repeat(), "}").As("exprSwitchStmt")
	exprCaseClause = con(exprSwitchCase, ":", stmtList).As("exprCaseClause")
	exprSwitchCase = con("case", or(exprList, "default")).As("exprSwitchCase")

	typeSwitchStmt  = con("switch", opt(simpleStmt, ";"), typeSwitchGuard, "{", typeCaseClause.Repeat(), "}").As("typeSwitchStmt")
	typeSwitchGuard = con(opt(identifier, ":="), primaryExpr, ".", "(", "type", ")").As("typeSwitchGuard")
	typeCaseClause  = con(typeSwitchCase, ":", stmtList).As("typeCaseClause")
	typeSwitchCase  = or(con("case", typeList), "default").As("typeSwitchCase")
	typeList        = con(type_, repeat(",", type_)).As("typeList")

	forStmt     = con("for", opt(or(condition, forClause, rangeClause)), block).As("forStmt")
	condition   = expr
	forClause   = con(opt(initStmt), ";", opt(condition), ";", opt(postStmt)).As("forClause")
	initStmt    = simpleStmt
	postStmt    = simpleStmt
	rangeClause = con(or(con(exprList, "="), con(identifierList, ":=")), "range", expr).As("rangeClause")

	goStmt = con("go", expr).As("goStmt")

	selectStmt = con("select", "{", repeat(commClause), "}").As("selectStmt")
	commClause = con(commCase, ":", stmtList).As("commClause")
	commCase   = or(con("case", or(sendStmt, recvStmt)), "default").As("commCase")
	recvStmt   = con(opt(or(con(exprList, "="), con(identifierList, ":="))), recvExpr).As("recvStmt")
	recvExpr   = expr

	returnStmt = con("return", opt(exprList)).As("returnStmt")

	breakStmt = con("break", opt(label)).As("breakStmt")

	continueStmt = con("continue", opt(label)).As("continueStmt")

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
	_            = literalValue.Define(con("{", elementList, opt(element), "}"))
	elementList  = repeat(element, ",").As("elementList")
	element      = con(opt(key, ":"), value).As("element")
	key          = expr
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
	slice    = con("[", or(con(opt(expr), ":", opt(expr)),
		con(opt(expr), ":", expr, ":", expr), "]")).As("slice")
	typeAssertion = con(".", "(", type_, ")").As("typeAssertion")
	call          = con("(", opt(argumentList, opt(",")), ")").As("call")
	argumentList  = con(exprList, opt("...")).As("argumentList")

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

	builtinCall = con(identifier, "(", opt(builtinArgs, opt(",")), ")").As("builtinCall")
	builtinArgs = or(con(type_, opt(",", argumentList)), argumentList).As("builtinArgs")

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

	structType     = con("struct", "{", repeat(fieldDecl, ";"), opt(fieldDecl), "}").As("structType")
	fieldDecl      = con(or(con(identifierList, type_), anonymousField), opt(tag)).As("fieldDecl")
	anonymousField = con(term("*").Optional(), typeName).As("anonymousField")
	tag            = stringLit

	pointerType = con("*", baseType).As("pointerType")
	baseType    = type_

	funcType      = con("func", signature).As("funcType")
	signature     = con(parameters, opt(result)).As("signature")
	result        = or(parameters, type_).As("result")
	parameters    = con("(", parameterList, opt(parameterDecl), ")").As("parameters")
	parameterList = repeat(parameterDecl, ",").As("parameterList")
	parameterDecl = con(opt(identifierList), opt("..."), type_).As("parameterDecl")

	interfaceType     = con("interface", "{", repeat(methodSpec, ";"), opt(methodSpec), "}").As("interfaceType")
	methodSpec        = or(con(methodName, signature), interfaceTypeName).As("methodSpec")
	methodName        = identifier
	interfaceTypeName = typeName

	mapType = con("map", "[", keyType, "]", elementType).As("mapType")
	keyType = type_

	channelType = con(or(con("chan", opt("<-")), con("<-", "chan")), elementType).As("channelType")

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
