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

	// goExpr
	goExpr = con(or(expression, type_), ";", EOF).As("goExpr")

	// Packages

	sourceFile = con(packageClause, ";", repeat(importDecl, ";"), repeat(topLevelDecl, ";")).As("sourceFile")

	packageClause = con("package", packageName).As("packageClause")
	packageName   = identifier // As??

	importDecl = con("import", declList(importSpec)).As("importDecl")
	importSpec = con(or("dot", packageName).Optional(), importPath).As("importSpec")
	importPath = stringLit

	// Declarations

	declaration  = or(constDecl, typeDecl, varDecl)
	topLevelDecl = or(declaration, functionDecl, methodDecl)

	constDecl      = con("const", declList(constSpec))
	constSpec      = con(identifierList, opt(opt("type"), "=", expressionList))
	identifierList = con(identifier, repeat(";", identifier))
	expressionList = con(expression, repeat(";", expression))

	typeDecl = con("type", declList(typeSpec))
	typeSpec = con(identifier, type_)

	varDecl = con("var", declList(varSpec))
	varSpec = con(identifierList, or(type_, con(opt(type_), "=", expressionList)))

	shortVarDecl = con(identifierList, ":=", expressionList)

	functionDecl = con("func", functionName, or(function, signature))
	functionName = identifier
	function     = con(signature, functionBody)
	functionBody = block

	methodDecl   = con("func", receiver, methodName, or(function, signature))
	receiver     = con("(", opt(identifier), opt("*"), baseTypeName, ")")
	baseTypeName = identifier

	// block

	block         = con("{", statementList, opt(statement), "}").As("block")
	statementList = repeat(statement, ";").As("statementList")

	// Statements
	statement = newRule()
	_         = statement.Define(or(
		declaration, labeledStmt, simpleStmt,
		goStmt, returnStmt, breakStmt, continueStmt, gotoStmt,
		fallthroughStmt, block, ifStmt, switchStmt, selectStmt, forStmt,
		deferStmt))
	simpleStmt = or(emptyStmt, expressionStmt, sendStmt, incDecStmt, assignment, shortVarDecl)

	emptyStmt = null

	labeledStmt = con(label, ":", statement)
	label       = identifier

	expressionStmt = expression

	sendStmt = con(channel, "<-", expression)
	channel  = expression

	incDecStmt = con(expression, or("++", "--"))

	assignment = con(expressionList, assignOp, expressionList)
	assignOp   = con(or(addOp, mulOp).Optional(), "=")

	ifStmt = newRule()
	_      = ifStmt.Define(con("if", opt(simpleStmt, ";"), expression, block, con("else", or(ifStmt, block)).Optional()))

	switchStmt     = or(exprSwitchStmt, typeSwitchStmt)
	exprSwitchStmt = con("switch", opt(simpleStmt, ";"), opt(expression), "{", exprCaseClause.Repeat(), "}")
	exprCaseClause = con(exprSwitchCase, ":", statementList)
	exprSwitchCase = con("case", or(expressionList, "default"))

	typeSwitchStmt  = con("switch", opt(simpleStmt, ";"), typeSwitchGuard, "{", typeCaseClause.Repeat(), "}")
	typeSwitchGuard = con(opt(identifier, ":="), primaryExpr, ".", "(", "type", ")")
	typeCaseClause  = con(typeSwitchCase, ":", statementList)
	typeSwitchCase  = or(con("case", typeList), "default")
	typeList        = con(type_, repeat(",", type_))

	forStmt     = con("for", opt(or(condition, forClause, rangeClause)), block)
	condition   = expression
	forClause   = con(opt(initStmt), ";", opt(condition), ";", opt(postStmt))
	initStmt    = simpleStmt
	postStmt    = simpleStmt
	rangeClause = con(or(con(expressionList, "="), con(identifierList, ":=")), "range", expression)

	goStmt = con("go", expression)

	selectStmt = con("select", "{", repeat(commClause), "}")
	commClause = con(commCase, ":", statementList)
	commCase   = or(con("case", or(sendStmt, recvStmt)), "default")
	recvStmt   = con(opt(or(con(expressionList, "="), con(identifierList, ":="))), recvExpr)
	recvExpr   = expression

	returnStmt = con("return", opt(expressionList))

	breakStmt = con("break", opt(label))

	continueStmt = con("continue", opt(label))

	gotoStmt = con("goto", label)

	fallthroughStmt = term("fallthrough")

	deferStmt = con("defer", expression)

	// Expression

	expression = newRule().As("expression")
	_          = expression.Define(or(unaryExpr, con(expression, binaryOp, unaryExpr)))
	unaryExpr  = newRule().As("unaryExpr")
	_          = unaryExpr.Define(or(primaryExpr, con(unaryOp, unaryExpr)))

	operand     = or(literal, operandName, methodExpr, con("(", expression, ")")).As("operand")
	literal     = or(basicLit, compositeLit, functionLit).As("literal")
	basicLit    = or(intLit, floatLit, imaginaryLit, runeLit, stringLit).As("basicLit")
	operandName = or(identifier, qualifiedIdent).As("operandName")

	qualifiedIdent = con(packageName, ".", identifier).As("qualifiedIdent")

	compositeLit = con(literalType, literalValue).As("compositeLit")
	literalType  = or(structType, arrayType, con("[", "...", "]", elementType),
		sliceType, mapType, typeName).As("literalType")
	literalValue = newRule().As("literalValue")
	_            = literalValue.Define(con("{", opt(elementList, opt(",")), "}"))
	elementList  = con(element, repeat(",", element)).As("elementList")
	element      = con(opt(key, ":"), value).As("element")
	key          = or(fieldName, elementIndex).As("key")
	fieldName    = identifier
	elementIndex = expression
	value        = or(expression, literalValue).As("value")

	functionLit = con("func", function).As("functionLit")

	primaryExpr = newRule().As("primaryExpr")
	_           = primaryExpr.Define(or(operand, conversion, builtinCall, con(primaryExpr, selector), con(primaryExpr, index),
		con(primaryExpr, slice),
		con(primaryExpr, typeAssertion),
		con(primaryExpr, call)))
	selector = con(".", identifier).As("selector")
	index    = con("[", expression, "]").As("index")
	slice    = con("[", or(con(opt(expression), ":", opt(expression)),
		con(opt(expression), ":", expression, ":", expression), "]")).As("slice")
	typeAssertion = con(".", "(", type_, ")").As("typeAssertion")
	call          = con("(", opt(argumentList, opt(",")), ")").As("call")
	argumentList  = con(expressionList, opt("...")).As("argumentList")

	binaryOp = or("||", "&&", relOp, addOp, mulOp).As("binaryOp")
	relOp    = or("==", "!=", "<", "<=", ">", ">=")
	addOp    = or("+", "-", "|", "^")
	mulOp    = or("*", "/", "%", "<<", ">>", "&", "&^")
	unaryOp  = or("+", "-", "!", "^", "*", "&", "<-")

	methodExpr   = con(receiverType, ".", methodName).As("methodExpr")
	receiverType = newRule().As("receiverType")
	_            = receiverType.Define(or(typeName, con("(", "*", typeName, ")"), con("(", receiverType, ")")))

	conversion = con(type_, "(", expression, opt(","), ")").As("conversion")

	// Built-in functions

	builtinCall = con(identifier, "(", opt(builtinArgs, opt(",")), ")")
	builtinArgs = or(con(type_, opt(",", argumentList)), argumentList)

	// Types

	type_    = newRule().As("type")
	_        = type_.Define(or(typeName, typeLit, con("(", type_, ")")))
	typeName = or(identifier, qualifiedIdent).As("typeName")
	typeLit  = or(arrayType, structType, pointerType, functionType, interfaceType,
		sliceType, mapType, channelType).As("typeLit")

	arrayType   = con("[", arrayLength, "]", elementType).As("arrayType")
	arrayLength = expression
	elementType = type_

	sliceType = con("[", "]", elementType).As("sliceType")

	structType     = con("struct", "{", repeat(fieldDecl, ";"), opt(fieldDecl), "}").As("structType")
	fieldDecl      = con(or(con(identifierList, type_), anonymousField), opt(tag)).As("fieldDecl")
	anonymousField = con(term("*").Optional(), typeName).As("anonymousField")
	tag            = stringLit

	pointerType = con("*", baseType).As("pointerType")
	baseType    = type_

	functionType  = con("func", signature).As("functionType")
	signature     = con(parameters, opt(result)).As("siginature")
	result        = or(parameters, type_).As("result")
	parameters    = con("(", opt(parameterList, opt(",")), ")").As("parameters")
	parameterList = con(parameterDecl, repeat(",", parameterDecl)).As("parameterList")
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
