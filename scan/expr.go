package scan

import "bytes"

type Expr interface {
	String() string
}

type exprs []Expr

func (es exprs) capture(captured bool) Pattern {
	var w exprWriter
	for i, e := range es {
		if i > 0 {
			w.WriteByte('|')
		}
		if captured {
			w.capture(e)
		} else {
			w.group(e)
		}
	}
	return w.pat()
}

type exprWriter struct {
	bytes.Buffer
}

func (w *exprWriter) group(e Expr) {
	w.WriteString("(?:")
	w.WriteString(e.String())
	w.WriteString(")")
}

func (w *exprWriter) capture(e Expr) {
	w.WriteString("(")
	w.WriteString(e.String())
	w.WriteString(")")
}

func (w *exprWriter) pat() Pattern {
	return Pat(w.String())
}
