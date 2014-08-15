package scan

type Pattern struct {
	rxSyntax
}

func Pat(pat string) Pattern {
	p := Pattern{rxSyntax{parse(pat)}}
	return p
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

func (p Pattern) OneOrMore() Pattern {
	var w exprWriter
	w.group(p)
	w.WriteByte('+')
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
