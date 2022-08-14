package puyo2

import (
	"fmt"
	"math/bits"
	"strings"
)

type ShapeBitField struct {
	Shapes             []*FieldBits
	originalShapes     []*FieldBits
	ChainOrderedShapes [][]*FieldBits
}

func NewShapeBitField() *ShapeBitField {
	shapeBitField := new(ShapeBitField)
	return shapeBitField
}

func NewShapeBitFieldWithFieldString(param string) *ShapeBitField {
	shapeBitField := new(ShapeBitField)

	max := 0
	for _, c := range param {
		if c == 'a' {
			continue
		}
		n := int(c) - int('0')
		if n > max {
			max = n
		}
	}
	shapes := make([]*FieldBits, max)
	for i := 0; i < max; i++ {
		shapes[i] = NewFieldBits()
	}
	for i, c := range param {
		if c == 'a' {
			continue
		}
		n := int(c) - int('0')
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

func (bf *ShapeBitField) Clone() *ShapeBitField {
	shapeBitField := new(ShapeBitField)
	shapeBitField.Shapes = make([]*FieldBits, len(bf.Shapes))
	for i, shape := range bf.Shapes {
		shapeBitField.Shapes[i] = shape.Clone()
	}
	shapeBitField.originalShapes = make([]*FieldBits, len(bf.originalShapes))
	for i, originalShape := range bf.originalShapes {
		shapeBitField.originalShapes[i] = originalShape.Clone()
	}
	return shapeBitField
}

func (sbf *ShapeBitField) Drop(fb *FieldBits) {
	for _, shape := range sbf.Shapes {
		r0 := Extract(shape.m[0], ^fb.m[0])
		r1 := Extract(shape.m[1], ^fb.m[1])
		var dropmask1 [2]uint64
		for x := 0; x < 6; x++ {
			idx := x >> 2
			vc := bits.OnesCount64(fb.ColBits(x))
			dropmask1[idx] |= bits.RotateLeft64((1<<vc)-1, 14-vc) << (x & 3 * 16)
		}
		shape.m[0] = Deposit(r0, ^dropmask1[0])
		shape.m[1] = Deposit(r1, ^dropmask1[1])
	}
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
			unTarget := NewFieldBitsWithM([2]uint64{^(shiftTarget.m[0] | 1<<(colNum*16)), 0})
			unTarget.m[0] = unTarget.ColBits(colNum)
			unTarget = unTarget.And(shape)
			switch colNum {
			case 0:
				shape.m[0] &= ^uint64(0xffff)
			case 1:
				shape.m[0] &= ^uint64(0xffff0000)
			case 2:
				shape.m[0] &= ^uint64(0xffff00000000)
			case 3:
				shape.m[0] &= ^uint64(0xffff000000000000)
			}

			shape = shape.Or(shift)
			shape = shape.Or(unTarget)
		} else {
			shift := NewFieldBitsWithM([2]uint64{0, shiftCol})
			unTarget := NewFieldBitsWithM([2]uint64{0, ^(shiftTarget.m[1] | 1<<((colNum-4)*16))})
			unTarget.m[1] = unTarget.ColBits(colNum)
			unTarget = unTarget.And(shape)
			switch colNum {
			case 4:
				shape.m[1] &= ^uint64(0xffff)
			case 5:
				shape.m[1] &= ^uint64(0xffff0000)
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
		for y := bits.Len64(col >> (x * 16)); y < 16; y++ {
			shiftTarget.SetOnebit(x, y)
		}
		sbf.shiftCol(fb, shiftTarget, x)
	}
	sbf.AddShape(fb)
}

func (sbf *ShapeBitField) FieldString() string {
	var b strings.Builder
	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			e := true
			for i, shape := range sbf.Shapes {
				if shape.Onebit(x, y) > 0 {
					fmt.Fprint(&b, i+1)
					e = false
					break
				}
			}
			if e {
				fmt.Fprint(&b, "a")
			}
		}
	}
	return b.String()
}

func (sbf *ShapeBitField) ChainOrderedFieldString() string {
	var b strings.Builder
	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			e := true
			for i, shape := range sbf.ChainOrderedShapes {
				if shape[0].Onebit(x, y) > 0 {
					fmt.Fprint(&b, i+1)
					e = false
					break
				}
			}
			if e {
				fmt.Fprint(&b, "a")
			}
		}
	}
	return b.String()
}

func (sbf *ShapeBitField) ShowDebug() {
	var b strings.Builder
	b.Grow(165)
	for y := 14; y >= 0; y-- {
		fmt.Fprintf(&b, "%02d: ", y)
		for x := 0; x < 6; x++ {
			e := true
			for i, shape := range sbf.Shapes {
				if shape.Onebit(x, y) > 0 {
					fmt.Fprint(&b, i)
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
	fmt.Print(b.String())
}

func (sbf *ShapeBitField) ToChainShapes() []*FieldBits {
	csbf := sbf.Clone()
	shapes := make([]*FieldBits, 0)
	vfbn := csbf.FindVanishingFieldBitsNum()
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
