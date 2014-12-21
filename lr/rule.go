package parse

type (
	ruleType int
	// Rule is a BNF production rule
	Rule struct {
		id   int
		Name string
		typ  ruleType
		Rules
		builder *Builder
	}
	Rules   []*Rule
	Builder struct {
		EOF   *Rule
		rules []Rule
		terms map[string]*Rule
	}
)

const (
	conRule ruleType = iota
	orRule
	termRule
)

func NewBuilder() *Builder {
	b := &Builder{terms: make(map[string]*Rule)}
	b.EOF = b.Term("EOF")
	return b
}

func (b *Builder) newRule(typ ruleType, name string) *Rule {
	b.rules = append(b.rules, Rule{id: len(b.rules), Name: name, typ: typ, builder: b})
	return &b.rules[len(b.rules)-1]
}

func (b *Builder) Rule(name string) *Rule {
	if name == "" {
		panic("rule name should not be empty")
	}
	return b.newRule(conRule, name)
}

func (b *Builder) Term(name string) *Rule {
	if r, ok := b.terms[name]; ok {
		return r
	}
	r := b.newRule(termRule, name)
	b.terms[name] = r
	return r
}

func (b *Builder) toRules(a []interface{}) []*Rule {
	rs := make([]*Rule, len(a))
	for i := range rs {
		switch o := a[i].(type) {
		case *Rule:
			rs[i] = o
		case string:
			rs[i] = b.Term(o)
		default:
			panic("type of argument should be either string or *Rule")
		}
	}
	return rs
}

func (r *Rule) Define(o *Rule) *Rule {
	r.typ = o.typ
	r.Rules = o.Rules
	return r
}

func (b *Builder) Con(rs ...interface{}) *Rule {
	return b.con(b.toRules(rs)...)
}
func (b *Builder) con(rules ...*Rule) *Rule {
	if len(rules) == 1 {
		return rules[0]
	}
	r := b.newRule(conRule, "")
	r.Rules = Rules(rules)
	return r
}

func (b *Builder) Or(rs ...interface{}) *Rule {
	return b.or(b.toRules(rs)...)
}
func (b *Builder) or(rules ...*Rule) *Rule {
	if len(rules) == 1 {
		return rules[0]
	}
	r := b.newRule(orRule, "")
	r.Rules = Rules(rules)
	return r
}

func (r *Rule) As(name string) *Rule {
	if r.Name != "" {
		rules := Rules{r}
		r = r.builder.newRule(conRule, name)
		r.Rules = rules
	}
	r.Name = name
	return r
}

func (r *Rule) AtLeast(n int) *Rule {
	if n == 1 {
		return r.oneOrMore()
	} else if n > 1 {
		rs := make(Rules, n)
		for i := range rs {
			rs[i] = r
		}
		rs[n-1] = r.oneOrMore()
		return r.builder.con(rs...)
	}
	panic("n should be at larger than 0")
}

func (r *Rule) oneOrMore() *Rule {
	x := r.builder.newRule(orRule, r.nameOrDef()+"+")
	x.Define(r.builder.or(r, r.builder.con(r, x)))
	return x
}
