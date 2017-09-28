package ua

import (
	"fmt"
	"reflect"

	"h12.me/gombi/parse"
	ogdl "github.com/ogdl/flow"
)

var (
	term = parse.Term
	rule = parse.Rule
	or   = parse.Or
	con  = parse.Con
	self = parse.Self
)

func op(v interface{}) {
	buf, _ := ogdl.MarshalIndent(v, "    ", "    ")
	typ := ""
	if v != nil {
		typ = reflect.TypeOf(v).String() + "\n"
	}
	fmt.Println("\n" +
		typ +
		string(buf))
}
