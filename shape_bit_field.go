package puyo2

import (
	"fmt"
	"math/bits"
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

func (sbf *ShapeBitField) AddShape(shape *FieldBits) {
	sbf.Shapes = append(sbf.Shapes, shape)
	sbf.originalShapes = append(sbf.originalShapes, shape.Clone())
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

func (sbf *ShapeBitField) ShapeCount() int {
	return len(sbf.Shapes)
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
	return vfbn
}

func (sbf *ShapeBitField) Simulate() *ShapeRensaResult {
	result := NewShapeRensaResult()
	for {
		vfbn := sbf.Simulate1()
		if len(vfbn) == 0 {
			break
		}
		chainShapes := []*FieldBits{}
		for _, n := range vfbn {
			chainShapes = append(chainShapes, sbf.originalShapes[n])
		}
		sbf.ChainOrderedShapes = append(sbf.ChainOrderedShapes, chainShapes)
		result.AddChain()
	}
	result.SetShapeBitField(sbf)
	return result
}

func (sbf *ShapeBitField) ShowDebug() {
	var s string
	for y := 14; y >= 0; y-- {
		s += fmt.Sprintf("%02d: ", y)
		for x := 0; x < 6; x++ {
			e := true
			for i, shape := range sbf.Shapes {
				if shape.Onebit(x, y) > 0 {
					s += fmt.Sprint(i)
					e = false
					break
				}
			}
			if e {
				s += "."
			}
		}
		s += fmt.Sprintln()
	}
	fmt.Println(s)
}
