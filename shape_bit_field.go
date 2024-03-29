package puyo2

import (
	"fmt"
	"math/bits"
	"os"
	"strings"
)

type ShapeBitField struct {
	Shapes             []*FieldBits
	originalShapes     []*FieldBits
	ChainOrderedShapes [][]*FieldBits
	keyString          *string
}

func NewShapeBitField() *ShapeBitField {
	shapeBitField := new(ShapeBitField)
	return shapeBitField
}

var letter2Num = map[rune]int{
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'a': 10,
	'b': 11,
	'c': 12,
	'd': 13,
	'e': 14,
	'f': 15,
	'g': 16,
	'h': 17,
	'i': 18,
	'j': 19,
	'k': 20,
}

var num2Letter = map[int]string{
	1:  "1",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "a",
	11: "b",
	12: "c",
	13: "d",
	14: "e",
	15: "f",
	16: "g",
	17: "h",
	18: "i",
	19: "j",
	20: "k",
}

func NewShapeBitFieldWithFieldString(param string) *ShapeBitField {
	shapeBitField := new(ShapeBitField)

	max := 0
	for _, c := range param {
		if c == '.' {
			continue
		}
		n := letter2Num[c]
		if n > max {
			max = n
		}
	}
	shapes := make([]*FieldBits, max)
	for i := 0; i < max; i++ {
		shapes[i] = NewFieldBits()
	}
	for i, c := range param {
		if c == '.' {
			continue
		}
		n := letter2Num[c]
		x := i % 6
		y := 13 - i/6

		shapes[n-1].SetOnebit(x, y)
	}

	for _, shape := range shapes {
		shapeBitField.AddShape(shape)
	}

	return shapeBitField
}

func (sbf *ShapeBitField) AddShape(shape *FieldBits) {
	sbf.Shapes = append(sbf.Shapes, shape)
	sbf.originalShapes = append(sbf.originalShapes, shape.Clone())
}

func (sbf *ShapeBitField) Clone() *ShapeBitField {
	shapeBitField := new(ShapeBitField)
	shapeBitField.Shapes = make([]*FieldBits, len(sbf.Shapes))
	for i, shape := range sbf.Shapes {
		shapeBitField.Shapes[i] = shape.Clone()
	}
	shapeBitField.originalShapes = make([]*FieldBits, len(sbf.originalShapes))
	for i, originalShape := range sbf.originalShapes {
		shapeBitField.originalShapes[i] = originalShape.Clone()
	}
	return shapeBitField
}

func (sbf *ShapeBitField) Drop(fb *FieldBits) {
	for _, shape := range sbf.Shapes {
		r0 := Extract(shape.M[0], ^fb.M[0])
		r1 := Extract(shape.M[1], ^fb.M[1])
		var dropmask1 [2]uint64
		for x := 0; x < 6; x++ {
			idx := x >> 2
			vc := bits.OnesCount64(fb.ColBits(x))
			dropmask1[idx] |= bits.RotateLeft64((1<<vc)-1, 14-vc) << (x & 3 * 16)
		}
		shape.M[0] = Deposit(r0, ^dropmask1[0])
		shape.M[1] = Deposit(r1, ^dropmask1[1])
	}
}

func (sbf *ShapeBitField) findPossibilities(colors []Color, num int) []Color {
	if colors[num] != Empty {
		return []Color{colors[num]}
	}
	palet := map[Color]struct{}{}
	order := []Color{Red, Blue, Yellow, Green}
	for _, c := range order {
		palet[c] = struct{}{}
	}
	for i, shape := range sbf.Shapes {
		if i == num {
			continue
		}
		ex := sbf.Shapes[num].Expand(shape)
		if ex.IsEmpty() == false && ex.FindVanishingBits().IsEmpty() == false {
			delete(palet, colors[i])
		}
	}
	possibilities := []Color{}
	for _, c := range order {
		if _, ok := palet[c]; ok {
			possibilities = append(possibilities, c)
		}
	}
	return possibilities
}

func (sbf *ShapeBitField) findColor(targets []Color, idx int) *BitField {
	if idx == sbf.ShapeCount() {
		sbfArray := sbf.ToChainShapesUInt64Array()
		overall := sbf.OverallShape()

		bf := NewBitField()
		for x := 0; x < 6; x++ {
			for y := 0; y < 14; y++ {
				if overall.Onebit(x, y) > 0 {
					for i, shape := range sbf.Shapes {
						if shape.Onebit(x, y) > 0 {
							bf.SetColor(targets[i], x, y)
							break
						}
					}
				}
			}
		}
		cbf := bf.Clone()
		result := cbf.Simulate()
		if result.Chains == sbf.ShapeCount() {
			sameChain := true
			for i, array := range bf.ToChainShapesUInt64Array() {
				if sbfArray[i][0] != array[0] || sbfArray[i][1] != array[1] {
					sameChain = false
					break
				}
			}
			if sameChain {
				return bf
			} else {
				fmt.Fprintf(os.Stderr, "Whats!!!!!!!!! %s", bf.MattulwanEditorParam())
				return nil
			}
		}
		return nil
	}

	colors := sbf.findPossibilities(targets, idx)
	for _, c := range colors {
		newTargets := make([]Color, len(targets))
		copy(newTargets, targets)
		newTargets[idx] = c
		bf := sbf.findColor(newTargets, idx+1)
		if bf != nil {
			return bf
		}
	}
	return nil
}

func (sbf *ShapeBitField) FillChainableColor() *BitField {
	targets := make([]Color, sbf.ShapeCount())
	targets[0] = Red
	return sbf.findColor(targets, 0)
}

func (sbf *ShapeBitField) FindVanishingFieldBitsNum() []int {
	vfbn := []int{}
	for i, shape := range sbf.Shapes {
		vs := shape.MaskField12().FindVanishingBits()
		if !vs.IsEmpty() {
			vfbn = append(vfbn, i)
		}
	}
	return vfbn
}

func (sbf *ShapeBitField) IsEmpty() bool {
	for _, shape := range sbf.Shapes {
		if shape.IsEmpty() == false {
			return false
		}
	}
	return true
}

func (sbf *ShapeBitField) OriginalOverallShape() *FieldBits {
	fb := NewFieldBits()
	for _, shape := range sbf.originalShapes {
		fb = fb.Or(shape)
	}
	return fb
}

func (sbf *ShapeBitField) OverallShape() *FieldBits {
	fb := NewFieldBits()
	for _, shape := range sbf.Shapes {
		fb = fb.Or(shape)
	}
	return fb
}

func (sbf *ShapeBitField) ShapeCount() int {
	return len(sbf.Shapes)
}

func (sbf *ShapeBitField) ShapeNum(x int, y int) int {
	for i, shape := range sbf.Shapes {
		if shape.Onebit(x, y) > 0 {
			return i
		}
	}
	return -1
}

func (sbf *ShapeBitField) Simulate1() []int {
	vfbn := sbf.FindVanishingFieldBitsNum()
	v := NewFieldBits()
	for _, n := range vfbn {
		vf := sbf.Shapes[n]
		v = v.Or(vf.MaskField12().FindVanishingBits())
	}
	if v.IsEmpty() {
		return []int{}
	}
	sbf.Drop(v)

	chainShapes := []*FieldBits{}
	for _, n := range vfbn {
		chainShapes = append(chainShapes, sbf.originalShapes[n])
	}
	sbf.ChainOrderedShapes = append(sbf.ChainOrderedShapes, chainShapes)
	return vfbn
}

func (sbf *ShapeBitField) Simulate() *ShapeRensaResult {
	result := NewShapeRensaResult()
	for {
		vfbn := sbf.Simulate1()
		if len(vfbn) == 0 {
			break
		}
		result.AddChain()
	}
	result.SetShapeBitField(sbf)
	return result
}

func (sbf *ShapeBitField) shiftCol(insertShape *FieldBits, shiftTarget *FieldBits, colNum int) {
	for i, shape := range sbf.Shapes {
		and := shape.And(shiftTarget)

		shiftBits := 0
		colBits := and.ColBits(colNum)
		if colBits == 0 {
			continue
		}

		shiftBits = bits.OnesCount(uint(insertShape.ColBits(colNum)))
		shiftCol := colBits << shiftBits
		if colNum < 4 {
			shift := NewFieldBitsWithM([2]uint64{shiftCol, 0})
			unTarget := NewFieldBitsWithM([2]uint64{^(shiftTarget.M[0] | 1<<(colNum*16)), 0})
			unTarget.M[0] = unTarget.ColBits(colNum)
			unTarget = unTarget.And(shape)
			switch colNum {
			case 0:
				shape.M[0] &= ^uint64(0xffff)
			case 1:
				shape.M[0] &= ^uint64(0xffff0000)
			case 2:
				shape.M[0] &= ^uint64(0xffff00000000)
			case 3:
				shape.M[0] &= ^uint64(0xffff000000000000)
			}

			shape = shape.Or(shift)
			shape = shape.Or(unTarget)
		} else {
			shift := NewFieldBitsWithM([2]uint64{0, shiftCol})
			unTarget := NewFieldBitsWithM([2]uint64{0, ^(shiftTarget.M[1] | 1<<((colNum-4)*16))})
			unTarget.M[1] = unTarget.ColBits(colNum)
			unTarget = unTarget.And(shape)
			switch colNum {
			case 4:
				shape.M[1] &= ^uint64(0xffff)
			case 5:
				shape.M[1] &= ^uint64(0xffff0000)
			}

			shape = shape.Or(shift)
			shape = shape.Or(unTarget)

		}
		sbf.Shapes[i] = shape
		sbf.originalShapes[i] = shape.Clone()
	}
}

func (sbf *ShapeBitField) InsertShape(fb *FieldBits) {
	overall := sbf.OverallShape()
	and := overall.And(fb)
	for x := 0; x < 6; x++ {
		col := and.ColBits(x)
		if col == 0 {
			continue
		}
		shiftTarget := fb.Clone()
		sb := x * 16
		if x > 3 {
			sb = (x - 4) * 16
		}
		for y := bits.Len64(col >> sb); y < 16; y++ {
			shiftTarget.SetOnebit(x, y)
		}
		sbf.shiftCol(fb, shiftTarget, x)
	}
	sbf.AddShape(fb)
}

func (sbf *ShapeBitField) Expand3PuyoShapes() *ShapeBitField {
	csbf := sbf.Clone()
	for i, shape := range csbf.Shapes {
		if shape.PopCount() == 3 {
			overall := csbf.OverallShape()
			overall.M[0] = ^overall.M[0]
			overall.M[1] = ^overall.M[1]
			overall = overall.MaskField13()
			expand := shape.Expand1(overall)
			csbf.Shapes[i] = shape.Or(expand)
		}
	}
	return csbf
}

func (sbf *ShapeBitField) KeyString() string {
	if sbf.keyString != nil {
		return *sbf.keyString
	}

	var b = strings.Builder{}
	for _, shape := range sbf.Shapes {
		fmt.Fprintf(&b, "_%x:%x", shape.M[0], shape.M[1])
	}
	ks := b.String()
	sbf.keyString = &ks
	return ks

}

func (sbf *ShapeBitField) FieldString() string {
	buf := []string{".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", ".", "."}
	overall := sbf.OverallShape()
	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			if overall.Onebit(x, y) == 0 {
				continue
			}
			idx := (13-y)*6 + x
			for i, shape := range sbf.Shapes {
				if shape.Onebit(x, y) > 0 {
					buf[idx] = num2Letter[i+1]
					break
				}
			}
		}
	}
	return strings.Join(buf, "")
}

func (sbf *ShapeBitField) ChainOrderedFieldString() string {
	shapes := []*FieldBits{}
	for _, shape := range sbf.ChainOrderedShapes {
		shapes = append(shapes, shape[0])
	}
	for i, shape := range sbf.Shapes {
		if shape.IsEmpty() == false {
			shapes = append(shapes, sbf.originalShapes[i])
		}
	}
	var b strings.Builder
	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			e := true
			for i, shape := range shapes {
				if shape.Onebit(x, y) > 0 {
					fmt.Fprint(&b, num2Letter[i+1])
					e = false
					break
				}
			}
			if e {
				fmt.Fprint(&b, ".")
			}
		}
	}
	return b.String()
}

func (sbf *ShapeBitField) ShowDebug() {
	fmt.Print(sbf.ToString())
}

func (sbf *ShapeBitField) ToString() string {
	var b strings.Builder
	b.Grow(165)
	for y := 14; y > 0; y-- {
		fmt.Fprintf(&b, "%02d: ", y)
		for x := 0; x < 6; x++ {
			e := true
			for i, shape := range sbf.Shapes {
				if shape.Onebit(x, y) > 0 {
					fmt.Fprint(&b, num2Letter[i+1])
					e = false
					break
				}
			}
			if e {
				fmt.Fprint(&b, ".")
			}
		}
		fmt.Fprintln(&b)
	}
	return b.String()
}

func (sbf *ShapeBitField) ToChainShapes() []*FieldBits {
	csbf := sbf.Clone()
	shapes := make([]*FieldBits, 0)
	vfbn := csbf.FindVanishingFieldBitsNum()
	if len(vfbn) == 0 {
		return shapes
	}
	shapes = append(shapes, sbf.Shapes[vfbn[0]])
	for {
		csbf.Simulate1()
		vfbn := csbf.FindVanishingFieldBitsNum()
		if len(vfbn) != 0 {
			shapes = append(shapes, csbf.Shapes[vfbn[0]].Clone())
		} else {
			break
		}
	}
	return shapes
}

func (sbf *ShapeBitField) ToChainShapesUInt64Array() [][2]uint64 {
	shapes := sbf.ToChainShapes()
	array := make([][2]uint64, 0, len(shapes))
	for _, v := range shapes {
		array = append(array, v.ToIntArray())
	}
	return array
}
