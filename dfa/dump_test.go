package dfa

import (
	"bytes"

	"github.com/hailiang/gspec"

	"testing"
)

func TestDotFormat(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	m := threeToken()
	expect(m.dotFormat()).Equal(gspec.Unindent(`
		digraph g {
			rankdir=LR;
			node [fontname="Source Code Pro"];
			edge [fontname="Source Code Pro"];
			edge [arrowhead=lnormal];
			node [shape=point];
			ENTRY;
			node [shape=circle, height=0.2];
			ENTRY -> 0 [label="(input)"];
			0 -> 1 [label="[0]"];
			0 -> 2 [label="[1-9]"];
			0 -> 3 [label="[A-Za-z]"];
			1 [style="filled"];
			node [style="solid"];
			1 -> 2 [label="[0-9]"];
			1 -> 4 [label="[Xx]"];
			2 [style="filled"];
			node [style="solid"];
			2 -> 2 [label="[0-9]"];
			3 [style="filled"];
			node [style="solid"];
			3 -> 3 [label="[0-9A-Za-z]"];
			4 -> 5 [label="[0-9A-Fa-f]"];
			5 [style="filled"];
			node [style="solid"];
			5 -> 5 [label="[0-9A-Fa-f]"];
		}
	`))
}

func (m *machine) dotFormat() string {
	var w bytes.Buffer
	w.WriteByte('\n')
	m.writeDotFormat(&w)
	w.WriteByte('\n')
	return w.String()
}
