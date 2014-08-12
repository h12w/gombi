package parse

type (
	// rule is a BNF production rule
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

var ruleEOF = term("EOF")

func term(s string) *rule {
	return &rule{name: s}
}

func (r *rule) isEOF() bool {
	return r == ruleEOF
}

func (a alt) last() *rule {
	if len(a.rules) > 0 {
		return a.rules[len(a.rules)-1]
	}
	return nil
}
