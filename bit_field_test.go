package puyo2

import (
	"testing"
)

func benchmarkSimulate(num int) {
	for i := 0; i < num; i++ {
		bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
		bf.Simulate()
	}
}
func BenchmarkSimulate(b *testing.B) {
	benchmarkSimulate(1_000_000)
}

func benchmarkSimulateDetail(num int) {
	for i := 0; i < num; i++ {
		bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
		bf.SimulateDetail()
	}
}
func BenchmarkSimulateDetail(b *testing.B) {
	benchmarkSimulateDetail(1_000_000)
}

func TestToChainShapes(t *testing.T) {
	bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
	expect := [][2]uint64{
		{262172, 0},
		{17180262402, 0},
		{1125925676646400, 4},
		{562949953421312, 131078},
		{562949953421312, 393218},
	}
	for i, v := range bf.ToChainShapes() {
		va := v.ToIntArray()
		if va[0] != expect[i][0] || va[1] != expect[i][1] {
			t.Fatalf("ToChainShapes() %d chain must be %v but %v", i, expect[i], va)
		}
	}
}

func TestDrop(t *testing.T) {
	// 13 段目のぷよが落ちてくるテスト
	bf := NewBitField()
	bf.SetColor(Red, 0, 13)
	v := NewFieldBits()
	v.SetOnebit(0, 12)
	bf.Drop(v)
	c := bf.Color(0, 12)
	if c != Red {
		bf.ShowDebug()
		t.Fatalf("row 12 puyo must be red, but %d", c)
	}
}

func TestMaskField(t *testing.T) {
	// 13 段目のぷよが落ちてくるテスト
	bf := NewBitFieldWithMattulwanC("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaeaaaaacccaaaeeeaaabcdaeabbcddaccdeee")
	fb := NewShapeBitFieldWithFieldString("............................................................111.1.111111111111").Shapes[0]
	mbf := bf.MaskField(fb)
	if mbf.MattulwanEditorParam() != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdaeabbcddaccdeee" {
		t.Fatalf("param must be aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdaeabbcddaccdeee but %s\n", mbf.MattulwanEditorParam())
	}
}

func TestEqualChain(t *testing.T) {
	bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
	if bf.EqualChain(bf) == false {
		t.Fatalf("EqualChain() same bf must be true")
	}
	bf2 := NewBitFieldWithMattulwan("a54ba3babcebdb3ce2bc2ed3")
	if bf.EqualChain(bf2) == false {
		t.Fatalf("EqualChain() color diff must be true")
	}
	bf3 := NewBitFieldWithMattulwan("a54ba3b3cebdb3ce3c2ed3")
	if bf.EqualChain(bf3) {
		t.Fatalf("EqualChain() diff chain shape must be false")
	}
}

func TestNewBitFieldWithMattulwanCAndHaipuyo(t *testing.T) {
	param := "ba77"
	haipuyo := "pprr"
	bf := NewBitFieldWithMattulwanCAndHaipuyo(param, haipuyo)
	if bf.Color(0, 13) != Red {
		t.Fatalf("field must keep existing puyo. actual=%v", bf.Color(0, 13))
	}
	bf.SetColor(Purple, 0, 1)
	if bf.Color(0, 1) != Purple {
		t.Fatalf("set purple must succeed. actual=%v", bf.Color(0, 1))
	}
}

func TestSetMattulwan(t *testing.T) {
	field := "a54ea3eaebdece3bd2eb2dc3"
	bf := NewBitField()
	bf.SetMattulwanParam(field)
	bf2 := NewBitFieldWithMattulwanC(field)
	if bf.MattulwanEditorParam() != bf2.MattulwanEditorParam() {
		t.Fatalf("mattulwan editor param must be same. actual %s\nexpected %s", bf.MattulwanEditorParam(), bf2.MattulwanEditorParam())
	}

	field = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaacaaaaabfaaaadbdaabbffaa"
	bf = NewBitFieldWithColors([]Color{Red, Blue, Yellow, Purple})
	bf.SetMattulwanParam(field)
	bf2 = NewBitFieldWithMattulwanC(field)
	if bf.MattulwanEditorParam() != bf2.MattulwanEditorParam() {
		t.Fatalf("mattulwan editor param must be same. actual %s\nexpected %s", bf.MattulwanEditorParam(), bf2.MattulwanEditorParam())
	}
}

func TestSimulateDetail(t *testing.T) {
	// 通常連鎖
	bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
	result := bf.SimulateDetail()
	if result.Chains != 5 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 5 %d", result.Chains)
	}
	if result.Score != 4_840 {
		bf.ShowDebug()
		t.Fatalf("result.Score must be 4,840 %v", result)
	}
	if result.BitField.IsEmpty() == false {
		bf.ShowDebug()
		t.Fatalf("result.BitField.IsEmpty() must be true %t", result.BitField.IsEmpty())
	}

	// 19 連鎖
	bf = NewBitFieldWithMattulwan("a2bdeb2c2bdcecbde2bcbdecedcedcdcedbcedcedbedcedbeb2cbcbcdec2ebcdebebcdebebcdeb")
	result = bf.SimulateDetail()
	if result.Chains != 19 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 19 %d", result.Chains)
	}
	if result.Score != 175_080 {
		bf.ShowDebug()
		t.Fatalf("result.Score must be 175,080 %v", result)
	}
	if result.BitField.IsEmpty() == false {
		bf.ShowDebug()
		t.Fatalf("result.BitField.IsEmpty() must be true %t", result.BitField.IsEmpty())
	}

	// 2 連鎖トリプル（3 色）
	bf = NewBitFieldWithMattulwan("a34ca5dca4dca4dca4bda4eba4eba3e2b")
	result = bf.SimulateDetail()
	if result.Chains != 2 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 2 %d", result.Chains)
	}
	if result.Score != 1_720 {
		bf.ShowDebug()
		t.Fatalf("result.Score must be 1,720 %v", result)
	}
	if result.BitField.IsEmpty() == false {
		bf.ShowDebug()
		t.Fatalf("result.BitField.IsEmpty() must be true %t", result.BitField.IsEmpty())
	}

	// 2 連鎖トリプル（2 色）
	bf = NewBitFieldWithMattulwan("a34ba5dba4dba4dba4bda4eba4eba3e2b")
	result = bf.SimulateDetail()
	if result.Chains != 2 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 2 %d", result.Chains)
	}
	if result.Score != 1_360 {
		bf.ShowDebug()
		t.Fatalf("result.Score must be 1,360 %v", result)
	}
	if result.BitField.IsEmpty() == false {
		bf.ShowDebug()
		t.Fatalf("result.BitField.IsEmpty() must be true %t", result.BitField.IsEmpty())
	}

	// おじゃまあり
	bf = NewBitFieldWithMattulwan("ga2g2c2a2g2dca2dgegae2bcgae2dcgaed2egcbeb2gdbedbcgbde2dg2c2edbgedcdcgd2c3ge2c")
	result = bf.SimulateDetail()
	if result.Chains != 9 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 9 %d", result.Chains)
	}
	if result.Score != 42_540 {
		bf.ShowDebug()
		t.Fatalf("result.Score must be 42,540 %v", result)
	}
}

func TestSimulate(t *testing.T) {
	// 通常連鎖
	bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
	result := bf.Simulate()
	if result.Chains != 5 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 5 %d", result.Chains)
	}
	if result.BitField.IsEmpty() == false {
		bf.ShowDebug()
		t.Fatalf("result.BitField.IsEmpty() must be true %t", result.BitField.IsEmpty())
	}

	// 19 連鎖
	bf = NewBitFieldWithMattulwan("a2bdeb2c2bdcecbde2bcbdecedcedcdcedbcedcedbedcedbeb2cbcbcdec2ebcdebebcdebebcdeb")
	result = bf.Simulate()
	if result.Chains != 19 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 19 %d", result.Chains)
	}

	// おじゃまあり
	bf = NewBitFieldWithMattulwan("ga2g2c2a2g2dca2dgegae2bcgae2dcgaed2egcbeb2gdbedbcgbde2dg2c2edbgedcdcgd2c3ge2c")
	result = bf.Simulate()
	if result.Chains != 9 {
		bf.ShowDebug()
		t.Fatalf("result.Chains must be 9 %d", result.Chains)
	}
}

// func TestExportSimulateImage(t *testing.T) {
// 	bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
// 	bf.ExportSimulateImage("simulate_images")
// }

func TestNewBitFieldWithMattulwanC(t *testing.T) {
	bf := NewBitFieldWithMattulwanC("a54ba5bedafab2ed2ae2df3")
	purple := bf.Color(3, 1)
	if purple != Purple {
		t.Fatalf("must be purple, but -> %v", purple)
	}
	// bf.ExportSimulateImage("simulate_images2")
}

func TestCloneDoesNotShareTableAndColors(t *testing.T) {
	original := NewBitFieldWithMattulwanC("ba77")
	cloned := original.Clone()
	oldTableColor := original.Table[Red]
	oldColor := original.Colors[0]

	cloned.Table[Red] = Blue
	cloned.Colors[0] = Green

	if original.Table[Red] != oldTableColor {
		t.Fatalf("table must not be shared between clone and original")
	}
	if original.Colors[0] != oldColor {
		t.Fatalf("colors must not be shared between clone and original")
	}
}

func TestPlacePuyoWithPlacementValidation(t *testing.T) {
	bf := NewBitFieldWithMattulwanC("ba77")
	if bf.PlacePuyoWithPlacement(nil) {
		t.Fatalf("nil placement must be rejected")
	}

	overlap := &PuyoSetPlacement{
		PuyoSet: &PuyoSet{Axis: Red, Child: Blue},
		AxisX:   0,
		AxisY:   13,
		ChildX:  1,
		ChildY:  13,
	}
	if bf.PlacePuyoWithPlacement(overlap) {
		t.Fatalf("placement on occupied cell must be rejected")
	}

	outOfRange := &PuyoSetPlacement{
		PuyoSet: &PuyoSet{Axis: Red, Child: Blue},
		AxisX:   -1,
		AxisY:   1,
		ChildX:  0,
		ChildY:  1,
	}
	if bf.PlacePuyoWithPlacement(outOfRange) {
		t.Fatalf("out-of-range placement must be rejected")
	}
}

func TestSetColorWithFieldBitsPurple(t *testing.T) {
	bf := NewBitFieldWithColors([]Color{Red, Blue, Yellow, Purple})
	fb := NewFieldBits()
	fb.SetOnebit(0, 1)
	bf.SetColorWithFieldBits(Purple, fb)
	if bf.Color(0, 1) != Purple {
		t.Fatalf("purple bit set must keep purple color. actual=%v", bf.Color(0, 1))
	}
}

func TestToChainShapesReturnsEmptyWhenNoChain(t *testing.T) {
	bf := NewBitField()
	bf.SetColor(Red, 0, 1)
	shapes := bf.ToChainShapes()
	if len(shapes) != 0 {
		t.Fatalf("ToChainShapes must return empty when no chain, actual=%d", len(shapes))
	}
}
