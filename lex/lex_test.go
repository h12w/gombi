package lex

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hailiang/gspec/core"
	exp "github.com/hailiang/gspec/expectation"
	"github.com/hailiang/gspec/suite"

	ogdl "github.com/ogdl/flow"
)

var _ = suite.Add(func(s core.S) {
	describe, given, it, they := suite.Alias4("describe", "given", "it", "they", s)
	expect := exp.Alias(s.FailNow)
	describe("lexer patterns", func() {
		describe("character patterns", func() {
			given("a character range pattern", func() {
				c := Char(`abcde`)
				it("can be converted to a canonical pattern", func() {
					expect(c.String()).Equal(`[a-e]`)
				})
				it("can be negated", func() {
					nc := c.Negate()
					expect(nc.String()).Equal(`[^a-e]`)
					expect(nc.Negate().String()).Equal(`[a-e]`)
				})
				it("can exclude a subset", func() {
					expect(c.Exclude(Char(`bd`)).String()).Equal(`[ace]`)
				})
			})
			given("a single character pattern (OpLiteral)", func() {
				a := Char(`a`)
				it("can be negated", func() {
					na := a.Negate()
					expect(na.String()).Equal(`[^a]`)
					expect(na.Negate().String()).Equal(`[a]`)
				})
			})
			given("multiple character patterns", func() {
				a := Char(`a`)
				mn := Char(`mn`)
				z := Char(`z`)
				they("can be merged into one", func() {
					one := Or(a, mn, z)
					expect(one.String()).Equal(`[am-nz]`)
				})
			})
		})
	})
})

// func TestIt(t *testing.T) {
// 	nonctrl := Char(`[:^cntrl:]`)
// 	linebreak := Char(`\r\n`)
// 	//const delemiter = `{},`
// 	inline := Or(nonctrl, Char(` \t`))
// 	any := Or(inline, linebreak)
// 	invalid := any.Negate()
// 	space := Or(Char(` \t`), linebreak)
//
// 	newline := Or(linebreak, Pat(`\r\n`))
//
// 	p(inline)
// 	p(any)
// 	p(invalid)
// 	p(space)
// 	p(newline)
// }

func TestAll(t *testing.T) {
	suite.Test(t)
}

func p(v ...interface{}) {
	fmt.Println(v...)
}

func op(v interface{}) {
	buf, _ := ogdl.MarshalIndent(v, "    ", "    ")
	typ := ""
	if v != nil {
		typ = reflect.TypeOf(v).String() + "\n"
	}
	p("\n" +
		typ +
		string(buf) +
		"\n")
}
