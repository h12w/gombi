package re

type token struct {
	id    int
	value []byte
}

type context struct {
	tokens []*token
}

func (c *context) capture(t *token) {
	c.tokens = append(c.tokens, t)
}

func (c *context) reset() {
	c.tokens = c.tokens[0:0]
}
