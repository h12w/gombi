package ua

import "github.com/hailiang/gombi/parse"

type Product struct {
	Name     string
	Version  Version
	Comments []Comment
}
type Version struct {
	Text    string
	Numbers []int
}
type Comment struct {
	Items    []string
	Comments []Comment
}

func ListIter(n *parse.Node) func() *parse.Node {
	cur := n
	return func() *parse.Node {
		if cur == nil {
			return nil
		}
		item := cur.Child(0)
		cur = cur.Child(1)
		return item
	}
}
