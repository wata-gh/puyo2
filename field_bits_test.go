package puyo2

import (
	"testing"
)

func TestAndNot(t *testing.T) {
	fb := NewFieldBits()
	fb.Onebit(0, 4)
	fb.Onebit(0, 3)
	fb.Onebit(0, 2)
	fb.Onebit(1, 2)

	fb.Onebit(4, 4)
	fb.Onebit(4, 3)
	fb.Onebit(4, 2)
	fb.Onebit(5, 2)

	fb2 := NewFieldBits()
	fb2.Onebit(0, 4)
	fb2.Onebit(0, 3)
	fb2.Onebit(0, 2)
	fb2.Onebit(1, 2)

	fb.AndNot(fb2)

	expect := NewFieldBits()
	expect.Onebit(4, 4)
	expect.Onebit(4, 3)
	expect.Onebit(4, 2)
	expect.Onebit(5, 2)

	if fb.Equals(expect) == false {
		t.Fatalf("AndNot() expected\n%s\n actual\n\n%s", fb.ToString(), expect.ToString())
	}
}

func TestEquqls(t *testing.T) {
	fb := NewFieldBits()
	fb.Onebit(0, 1)
	fb2 := NewFieldBits()
	fb2.Onebit(0, 1)
	if fb.Equals(fb2) == false {
		t.Fatal("Equals() fb and fb2 must be equal")
	}
	fb3 := NewFieldBits()
	fb.Onebit(1, 1)
	if fb.Equals(fb3) {
		t.Fatal("Equals() fb and fb3 must not be equal")
	}
}

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

func TestIterateBitWithMasking(t *testing.T) {
	fb := NewFieldBits()
	fb.Onebit(0, 4)
	fb.Onebit(0, 3)
	fb.Onebit(0, 2)
	fb.Onebit(1, 2)

	fb.Onebit(4, 4)
	fb.Onebit(4, 3)
	fb.Onebit(4, 2)
	fb.Onebit(5, 2)

	cnt := 0
	fb.IterateBitWithMasking(func(c *FieldBits) *FieldBits {
		c = c.expand(fb)
		v := c.FindVanishingBits()
		if cnt == 0 {
			e := NewFieldBits()
			e.Onebit(0, 4)
			e.Onebit(0, 3)
			e.Onebit(0, 2)
			e.Onebit(1, 2)
			if v.Equals(e) == false {
				t.Fatalf("IterateBitWithMasking() first expected\n%s\nactual\n\n%s", e.ToString(), v.ToString())
			}
		} else if cnt == 1 {
			e := NewFieldBits()
			e.Onebit(4, 4)
			e.Onebit(4, 3)
			e.Onebit(4, 2)
			e.Onebit(5, 2)
			if v.Equals(e) == false {
				t.Fatalf("IterateBitWithMasking() second expected\n%s\nactual\n\n%s", e.ToString(), v.ToString())
			}
		}
		cnt++
		return v
	})
	if cnt != 2 {
		t.Fatalf("IterateBitWithMasking() callback must be called 2 times but %d", cnt)
	}
}
