package puyo2

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/bits"
	"os"
)

type BitField struct {
	m [3][2]uint64
}

func NewBitField() *BitField {
	bitField := new(BitField)
	return bitField
}

func NewBitFieldWithMattulwan(field string) *BitField {
	bf := new(BitField)

	for i, c := range ExpandMattulwanParam(field) {
		x := i % 6
		y := 13 - i/6
		switch c {
		case 'b':
			bf.SetColor(Red, x, y)
		case 'c':
			bf.SetColor(Blue, x, y)
		case 'd':
			bf.SetColor(Yellow, x, y)
		case 'e':
			bf.SetColor(Green, x, y)
		case 'g':
			bf.SetColor(Ojama, x, y)
		}
	}
	return bf
}

func (bf *BitField) colorBits(c Color) []uint64 {
	switch c {
	case Empty:
		return []uint64{0, 0, 0}
	case Ojama:
		return []uint64{1, 0, 0}
	case Wall:
		return []uint64{0, 1, 0}
	case Iron:
		return []uint64{1, 1, 0}
	case Red:
		return []uint64{0, 0, 1}
	case Blue:
		return []uint64{1, 0, 1}
	case Yellow:
		return []uint64{0, 1, 1}
	case Green:
		return []uint64{1, 1, 1}
	}
	panic(fmt.Sprintf("Color must be valid. passed %d", c))
}

func (bf *BitField) colorChar(c Color) string {
	switch c {
	case Empty:
		return "."
	case Ojama:
		return "O"
	case Wall:
		return "W"
	case Iron:
		return "I"
	case Red:
		return "R"
	case Blue:
		return "B"
	case Yellow:
		return "Y"
	case Green:
		return "G"
	default:
		return "U"
	}
}

func (bf *BitField) drawPuyo(c Color, x int, y int, puyo *image.Image, out *image.NRGBA) {
	ix := 0
	iy := 0
	u, d, r, l := false, false, false, false
	if bf.Color(x, y+1) == c {
		u = true
	}
	if bf.Color(x, y-1) == c {
		d = true
	}
	if x > 0 && bf.Color(x-1, y) == c {
		l = true
	}
	if x < 5 && bf.Color(x+1, y) == c {
		r = true
	}
	if u == false && d == false && l == false && r == false {
		iy = 0
	} else if u && d == false && l == false && r == false {
		iy = 32
	} else if u == false && d && l == false && r == false {
		iy = 64
	} else if u && d && l == false && r == false {
		iy = 96
	} else if u == false && d == false && l && r == false {
		iy = 128
	} else if u && d == false && l && r == false {
		iy = 160
	} else if u == false && d && l && r == false {
		iy = 192
	} else if u && d && l && r == false {
		iy = 224
	} else if u == false && d == false && l == false && r {
		iy = 256
	} else if u && d == false && l == false && r {
		iy = 288
	} else if u == false && d && l == false && r {
		iy = 320
	} else if u && d && l == false && r {
		iy = 352
	} else if u == false && d == false && l && r {
		iy = 384
	} else if u && d == false && l && r {
		iy = 416
	} else if u && d == false && l && r {
		iy = 448
	} else {
		iy = 480
	}
	switch c {
	case Red:
		ix = 0
	case Blue:
		ix = 64
	case Green:
		ix = 32
	case Yellow:
		ix = 96
	case Empty:
		return
	}
	point := image.Pt((x+2)*32, (14-y)*32)
	draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, *puyo, image.Pt(ix, iy), draw.Over)
}

func (bf *BitField) Bits(c Color) *FieldBits {
	switch c {
	case Empty:
		return NewFieldBitsWithM([2]uint64{^(bf.m[0][0] | bf.m[1][0] | bf.m[2][0]), ^(bf.m[0][1] | bf.m[1][1] | bf.m[2][1])})
	case Ojama:
		return NewFieldBitsWithM([2]uint64{bf.m[0][0] &^ (bf.m[1][0] | bf.m[2][0]), bf.m[0][1] &^ (bf.m[1][1] | bf.m[2][1])})
	case Wall:
		return NewFieldBitsWithM([2]uint64{bf.m[1][0] &^ (bf.m[0][0] | bf.m[2][0]), bf.m[1][1] &^ (bf.m[0][1] | bf.m[2][1])})
	case Iron:
		return NewFieldBitsWithM([2]uint64{bf.m[0][0] & bf.m[1][0] &^ bf.m[2][0], bf.m[0][1] & bf.m[1][1] &^ bf.m[2][1]})
	case Red:
		return NewFieldBitsWithM([2]uint64{bf.m[2][0] &^ (bf.m[0][0] | bf.m[1][0]), bf.m[2][1] &^ (bf.m[0][1] | bf.m[1][1])})
	case Blue:
		return NewFieldBitsWithM([2]uint64{bf.m[2][0] & bf.m[0][0] &^ bf.m[1][0], bf.m[2][1] & bf.m[0][1] &^ bf.m[1][1]})
	case Yellow:
		return NewFieldBitsWithM([2]uint64{bf.m[2][0] & bf.m[1][0] &^ bf.m[0][0], bf.m[2][1] & bf.m[1][1] &^ bf.m[0][1]})
	case Green:
		return NewFieldBitsWithM([2]uint64{bf.m[2][0] & bf.m[1][0] & bf.m[0][0], bf.m[2][1] & bf.m[1][1] & bf.m[0][1]})
	}
	panic(fmt.Sprintf("Color must be valid. passed %d", c))
}

func (bf *BitField) Clone() *BitField {
	bitField := new(BitField)
	bitField.m[0] = bf.m[0]
	bitField.m[1] = bf.m[1]
	bitField.m[2] = bf.m[2]
	return bitField
}

func (bf *BitField) Color(x int, y int) Color {
	idx := x >> 2
	pos := x&3*16 + y
	c := ((bf.m[0][idx] >> pos) & 1) + ((bf.m[1][idx] >> pos) & 1 << 1) + ((bf.m[2][idx] >> pos) & 1 << 2)
	return Color(c)
}

func (bf *BitField) Drop(fb *FieldBits) {
	for i := 0; i < len(bf.m); i++ {
		r0 := Extract(bf.m[i][0], ^fb.m[0])
		r1 := Extract(bf.m[i][1], ^fb.m[1])
		var dropmask1 [2]uint64
		for x := 0; x < 6; x++ {
			idx := x >> 2
			vc := bits.OnesCount64(fb.ColBits(x))
			dropmask1[idx] |= bits.RotateLeft64((1<<vc)-1, 14-vc) << (x & 3 * 16)
		}
		bf.m[i][0] = Deposit(r0, ^dropmask1[0])
		bf.m[i][1] = Deposit(r1, ^dropmask1[1])
	}
}

func (bf *BitField) Equals(bf2 *BitField) bool {
	return bf.m[0] == bf2.m[0] && bf.m[1] == bf2.m[1] && bf.m[2] == bf2.m[2]
}

func (bf *BitField) EqualChain(bf2 *BitField) bool {
	cs := bf.ToChainShapesUInt64Array()
	cs2 := bf2.ToChainShapesUInt64Array()
	if len(cs) != len(cs2) {
		return false
	}
	for i, v := range cs {
		v2 := cs2[i]
		if v[0] != v2[0] || v[1] != v2[1] {
			return false
		}
	}
	return true
}

func (bf *BitField) ExportImage(name string) {
	fpuyo, _ := os.Open("images/puyos.png")
	defer fpuyo.Close()

	puyo, _, _ := image.Decode(fpuyo)
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*14)})

	for y := 13; y >= 0; y-- {
		draw.Draw(out, image.Rectangle{image.Pt(0, (13-y)*32), image.Pt(32, (14-y)*32)}, puyo, image.Pt(5*32, 0), draw.Src)
		draw.Draw(out, image.Rectangle{image.Pt(32*7, (13-y)*32), image.Pt(32*8, (14-y)*32)}, puyo, image.Pt(5*32, 0), draw.Src)
		for x := 0; x < 8; x++ {
			// block
			if (x == 0 || x == 7 || y == 0) && y != 13 {
				draw.Draw(out, image.Rectangle{image.Pt(x*32, 13*32), image.Pt((x+1)*32, 14*32)}, puyo, image.Pt(5*32, 0), draw.Src)
			} else { // background
				draw.Draw(out, image.Rectangle{image.Pt(x*32, (13-y)*32), image.Pt((x+1)*32, (14-y)*32)}, puyo, image.Pt(5*32, 32), draw.Src)
			}
		}
	}
	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			c := bf.Color(x, y)
			if c == Empty {
				continue
			}
			if c == Ojama {
				point := image.Pt((x+2)*32, (14-y)*32)
				draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, puyo, image.Pt(5*32, 2*32), draw.Over)
			} else {
				bf.drawPuyo(c, x, y, &puyo, out)
			}
		}
	}

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}

func (bf *BitField) ExportImageWithTransparent(name string, trans *FieldBits) {
	fpuyo, _ := os.Open("images/puyos.png")
	fpuyot, _ := os.Open("images/puyos_transparent.png")
	defer fpuyo.Close()
	defer fpuyot.Close()

	puyo, _, _ := image.Decode(fpuyo)
	puyot, _, _ := image.Decode(fpuyot)
	var dpuyo image.Image
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*14)})

	for y := 13; y >= 0; y-- {
		draw.Draw(out, image.Rectangle{image.Pt(0, (13-y)*32), image.Pt(32, (14-y)*32)}, puyo, image.Pt(5*32, 0), draw.Src)
		draw.Draw(out, image.Rectangle{image.Pt(32*7, (13-y)*32), image.Pt(32*8, (14-y)*32)}, puyo, image.Pt(5*32, 0), draw.Src)
		for x := 0; x < 8; x++ {
			// block
			if (x == 0 || x == 7 || y == 0) && y != 13 {
				draw.Draw(out, image.Rectangle{image.Pt(x*32, 13*32), image.Pt((x+1)*32, 14*32)}, puyo, image.Pt(5*32, 0), draw.Src)
			} else { // background
				draw.Draw(out, image.Rectangle{image.Pt(x*32, (13-y)*32), image.Pt((x+1)*32, (14-y)*32)}, puyo, image.Pt(5*32, 32), draw.Src)
			}
		}
	}
	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			c := bf.Color(x, y)
			if c == Empty {
				continue
			}
			if c == Ojama {
				point := image.Pt((x+2)*32, (14-y)*32)
				draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, puyo, image.Pt(5*32, 2*32), draw.Over)
			} else {
				if trans.Onebit(x, y) == 0 {
					dpuyo = puyo
				} else {
					dpuyo = puyot
				}
				bf.drawPuyo(c, x, y, &dpuyo, out)
			}
		}
	}

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}

func (bf *BitField) FindVanishingBits() *FieldBits {
	return bf.Bits(Green).FindVanishingBits().Or(bf.Bits(Red).FindVanishingBits()).Or(bf.Bits(Yellow).FindVanishingBits()).Or(bf.Bits(Blue).FindVanishingBits())
}

func (bf *BitField) IsEmpty() bool {
	return bf.m[0][0] == 0 && bf.m[1][0] == 0 && bf.m[2][0] == 0 && bf.m[0][1] == 0 && bf.m[1][1] == 0 && bf.m[2][1] == 0
}

func (bf *BitField) MattulwanEditorParam() string {
	table := map[string]string{"R": "b", "B": "c", "Y": "d", "G": "e", ".": "a"}
	var s string
	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			s += fmt.Sprint(table[bf.colorChar(bf.Color(x, y))])
		}
	}
	return s
}

func (bf *BitField) MattulwanEditorUrl() string {
	return "https://pndsng.com/puyo/index.html?" + bf.MattulwanEditorParam()
}

func (bf *BitField) Normalize() *BitField {
	ns := ""
	table := map[rune]string{}
	colors := []string{"b", "c", "d", "e", "f"}
	for _, c := range bf.MattulwanEditorParam() {
		if c == 'a' {
			ns += "a"
		} else {
			v, e := table[c]
			if e {
				ns += v
			} else {
				cv := colors[0]
				table[c] = cv
				ns += cv
				colors = colors[1:]
			}
		}
	}
	return NewBitFieldWithMattulwan(ns)
}

func (bf *BitField) SetColor(c Color, x int, y int) {
	b := bf.colorBits(c)
	pos := x&3*16 + y
	idx := x >> 2
	for i := 0; i < len(bf.m); i++ {
		if b[i] == 1 {
			bf.m[i][idx] |= b[i] << pos
		} else {
			posbit := uint64(1) << pos
			if bf.m[i][idx]&posbit > 0 {
				bf.m[i][idx] -= posbit
			}
		}
	}
}

func (bf *BitField) SetColorWithFieldBits(c Color, fb *FieldBits) {
	b := bf.colorBits(c)
	mask := [2]uint64{^fb.m[0], ^fb.m[1]}
	for i := 0; i < len(fb.m); i++ {
		bf.m[i][0] &= mask[0]
		bf.m[i][1] &= mask[1]

		if b[i] == 1 {
			bf.m[i][0] |= fb.m[0]
			bf.m[i][1] |= fb.m[1]
		}
	}
}

func (bf *BitField) Simulate() *RensaResult {
	result := NewRensaResult()
	for bf.Simulate1() {
		result.AddChain()
	}
	result.SetBitField(bf)
	return result
}

func (bf *BitField) SimulateDetail() *RensaResult {
	result := NewRensaResult()
	for {
		numColors := 0
		numErased := 0
		longBonusCoef := 0
		vanished := NewFieldBits()

		for _, color := range []Color{Red, Blue, Yellow, Green} {
			vb := bf.Bits(color).FindVanishingBits()
			if vb.IsEmpty() == false {
				numColors++
				popCount := vb.PopCount()
				numErased += popCount
				vanished = vanished.Or(vb)

				popcount := vb.PopCount()
				if popcount <= 7 {
					longBonusCoef += LongBonus(popCount)
					continue
				}

				vb.IterateBitWithMasking(func(c *FieldBits) *FieldBits {
					v := c.expand(vb)
					longBonusCoef += LongBonus(c.PopCount())
					return v
				})
			}
		}

		if numColors == 0 {
			break
		}

		// TODO 同色でマルチの場合の判定
		// 参考： https://github.com/puyoai/puyoai/blob/575068dffab021cd42b0178699d480a743ca4e1f/src/core/bit_field_inl.h#L122-L127
		result.AddChain()
		colorBonusCoef := ColorBonus(numColors)
		coef := CalcRensaBonusCoef(RensaBonus(result.Chains), longBonusCoef, colorBonusCoef)
		result.AddScore(10 * numErased * coef)
		bf.Drop(vanished)
	}
	result.SetBitField(bf)
	return result
}

func (bf *BitField) SimulateWithNewBitField() *RensaResult {
	return bf.Clone().Simulate()
}

func (bf *BitField) Simulate1() bool {
	// TODO お邪魔処理
	v := bf.FindVanishingBits()
	if v.IsEmpty() {
		return false
	}
	bf.Drop(v)
	return true
}

func (bf *BitField) ToChainShapes() []*FieldBits {
	cbf := bf.Clone()
	shapes := make([]*FieldBits, 0)
	v := cbf.FindVanishingBits()
	shapes = append(shapes, v)
	for cbf.Simulate1() {
		v := cbf.FindVanishingBits()
		if v.IsEmpty() == false {
			shapes = append(shapes, v)
		} else {
			break
		}
	}
	return shapes
}

func (bf *BitField) ToChainShapesUInt64Array() [][2]uint64 {
	shapes := bf.ToChainShapes()
	array := make([][2]uint64, 0, len(shapes))
	for _, v := range shapes {
		array = append(array, v.ToIntArray())
	}
	return array
}

func (bf *BitField) ToString() string {
	var s string
	for y := 14; y > 0; y-- {
		s += fmt.Sprintf("%02d: ", y)
		for x := 0; x < 6; x++ {
			s += fmt.Sprint(bf.colorChar(bf.Color(x, y)))
		}
		s += fmt.Sprintln()
	}
	return s
}

func (bf *BitField) ShowDebug() {
	fmt.Print(bf.ToString())
}
