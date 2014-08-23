package scan

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func (l *Literal) String() string {
	var w bytes.Buffer
	for _, r := range string(l.a) {
		w.WriteString(escapeLiteral(r))
	}
	return w.String()
}
func escapeLiteral(r rune) string {
	switch r {
	case '(', ')', '[', ']', '-', '{', '}', '+', '*', '?':
		return `\` + string(r)
	}
	return strings.Trim(strconv.QuoteRune(r), `'`)
}

func (s *CharSet) String() string {
	var w bytes.Buffer
	w.WriteByte('[')
	for _, r := range s.ranges {
		if r.s == r.e {
			w.WriteString(escapeCharset(r.s))
		} else {
			w.WriteString(escapeCharset(r.s))
			w.WriteByte('-')
			w.WriteString(escapeCharset(r.e))
		}
	}
	w.WriteByte(']')
	return w.String()
}
func escapeCharset(r rune) string {
	switch r {
	case '[', ']', '-':
		return `\` + string(r)
	}
	return strings.Trim(strconv.QuoteRune(r), `'`)
}

func (r *Repetition) String() string {
	sentinel := ""
	if r.sentinel != nil {
		sentinel = "?" + r.sentinel.String()
	}
	max := ""
	if r.max == maxInt {
		if r.min == 0 {
			return parens(r.m.String()) + "*" + sentinel
		} else if r.min == 1 {
			return parens(r.m.String()) + "+" + sentinel
		}
	} else if r.min == r.max {
		return parens(r.m.String()) + fmt.Sprintf("{%d}", r.min) + sentinel
	} else {
		max = strconv.Itoa(r.max)
	}
	min := strconv.Itoa(r.min)
	return parens(r.m.String()) + fmt.Sprintf("{%s,%s}", min, max) + sentinel
}

func (s *Sequence) String() string {
	return matchers(s.ms).joinString("")
}

func (c *Choice) String() string {
	return parens(matchers(c.ms).joinString("|"))
}

type matchers []Matcher

func (ms matchers) joinString(sep string) string {
	ss := make([]string, len(ms))
	for i := range ss {
		ss[i] = ms[i].String()
	}
	return strings.Join(ss, sep)

}

func parens(s string) string {
	return "(" + s + ")"
}
