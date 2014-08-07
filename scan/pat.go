package scan

type Pattern struct {
	rxSyntax
}

func Pat(pat string) Pattern {
	p := Pattern{rxSyntax{parse(pat)}}
	return p
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
	return exprs(es).or(false)
}

func Con(es ...Expr) Pattern {
	var w exprWriter
	for _, e := range es {
		w.group(e)
	}
	return w.pat()
}

func Tokens(es ...Expr) Pattern {
	return Con(Pat(`\A`), exprs(es).or(true))
}
