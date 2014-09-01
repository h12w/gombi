package dfa

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"time"
	"unicode/utf8"
)

type GraphOption struct {
	FontName  string
	TimeLabel bool
}

func (s *state) dump(ss []state, sid int) string {
	var w bytes.Buffer
	w.WriteByte('s')
	w.WriteString(strconv.Itoa(sid))
	if s.final() {
		w.WriteByte('$')
		if s.label > defaultFinal {
			w.WriteString(strconv.Itoa(int(s.label)))
		}
	}
	for _, trans := range s.table {
		w.WriteByte('\n')
		w.WriteString("\t")
		w.WriteString(trans.dump())
	}
	w.WriteByte('\n')
	return w.String()
}

func (t *trans) dump() string {
	var w bytes.Buffer
	fmt.Fprintf(&w, "%-7s", t.rangeString())
	w.WriteString(" s")
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
	if b < utf8.RuneSelf && strconv.IsPrint(rune(b)) {
		return strconv.QuoteRune(rune(b))
	}
	return fmt.Sprintf(`%.2x`, b)
}

func (m *Machine) dump() string {
	var w bytes.Buffer
	w.WriteByte('\n')
	for i := range m.states {
		w.WriteString(m.states[i].dump(m.states, i))
	}
	return w.String()
}

func (m *Machine) SaveSVG(file string, opt ...*GraphOption) error {
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
	if len(opt) == 0 {
		opt = []*GraphOption{{}}
	}
	return m.writeDotFormat(w, opt[0])
}

func (m *Machine) writeDotFormat(writer io.Writer, opt *GraphOption) error {
	var w bytes.Buffer
	w.WriteString("digraph g {\n")
	if opt.TimeLabel {
		fmt.Fprintf(&w, "graph [label=\"(%s)\", labeljust=right, fontsize=12];", time.Now().Format("2006-01-02 15:04:05"))
	}
	w.WriteString("\trankdir=LR;\n")
	if opt.FontName != "" {
		fmt.Fprintf(&w, "\tnode [fontname=\"%s\"];\n", opt.FontName)
		fmt.Fprintf(&w, "\tedge [fontname=\"%s\"];\n", opt.FontName)
	}
	w.WriteString("\tnode [fontsize=12, shape=circle, fixedsize=true, width=\".25\"];\n")
	w.WriteString("\tedge [fontsize=12];\n")
	w.WriteString("\tedge [arrowhead=lnormal];\n")
	w.WriteString("\tENTRY [shape=point, fixedsize=false, width=\".05\"];\n")
	w.WriteString("\tENTRY -> 0 [label=\"(input)\"];\n")
	if len(m.states) > 0 {
		for i := range m.states {
			s := &m.states[i]
			s.writeDotFormat(&w, i)
		}
	}
	w.WriteString("}")

	_, err := w.WriteTo(writer)
	return err
}
func (s *state) writeDotFormat(w io.Writer, sid int) {
	if s.final() {
		fmt.Fprintf(w, "\t%d [shape=doublecircle, width=\".18\"];\n", sid)
	}
	m := make(map[int]bool)
	for _, trans := range s.table {
		if !m[trans.next] {
			fmt.Fprintf(w, "\t%d -> %d [label=\"%s\"];\n", sid, trans.next, s.table.description(trans.next))
			m[trans.next] = true
		}
	}
}
func (table transTable) description(sid int) (l string) {
	for _, trans := range table {
		if trans.next == sid {
			if l != "" {
				l += `\n`
			}
			l += trans.rangeString()
		}
	}
	return
}
