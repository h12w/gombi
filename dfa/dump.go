package dfa

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strconv"
)

func (s *state) id(start *state) string {
	return strconv.Itoa((s.intPtr() - start.intPtr()) / s.size())
}
func (s *state) intPtr() int {
	if i, err := strconv.ParseInt(fmt.Sprintf("%p", s)[2:], 16, 64); err == nil {
		return int(i)
	}
	return 0
}
func (s *state) size() int {
	return int(reflect.TypeOf(*s).Size())
}

func (s *state) dump(start *state) string {
	var w bytes.Buffer
	w.WriteByte('s')
	w.WriteString(s.id(start))
	w.WriteByte('\n')
	for _, trans := range s.tt {
		w.WriteString("    ")
		w.WriteString(trans.dump(start))
		w.WriteByte('\n')
	}
	w.WriteByte('\n')
	return w.String()
}

func (t *trans) dump(start *state) string {
	var w bytes.Buffer
	w.WriteString(t.rangeString())
	w.WriteString(": s")
	w.WriteString(t.next.id(start))
	return w.String()
}
func (t *trans) rangeString() string {
	if t.s == t.e {
		return quote(t.s)
	}
	return quote(t.s) + "-" + quote(t.e)
}
func quote(b byte) string {
	s := strconv.QuoteRune(rune(b))
	return s[1 : len(s)-1]
}

func (d *dfa) saveSvg(file string) error {
	dotCmd := exec.Command("dot", "-Tsvg", "-o", file)
	w, err := dotCmd.StdinPipe()
	if err != nil {
		return err
	}
	defer w.Close()
	go dotCmd.Run()
	return d.writeDotFormat(w)
}

func (d *dfa) writeDotFormat(writer io.Writer) error {
	var w bytes.Buffer
	const fontName = "Ubuntu Mono"
	w.WriteString("digraph g {\n")
	w.WriteString("    rankdir=LR;\n")
	fmt.Fprintf(&w, "    node [fontname=\"%s\"];\n", fontName)
	fmt.Fprintf(&w, "    edge [fontname=\"%s\"];\n", fontName)
	w.WriteString("    edge [arrowhead=lnormal];\n")
	w.WriteString("    node [shape=point];\n")
	w.WriteString("    ENTRY;\n")
	w.WriteString("    node [shape=circle, height=0.2];\n")
	w.WriteString("    ENTRY -> 0 [label=\"(input)\"];\n")
	if len(d.ss) > 0 {
		s0 := &d.ss[0]
		for i := range d.ss {
			s := &d.ss[i]
			s.writeDotFormat(&w, s0)
		}
	}
	w.WriteString("}")

	_, err := w.WriteTo(writer)
	return err
}
func (s *state) writeDotFormat(w io.Writer, s0 *state) {
	if s.accept {
		fmt.Fprintf(w, "%s [style=\"filled\"];", s.id(s0))
		fmt.Fprint(w, "node [style=\"solid\"];")
	}
	m := make(map[*state]bool)
	for _, trans := range s.tt {
		if !m[trans.next] {
			fmt.Fprintf(w, "    %s -> %s [label=\"[%s]\"];\n", s.id(s0), trans.next.id(s0), s.tt.label(trans.next))
			m[trans.next] = true
		}
	}
}

func (tt transTable) label(s *state) (l string) {
	for _, trans := range tt {
		if trans.next == s {
			l += trans.rangeString()
		}
	}
	return
}
