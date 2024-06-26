package util

import (
	"fmt"
	"sort"
)

type Decoder struct {
	buf   []byte
	pos   int
	magic []int
	word  byte
	bit   byte
}

func asSigned(i int) int {
	if (i & 1) != 0 {
		return ^i >> 1
	} else {
		return i >> 1
	}
}

func NewDecoder(v []byte) *Decoder {
	var d = &Decoder{}
	d.buf = v
	d.magic = d.readMagic()
	return d
}

func (d *Decoder) AssertEOF() {
	if d.pos < len(d.buf) {
		panic(fmt.Sprintf("expected eof: %d/%d", d.pos, len(d.buf)))
	}
}

func (d *Decoder) readMagic() []int {
	var list []int
	w := 0
	for {
		dw := d.readUnary()
		if dw == 0 {
			break
		}
		w += dw
		list = append(list, w)
	}
	return list
}

func (d *Decoder) readBit() bool {
	if d.bit == 0 {
		d.word = d.buf[d.pos]
		d.pos++
		d.bit = 1
	}
	bit := (d.word & d.bit) != 0
	d.bit <<= 1
	return bit
}

func (d *Decoder) readUnary() int {
	x := 0
	for d.readBit() {
		x++
	}
	return x
}

func (d *Decoder) readBinary(w int) int {
	x := 0
	for b := 1 << (w - 1); b != 0; b >>= 1 {
		if d.readBit() {
			x |= b
		}
	}
	return x
}

func (d *Decoder) ReadUnsigned() int {
	a := 0
	var w int
	for i := 0; ; i++ {
		w = d.magic[i]
		n := 1 << w
		if i+1 == len(d.magic) || !d.readBit() {
			break
		}
		a += n
	}
	return a + d.readBinary(w)
}

func (d *Decoder) readArray(n int, fn func(prev, x int) int) []int {
	v := make([]int, n)
	prev := -1
	for i := 0; i < n; i++ {
		v[i] = fn(prev, d.ReadUnsigned())
		prev = v[i]
	}
	return v
}

func (d *Decoder) ReadSortedAscending(n int) []int {
	return d.readArray(n, func(prev, x int) int { return prev + 1 + x })
}

func (d *Decoder) ReadUnsortedDeltas(n int) []int {
	return d.readArray(n, func(prev, x int) int { return prev + asSigned(x) })
}

func (d *Decoder) ReadString() string {
	v := d.ReadUnsortedDeltas(d.ReadUnsigned())
	cps := make([]rune, len(v))
	for i, x := range v {
		cps[i] = rune(x)
	}
	return string(cps)
}

func (d *Decoder) ReadUnique() []int {
	v := d.ReadSortedAscending(d.ReadUnsigned())
	n := d.ReadUnsigned()
	if n > 0 {
		vX := d.ReadSortedAscending(n)
		vS := d.ReadUnsortedDeltas(n)
		for i := 0; i < n; i++ {
			for x, e := vX[i], vX[i]+vS[i]; x < e; x++ {
				v = append(v, x)
			}
		}
	}
	return v
}

func (d *Decoder) ReadSortedUnique() []int {
	v := d.ReadUnique()
	sort.Ints(v)
	return v
}
