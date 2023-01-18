package puyo2

import (
	"fmt"
	"math/bits"
)

var Rune2ColorTable = map[rune]Color{
	'a': Empty,
	'b': Red,
	'c': Blue,
	'd': Yellow,
	'e': Green,
	'f': Purple,
	'g': Ojama,
}

type BitField struct {
	M      [3][2]uint64
	Table  map[Color]Color
	colors []Color
}

func NewBitField() *BitField {
	bitField := new(BitField)
	bitField.colors = []Color{Red, Blue, Yellow, Green}
	return bitField
}

func NewBitFieldWithColors(colors []Color) *BitField {
	bitField := new(BitField)
	hasPurple := false
	for _, c := range colors {
		if c == Purple {
			hasPurple = true
			break
		}
	}
	table := map[Color]Color{}
	if hasPurple {
		for _, c := range []Color{Red, Blue, Yellow, Green} {
			found := false
			for _, c2 := range colors {
				if c2 == Purple {
					continue
				}
				if c == c2 {
					table[c] = c
					found = true
					break
				}
			}
			if found == false {
				table[Purple] = c
				table[c] = Purple
			}
		}
	} else {
		for _, c := range []Color{Red, Blue, Yellow, Green} {
			table[c] = c
		}
	}
	bitField.Table = table
	bitField.resetColor()
	return bitField
}

func NewBitFieldWithTable(table map[Color]Color) *BitField {
	bitField := new(BitField)
	bitField.Table = table
	bitField.resetColor()
	return bitField
}

func NewBitFieldWithTableAndColors(table map[Color]Color, colors []Color) *BitField {
	bitField := new(BitField)
	bitField.Table = table
	bitField.colors = colors
	return bitField
}

func NewBitFieldWithM(m [3][2]uint64) *BitField {
	bitField := new(BitField)
	bitField.M = m
	bitField.colors = []Color{Red, Blue, Yellow, Green}
	return bitField
}

func NewBitFieldWithMattulwan(field string) *BitField {
	bf := new(BitField)
	bf.colors = []Color{Red, Blue, Yellow, Green}

	for i, c := range ExpandMattulwanParam(field) {
		x := i % 6
		y := 13 - i/6
		switch c {
		case 'a':
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
		default:
			panic("only supports puyo color a,b,c,d,e,g")
		}
	}
	return bf
}

func NewBitFieldWithMattulwanC(field string) *BitField {
	bf := new(BitField)
	bf.SetMattulwanParam(field)
	return bf
}

func NewBitFieldWithMattulwanCAndHaipuyo(field string, haipuyo string) *BitField {
	bf := new(BitField)
	bf.SetMattulwanParam(field)
	tableColors := bf.tableColors(bf.Table)
	puyoSets := Haipuyo2PuyoSets(haipuyo)
	_, colors := bf.createTableAndColors(puyoSets)
	bf.mergeTableColors(tableColors, colors)
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
	case Purple:
		return "P"
	default:
		return "U"
	}
}

func (bf *BitField) converColor(puyo Color) Color {
	if bf.Table != nil && puyo != Empty && puyo != Ojama && puyo != Iron && puyo != Wall {
		c, ok := bf.Table[puyo]
		if ok == false {
			panic(fmt.Sprintf("conver table is not nil, but can not convert color. %v %v", puyo, bf.Table))
		}
		return c
	}
	return puyo
}

func (bf *BitField) createTableAndColors(puyoSets []*PuyoSet) (map[Color]Color, []Color) {
	colors := []Color{}
	table := map[Color]Color{
		Red:    Empty,
		Blue:   Empty,
		Green:  Empty,
		Yellow: Empty,
		Purple: Empty,
	}

	for _, puyoSet := range puyoSets {
		table[puyoSet.Axis] = puyoSet.Axis
		table[puyoSet.Child] = puyoSet.Child
	}
	// contains purple
	if table[Purple] == Purple {
		for _, c := range []Color{Red, Blue, Yellow, Green} {
			if table[c] == Empty {
				table[c] = Purple
				table[Purple] = c
			} else {
				colors = append(colors, c)
			}
		}
		colors = append(colors, Purple)
	} else {
		for k, v := range table {
			if v != Empty {
				colors = append(colors, k)
			}
		}
	}
	return table, colors
}

func (bf *BitField) mergeTableColors(tableColors []Color, colors []Color) {
	for _, c := range tableColors {
		found := false
		for _, c2 := range colors {
			if c == c2 {
				found = true
				break
			}
		}
		if found == false {
			colors = append(colors, c)
		}
	}

	ntable := map[Color]Color{
		Red:    Empty,
		Blue:   Empty,
		Green:  Empty,
		Yellow: Empty,
		Purple: Empty,
	}

	for _, c := range colors {
		ntable[c] = c
	}

	// contains purple
	if ntable[Purple] == Purple {
		for _, c := range []Color{Red, Blue, Yellow, Green} {
			if ntable[c] == Empty {
				ntable[c] = Purple
				ntable[Purple] = c
				break
			}
		}
	}
	bf.Table = ntable
	bf.resetColor()
}

func (bf *BitField) resetColor() {
	bf.colors = bf.tableColors(bf.Table)
}

func (bf *BitField) tableColors(table map[Color]Color) []Color {
	colors := []Color{}
	for _, c := range []Color{Red, Blue, Yellow, Green} {
		v := table[c]
		if v == Purple {
			colors = append(colors, Purple)
		} else if c == v {
			colors = append(colors, c)
		}
	}
	return colors
}

func (bf *BitField) Bits(c Color) *FieldBits {
	switch bf.converColor(c) {
	case Empty:
		return NewFieldBitsWithM([2]uint64{^(bf.M[0][0] | bf.M[1][0] | bf.M[2][0]), ^(bf.M[0][1] | bf.M[1][1] | bf.M[2][1])})
	case Ojama:
		return NewFieldBitsWithM([2]uint64{bf.M[0][0] &^ (bf.M[1][0] | bf.M[2][0]), bf.M[0][1] &^ (bf.M[1][1] | bf.M[2][1])})
	case Wall:
		return NewFieldBitsWithM([2]uint64{bf.M[1][0] &^ (bf.M[0][0] | bf.M[2][0]), bf.M[1][1] &^ (bf.M[0][1] | bf.M[2][1])})
	case Iron:
		return NewFieldBitsWithM([2]uint64{bf.M[0][0] & bf.M[1][0] &^ bf.M[2][0], bf.M[0][1] & bf.M[1][1] &^ bf.M[2][1]})
	case Red:
		return NewFieldBitsWithM([2]uint64{bf.M[2][0] &^ (bf.M[0][0] | bf.M[1][0]), bf.M[2][1] &^ (bf.M[0][1] | bf.M[1][1])})
	case Blue:
		return NewFieldBitsWithM([2]uint64{bf.M[2][0] & bf.M[0][0] &^ bf.M[1][0], bf.M[2][1] & bf.M[0][1] &^ bf.M[1][1]})
	case Yellow:
		return NewFieldBitsWithM([2]uint64{bf.M[2][0] & bf.M[1][0] &^ bf.M[0][0], bf.M[2][1] & bf.M[1][1] &^ bf.M[0][1]})
	case Green:
		return NewFieldBitsWithM([2]uint64{bf.M[2][0] & bf.M[1][0] & bf.M[0][0], bf.M[2][1] & bf.M[1][1] & bf.M[0][1]})
	}
	panic(fmt.Sprintf("Color must be valid. passed %d, converted to -> %d, %+v", c, bf.converColor(c), bf.Table))
}

func (bf *BitField) Clone() *BitField {
	bitField := new(BitField)
	bitField.M[0] = bf.M[0]
	bitField.M[1] = bf.M[1]
	bitField.M[2] = bf.M[2]
	bitField.Table = bf.Table
	bitField.colors = bf.colors
	return bitField
}

func (bf *BitField) Color(x int, y int) Color {
	idx := x >> 2
	pos := x&3*16 + y
	c := ((bf.M[0][idx] >> pos) & 1) + ((bf.M[1][idx] >> pos) & 1 << 1) + ((bf.M[2][idx] >> pos) & 1 << 2)
	return bf.converColor(Color(c))
}

func (bf *BitField) Drop(fb *FieldBits) {
	for i := 0; i < len(bf.M); i++ {
		r0 := Extract(bf.M[i][0], ^fb.M[0])
		r1 := Extract(bf.M[i][1], ^fb.M[1])
		var dropmask1 [2]uint64
		for x := 0; x < 6; x++ {
			idx := x >> 2
			vc := bits.OnesCount64(fb.ColBits(x))
			dropmask1[idx] |= bits.RotateLeft64((1<<vc)-1, 14-vc) << (x & 3 * 16)
		}
		bf.M[i][0] = Deposit(r0, ^dropmask1[0])
		bf.M[i][1] = Deposit(r1, ^dropmask1[1])
	}
}

func (bf *BitField) Equals(bf2 *BitField) bool {
	return bf.M[0] == bf2.M[0] && bf.M[1] == bf2.M[1] && bf.M[2] == bf2.M[2]
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

func (bf *BitField) FindVanishingBits() *FieldBits {
	v := NewFieldBits()
	// v := bf.Bits(bf.colors[0]).MaskField12().FindVanishingBits()
	for _, c := range bf.colors {
		v = v.Or(bf.Bits(c).MaskField12().FindVanishingBits())
	}
	// v := bf.Bits(Green).MaskField12().FindVanishingBits().Or(bf.Bits(Red).MaskField12().FindVanishingBits()).Or(bf.Bits(Yellow).MaskField12().FindVanishingBits()).Or(bf.Bits(Blue).MaskField12().FindVanishingBits())
	o := v.Expand1(bf.Bits(Ojama))
	return v.Or(o)
}

func (bf *BitField) FlipHorizontal() *BitField {
	m := [3][2]uint64{{0, 0}, {0, 0}, {0, 0}}
	for i := 0; i < 3; i++ {
		m[i][1] = (bf.M[i][0] & 0xffff) << 16              // 1 -> 6
		m[i][1] |= (bf.M[i][0] & 0xffff0000) >> 16         // 2 -> 5
		m[i][0] = (bf.M[i][0] & 0xffff00000000) << 16      // 3 -> 4
		m[i][0] |= (bf.M[i][0] & 0xffff000000000000) >> 16 // 4 -> 3
		m[i][0] |= (bf.M[i][1] & 0xffff) << 16             // 5 -> 2
		m[i][0] |= (bf.M[i][1] & 0xffff0000) >> 16         // 6 -> 1
	}
	bf.M = m
	return bf
}

func (bf *BitField) IsEmpty() bool {
	return bf.M[0][0] == 0 && bf.M[1][0] == 0 && bf.M[2][0] == 0 && bf.M[0][1] == 0 && bf.M[1][1] == 0 && bf.M[2][1] == 0
}

func (bf *BitField) MattulwanEditorParam() string {
	table := map[string]string{"R": "b", "B": "c", "Y": "d", "G": "e", "P": "f", ".": "a", "O": "g"}
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

func (bf *BitField) PlacePuyoWithPlacement(placement *PuyoSetPlacement) bool {
	// TODO return false if invalid placement.
	bf.SetColor(placement.PuyoSet.Axis, placement.AxisX, placement.AxisY)
	bf.SetColor(placement.PuyoSet.Child, placement.ChildX, placement.ChildY)
	return true
}

func (bf *BitField) SetColor(c Color, x int, y int) {
	if len(bf.colors) < 4 && c != Empty && c != Ojama && c != Iron && c != Wall {
		found := false
		for _, color := range bf.colors {
			if c == color {
				found = true
				break
			}
		}
		if found == false {
			bf.colors = append(bf.colors, c)
			if c == Purple {
				for _, c := range []Color{Red, Blue, Yellow, Green} {
					if bf.Table[c] == Empty {
						bf.Table[c] = Purple
						bf.Table[Purple] = c
						break
					}
				}
			} else {
				bf.Table[c] = c
			}
		}
	}
	b := bf.colorBits(bf.converColor(c))
	pos := x&3*16 + y
	idx := x >> 2
	for i := 0; i < len(bf.M); i++ {
		if b[i] == 1 {
			bf.M[i][idx] |= b[i] << pos
		} else {
			posbit := uint64(1) << pos
			if bf.M[i][idx]&posbit > 0 {
				bf.M[i][idx] -= posbit
			}
		}
	}
}

func (bf *BitField) SetColorWithFieldBits(c Color, fb *FieldBits) {
	b := bf.colorBits(c)
	mask := [2]uint64{^fb.M[0], ^fb.M[1]}
	for i := 0; i < len(fb.M); i++ {
		bf.M[i][0] &= mask[0]
		bf.M[i][1] &= mask[1]

		if b[i] == 1 {
			bf.M[i][0] |= fb.M[0]
			bf.M[i][1] |= fb.M[1]
		}
	}
}

func (bf *BitField) SetMattulwanParam(field string) {
	expandField := ExpandMattulwanParam(field)
	bf.Table = map[Color]Color{
		Red:    Empty,
		Blue:   Empty,
		Green:  Empty,
		Yellow: Empty,
		Purple: Empty,
	}
	colorCnt := 0
	for _, c := range expandField {
		switch c {
		case 'a':
		case 'b':
			if bf.Table[Red] == Empty {
				bf.Table[Red] = Red
				bf.colors = append(bf.colors, Red)
				colorCnt++
			}
		case 'c':
			if bf.Table[Blue] == Empty {
				bf.Table[Blue] = Blue
				bf.colors = append(bf.colors, Blue)
				colorCnt++
			}
		case 'd':
			if bf.Table[Yellow] == Empty {
				bf.Table[Yellow] = Yellow
				bf.colors = append(bf.colors, Yellow)
				colorCnt++
			}
		case 'e':
			if bf.Table[Green] == Empty {
				bf.Table[Green] = Green
				bf.colors = append(bf.colors, Green)
				colorCnt++
			}
		case 'f':
			if bf.Table[Purple] == Empty {
				bf.Table[Purple] = Purple
				bf.colors = append(bf.colors, Purple)
				colorCnt++
			}
		}
		if colorCnt == 4 {
			break
		}
	}

	// contains purple
	if bf.Table[Purple] == Purple {
		for _, c := range []Color{Red, Blue, Yellow, Green} {
			if bf.Table[c] == Empty {
				bf.Table[c] = Purple
				bf.Table[Purple] = c
				break
			}
		}
	}

	for i, c := range expandField {
		x := i % 6
		y := 13 - i/6
		puyo, ok := Rune2ColorTable[c]
		if ok == false {
			fmt.Printf("%c %v\n", c, puyo)
			panic("only supports puyo color a,b,c,d,e,f,g")
		}
		if puyo != Empty {
			bf.SetColor(puyo, x, y)
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
		nth := NewNthResult(result.Chains + 1)

		for _, color := range bf.colors {
			vb := bf.Bits(color).MaskField12().FindVanishingBits()
			if vb.IsEmpty() == false {
				numColors++
				popCount := vb.PopCount()
				numErased += popCount
				vanished = vanished.Or(vb)

				if popCount <= 7 {
					sr := SingleResult{
						Color:     color,
						Connected: popCount,
					}
					nth.ErasedPuyos = append(nth.ErasedPuyos, sr)
					longBonusCoef += LongBonus(popCount)
					continue
				}

				vb.IterateBitWithMasking(func(c *FieldBits) *FieldBits {
					v := c.Expand(vb)
					popCount := v.PopCount()
					sr := SingleResult{
						Color:     color,
						Connected: popCount,
					}
					nth.ErasedPuyos = append(nth.ErasedPuyos, sr)
					longBonusCoef += LongBonus(popCount)
					return v
				})
			}
		}

		if numColors == 0 {
			break
		}

		result.NthResults = append(result.NthResults, nth)
		vanished = vanished.Or(vanished.Expand1(bf.Bits(Ojama)))
		result.AddChain()
		colorBonusCoef := ColorBonus(numColors)
		coef := CalcRensaBonusCoef(RensaBonus(result.Chains), longBonusCoef, colorBonusCoef)
		result.AddErased(numErased)
		result.AddScore(10 * numErased * coef)
		heights := bf.CreateHeights()
		result.Quick = true
		for x, height := range heights {
			col := vanished.ShiftedColBits(x)
			vh := bits.Len64(col) - 1
			if vh != -1 && height > vh {
				result.Quick = false
				break
			}
		}
		bf.Drop(vanished)
	}
	result.SetBitField(bf)
	return result
}

func (bf *BitField) SimulateWithNewBitField() *RensaResult {
	return bf.Clone().Simulate()
}

func (bf *BitField) Simulate1() bool {
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

func (bf *BitField) ToShapeBitField() *ShapeBitField {
	sbf := NewShapeBitField()
	for _, ary := range bf.ToChainShapesUInt64Array() {
		shape := NewFieldBitsWithM(ary)
		sbf.AddShape(shape)
	}
	return sbf
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

func (bf *BitField) TrimLeft() *BitField {
	mv := 0
	if bf.M[2][0]&0xffff == 0 { // row 1
		mv++
		if bf.M[2][0]&0xffff0000 == 0 { // row 2
			mv++
			if bf.M[2][0]&0xffff00000000 == 0 { // row 3
				mv++
				if bf.M[2][0]&0xffff000000000000 == 0 { // row 4
					mv++
					if bf.M[2][1]&0xffff == 0 { // row 5
						mv++
						if bf.M[2][1]&0xffff0000 == 0 { // row 6
							// all clear
							return bf
						}
					}
				}
			}
		}
	} else {
		return bf
	}

	switch mv {
	case 1:
		for i := 0; i < 3; i++ {
			bf.M[i][0] >>= 16
			bf.M[i][0] |= (bf.M[i][1] << 48) & 0xffff000000000000
			bf.M[i][1] >>= 16
		}
	case 2:
		for i := 0; i < 3; i++ {
			bf.M[i][0] >>= 32
			bf.M[i][0] |= ((bf.M[i][1] << 32) & 0xffffffff00000000)
			bf.M[i][1] = 0
		}
	case 3:
		for i := 0; i < 3; i++ {
			bf.M[i][0] >>= 48
			bf.M[i][0] |= (bf.M[i][1] << 16) & 0x0000ffffffff0000
			bf.M[i][1] = 0
		}
	case 4:
		for i := 0; i < 3; i++ {
			bf.M[i][0] = bf.M[i][1]
			bf.M[i][1] = 0
		}
	case 5:
		for i := 0; i < 3; i++ {
			bf.M[i][0] = bf.M[i][1] >> 16
			bf.M[i][1] = 0
		}
	}
	return bf
}

func (bf *BitField) ShowDebug() {
	fmt.Print(bf.ToString())
}
