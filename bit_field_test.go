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
	v.Onebit(0, 12)
	bf.Drop(v)
	c := bf.Color(0, 12)
	if c != Red {
		bf.ShowDebug()
		t.Fatalf("row 12 puyo must be red, but %d", c)
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
		t.Fatalf("result.Chains must be 4,840 %v", result)
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
}
