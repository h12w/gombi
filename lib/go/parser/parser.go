package parser

import (
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"

	"github.com/hailiang/gombi/parse"
)

type parser struct {
	rule    *parse.R
	file    *token.File
	errors  scanner.ErrorList
	scanner scanner.Scanner

	// Tracing/debugging
	mode  Mode // parsing mode
	trace bool // == (mode & Trace != 0)
	//indent int  // indentation used for tracing output

	// Comments
	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup // last lead comment
	lineComment *ast.CommentGroup // last line comment

	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	//// Error recovery
	//// (used to limit the number of calls to syncXXX functions
	//// w/o making scanning progress - avoids potential endless
	//// loops across multiple parser functions during error recovery)
	//syncPos token.Pos // last synchronization position
	//syncCnt int       // number of calls to syncXXX without progress

	//// Non-syntactic parser control
	//exprLev int  // < 0: in control clause, >= 0: in expression
	//inRhs   bool // if set, the parser is parsing a rhs expression

	// Ordinary identifier scopes
	pkgScope   *ast.Scope   // pkgScope.Outer == nil
	topScope   *ast.Scope   // top-most scope; may be pkgScope
	unresolved []*ast.Ident // unresolved identifiers
	//imports    []*ast.ImportSpec // list of imports

	// Label scopes
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels
}

func (p *parser) error(pos token.Pos, msg string) {
	epos := p.file.Position(pos)

	// If AllErrors is not set, discard errors reported on the same line
	// as the last recorded error and stop parsing if there are more than
	// 10 errors.
	if p.mode&AllErrors == 0 {
		n := len(p.errors)
		if n > 0 && p.errors[n-1].Pos.Line == epos.Line {
			return // discard - likely a spurious error
		}
		if n > 10 {
			panic(bailout{})
		}
	}

	p.errors.Add(epos, msg)
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		if p.tok == token.SEMICOLON && p.lit == "\n" {
			msg += ", found newline"
		} else {
			msg += ", found '" + p.tok.String() + "'"
			if p.tok.IsLiteral() {
				msg += " " + p.lit
			}
		}
	}
	p.error(pos, msg)
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *ast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &ast.Comment{Slash: p.pos, Text: p.lit}

	p.pos, p.tok, p.lit = p.scanner.Scan()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
//
func (p *parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

// ----------------------------------------------------------------------------
// Scoping support

func (p *parser) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

func (p *parser) openLabelScope() {
	p.labelScope = ast.NewScope(p.labelScope)
	p.targetStack = append(p.targetStack, nil)
}

func (p *parser) closeLabelScope() {
	// resolve labels
	n := len(p.targetStack) - 1
	scope := p.labelScope
	for _, ident := range p.targetStack[n] {
		ident.Obj = scope.Lookup(ident.Name)
		if ident.Obj == nil && p.mode&DeclarationErrors != 0 {
			p.error(ident.Pos(), fmt.Sprintf("label %s undefined", ident.Name))
		}
	}
	// pop label scope
	p.targetStack = p.targetStack[0:n]
	p.labelScope = p.labelScope.Outer
}

func (p *parser) declare(decl, data interface{}, scope *ast.Scope, kind ast.ObjKind, idents ...*ast.Ident) {
	for _, ident := range idents {
		assert(ident.Obj == nil, "identifier already declared or resolved")
		obj := ast.NewObj(kind, ident.Name)
		// remember the corresponding declaration for redeclaration
		// errors and global variable resolution/typechecking phase
		obj.Decl = decl
		obj.Data = data
		ident.Obj = obj
		if ident.Name != "_" {
			if alt := scope.Insert(obj); alt != nil && p.mode&DeclarationErrors != 0 {
				prevDecl := ""
				if pos := alt.Pos(); pos.IsValid() {
					prevDecl = fmt.Sprintf("\n\tprevious declaration at %s", p.file.Position(pos))
				}
				p.error(ident.Pos(), fmt.Sprintf("%s redeclared in this block%s", ident.Name, prevDecl))
			}
		}
	}
}

func (p *parser) declareShortVar(decl *ast.AssignStmt, list []ast.Expr) {
	// Go spec: A short variable declaration may redeclare variables
	// provided they were originally declared in the same block with
	// the same type, and at least one of the non-blank variables is new.
	n := 0 // number of new variables
	for _, x := range list {
		if ident, isIdent := x.(*ast.Ident); isIdent {
			assert(ident.Obj == nil, "identifier already declared or resolved")
			obj := ast.NewObj(ast.Var, ident.Name)
			// remember corresponding assignment for other tools
			obj.Decl = decl
			ident.Obj = obj
			if ident.Name != "_" {
				if alt := p.topScope.Insert(obj); alt != nil {
					ident.Obj = alt // redeclaration
				} else {
					n++ // new declaration
				}
			}
		} else {
			p.errorExpected(x.Pos(), "identifier on left side of :=")
		}
	}
	if n == 0 && p.mode&DeclarationErrors != 0 {
		p.error(list[0].Pos(), "no new variables on left side of :=")
	}
}

//// The unresolved object is a sentinel to mark identifiers that have been added
//// to the list of unresolved identifiers. The sentinel is only used for verifying
//// internal consistency.
//var unresolved = new(ast.Object)

// If x is an identifier, tryResolve attempts to resolve x by looking up
// the object it denotes. If no object is found and collectUnresolved is
// set, x is marked as unresolved and collected in the list of unresolved
// identifiers.
//
func (p *parser) tryResolve(x ast.Expr, collectUnresolved bool) {
	// nothing to do if x is not an identifier or the blank identifier
	ident, _ := x.(*ast.Ident)
	if ident == nil {
		return
	}
	assert(ident.Obj == nil, "identifier already declared or resolved")
	if ident.Name == "_" {
		return
	}
	// try to resolve the identifier
	for s := p.topScope; s != nil; s = s.Outer {
		if obj := s.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
			return
		}
	}
	// all local scopes are known, so any unresolved identifier
	// must be found either in the file scope, package scope
	// (perhaps in another file), or universe scope --- collect
	// them so that they can be resolved later
	if collectUnresolved {
		ident.Obj = unresolved
		p.unresolved = append(p.unresolved, ident)
	}
}

func (p *parser) resolve(x ast.Expr) {
	p.tryResolve(x, true)
}

func (p *parser) init(fset *token.FileSet, filename string, src []byte, mode Mode) {
	p.file = fset.AddFile(filename, -1, len(src))
	var m scanner.Mode
	if mode&ParseComments != 0 {
		m = scanner.ScanComments
	}
	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner.Init(p.file, src, eh, m)

	p.mode = mode
	p.trace = mode&Trace != 0 // for convenience (p.trace is used frequently)

	//p.next()
}

func (p *parser) parseExprOrType() ast.Expr {
	pp := parse.New(sourceExpr)
	for {
		pos, tok, str := p.scanner.Scan()
		if r := tokenTable[tok]; r != nil {
			if !pp.Parse(&parse.Token{ID: int(tok), Value: []byte(str), Pos: int(pos)}, r) {
				break
			}
		}
		if tok == token.EOF {
			break
		}
	}
	if len(pp.Results()) != 1 {
		p.errors.Add(p.file.Position(0), "gombi parse error")
		return nil
	}
	return p.parseGoExpr(pp.Results()[0])
}

func (p *parser) parseFile() *ast.File {
	pp := parse.New(sourceFile)
	skipScan := false
	for {
		p.leadComment = nil
		p.lineComment = nil
		prev := p.pos
		if skipScan {
			skipScan = false
		} else {
			p.pos, p.tok, p.lit = p.scanner.Scan()
		}
		if r := tokenTable[p.tok]; r != nil {
			if !pp.Parse(&parse.Token{ID: int(p.tok), Value: []byte(p.lit), Pos: int(p.pos)}, r) {
				break
			}
		}
		switch p.tok {
		case token.COMMENT:
			var comment *ast.CommentGroup
			var endline int

			if p.file.Line(p.pos) == p.file.Line(prev) {
				// The comment is on same line as the previous token; it
				// cannot be a lead comment but may be a line comment.
				comment, endline = p.consumeCommentGroup(0)
				if p.file.Line(p.pos) != endline {
					// The next token is on a different line, thus
					// the last comment group is a line comment.
					p.lineComment = comment
				}
			}

			// consume successor comments, if any
			endline = -1
			for p.tok == token.COMMENT {
				comment, endline = p.consumeCommentGroup(1)
			}

			if endline+1 == p.file.Line(p.pos) {
				// The next token is following on the line immediately after the
				// comment group, thus the last comment group is a lead comment.
				p.leadComment = comment
			}
			skipScan = true

		case token.EOF:
			break
		}
	}
	if len(pp.Results()) != 1 {
		if len(pp.Results()) > 1 {
			fmt.Println("Ambiguous", len(pp.Results()))
		}
		p.errors.Add(p.file.Position(0), "")
		return nil
	}
	return p.parseSourceFile(pp.Results()[0])
}

func (p *parser) parseSourceFile(n *parse.Node) *ast.File {
	p.openScope()
	pac, name := p.parsePackageClause(n.Child(0))
	importDecls, importSpecs := p.parseImports(n.Child(2))
	decls := append(importDecls, p.parseTopLevelDecls(n.Child(3))...)
	p.closeScope()
	return &ast.File{
		Package:  pac,
		Name:     name,
		Decls:    decls,
		Scope:    p.topScope,
		Imports:  importSpecs,
		Comments: p.comments,
	}
}

func (p *parser) parseTopLevelDecls(n *parse.Node) (decls []ast.Decl) {
	n.Each(func(item *parse.Node) {
		decls = append(decls, p.parseTopLevelDecl(item.Child(0)))
	})
	return
}

func (p *parser) parseTopLevelDecl(n *parse.Node) ast.Decl {
	n = n.Child(0)
	switch n.Rule() {
	case decl:
		return p.parseDecl(n)
	case funcDecl:
		return p.parseFuncDecl(n)
	case methodDecl:
		return p.parseMethodDecl(n)
	}
	return nil
}

func (p *parser) parseDecl(n *parse.Node) ast.Decl {
	n = n.Child(0)
	switch n.Rule() {
	case constDecl:
		return p.parseConstDecl(n)
	case typeDecl:
		return p.parseTypeDecl(n)
	case varDecl:
		return p.parseVarDecl(n)
	}
	return nil
}

func (p *parser) parseConstDecl(n *parse.Node) ast.Decl {
	fmt.Println(n)
	return nil
}

func (p *parser) parseTypeDecl(n *parse.Node) ast.Decl {
	fmt.Println(n)
	return nil
}

func (p *parser) parseVarDecl(n *parse.Node) ast.Decl {
	var specs []ast.Spec
	lParen, rParen := declListEach(n.Child(1), func(item *parse.Node) {
		specs = append(specs, p.parseVarSpec(item))
	})
	return &ast.GenDecl{
		TokPos: token.Pos(n.Child(0).Pos()),
		Tok:    token.Token(n.Child(0).ID()),
		Lparen: token.Pos(lParen),
		Specs:  specs,
		Rparen: token.Pos(rParen),
	}
}

func declListEach(n *parse.Node, visit func(*parse.Node)) (lParen, rParen int) {
	if n.Child(0).Child(0).Is(term("(")) {
		n.Child(0).Child(1).Each(func(item *parse.Node) {
			visit(item.Child(0))
		})
		if n.Child(0).Child(2) != nil {
			visit(n.Child(0).Child(2).Child(0))
		}
		return n.Child(0).Child(0).Pos(), n.Child(0).Child(3).Pos()
	}
	visit(n.Child(0))
	return 0, 0
}

func (p *parser) parseVarSpec(n *parse.Node) *ast.ValueSpec {
	spec := ast.ValueSpec{
		Names: p.parseIdentList(n.Child(0)),
	}
	def := n.Child(1).Child(0)
	if def.Is(type_) {
		spec.Type = p.parseType(def)
	} else {
		if def.Child(0) != nil {
			spec.Type = p.parseType(def.Child(0).Child(0))
		}
		spec.Values = p.parseExprList(def.Child(2))
	}
	return &spec
}

func (p *parser) parseMethodDecl(n *parse.Node) ast.Decl {
	fmt.Println(n)
	return nil
}

func (p *parser) parseLabeledStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	return nil
}

func (p *parser) parseGoStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	return nil
}

func (p *parser) parseReturnStmt(n *parse.Node) ast.Stmt {
	return &ast.ReturnStmt{
		Return:  token.Pos(n.Child(0).Pos()),
		Results: p.parseExprList(n.Child(1).Child(0)),
	}
}

func (p *parser) parseBreakStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	return nil
}

func (p *parser) parseContinueStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	return nil
}

func (p *parser) parseGotoStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	return nil
}

func (p *parser) parseFallthroughStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	return nil
}

func (p *parser) parseSwitchStmt(n *parse.Node) ast.Stmt {
	p.openScope()
	n = n.Child(0)
	switch n.Rule() {
	case exprSwitchStmt:
		return p.parseExprSwitchStmt(n)
	case typeSwitchStmt:
		return p.parseTypeSwitchStmt(n)
	}
	p.closeScope()
	return nil
}

func (p *parser) parseExprSwitchStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	// for each case clause/ open/close scope
	return nil
}

func (p *parser) parseTypeSwitchStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	// type switch may introduce a variable that needs extra scope
	// for each case clause/ open/close scope
	return nil
}

func (p *parser) parseSelectStmt(n *parse.Node) ast.Stmt {
	fmt.Println(n)
	return nil
}

func (p *parser) parseFuncDecl(n *parse.Node) ast.Decl {
	scope := ast.NewScope(p.topScope)
	funcDecl := ast.FuncDecl{
		Name: p.parseIdent(n.Child(1)),
	}
	funcOrSig := n.Child(2).Child(0)
	if funcOrSig.Rule() == signature {
		funcDecl.Type = p.parseSignature(funcOrSig, scope)
	} else {
		funcDecl.Type, funcDecl.Body = p.parseFunc(funcOrSig, scope)
	}
	funcDecl.Type.Func = token.Pos(n.Child(0).Pos())
	return &funcDecl
}

func (p *parser) parseFunc(n *parse.Node, scope *ast.Scope) (*ast.FuncType, *ast.BlockStmt) {
	return p.parseSignature(n.Child(0), scope),
		p.parseBody(n.Child(1), scope)
}

func (p *parser) parseBody(n *parse.Node, scope *ast.Scope) *ast.BlockStmt {
	p.topScope = scope
	p.openLabelScope()
	block := p.parseBlock(n, scope)
	p.closeLabelScope()
	p.closeScope()
	return block
}

func (p *parser) parseSignature(n *parse.Node, scope *ast.Scope) *ast.FuncType {
	return &ast.FuncType{
		Params:  p.parseParams(n.Child(0), scope),
		Results: p.parseResults(n.Child(1), scope),
	}
}

func (p *parser) parseParams(n *parse.Node, scope *ast.Scope) *ast.FieldList {
	fieldList := ast.FieldList{
		Opening: token.Pos(n.Child(0).Pos()),
		Closing: token.Pos(n.Child(3).Pos()),
	}
	n.Child(1).Each(func(item *parse.Node) {
		fieldList.List = append(fieldList.List, p.parseParamDecl(item.Child(0), scope))
	})
	if n.Child(2).Child(0) != nil {
		fieldList.List = append(fieldList.List, p.parseParamDecl(n.Child(2).Child(0), scope))
	}
	return &fieldList
}

func (p *parser) parseParamDecl(n *parse.Node, scope *ast.Scope) *ast.Field {
	idents := p.parseIdentList(n.Child(0).Child(0))
	typ := p.parseType(n.Child(2))
	field := ast.Field{
		Names: idents,
		Type:  typ,
	}
	p.declare(field, nil, scope, ast.Var, idents...)
	p.resolve(typ)
	return &field
}

func (p *parser) parseIdentList(n *parse.Node) (idents []*ast.Ident) {
	idents = append(idents, p.parseIdent(n.Child(0)))
	n.Child(1).Each(func(item *parse.Node) {
		idents = append(idents, p.parseIdent(item.Child(1)))
	})
	return
}

func (p *parser) parseResults(n *parse.Node, scope *ast.Scope) *ast.FieldList {
	n = n.Child(0).Child(0)
	switch n.Rule() {
	case parameters:
		return p.parseParams(n, scope)
	case type_:
		return &ast.FieldList{List: []*ast.Field{{Type: p.parseType(n)}}}
	}
	return nil
}

func (p *parser) parseBlockStmt(n *parse.Node) *ast.BlockStmt {
	p.openScope()
	block := p.parseBlock(n, p.topScope)
	p.closeScope()
	return block
}

func (p *parser) parseBlock(n *parse.Node, scope *ast.Scope) *ast.BlockStmt {
	block := ast.BlockStmt{
		Lbrace: token.Pos(n.Child(0).Pos()),
		Rbrace: token.Pos(n.Child(3).Pos()),
	}
	n.Child(1).Each(func(item *parse.Node) {
		block.List = append(block.List, p.parseStmt(item.Child(0)))
	})
	if item := n.Child(2).Child(0); item != nil {
		block.List = append(block.List, p.parseStmt(item))
	}
	return &block
}

func (p *parser) parseStmt(n *parse.Node) ast.Stmt {
	n = n.Child(0)
	switch n.Rule() {
	case decl:
		return p.parseDeclStmt(n)
	case labeledStmt:
		return p.parseLabeledStmt(n)
	case simpleStmt:
		return p.parseSimpleStmt(n)
	case goStmt:
		return p.parseGoStmt(n)
	case returnStmt:
		return p.parseReturnStmt(n)
	case breakStmt:
		return p.parseBreakStmt(n)
	case continueStmt:
		return p.parseContinueStmt(n)
	case gotoStmt:
		return p.parseGotoStmt(n)
	case fallthroughStmt:
		return p.parseFallthroughStmt(n)
	case block:
		return p.parseBlockStmt(n)
	case ifStmt:
		return p.parseIfStmt(n)
	case switchStmt:
		return p.parseSwitchStmt(n)
	case selectStmt:
		return p.parseSelectStmt(n)
	case forStmt:
		return p.parseForStmt(n)
	case deferStmt:
		return p.parseDeferStmt(n)
	}
	return nil
}

func (p *parser) parseDeclStmt(n *parse.Node) ast.Stmt {
	return &ast.DeclStmt{p.parseDecl(n)}
}

func (p *parser) parseIfStmt(n *parse.Node) ast.Stmt {
	p.openScope()
	ifStmt := &ast.IfStmt{
		If:   token.Pos(n.Child(0).Pos()),
		Init: p.parseInit(n.Child(1)),
		Cond: p.parseCond(n.Child(2)),
		Body: p.parseBlock(n.Child(3), p.topScope),
		Else: p.parseElse(n.Child(4)),
	}
	p.closeScope()
	return ifStmt
}

func (p *parser) parseElse(n *parse.Node) ast.Stmt {
	if n == nil {
		return nil
	}
	n = n.Child(1).Child(0)
	switch n.Rule() {
	case ifStmt:
		return p.parseIfStmt(n)
	case block:
		return p.parseBlock(n, p.topScope)
	}
	return nil
}

func (p *parser) parseForStmt(n *parse.Node) (r ast.Stmt) {
	p.openScope()
	forPos := token.Pos(n.Child(0).Pos())
	option := n.Child(1).Child(0).Child(0)
	body := p.parseBlock(n.Child(2), p.topScope)
	switch option.Rule() {
	case condition:
		fmt.Println(option)
	case forClause:
		fmt.Println(option)
		r = &ast.ForStmt{
			For:  token.Pos(n.Child(0).Pos()),
			Init: p.parseInit(n.Child(1)),
			Cond: p.parseCond(n.Child(1)),
			Post: p.parsePost(n.Child(1)),
		}
	case rangeClause:
		forStmt := p.parseRangeStmt(option)
		forStmt.For, forStmt.Body = forPos, body
		r = forStmt
	}
	p.closeScope()
	return
}
func (p *parser) parseInit(n *parse.Node) ast.Stmt {
	if n == nil {
		return nil
	}
	n = n.Child(0)
	fmt.Println(n)
	return nil
}
func (p *parser) parseCond(n *parse.Node) ast.Expr {
	return p.parseExpr(n)
}

func (p *parser) parsePost(n *parse.Node) ast.Stmt {
	if n == nil {
		return nil
	}
	n = n.Child(0)
	if !n.Is(forClause) {
		return nil
	}
	fmt.Println(n)
	return nil
}

func (p *parser) parseDeferStmt(n *parse.Node) (r ast.Stmt) {
	return &ast.DeferStmt{
		Defer: token.Pos(n.Child(0).Pos()),
		Call:  p.parseCallExpr(n.Child(1)),
	}
}

func (p *parser) parseRangeStmt(n *parse.Node) *ast.RangeStmt {
	kv := n.Child(0).Child(0)
	tok := n.Child(0).Child(1)
	rangeStmt := &ast.RangeStmt{
		TokPos: token.Pos(tok.Pos()),
		Tok:    token.Token(tok.ID()),
		X:      p.parseExpr(n.Child(2)),
	}
	if kv.Is(exprList) {
		es := p.parseExprList(kv)
		rangeStmt.Key = es[0]
		if len(es) > 1 {
			rangeStmt.Value = es[1]
		}
	} else {
		es := p.parseIdentList(kv)
		rangeStmt.Key = es[0]
		if len(es) > 1 {
			rangeStmt.Value = es[1]
		}
	}
	return rangeStmt
}

func (p *parser) parseSimpleStmt(n *parse.Node) ast.Stmt {
	n = n.Child(0)
	switch n.Rule() {
	case exprStmt:
		return p.parseExprStmt(n)
	case sendStmt:
	case incDecStmt:
	case assignment:
		return p.parseAsignment(n)
	case shortVarDecl:
		return p.parseShortVarDecl(n)
	}
	fmt.Println("simpleStmt", n)
	return nil
}

func (p *parser) parseExprStmt(n *parse.Node) ast.Stmt {
	return &ast.ExprStmt{p.parseExpr(n)}
}

func (p *parser) parseAsignment(n *parse.Node) ast.Stmt {
	return &ast.AssignStmt{
		Lhs: p.parseExprList(n.Child(0)),
		Tok: token.Token(n.Child(1).Child(0).ID()),
		Rhs: p.parseExprList(n.Child(2)),
	}
}

func (p *parser) parseShortVarDecl(n *parse.Node) ast.Stmt {
	idents := p.parseIdentList(n.Child(0))
	exprs := make([]ast.Expr, len(idents))
	for i := range exprs {
		exprs[i] = idents[i]
	}
	ast := &ast.AssignStmt{
		Lhs:    exprs,
		TokPos: token.Pos(n.Child(1).Pos()),
		Tok:    token.Token(n.Child(1).ID()),
		Rhs:    p.parseExprList(n.Child(2)),
	}
	p.declareShortVar(ast, exprs)
	return ast
}

func (p *parser) parseExprList(n *parse.Node) (exprs []ast.Expr) {
	exprs = append(exprs, p.parseExpr(n.Child(0)))
	n.Child(1).Each(func(item *parse.Node) {
		exprs = append(exprs, p.parseExpr(item.Child(1)))
	})
	return
}

func (p *parser) parsePackageClause(n *parse.Node) (token.Pos, *ast.Ident) {
	pac, name := n.Child(0), n.Child(1)
	return token.Pos(pac.Pos()), &ast.Ident{
		NamePos: token.Pos(name.Pos()),
		Name:    string(name.Value()),
	}
}

func (p *parser) parseImports(n *parse.Node) (decls []ast.Decl, specs []*ast.ImportSpec) {
	n.Each(func(item *parse.Node) {
		decl, ss := p.parseImportDecl(item.Child(0))
		decls = append(decls, decl)
		specs = append(specs, ss...)
	})
	return
}

func (p *parser) parseImportDecl(n *parse.Node) (decl *ast.GenDecl, specs []*ast.ImportSpec) {
	decl = &ast.GenDecl{
		TokPos: token.Pos(n.Child(0).Pos()),
		Tok:    token.IMPORT,
		Lparen: token.Pos(n.Child(1).Child(0).Pos()),
		Rparen: token.Pos(n.Child(1).Child(3).Pos()),
	}
	n = n.Child(1)
	if n.Child(0).Is(importSpec) {
		spec := p.parseImportSpec(n.Child(0))
		specs = append(specs, spec)
		decl.Specs = append(decl.Specs, spec)
		return
	}
	n = n.Child(1) // skip (
	n.Each(func(item *parse.Node) {
		spec := p.parseImportSpec(item.Child(0))
		specs = append(specs, spec)
		decl.Specs = append(decl.Specs, spec)
	})
	return
}

func (p *parser) parseImportSpec(n *parse.Node) *ast.ImportSpec {
	spec := ast.ImportSpec{}
	if name := n.Child(0); name != nil {
		name = name.Child(0).Child(0)
		switch name.Rule() {
		case identifier:
			spec.Name = p.parseIdent(name)
		case term("."):
			spec.Name = &ast.Ident{
				NamePos: token.Pos(name.Pos()),
				Name:    ".",
			}
		}
	}
	spec.Path = p.parseBasicLit(n.Child(1))
	return &spec
}

func (p *parser) parseIdent(n *parse.Node) *ast.Ident {
	return &ast.Ident{
		NamePos: token.Pos(n.Pos()),
		Name:    string(n.Value()),
	}
}

func (p *parser) parseBasicLit(n *parse.Node) *ast.BasicLit {
	return &ast.BasicLit{
		ValuePos: token.Pos(n.Pos()),
		Kind:     token.Token(n.ID()),
		Value:    string(n.Value()),
	}
}

func (p *parser) parseGoExpr(n *parse.Node) ast.Expr {
	n = n.Child(0).Child(0)
	switch n.Rule() {
	case expr:
		return p.parseExpr(n)
	case type_:
		return p.parseType(n)
	}
	return nil
}

func (p *parser) parseExpr(n *parse.Node) ast.Expr {
	switch n.ChildCount() {
	case 1:
		return p.parseUnaryExpr(n.Child(0))
	case 3:
		op := n.Child(1).Child(0).Child(0)
		return &ast.BinaryExpr{
			X:     p.parseExpr(n.Child(0)),
			OpPos: token.Pos(op.Pos()),
			Op:    token.Token(op.ID()),
			Y:     p.parseUnaryExpr(n.Child(2)),
		}
	}
	return nil
}

func (p *parser) parseUnaryExpr(n *parse.Node) ast.Expr {
	if n.Child(0).Is(primaryExpr) {
		return p.parsePrimaryExpr(n.Child(0))
	}
	return &ast.UnaryExpr{
		OpPos: token.Pos(n.Child(0).Child(0).Pos()),
		Op:    token.Token(n.Child(0).Child(0).ID()),
		X:     p.parseUnaryExpr(n.Child(1)),
	}
}

func (p *parser) parsePrimaryExpr(n *parse.Node) ast.Expr {
	if n.Child(0).Is(operand) {
		return p.parseOperand(n.Child(0))
	}
	switch n.Child(1).Rule() {
	case selector:
		return p.parseSelector(n)
	case index:
		return p.parseIndex(n)
	case slice:
	case typeAssertion:
	case call:
		return p.parseCallExpr(n)
	}
	fmt.Println(n)
	return nil
}

func (p *parser) parseIndex(n *parse.Node) ast.Expr {
	index := n.Child(1)
	return &ast.IndexExpr{
		X:      p.parsePrimaryExpr(n.Child(0)),
		Lbrack: token.Pos(index.Child(0).Pos()),
		Index:  p.parseExpr(index.Child(1)),
		Rbrack: token.Pos(index.Child(2).Pos()),
	}
}

func (p *parser) parseSelector(n *parse.Node) ast.Expr {
	return &ast.SelectorExpr{
		X:   p.parsePrimaryExpr(n.Child(0)),
		Sel: p.parseIdent(n.Child(1).Child(1)),
	}
}

func (p *parser) parseCallExpr(n *parse.Node) *ast.CallExpr {
	call := n.Child(1)
	argList := call.Child(1)
	return &ast.CallExpr{
		Fun:      p.parsePrimaryExpr(n.Child(0)),
		Lparen:   token.Pos(call.Child(0).Pos()),
		Args:     p.parseArgs(argList.Child(0)),
		Ellipsis: token.Pos(argList.Child(1).Pos()),
		Rparen:   token.Pos(call.Child(2).Pos()),
	}
}

func (p *parser) parseArgs(n *parse.Node) []ast.Expr {
	n = n.Child(0)
	if n == nil {
		return nil
	}
	return p.parseExprList(n)
}

func (p *parser) parseOperand(n *parse.Node) ast.Expr {
	n = n.Child(0)
	switch n.Rule() {
	case literal:
		return p.parseLiteral(n)
	case operandName:
		return p.parseIdent(n)
	default:
		return p.parseExpr(n.Child(1))
	}
	return nil
}

func (p *parser) parseLiteral(n *parse.Node) ast.Expr {
	n = n.Child(0)
	switch n.Rule() {
	case basicLit:
		return p.parseBasicLit(n.Child(0))
	case compositeLit:
		return p.parseCompositeLit(n)
	case funcLit:
		return p.parseFuncLit(n)
	}
	return nil
}

func (p *parser) parseCompositeLit(n *parse.Node) ast.Expr {
	litValue := n.Child(1)
	return &ast.CompositeLit{
		Type:   p.parseLiteralType(n.Child(0)),
		Lbrace: token.Pos(litValue.Child(0).Pos()),
		Elts:   p.parseLiteralValue(litValue),
		Rbrace: token.Pos(litValue.Child(3).Pos()),
	}
}

func (p *parser) parseLiteralType(n *parse.Node) ast.Expr {
	n = n.Child(0)
	switch n.Rule() {
	case arrayType:
	case structType:
		return p.parseStructType(n)
	case sliceType:
		return p.parseSliceType(n)
	case mapType:
	case typeName:
	default:
		// array
	}
	fmt.Println(n)
	return nil
}

func (p *parser) parseSliceType(n *parse.Node) ast.Expr {
	return &ast.ArrayType{
		Lbrack: token.Pos(n.Child(0).Pos()),
		Len:    nil,
		Elt:    p.parseType(n.Child(2)),
	}
}

func (p *parser) parseLiteralValue(n *parse.Node) (exprs []ast.Expr) {
	n.Child(1).Each(func(item *parse.Node) {
		exprs = append(exprs, p.parseElement(item.Child(0)))
	})
	if n.Child(2) != nil {
		exprs = append(exprs, p.parseElement(n.Child(2).Child(0)))
	}
	return
}

func (p *parser) parseElement(n *parse.Node) ast.Expr {
	if n.Child(0) == nil {
		return p.parseValue(n.Child(1))
	}
	return p.parseKeyValue(n)
}

func (p *parser) parseKeyValue(n *parse.Node) *ast.KeyValueExpr {
	return &ast.KeyValueExpr{
		Key:   p.parseExpr(n.Child(0).Child(0)),
		Colon: token.Pos(n.Child(0).Child(1).Pos()),
		Value: p.parseValue(n.Child(1)),
	}
}

func (p *parser) parseKey(n *parse.Node) ast.Expr {
	return p.parseExpr(n)
}

func (p *parser) parseValue(n *parse.Node) ast.Expr {
	n = n.Child(0)
	switch n.Rule() {
	case expr:
		return p.parseExpr(n)
	case literalValue:
	}
	fmt.Println("parseValue", n)
	return nil
}

func (p *parser) parseFuncLit(n *parse.Node) ast.Expr {
	scope := ast.NewScope(p.topScope)
	funcLit := ast.FuncLit{}
	funcLit.Type, funcLit.Body = p.parseFunc(n.Child(1), scope)
	return &funcLit
}

func (p *parser) parseType(n *parse.Node) ast.Expr {
	n = n.Child(0)
	switch n.Rule() {
	case typeName:
		return p.parseTypeName(n)
	case typeLit:
		return p.parseTypeLit(n)
	}
	return p.parseType(n.Child(1))
}

func (p *parser) parseTypeName(n *parse.Node) ast.Expr {
	n = n.Child(0)
	if n.Rule() == identifier {
		return p.parseIdent(n)
	}
	return p.parseQualifiedIdent(n)
}

func (p *parser) parseQualifiedIdent(n *parse.Node) ast.Expr {
	return &ast.SelectorExpr{
		X:   p.parseIdent(n.Child(0)),
		Sel: p.parseIdent(n.Child(2)),
	}
}

func (p *parser) parseTypeLit(n *parse.Node) ast.Expr {
	n = n.Child(0)
	switch n.Rule() {
	case arrayType:
	case structType:
		return p.parseStructType(n)
	case pointerType:
		return p.parsePointerType(n)
	case funcType:
	case interfaceType:
		return p.parseInterfaceType(n)
	case sliceType:
	case mapType:
	case channelType:
	}
	fmt.Println(n)
	return nil
}

func (p *parser) parseInterfaceType(n *parse.Node) ast.Expr {
	specs := ast.FieldList{
		Opening: token.Pos(n.Child(1).Pos()),
		Closing: token.Pos(n.Child(4).Pos()),
	}
	n.Child(2).Each(func(item *parse.Node) {
		specs.List = append(specs.List, p.parseMethodSpec(item.Child(0)))
	})
	if n.Child(3) != nil {
		specs.List = append(specs.List, p.parseMethodSpec(n.Child(3).Child(0)))
	}
	return &ast.InterfaceType{
		Interface: token.Pos(n.Child(0).Pos()),
		Methods:   &specs,
	}
}

func (p *parser) parseMethodSpec(n *parse.Node) *ast.Field {
	fmt.Println(n)
	return nil
}

func (p *parser) parsePointerType(n *parse.Node) ast.Expr {
	return &ast.StarExpr{
		Star: token.Pos(n.Child(0).Pos()),
		X:    p.parseType(n.Child(1)),
	}
}

func (p *parser) parseStructType(n *parse.Node) ast.Expr {
	struct_ := n.Child(0)
	return &ast.StructType{
		Struct: token.Pos(struct_.Pos()),
		// TODO: Fields
	}
}
