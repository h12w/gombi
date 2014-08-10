package main

import (
	"fmt"
	"strings"
)

type (
	rule struct {
		id   string
		term bool
		alts
	}
	alt struct {
		rules
	}
	rules []*rule
	alts  []*alt
)

func term(s string) *rule {
	return &rule{id: s, term: true, alts: nil}
}

func (r *rule) expr() string {
	return r.id
}

func (r *rule) def() string {
	return fmt.Sprintf("%s ::= %s", r.id, r.alts.expr())
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
