package scan

import (
	"fmt"
	"io"
	"unicode/utf8"
)

type runeBuffer struct {
	rd  io.Reader
	buf []byte
	r   int // position of next unread rune
	t   int // position of next unread token
	p   int // absolute position of next unread token
}

const defaultBufCap = 4096

func newRuneBuffer(rd io.Reader) *runeBuffer {
	return &runeBuffer{rd: rd, buf: make([]byte, 0, defaultBufCap)}
}

func (b *runeBuffer) ReadRune() (ru rune, size int, err error) {
	for !utf8.FullRune(b.buf[b.r:]) {
		if err := b.fill(); err != nil {
			return 0, 0, err
		}
	}
	ru, size = utf8.DecodeRune(b.buf[b.r:])
	b.r += size
	return ru, size, nil
}

func (b *runeBuffer) ReadToken(n int) (token []byte, pos int, err error) {
	if t := b.t + n; t > 0 && t <= len(b.buf) {
		token = b.buf[b.t : b.t+n]
		pos = b.p
		b.t = t
		b.r = b.t // reset rune position to prepare for next match.
		b.p += n
		return token, pos, nil
	}
	return nil, -1, fmt.Errorf("invalid n %d", n)
}

func (b *runeBuffer) bytes() []byte {
	return b.buf[b.t:]
}

func (b *runeBuffer) shift() {
	copy(b.buf, b.buf[b.t:])
	b.buf = b.buf[:len(b.buf)-b.t]
	b.r -= b.t
	b.t = 0
}

func (b *runeBuffer) fill() error {
	buf := b.buf
	if len(buf) == cap(buf) {
		b.shift()
		if len(buf) == cap(buf) {
			buf = append(buf, 0)[:len(buf)]
		}
	}
	n, err := b.rd.Read(buf[len(buf):cap(buf)])
	b.buf = buf[:len(buf)+n] // This is correct, err should be handled afterwards.
	return err
}
