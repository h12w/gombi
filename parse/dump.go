package parse

import (
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

func (s *state) String() string {
	if s.token != nil {
		if s.d == 0 {
			return fmt.Sprintf("%s ::= •%v", s.name(), escape(string(s.token.Value)))
		} else if s.d == 1 {
			return fmt.Sprintf("%s ::= %v•", s.name(), escape(string(s.token.Value)))
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

func (ss *stateSet) String() string {
	strs := make([]string, 0, len(ss.a))
	for _, s := range ss.a {
		strs = append(strs, newNode(s).String())
	}
	return strings.Join(strs, "\n")
}

func (n *Node) String() string {
	output := ""
	indent := "\t"
	n.traverse(0, func(s *state, level int) {
		output += fmt.Sprintf("%s%s\n", strings.Repeat(indent, level), s.String())
	})
	return strings.TrimSuffix(output, "\n")
}

func parens(s string) string {
	if strings.ContainsAny(s, " |") {
		if !strings.HasPrefix(s, "(") {
			return "(" + s + ")"
		}
	}
	return s
}
