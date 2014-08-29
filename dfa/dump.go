package dfa

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"unicode/utf8"
)

var FontName = "Source Code Pro"

func (s *state) dump(ss []state, sid int) string {
	var w bytes.Buffer
	w.WriteByte('s')
	w.WriteString(strconv.Itoa(sid))
	if s.final() {
		w.WriteByte('$')
		if s.finalLabel > defaultFinal {
			w.WriteString(strconv.Itoa(s.finalLabel))
		}
	}
	for _, trans := range s.tt {
		w.WriteByte('\n')
		w.WriteString("\t")
		w.WriteString(trans.dump())
	}
	w.WriteByte('\n')
	return w.String()
}

func (t *trans) dump() string {
	var w bytes.Buffer
	w.WriteString(t.rangeString())
	w.WriteString("\ts")
	w.WriteString(strconv.Itoa(t.next))
	return w.String()
}
func (t *trans) rangeString() string {
	if t.s == t.e {
		return quote(t.s)
	}
	return quote(t.s) + "-" + quote(t.e)
}
func quote(b byte) string {
	if b < utf8.RuneSelf {
		s := strconv.QuoteRune(rune(b))
		return s[1 : len(s)-1]
	}
	return fmt.Sprintf(`\\x%X`, b)
}

func (m *machine) dump() string {
	var w bytes.Buffer
	w.WriteByte('\n')
	for i := range m.ss {
		w.WriteString(m.ss[i].dump(m.ss, i))
	}
	return w.String()
}

func (m *machine) saveSvg(file string) error {
	dotCmd := exec.Command("dot", "-Tsvg", "-o", file)
	w, err := dotCmd.StdinPipe()
	if err != nil {
		return err
	}
	ew, err := dotCmd.StderrPipe()
	if err != nil {
		return err
	}
	defer w.Close()
	go dotCmd.Run()
	go func() {
		buf, _ := ioutil.ReadAll(ew)
		fmt.Println(string(buf))
	}()
	return m.writeDotFormat(w)
}

func (m *machine) writeDotFormat(writer io.Writer) error {
	var w bytes.Buffer
	w.WriteString("digraph g {\n")
	w.WriteString("\trankdir=LR;\n")
	fmt.Fprintf(&w, "\tnode [fontname=\"%s\"];\n", FontName)
	fmt.Fprintf(&w, "\tedge [fontname=\"%s\"];\n", FontName)
	w.WriteString("\tedge [arrowhead=lnormal];\n")
	w.WriteString("\tnode [shape=point];\n")
	w.WriteString("\tENTRY;\n")
	w.WriteString("\tnode [shape=circle, height=0.2];\n")
	w.WriteString("\tENTRY -> 0 [label=\"(input)\"];\n")
	if len(m.ss) > 0 {
		for i := range m.ss {
			s := &m.ss[i]
			s.writeDotFormat(&w, i)
		}
	}
	w.WriteString("}")

	_, err := w.WriteTo(writer)
	return err
}
func (s *state) writeDotFormat(w io.Writer, sid int) {
	if s.final() {
		fmt.Fprintf(w, "\t%d [style=\"filled\"];\n", sid)
		fmt.Fprint(w, "\tnode [style=\"solid\"];\n")
	}
	m := make(map[int]bool)
	for _, trans := range s.tt {
		if !m[trans.next] {
			fmt.Fprintf(w, "\t%d -> %d [label=\"[%s]\"];\n", sid, trans.next, s.tt.label(trans.next))
			m[trans.next] = true
		}
	}
}

func (tt transTable) label(sid int) (l string) {
	for _, trans := range tt {
		if trans.next == sid {
			l += trans.rangeString()
		}
	}
	return
}
