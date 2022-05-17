package puyo2

import "testing"

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
