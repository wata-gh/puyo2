package puyo2

import (
	"testing"
)

func TestFastLift(t *testing.T) {
	fb := NewFieldBits()
	fb.SetOnebit(0, 1)
	fb.SetOnebit(0, 2)
	fb.SetOnebit(1, 1)
	fb.SetOnebit(2, 1)
	lift1 := fb.FastLift(1)
	expect1 := NewFieldBits()
	expect1.SetOnebit(0, 2)
	expect1.SetOnebit(0, 3)
	expect1.SetOnebit(1, 2)
	expect1.SetOnebit(2, 2)
	if lift1.Equals(expect1) == false {
		t.Fatalf("lift shape must be\n%s\nbut\n%s", lift1.ToString(), expect1.ToString())
	}
	lift2 := fb.FastLift(18)
	expect2 := NewFieldBits()
	expect2.SetOnebit(1, 3)
	expect2.SetOnebit(1, 4)
	expect2.SetOnebit(2, 3)
	expect2.SetOnebit(3, 3)
	if lift2.Equals(expect2) == false {
		t.Fatalf("lift shape must be\n%s\nbut\n%s", lift2.ToString(), expect2.ToString())
	}
}

func TestAnd(t *testing.T) {
	fb := NewFieldBits()
	fb.SetOnebit(0, 4)
	fb.SetOnebit(0, 3)
	fb.SetOnebit(0, 2)
	fb.SetOnebit(1, 2)

	fb2 := NewFieldBits()
	fb2.SetOnebit(0, 4)
	fb2.SetOnebit(0, 2)
	fb2.SetOnebit(2, 2)

	result := fb.And(fb2)

	expect := NewFieldBits()
	expect.SetOnebit(0, 4)
	expect.SetOnebit(0, 2)

	if result.Equals(expect) == false {
		t.Fatalf("And() expected\n%s\n actual\n\n%s", result.ToString(), expect.ToString())
	}
}

func TestAndNot(t *testing.T) {
	fb := NewFieldBits()
	fb.SetOnebit(0, 4)
	fb.SetOnebit(0, 3)
	fb.SetOnebit(0, 2)
	fb.SetOnebit(1, 2)

	fb.SetOnebit(4, 4)
	fb.SetOnebit(4, 3)
	fb.SetOnebit(4, 2)
	fb.SetOnebit(5, 2)

	fb2 := NewFieldBits()
	fb2.SetOnebit(0, 4)
	fb2.SetOnebit(0, 3)
	fb2.SetOnebit(0, 2)
	fb2.SetOnebit(1, 2)

	fb.AndNot(fb2)

	expect := NewFieldBits()
	expect.SetOnebit(4, 4)
	expect.SetOnebit(4, 3)
	expect.SetOnebit(4, 2)
	expect.SetOnebit(5, 2)

	if fb.Equals(expect) == false {
		t.Fatalf("AndNot() expected\n%s\n actual\n\n%s", fb.ToString(), expect.ToString())
	}
}

func TestEquqls(t *testing.T) {
	fb := NewFieldBits()
	fb.SetOnebit(0, 1)
	fb2 := NewFieldBits()
	fb2.SetOnebit(0, 1)
	if fb.Equals(fb2) == false {
		t.Fatal("Equals() fb and fb2 must be equal")
	}
	fb3 := NewFieldBits()
	fb.SetOnebit(1, 1)
	if fb.Equals(fb3) {
		t.Fatal("Equals() fb and fb3 must not be equal")
	}
}

func TestFindVanishingBits(t *testing.T) {
	bf := NewBitFieldWithMattulwan("a54ea3eaebdece3bd2eb2dc3")
	v := bf.Bits(Green).FindVanishingBits()
	expect := NewFieldBits()
	expect.SetOnebit(0, 4)
	expect.SetOnebit(0, 3)
	expect.SetOnebit(0, 2)
	expect.SetOnebit(1, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Red).FindVanishingBits()
	expect = NewFieldBits()
	expect.SetOnebit(0, 1)
	expect.SetOnebit(1, 1)
	expect.SetOnebit(1, 2)
	expect.SetOnebit(2, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Yellow).FindVanishingBits()
	expect = NewFieldBits()
	expect.SetOnebit(2, 1)
	expect.SetOnebit(2, 2)
	expect.SetOnebit(3, 2)
	expect.SetOnebit(4, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Blue).FindVanishingBits()
	expect = NewFieldBits()
	expect.SetOnebit(3, 1)
	expect.SetOnebit(4, 1)
	expect.SetOnebit(4, 2)
	expect.SetOnebit(5, 1)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}

	bf.Simulate1()
	v = bf.Bits(Green).FindVanishingBits()
	expect = NewFieldBits()
	expect.SetOnebit(3, 1)
	expect.SetOnebit(4, 1)
	expect.SetOnebit(5, 1)
	expect.SetOnebit(5, 2)
	if v.Equals(expect) == false {
		bf.ShowDebug()
		t.Fatalf("FindVanishingBits must be \n%s\nbut\n\n%s", expect.ToString(), v.ToString())
	}
}

func TestIterateBitWithMasking(t *testing.T) {
	fb := NewFieldBits()
	fb.SetOnebit(0, 4)
	fb.SetOnebit(0, 3)
	fb.SetOnebit(0, 2)
	fb.SetOnebit(1, 2)

	fb.SetOnebit(4, 4)
	fb.SetOnebit(4, 3)
	fb.SetOnebit(4, 2)
	fb.SetOnebit(5, 2)

	cnt := 0
	fb.IterateBitWithMasking(func(c *FieldBits) *FieldBits {
		v := c.Expand(fb)
		if cnt == 0 {
			e := NewFieldBits()
			e.SetOnebit(0, 4)
			e.SetOnebit(0, 3)
			e.SetOnebit(0, 2)
			e.SetOnebit(1, 2)
			if v.Equals(e) == false {
				t.Fatalf("IterateBitWithMasking() first expected\n%s\nactual\n\n%s", e.ToString(), v.ToString())
			}
		} else if cnt == 1 {
			e := NewFieldBits()
			e.SetOnebit(4, 4)
			e.SetOnebit(4, 3)
			e.SetOnebit(4, 2)
			e.SetOnebit(5, 2)
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
