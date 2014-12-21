package parse

import (
	"bytes"
	"strings"
)

func (r *Rule) nameOrDef() string {
	if r.Name == "" {
		return "(" + r.Definition() + ")"
	}
	return r.Name
}

func (r *Rule) Definition() string {
	switch r.typ {
	case conRule:
		return r.Rules.toString(" ")
	case orRule:
		return r.Rules.toString(" | ")
	}
	return ""
}

func (r *Rule) String() string {
	var w bytes.Buffer
	w.WriteString(r.Name)
	w.WriteString(" ::= ")
	w.WriteString(r.Definition())
	return w.String()
}

func (rs Rules) toString(sep string) string {
	ss := make([]string, len(rs))
	for i := range rs {
		ss[i] = rs[i].nameOrDef()
	}
	return strings.Join(ss, sep)
}

func (r *matchingRule) String() string {
	var w bytes.Buffer
	w.WriteString(r.Name)
	w.WriteString(" ::= ")
	switch r.typ {
	case conRule:
		w.WriteString(r.Rules[:r.pos].toString(" "))
		w.WriteString("•")
		w.WriteString(r.Rules[r.pos:].toString(" "))
	case orRule:
		if r.pos == 0 {
			w.WriteString("•")
		}
		w.WriteString("(" + r.Definition() + ")")
		if r.pos == 1 {
			w.WriteString("•")
		}
	}
	return w.String()
}

func (rs matchingRules) String() string {
	ss := make([]string, len(rs))
	for i := range rs {
		ss[i] = rs[i].String()
	}
	return strings.Join(ss, "\n")
}

func (t *trans) String() string {
	return t.input.nameOrDef() + " ->\n" + indent(t.next.kernel.String(), "\t")
}

func (table transTable) String() string {
	ss := make([]string, len(table))
	for i := range table {
		ss[i] = table[i].String()
	}
	return strings.Join(ss, "\n")
}

func indent(s, indent string) string {
	ss := strings.Split(strings.TrimSpace(s), "\n")
	for i := range ss {
		ss[i] = indent + ss[i]
	}
	return strings.Join(ss, "\n")
}
