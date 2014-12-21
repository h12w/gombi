package parse

import (
	"fmt"
	"testing"

	"github.com/hailiang/gspec"
)

func TestRule(t *testing.T) {
	var (
		b                   = NewBuilder()
		term, rule, con, or = b.Term, b.Rule, b.Con, b.Or

		X = rule("X")
		_ = X.Define(con(X, "a"))
	)
	testTable(t, Table{
		0: {
			con("b", "c").As("A"),
			"A ::= b c"},
		1: {
			or("b", "c").As("A"),
			"A ::= b | c"},
		2: {
			con(or("b", "c"), "d").As("A"),
			"A ::= (b | c) d"},
		3: {
			or(con("b", "c"), "d").As("A"),
			"A ::= (b c) | d"},
		4: {
			term("a").AtLeast(1).As("A"),
			"A ::= a+"},
		5: {
			term("a").AtLeast(2).As("A"),
			"A ::= a a+"},
		6: {
			con("a", "b").AtLeast(1).As("A"),
			"A ::= (a b)+"},
		7: {
			con("a", "b").AtLeast(2).As("A"),
			"A ::= (a b) (a b)+"},
		8: {
			X,
			"X ::= X a"},
	})
}

func TestMatchingRule(t *testing.T) {
	var (
		b       = NewBuilder()
		con, or = b.Con, b.Or
		//term, rule, con, or = b.Term, b.Rule, b.Con, b.Or
	)
	testTable(t, Table{
		0: {
			&matchingRule{con("b", "c").As("A"), 0},
			"A ::= •b c"},
		1: {
			&matchingRule{con("b", "c").As("A"), 1},
			"A ::= b•c"},
		2: {
			&matchingRule{con("b", "c").As("A"), 2},
			"A ::= b c•"},
		3: {
			&matchingRule{or("b", "c").As("A"), 0},
			"A ::= •(b | c)"},
		4: {
			&matchingRule{or("b", "c").As("A"), 1},
			"A ::= (b | c)•"},
	})
}

func TestTrans(t *testing.T) {
	var (
		b         = NewBuilder()
		con, term = b.Con, b.Term
		c         = term("c")
		d         = term("d")
		e         = term("e")
		A         = con(c, d).As("A")
		B         = con(c, e).As("B")
	)
	testTable(t, Table{
		0: {
			&trans{c, &state{kernel: matchingRules{
				{A, 1},
				{B, 1},
			}}}, `
			c ->
				A ::= c•d
				B ::= c•e`,
		},
	})
}

func TestSingle(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	var (
		b                   = NewBuilder()
		term, rule, con, or = b.Term, b.Rule, b.Con, b.Or

		T    = term("T")
		Plus = term(`+`)
		Mult = term(`*`)
		M    = rule("M")
		_    = M.Define(or(
			con(M, Mult, T),
			T,
		))
		S = rule("S")
		_ = S.Define(or(
			con(S, Plus, M),
			M,
		))
		P     = S.As("P")
		start = matchingRules{{P, 0}}
	)

	fmt.Println(P)
	fmt.Println(S)
	fmt.Println(M)
	fmt.Println(T)
	closure0 := start.closure()
	expect(len(closure0)).Equal(6)
	transTables := make([]transTable, 6)
	for i := range closure0 {
		transTables[i] = closure0[i].next()
	}
	testTable(t, Table{
		0: {
			closure0, `
			P ::= •S
			S ::= •((S + M) | M)
			 ::= •S + M
			M ::= •((M * T) | T)
			 ::= •M * T
			T ::= `},
		1: {
			transTables[0], `
			S ->
				P ::= S•`},
		2: {
			transTables[1], `
			(S + M) ->
				S ::= ((S + M) | M)•
			M ->
				S ::= ((S + M) | M)•`},
		3: {
			mergeTransTable(transTables), `
			S ->
				P ::= S•
				 ::= S• + M
			(S + M) ->
				S ::= ((S + M) | M)•
			M ->
				S ::= ((S + M) | M)•
				 ::= M• * T
			`},
	})
}

type stringer interface {
	String() string
}

type TableEntry struct {
	obj stringer
	str string
}
type Table []TableEntry

func testTable(t *testing.T, table Table) {
	expect := gspec.Expect(t.FailNow, 1)
	for i, testcase := range table {
		expect(fmt.Sprintf("(testcase %d)", i), testcase.obj.String()).Equal(gspec.Unindent(testcase.str))
	}
}
