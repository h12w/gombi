package dfa

import (
	"bytes"

	"github.com/hailiang/gspec"

	"testing"
)

func TestDotFormat(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	expect(threeToken.dotFormat()).Equal(gspec.Unindent(`
		digraph g {
			rankdir=LR;
			node [fontname="Ubuntu Mono"];
			edge [fontname="Ubuntu Mono"];
			node [fontsize=12, shape=circle, fixedsize=true, width=".25"];
			edge [fontsize=12];
			edge [arrowhead=lnormal];
			ENTRY [shape=point, fixedsize=false, width=".05"];
			ENTRY -> 0 [label="(input)"];
			0 -> 1 [label="'0'"];
			0 -> 2 [label="'1'-'9'"];
			0 -> 3 [label="'A'-'Z'\n'a'-'z'"];
			1 [shape=doublecircle, width=".18"];
			1 -> 2 [label="'0'-'9'"];
			1 -> 4 [label="'X'\n'x'"];
			2 [shape=doublecircle, width=".18"];
			2 -> 2 [label="'0'-'9'"];
			3 [shape=doublecircle, width=".18"];
			3 -> 3 [label="'0'-'9'\n'A'-'Z'\n'a'-'z'"];
			4 -> 5 [label="'0'-'9'\n'A'-'F'\n'a'-'f'"];
			5 [shape=doublecircle, width=".18"];
			5 -> 5 [label="'0'-'9'\n'A'-'F'\n'a'-'f'"];
		}
	`))
}

func (m *Machine) dotFormat() string {
	var w bytes.Buffer
	w.WriteByte('\n')
	m.writeDotFormat(&w, &GraphOption{"Ubuntu Mono", false})
	w.WriteByte('\n')
	return w.String()
}
