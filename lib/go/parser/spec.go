package parser

import "github.com/hailiang/gombi/parse"

var (
	builder = parse.NewBuilder()
	term    = builder.Term
	rule    = builder.Rule
	recur   = builder.Recur
	or      = builder.Or
	con     = builder.Con
	null    = parse.Null
	newRule = parse.NewRule
	opt     = func(rules ...interface{}) *parse.R {
		return con(rules...).Optional()
	}
	repeat = func(rules ...interface{}) *parse.R {
		return con(rules...).Repeat()
	}

	declList = func(item *parse.R) *parse.R {
		return or(item, con("(", repeat(item, ";"), ")"))
	}

	// Packages

	sourceFile = con(packageClause, ";", repeat(importDecl, ";"), repeat(topLevelDecl, ";"))

	packageClause = con("package", packageName)
	packageName   = identifier

	importDecl = con("import", declList(importSpec))
	importSpec = con(or("dot", packageName).Optional(), importPath)
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

	block         = con("{", statementList, "}")
	statementList = repeat(statement, ";")

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

	expression = newRule()
	_          = expression.Define(con(or(unaryExpr, con(expression, binaryOp, unaryExpr))))
	unaryExpr  = newRule()
	_          = unaryExpr.Define(or(primaryExpr, con(unaryOp, unaryExpr)))

	identifier   = term("identifier")
	stringLit    = term("string")
	runeLit      = term("rune")
	intLit       = term("int")
	floatLit     = term("float")
	imaginaryLit = term("imag")

	operand     = or(literal, operandName, methodExpr, "(", expression, ")")
	literal     = or(basicLit, compositeLit, functionLit)
	basicLit    = or(intLit, floatLit, imaginaryLit, runeLit, stringLit)
	operandName = or(identifier, qualifiedIdent)

	qualifiedIdent = con(packageName, ".", identifier)

	compositeLit = con(literalType, literalValue)
	literalType  = or(structType, arrayType, con("[", "...", "]", elementType),
		sliceType, mapType, typeName)
	literalValue = newRule()
	_            = literalValue.Define(con("{", opt(elementList, opt(",")), "}"))
	elementList  = con(element, repeat(",", element))
	element      = con(opt(key, ":"), value)
	key          = or(fieldName, elementIndex)
	fieldName    = identifier
	elementIndex = expression
	value        = or(expression, literalValue)

	functionLit = con("func", function)

	primaryExpr = newRule()
	_           = primaryExpr.Define(or(operand, conversion, builtinCall, con(primaryExpr, selector), con(primaryExpr, index),
		con(primaryExpr, slice),
		con(primaryExpr, typeAssertion),
		con(primaryExpr, call)))
	selector = con(".", identifier)
	index    = con("[", expression, "]")
	slice    = con("[", or(con(opt(expression), ":", opt(expression)),
		con(opt(expression), ":", expression, ":", expression), "]"))
	typeAssertion = con(".", "(", type_, ")")
	call          = con("(", opt(argumentList, opt(",")), ")")
	argumentList  = con(expressionList, opt("..."))

	binaryOp = or("||", "&&", relOp, addOp, mulOp)
	relOp    = or("==", "!=", "<", "<=", ">", ">=")
	addOp    = or("+", "-", "|", "^")
	mulOp    = or("*", "/", "%", "<<", ">>", "&", "&^")
	unaryOp  = or("+", "-", "!", "^", "*", "&", "<-")

	methodExpr   = con(receiverType, ".", methodName)
	receiverType = newRule()
	_            = receiverType.Define(or(typeName, con("(", "*", typeName, ")"), con("(", receiverType, ")")))

	conversion = con(type_, "(", expression, opt(","), ")")

	// Built-in functions

	builtinCall = con(identifier, "(", opt(builtinArgs, opt(",")), ")")
	builtinArgs = or(con(type_, opt(",", argumentList)), argumentList)

	// Types

	type_    = newRule()
	_        = type_.Define(or(typeName, typeLit, con("(", type_, ")")))
	typeName = or(identifier, qualifiedIdent)
	typeLit  = or(arrayType, structType, pointerType, functionType, interfaceType,
		sliceType, mapType, channelType)

	arrayType   = con("[", arrayLength, "]", elementType)
	arrayLength = expression
	elementType = type_

	sliceType = con("[", "]", elementType)

	structType     = con("struct", "{", repeat(fieldDecl, ";"), "}")
	fieldDecl      = con(or(con(identifierList, type_), anonymousField), opt(tag))
	anonymousField = con(term("*").Optional(), typeName)
	tag            = stringLit

	pointerType = con("*", baseType)
	baseType    = type_

	functionType  = con("func", signature)
	signature     = con(parameters, opt(result))
	result        = or(parameters, type_)
	parameters    = con("(", opt(parameterList, opt(",")), ")")
	parameterList = con(parameterDecl, repeat(",", parameterDecl))
	parameterDecl = con(opt(identifierList), opt("..."), type_)

	interfaceType     = con("interface", "{", repeat(methodSpec, ";"), "}")
	methodSpec        = or(con(methodName, signature), interfaceTypeName)
	methodName        = identifier
	interfaceTypeName = typeName

	mapType = con("map", "[", keyType, "]", elementType)
	keyType = type_

	channelType = con(or(con("chan", opt("<-")), con("<-", "chan")), elementType)
)
