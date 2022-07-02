package puyo2

import (
	"testing"
)

func TestShapeBitField(t *testing.T) {
	sbf := NewShapeBitField()
	shape := NewFieldBits()
	shape.SetOnebit(0, 1)
	shape.SetOnebit(0, 2)
	shape.SetOnebit(1, 1)
	shape.SetOnebit(1, 3)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(1, 2)
	shape.SetOnebit(2, 1)
	shape.SetOnebit(2, 2)
	shape.SetOnebit(3, 1)
	sbf.AddShape(shape)

	//sbf.ShowDebug()

	// sbf.Simulate1()
	result := sbf.Simulate()
	if result.Chains != 2 {
		t.Fatalf("result.Chain must be 2, but %d.", result.Chains)
	}
}

func TestShapeBitFieldSimulate(t *testing.T) {
	sbf := NewShapeBitField()
	shape := NewFieldBits()
	shape.SetOnebit(0, 2)
	shape.SetOnebit(0, 3)
	shape.SetOnebit(0, 4)
	shape.SetOnebit(1, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(0, 1)
	shape.SetOnebit(1, 1)
	shape.SetOnebit(1, 3)
	shape.SetOnebit(2, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(2, 1)
	shape.SetOnebit(2, 3)
	shape.SetOnebit(3, 2)
	shape.SetOnebit(4, 2)
	shape.SetOnebit(5, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(3, 1)
	shape.SetOnebit(4, 1)
	shape.SetOnebit(4, 3)
	shape.SetOnebit(5, 1)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(3, 3)
	shape.SetOnebit(4, 4)
	shape.SetOnebit(5, 3)
	shape.SetOnebit(5, 4)
	sbf.AddShape(shape)

	sbf.ShowDebug()

	// sbf.Simulate1()
	result := sbf.Simulate()
	if result.Chains != 5 {
		t.Fatalf("result.Chain must be 5, but %d.", result.Chains)
	}
}

/*
func TestShapeBitFieldExport(t *testing.T) {
	sbf := NewShapeBitField()
	shape := NewFieldBits()
	shape.SetOnebit(0, 2)
	shape.SetOnebit(0, 3)
	shape.SetOnebit(0, 4)
	shape.SetOnebit(1, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(2, 1)
	shape.SetOnebit(2, 3)
	shape.SetOnebit(3, 2)
	shape.SetOnebit(4, 2)
	shape.SetOnebit(5, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(0, 1)
	shape.SetOnebit(1, 1)
	shape.SetOnebit(1, 3)
	shape.SetOnebit(2, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(3, 1)
	shape.SetOnebit(4, 1)
	shape.SetOnebit(4, 3)
	shape.SetOnebit(5, 1)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(3, 3)
	shape.SetOnebit(4, 4)
	shape.SetOnebit(5, 3)
	shape.SetOnebit(5, 4)
	sbf.AddShape(shape)

	sbf.ExportImage("gray.png")
	sbf.Simulate()
	sbf.ExportChainImage("chain_ordered_gray.png")
}

func TestShapeBitFieldExport2(t *testing.T) {
	sbf := NewShapeBitField()
	shape := NewFieldBits()
	shape.SetOnebit(0, 2)
	shape.SetOnebit(0, 3)
	shape.SetOnebit(0, 4)
	shape.SetOnebit(1, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(2, 1)
	shape.SetOnebit(2, 3)
	shape.SetOnebit(3, 2)
	shape.SetOnebit(3, 1)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(0, 1)
	shape.SetOnebit(1, 1)
	shape.SetOnebit(1, 3)
	shape.SetOnebit(2, 2)
	sbf.AddShape(shape)

	shape = NewFieldBits()
	shape.SetOnebit(2, 4)
	shape.SetOnebit(2, 5)
	shape.SetOnebit(2, 6)
	shape.SetOnebit(3, 3)
	sbf.AddShape(shape)

	sbf.ExportImage("gray.png")
	sbf.Simulate()
	sbf.ExportChainImage("chain_ordered_gray.png")
}
*/
