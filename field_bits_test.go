package puyo2

import "testing"

func TestFindVanishingBits(t *testing.T) {
	bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
	v := bf.Bits(Green).FindVanishingBits()
	expect := NewFieldBits()
	expect.Onebit(0, 4)
	expect.Onebit(0, 3)
	expect.Onebit(0, 2)
	expect.Onebit(1, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Red).FindVanishingBits()
	expect = NewFieldBits()
	expect.Onebit(0, 1)
	expect.Onebit(1, 1)
	expect.Onebit(1, 2)
	expect.Onebit(2, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Yellow).FindVanishingBits()
	expect = NewFieldBits()
	expect.Onebit(2, 1)
	expect.Onebit(2, 2)
	expect.Onebit(3, 2)
	expect.Onebit(4, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Blue).FindVanishingBits()
	expect = NewFieldBits()
	expect.Onebit(3, 1)
	expect.Onebit(4, 1)
	expect.Onebit(4, 2)
	expect.Onebit(5, 1)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Green).FindVanishingBits()
	expect = NewFieldBits()
	expect.Onebit(3, 1)
	expect.Onebit(4, 1)
	expect.Onebit(5, 1)
	expect.Onebit(5, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}
}
