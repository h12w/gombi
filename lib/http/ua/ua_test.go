package ua

import (
	"reflect"
	"strconv"
	"testing"

	"h12.io/gspec"
	"github.com/ogdl/flow"
)

var testcases = []struct {
	str string
	ua  []*Product
}{
	{`*Product`, []*Product{
		{Name: "*Product"},
	}},
	{`*Product/1.0`, []*Product{
		{Name: "*Product",
			Version: Version{
				Text: "1.0",
			}},
	}},
	{`()`, []*Product{
		{},
	}},
	{`(text)`, []*Product{
		{
			Comment: Comment{
				Items: []string{"text"},
			},
		},
	}},
	{`(text; (nested))`, []*Product{
		{
			Comment: Comment{
				Items: []string{"text"},
				Comments: []Comment{
					{Items: []string{"nested"}},
				},
			},
		},
	}},
	{`Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:31.0) Gecko/20100101 Firefox/31.0`,
		[]*Product{
			{
				Name: "Mozilla",
				Version: Version{
					Text: "5.0",
				},
				Comment: Comment{
					Items: []string{"X11", "Ubuntu", "Linux x86_64", "rv:31.0"},
				},
			},
			{
				Name: "Gecko",
				Version: Version{
					Text: "20100101",
				},
			},
			{
				Name: "Firefox",
				Version: Version{
					Text: "31.0",
				},
			},
		},
	},
}

var _ = gspec.Add(func(s gspec.S) {
	testcase := s.Alias("testcase:")
	expect := gspec.Expect(s.FailNow)
	for i, tc := range testcases {
		testcase(strconv.Itoa(i), func() {
			r, err := ParseUserAgent(tc.str)
			expect(err).Equal(nil)
			expect(r).Equal(tc.ua)
		})
	}
})

func TestAll(t *testing.T) {
	gspec.Test(t)
}

func init() {
	gspec.SetSprint(flowPrint)
}

func flowPrint(v interface{}) string {
	buf, _ := flow.MarshalIndent(v, "    ", "    ")
	typ := ""
	if v != nil {
		typ = reflect.TypeOf(v).String() + "\n"
	}
	return "\n" +
		typ +
		string(buf) +
		"\n"
}
