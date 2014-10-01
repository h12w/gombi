package parse

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func (r *R) Name() string {
	if r.name == "" {
		if r.recursive {
			addr := fmt.Sprintf("%p", r)
			return "<" + addr[len(addr)-4:] + ">"
		}
		return parens(r.Alts.String())
	}
	return escape(r.name)
}

func (r *R) String() string {
	return fmt.Sprintf("%s ::= %s", r.Name(), r.Alts.String())
}

func (rs Rules) toString(sep string) string {
	ss := make([]string, len(rs))
	for i := range rs {
		ss[i] = rs[i].Name()
	}
	s := strings.Join(ss, sep)
	return s
}

func (a Alt) String() string {
	return a.Rules.toString(" ")
}

func (as Alts) String() string {
	ss := make([]string, len(as))
	for i := range as {
		ss[i] = as[i].String()
	}
	s := strings.Join(ss, " | ")
	if len(as) > 1 {
		s = parens(s)
	}
	return s
}

func (s *state) name() string {
	return s.Alt.R.Name()
}

func (s *state) tokenValue() string {
	if s.token != nil && string(s.token.Value) != "" {
		return escape(string(s.token.Value))
	}
	return ""
}

func (s *state) String() string {
	if s.rule().isTerm {
		if s.d == 0 {
			return fmt.Sprintf("%s ::= •%v", s.name(), s.tokenValue())
		} else if s.d == 1 {
			return fmt.Sprintf("%s ::= %v•", s.name(), s.tokenValue())
		}
	}
	return fmt.Sprintf("%s ::= %v•%v", s.name(), s.Alt.Rules[:s.d].toString(" "), s.Alt.Rules[s.d:].toString(" "))
}
func escape(s string) string {
	if strings.HasPrefix(s, `"`) || strings.HasPrefix(s, `(`) {
		return s
	}
	s = strconv.Quote(s)
	if !strings.ContainsAny(s, " ()|") {
		s = s[1 : len(s)-1]
	}
	return s
}

func (s *state) traverse(level int, visit func(*state, int)) {
	if s == nil {
		return
	}
	visit(s, level)
	for _, c := range s.values {
		c.traverse(level+1, visit)
	}
}

func (s *state) traverseUp(m map[*R]bool, level int, visit func(*state, int)) {
	if s == nil {
		return
	}
	visit(s, level)
	if !m[s.rule()] {
		m[s.rule()] = true
		for _, c := range s.parents {
			c.traverseUp(m, level+1, visit)
		}
	}
}

func (s *state) dumpUp(level int) string {
	var w bytes.Buffer
	s.traverseUp(make(map[*R]bool), 0, func(st *state, level int) {
		fmt.Fprintf(&w, "%s%s\n", strings.Repeat("\t", level), st.String())
	})
	return w.String()
}

func (ss states) dumpUp() string {
	strs := make([]string, 0, len(ss.a))
	for _, s := range ss.a {
		strs = append(strs, s.dumpUp(0))
	}
	return strings.Join(strs, "\n")
}

func (ss states) String() string {
	strs := make([]string, 0, len(ss.a))
	for _, s := range ss.a {
		strs = append(strs, s.String())
	}
	return strings.Join(strs, "\n")
}

func (n *Node) String() string {
	output := ""
	indent := "\t"
	n.traverse(0, func(s *state, level int) {
		output += fmt.Sprintf("%s%s\n", strings.Repeat(indent, level), s.String())
	})
	return output
}

func parens(s string) string {
	if strings.ContainsAny(s, "|") {
		if !strings.HasPrefix(s, "(") {
			return "(" + s + ")"
		}
	}
	return s
}
