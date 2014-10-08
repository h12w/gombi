package parser_test

import (
	"bytes"
	"fmt"
	std "go/parser"
	"go/printer"
	"go/token"
	"os"
	"reflect"
	"testing"

	gom "github.com/hailiang/gombi/lib/go/parser"
	"github.com/hailiang/gspec"
	"github.com/ogdl/flow"
)

func TestCompatible(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	srcList := []string{
		"compatible_test.go",
	}

	for _, srcFile := range srcList {
		fset := token.NewFileSet()
		stdAst, stdErr := std.ParseFile(fset, srcFile, nil, std.ParseComments)
		fset = token.NewFileSet()
		gomAst, gomErr := gom.ParseFile(fset, srcFile, nil, gom.ParseComments)
		expect("parse error", gomErr).Equal(stdErr)
		expect("decl count", len(gomAst.Decls)).Equal(len(stdAst.Decls))
		//fmt.Println(gomAst.Decls[1].(*ast.FuncDecl).Type.Results)
		//fmt.Println(stdAst.Decls[1].(*ast.FuncDecl).Type.Results)
		for i := range stdAst.Decls {
			expect(astStr(gomAst.Decls[i], fset)).Equal(astStr(stdAst.Decls[i], fset))
		}
		expect(astStr(gomAst, fset)).Equal(astStr(stdAst, fset))
	}
}

func printAst(o interface{}, fset *token.FileSet) {
	printer.Fprint(os.Stdout, fset, o)
}

func ogdlPrint(v interface{}) {
	buf, _ := flow.MarshalIndent(v, "    ", "    ")
	typ := ""
	if v != nil {
		typ = reflect.TypeOf(v).String() + "\n"
	}
	fmt.Print("\n" +
		typ +
		string(buf) +
		"\n")
}

func astStr(o interface{}, fset *token.FileSet) string {
	defer func() { recover() }()
	var w bytes.Buffer
	w.WriteByte('\n')
	printer.Fprint(&w, fset, o)
	return w.String()
}
