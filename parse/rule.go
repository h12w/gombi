package main

import (
	"fmt"
	"strings"
)

type (
	rule struct {
		name string
		alts
	}
	alt struct {
		rules
	}
	rules []*rule
	alts  []*alt
)

func term(s string) *rule {
	return &rule{name: s}
}

func (r *rule) expr() string {
	return r.name
}

func (r *rule) def() string {
	return fmt.Sprintf("%s ::= %s", r.name, r.alts.expr())
}

func (rs rules) expr() string {
	ss := make([]string, len(rs))
	for i := range rs {
		ss[i] = rs[i].expr()
	}
	return strings.Join(ss, " ")
}

func (as alts) expr() string {
	ss := make([]string, len(as))
	for i := range as {
		ss[i] = as[i].expr()
	}
	return strings.Join(ss, " | ")
}
