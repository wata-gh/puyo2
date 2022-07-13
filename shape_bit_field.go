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
		y := 14 - i/6

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

func (sbf *ShapeBitField) FieldString() string {
	var b strings.Builder
	for y := 14; y > 0; y-- {
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
	for y := 14; y > 0; y-- {
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
