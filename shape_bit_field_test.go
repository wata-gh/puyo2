package puyo2

import (
	"testing"
)

func TestFillChainableColor(t *testing.T) {
	sbf := NewShapeBitFieldWithFieldString("........................23....11....13....12....52....42....33....445...455...")
	bf := sbf.FillChainableColor()
	if bf.MattulwanEditorParam() != "aaaaaaaaaaaaaaaaaaaaaaaacdaaaabbaaaabdaaaabcaaaaecaaaabcaaaaddaaaabbeaaabeeaaa" {
		t.Fatalf("mattulwan editor param must be\naaaaaaaaaaaaaaaaaaaaaaaacdaaaabbaaaabdaaaabcaaaaecaaaabcaaaaddaaaabbeaaabeeaaa but\n%s", bf.MattulwanEditorParam())
	}

	sbf = NewShapeBitFieldWithFieldString("..............................5.....32....12....11....212...334...355...4454..")
	bf = sbf.FillChainableColor()
	if bf.MattulwanEditorParam() != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaacaaaaadcaaaabcaaaabbaaaacbcaaaddbaaadccaaabbcbaa" {
		t.Fatalf("mattulwan editor param must be\naaaaaaaaaaaaaaaaaaaaaaaaaaaaaacaaaaadcaaaabcaaaabbaaaacbcaaaddbaaadccaaabbcbaa but\n%s", bf.MattulwanEditorParam())
	}

	sbf = NewShapeBitFieldWithFieldString(".............................................................2....513....4....")
	bf = sbf.FillChainableColor()
	if bf != nil {
		t.Fatalf("BitField must be nil but %v", bf)
	}
}

func TestInsertShape2(t *testing.T) {
	sbf := NewShapeBitField()
	shape := NewFieldBits()
	shape.SetOnebit(0, 1)
	shape.SetOnebit(0, 2)
	shape.SetOnebit(1, 1)
	shape.SetOnebit(2, 1)
	sbf.AddShape(shape)

	insert := NewFieldBits()
	insert.SetOnebit(0, 1)
	insert.SetOnebit(0, 2)
	insert.SetOnebit(1, 1)
	insert.SetOnebit(2, 1)
	sbf.InsertShape(insert)
	sbf.ShowDebug()

	insert = NewFieldBits()
	insert.SetOnebit(0, 2)
	insert.SetOnebit(0, 3)
	insert.SetOnebit(1, 2)
	insert.SetOnebit(1, 3)
	sbf.InsertShape(insert)
	sbf.ShowDebug()

	insert = NewFieldBits()
	insert.SetOnebit(1, 3)
	insert.SetOnebit(1, 4)
	insert.SetOnebit(1, 5)
	insert.SetOnebit(2, 4)
	sbf.InsertShape(insert)
	sbf.ShowDebug()
}

func TestInsertShape(t *testing.T) {
	sbf := NewShapeBitField()
	shape := NewFieldBits()
	shape.SetOnebit(0, 1)
	shape.SetOnebit(0, 2)
	shape.SetOnebit(0, 3)
	shape.SetOnebit(1, 1)
	sbf.AddShape(shape)

	insert := NewFieldBits()
	insert.SetOnebit(0, 2)
	insert.SetOnebit(0, 3)
	insert.SetOnebit(1, 2)
	insert.SetOnebit(2, 2)
	sbf.InsertShape(insert)

	if sbf.Shapes[0].M[0] != 131122 || sbf.Shapes[0].M[1] != 0 {
		t.Fatalf("shape must be %d %d", 131122, 0)
	}

	sbf = NewShapeBitField()
	shape = NewFieldBits()
	shape.SetOnebit(0, 1)
	shape.SetOnebit(0, 2)
	shape.SetOnebit(1, 1)
	shape.SetOnebit(2, 1)
	sbf.AddShape(shape)

	insert = NewFieldBits()
	insert.SetOnebit(0, 1)
	insert.SetOnebit(0, 2)
	insert.SetOnebit(1, 1)
	insert.SetOnebit(2, 1)

	sbf.InsertShape(insert)
	if sbf.Shapes[0].M[0] != 17180131352 || sbf.Shapes[0].M[1] != 0 {
		t.Fatalf("shape must be %d %d", 17180131352, 0)
	}

	insert = NewFieldBits()
	insert.SetOnebit(0, 1)
	insert.SetOnebit(0, 2)
	insert.SetOnebit(1, 1)
	insert.SetOnebit(2, 1)

	sbf.InsertShape(insert)
	if sbf.Shapes[0].M[0] != 34360262752 || sbf.Shapes[0].M[1] != 0 {
		t.Fatalf("shape must be %d %d", 34360262752, 0)
	}

	insert = NewFieldBits()
	insert.SetOnebit(0, 1)
	insert.SetOnebit(1, 1)
	insert.SetOnebit(1, 2)
	insert.SetOnebit(2, 1)

	sbf.InsertShape(insert)
	if sbf.Shapes[0].M[0] != 68721574080 || sbf.Shapes[0].M[1] != 0 {
		t.Fatalf("shape must be %d %d", 68721574080, 0)
	}

	sbf = NewShapeBitField()
	shape = NewFieldBits()
	shape.SetOnebit(3, 1)
	shape.SetOnebit(3, 2)
	shape.SetOnebit(4, 1)
	shape.SetOnebit(5, 1)
	sbf.AddShape(shape)

	insert = NewFieldBits()
	insert.SetOnebit(3, 1)
	insert.SetOnebit(3, 2)
	insert.SetOnebit(4, 1)
	insert.SetOnebit(5, 1)

	sbf.InsertShape(insert)
	if sbf.Shapes[0].M[0] != 6755399441055744 || sbf.Shapes[0].M[1] != 262148 {
		t.Fatalf("shape must be %d %d", 6755399441055744, 262148)
	}

	insert = NewFieldBits()
	insert.SetOnebit(3, 1)
	insert.SetOnebit(3, 2)
	insert.SetOnebit(4, 1)
	insert.SetOnebit(5, 1)

	sbf.InsertShape(insert)
	if sbf.Shapes[0].M[0] != 27021597764222976 || sbf.Shapes[0].M[1] != 524296 {
		t.Fatalf("shape must be %d %d", 27021597764222976, 524296)
	}

	insert = NewFieldBits()
	insert.SetOnebit(3, 1)
	insert.SetOnebit(4, 1)
	insert.SetOnebit(4, 2)
	insert.SetOnebit(5, 1)

	sbf.InsertShape(insert)
	if sbf.Shapes[0].M[0] != 54043195528445952 || sbf.Shapes[0].M[1] != 1048608 {
		t.Fatalf("shape must be %d %d", 54043195528445952, 1048608)
	}
}

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

	// sbf.ShowDebug()

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
