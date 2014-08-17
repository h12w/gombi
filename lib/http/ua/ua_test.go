package ua

import (
	"fmt"
	"strings"
	//"strings"
	"testing"

	"github.com/hailiang/gspec/core"
	exp "github.com/hailiang/gspec/expectation"
	"github.com/hailiang/gspec/suite"
)

const (
	firefox = `Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:31.0) Gecko/20100101 Firefox/31.0`
	ie      = `Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0; GTB7.4; InfoPath.2; SV1; .NET CLR 3.3.69573; WOW64; en-US)`
)

var _ = suite.Add(func(s core.S) {
	testcase := s.Alias("testcase:")
	expect := exp.Alias(s.FailNow)
	testcase("firefox", func() {
		s := newScanner()
		s.Init(strings.NewReader(firefox))
		for s.Scan() {
			parser.Parse(s.parserToken())
		}
		expect(s.Error()).Equal(nil)
		expect(len(parser.Results())).Equal(1)
		r := parser.Results()[0]
		expect(r.Rule()).Equal(userAgent)

		//ps := []*Product{}
		//var p *Product
		next := ListIter(r)
		for {
			n := next()
			if n == nil {
				break
			}
			n = n.Child(0) // from (product | comment) to product or comment
			if n.Is(product) {
				p := &Product{
					Name: n.Get(productName),
					Version: Version{
						Text: n.Get(productVersion),
					},
				}
				fmt.Println(p)
			}
		}
	})
})

func TestAll(t *testing.T) {
	suite.Test(t)
}
