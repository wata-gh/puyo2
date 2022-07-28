package puyo2

import (
	"fmt"
	"math/bits"
)

type FieldBits struct {
	m [2]uint64
}

func NewFieldBits() *FieldBits {
	return new(FieldBits)
}

func NewFieldBitsWithM(m [2]uint64) *FieldBits {
	fb := new(FieldBits)
	fb.m = m
	return fb
}

func (fb *FieldBits) bitChar(b int) string {
	if b == 0 {
		return "."
	}
	return "o"
}

func (fb *FieldBits) expand(mask *FieldBits) *FieldBits {
	cfb := fb
	for {
		nfb := cfb.expand1(mask)
		if nfb.Equals(cfb) {
			return nfb
		}
		cfb = nfb
	}
}

func (fb *FieldBits) expand1(mask *FieldBits) *FieldBits {
	// 4 列目と 5 列目は変数をまたぐので個別に計算
	r4 := fb.m[0] & 0xffff000000000000
	r5 := fb.m[1] & 0xffff
	m0 := fb.m[0] | fb.m[0]<<1 | fb.m[0]>>1 | fb.m[0]<<16 | fb.m[0]>>16 | r5<<48
	m1 := fb.m[1] | fb.m[1]<<1 | fb.m[1]>>1 | fb.m[1]<<16 | fb.m[1]>>16 | r4>>48
	return NewFieldBitsWithM([2]uint64{m0 & mask.m[0], m1 & mask.m[1]})
}

func (fb *FieldBits) And(fb2 *FieldBits) *FieldBits {
	return NewFieldBitsWithM([2]uint64{fb.m[0] & fb2.m[0], fb.m[1] & fb2.m[1]})
}

func (fb *FieldBits) AndNot(fb2 *FieldBits) {
	fb.m[0] &= ^fb2.m[0]
	fb.m[1] &= ^fb2.m[1]
}

func (fb *FieldBits) Clone() *FieldBits {
	fieldBits := new(FieldBits)
	fieldBits.m[0] = fb.m[0]
	fieldBits.m[1] = fb.m[1]
	return fieldBits
}

func (fb *FieldBits) ColBits(c int) uint64 {
	switch c {
	case 0:
		return fb.m[0] & 0xffff
	case 1:
		return fb.m[0] & 0xffff0000
	case 2:
		return fb.m[0] & 0xffff00000000
	case 3:
		return fb.m[0] & 0xffff000000000000
	case 4:
		return fb.m[1] & 0xffff
	case 5:
		return fb.m[1] & 0xffff0000
	}
	panic(fmt.Sprintf("column number must be 0-5. passed %d", c))
}

func (fb *FieldBits) Equals(fb2 *FieldBits) bool {
	return fb.m[0] == fb2.m[0] && fb.m[1] == fb2.m[1]
}

func (fb *FieldBits) FindVanishingBits() *FieldBits {
	// 4 列目と 5 列目は変数をまたぐので個別に計算
	r4 := fb.m[0] & 0xffff000000000000
	r5 := fb.m[1] & 0xffff
	u := [2]uint64{fb.m[0] & (fb.m[0] << 1), fb.m[1] & (fb.m[1] << 1)}
	d := [2]uint64{fb.m[0] & (fb.m[0] >> 1), fb.m[1] & (fb.m[1] >> 1)}
	l := [2]uint64{fb.m[0] & (fb.m[0]>>16 | r5<<48), fb.m[1] & (fb.m[1] >> 16)}
	r := [2]uint64{fb.m[0] & (fb.m[0] << 16), fb.m[1] & (fb.m[1]<<16 | r4>>48)}

	ud_and := [2]uint64{u[0] & d[0], u[1] & d[1]}
	lr_and := [2]uint64{l[0] & r[0], l[1] & r[1]}
	ud_or := [2]uint64{u[0] | d[0], u[1] | d[1]}
	lr_or := [2]uint64{l[0] | r[0], l[1] | r[1]}
	threes_ud_and_and_lr_or := [2]uint64{ud_and[0] & lr_or[0], ud_and[1] & lr_or[1]}
	threes_lr_and_and_ud_or := [2]uint64{lr_and[0] & ud_or[0], lr_and[1] & ud_or[1]}
	threes := [2]uint64{threes_ud_and_and_lr_or[0] | threes_lr_and_and_ud_or[0], threes_ud_and_and_lr_or[1] | threes_lr_and_and_ud_or[1]}

	twos_ud_and_or_lr_and := [2]uint64{ud_and[0] | lr_and[0], ud_and[1] | lr_and[1]}
	twos_ud_or_and_lr_or := [2]uint64{ud_or[0] & lr_or[0], ud_or[1] & lr_or[1]}
	twos := [2]uint64{twos_ud_and_or_lr_and[0] | twos_ud_or_and_lr_or[0], twos_ud_and_or_lr_and[1] | twos_ud_or_and_lr_or[1]}

	two_d := [2]uint64{twos[0] >> 1 & twos[0], twos[1] >> 1 & twos[1]}
	two_l := [2]uint64{(twos[0]>>16 | ((twos[1] & 0xffff) << 48)) & twos[0], (twos[1] >> 16) & twos[1]}

	var vanishing FieldBits
	vanishing.m[0] |= threes[0] | two_d[0] | two_l[0]
	vanishing.m[1] |= threes[1] | two_d[1] | two_l[1]
	if vanishing.IsEmpty() {
		return &vanishing
	}

	two_u := [2]uint64{twos[0] << 1 & twos[0], twos[1] << 1 & twos[1]}
	two_r := [2]uint64{(twos[0] << 16) & twos[0], (twos[1]<<16 | twos[0]>>48) & twos[1]}

	vanishing.m[0] |= two_u[0] | two_r[0]
	vanishing.m[1] |= two_u[1] | two_r[1]
	return vanishing.expand1(fb)
}

func (fb *FieldBits) Onebit(x int, y int) uint64 {
	idx := x >> 2
	pos := x&3*16 + y
	return fb.m[idx] & (1 << pos)
}

func (fb *FieldBits) IsEmpty() bool {
	return fb.m[0] == 0 && fb.m[1] == 0
}

func (fb *FieldBits) IterateBitWithMasking(callback func(fb *FieldBits) *FieldBits) {
	current := fb.Clone()
	for current.IsEmpty() == false {
		len := bits.Len64(current.m[0])
		if len > 0 {
			mask := callback(NewFieldBitsWithM([2]uint64{1 << (len - 1), 0}))
			current.AndNot(mask)
		}
		len = bits.Len64(current.m[1])
		if len > 0 {
			mask := callback(NewFieldBitsWithM([2]uint64{0, 1 << (len - 1)}))
			current.AndNot(mask)
		}
	}
}

func (fb *FieldBits) MaskField13() *FieldBits {
	mfb := new(FieldBits)
	mfb.m[0] = fb.m[0] & 0xBFFEBFFEBFFEBFFE
	mfb.m[1] = fb.m[1] & 0xBFFEBFFEBFFEBFFE
	return mfb
}

func (fb *FieldBits) MaskField12() *FieldBits {
	mfb := new(FieldBits)
	mfb.m[0] = fb.m[0] & 0x1FFE1FFE1FFE1FFE
	mfb.m[1] = fb.m[1] & 0x1FFE1FFE1FFE1FFE
	return mfb
}

func (fb *FieldBits) SetOnebit(x int, y int) {
	idx := x >> 2
	pos := x&3*16 + y
	fb.m[idx] |= 1 << pos
}

func (fb *FieldBits) Or(fb2 *FieldBits) *FieldBits {
	return NewFieldBitsWithM([2]uint64{fb.m[0] | fb2.m[0], fb.m[1] | fb2.m[1]})
}

func (fb *FieldBits) PopCount() int {
	return bits.OnesCount64(fb.m[0]) + bits.OnesCount64(fb.m[1])
}

func (fb *FieldBits) ShowDebug() {
	fmt.Println(fb.ToString())
}

func (fb *FieldBits) ToIntArray() [2]uint64 {
	return fb.m
}

func (fb *FieldBits) ToString() string {
	var s string
	for y := 14; y >= 0; y-- {
		s += fmt.Sprintf("%02d: ", y)
		for x := 0; x < 6; x++ {
			idx := x >> 2
			pos := x&3*16 + y
			s += fmt.Sprint(fb.bitChar(int((fb.m[idx] >> pos) & 1)))
		}
		s += fmt.Sprintln()
	}
	return s
}
