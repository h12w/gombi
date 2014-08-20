package scan

import "strconv"

type Pattern struct {
	rxSyntax
}

func Pat(pat string) Pattern {
	return Pattern{rxSyntax{parsePat(pat)}}
}

func Str(str string) Pattern {
	return Pattern{rxSyntax{parseStr(str)}}
}

func (p Pattern) ZeroOrOne() Pattern {
	var w exprWriter
	w.group(p)
	w.WriteByte('?')
	return w.pat()
}

func (p Pattern) ZeroOrMore() Pattern {
	var w exprWriter
	w.group(p)
	w.WriteByte('*')
	return w.pat()
}

func (p Pattern) Ungreedy() Pattern {
	var w exprWriter
	w.WriteString("(?U:")
	w.WriteString(p.String())
	w.WriteByte(')')
	return w.pat()
}

func (p Pattern) OneOrMore() Pattern {
	var w exprWriter
	w.group(p)
	w.WriteByte('+')
	return w.pat()
}

func (p Pattern) Repeat(n int) Pattern {
	var w exprWriter
	w.group(p)
	w.WriteByte('{')
	w.WriteString(strconv.Itoa(n))
	w.WriteByte('}')
	return w.pat()

}

func Or(es ...Expr) Pattern {
	return exprs(es).capture(false)
}

func Con(es ...Expr) Pattern {
	var w exprWriter
	for _, e := range es {
		w.group(e)
	}
	return w.pat()
}
