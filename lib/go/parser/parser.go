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

	//// Comments
	//comments    []*ast.CommentGroup
	//leadComment *ast.CommentGroup // last lead comment
	//lineComment *ast.CommentGroup // last line comment

	//// Next token
	//pos token.Pos   // token position
	//tok token.Token // one token look-ahead
	//lit string      // token literal

	//// Error recovery
	//// (used to limit the number of calls to syncXXX functions
	//// w/o making scanning progress - avoids potential endless
	//// loops across multiple std_parser functions during error recovery)
	//syncPos token.Pos // last synchronization position
	//syncCnt int       // number of calls to syncXXX without progress

	//// Non-syntactic std_parser control
	//exprLev int  // < 0: in control clause, >= 0: in expression
	//inRhs   bool // if set, the std_parser is parsing a rhs expression

	// Ordinary identifier scopes
	pkgScope *ast.Scope // pkgScope.Outer == nil
	topScope *ast.Scope // top-most scope; may be pkgScope
	//unresolved []*ast.Ident      // unresolved identifiers
	//imports    []*ast.ImportSpec // list of imports

	//// Label scopes
	//// (maintained by open/close LabelScope)
	//labelScope  *ast.Scope     // label scope for current function
	//targetStack [][]*ast.Ident // stack of unresolved labels
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

func (p *parser) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

func (p *parser) parseExpr() ast.Expr {
	pp := parse.New(con(or(expression, type_), ";", parse.EOF).As("PExpr"))
	for {
		pos, tok, str := p.scanner.Scan()
		if r := tokenTable[tok]; r != nil {
			pp.Parse(&parse.Token{Value: []byte(str), Pos: int(pos)}, r)
			//fmt.Println(tok)
		}
		if tok == token.EOF {
			break
		}
	}
	fmt.Println(pp.Results())
	return nil
}
