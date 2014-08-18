package ua

import (
	"strings"

	"github.com/hailiang/gombi/parse"
)

type Product struct {
	Name    string
	Version Version
	Comment Comment
}
type Version struct {
	Text    string
	Numbers []int
}
type Comment struct {
	Items    []string
	Comments []Comment
}

func ParseUserAgent(s string) ([]*Product, error) {
	scanner.SetReader(strings.NewReader(s))
	parser.Reset()
	for scanner.Scan() &&
		parser.Parse(scanner.parserToken()) {
	}
	if scanner.Error() != nil {
		return nil, scanner.Error()
	}
	if parser.Error() != nil {
		return nil, parser.Error()
	}
	r := parser.Results()[0]

	ps := []*Product{}
	p := &Product{}
	r.Child(0).Each(func(item *parse.Node) {
		item = item.Child(0) // from (product | comment) to product or comment
		if item.Is(product) {
			p = newProduct(item)
			ps = append(ps, p)
		} else {
			p.Comment = p.Comment.append(item)
		}
	})
	if len(ps) == 0 {
		ps = append(ps, p)
	}
	return ps, nil
}

func newProduct(productNode *parse.Node) *Product {
	return &Product{
		Name: productNode.Get(productName),
		Version: Version{
			Text: productNode.Get(productVersion),
		},
	}
}

func (c Comment) append(commentNode *parse.Node) Comment {
	commentNode.Child(1).Each(func(citem *parse.Node) {
		citem = citem.Child(0)
		if citem.Is(commentText) {
			c.Items = append(c.Items, citem.Get(commentText))
		} else if citem.Is(comment) {
			c.Comments = append(c.Comments, Comment{}.append(citem))
		}
	})
	return c
}
