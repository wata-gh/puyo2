package puyo2

import (
	"fmt"
	"testing"
)

func TestPlacePuyo(t *testing.T) {
	bf := NewBitField()
	bf.placePuyo(PuyoSet{Red, Green}, [2]int{0, 0})
	bf.Drop(bf.Bits(Empty).MaskField12())
	if bf.Bits(Red).Equals(NewFieldBitsWithM([2]uint64{2, 0})) == false {
		bf.ShowDebug()
		t.Fatal("red must be 0,1")
	}
	if bf.Bits(Green).Equals(NewFieldBitsWithM([2]uint64{4, 0})) == false {
		bf.ShowDebug()
		t.Fatal("green must be 0,2")
	}

	bf = NewBitField()
	bf.placePuyo(PuyoSet{Red, Green}, [2]int{0, 1})
	bf.Drop(bf.Bits(Empty).MaskField12())
	if bf.Bits(Red).Equals(NewFieldBitsWithM([2]uint64{2, 0})) == false {
		bf.ShowDebug()
		t.Fatal("red must be 0,1")
	}
	if bf.Bits(Green).Equals(NewFieldBitsWithM([2]uint64{2 << 16, 0})) == false {
		bf.ShowDebug()
		t.Fatal("green must be 1,2")
	}

	bf = NewBitField()
	bf.placePuyo(PuyoSet{Red, Green}, [2]int{0, 2})
	bf.Drop(bf.Bits(Empty).MaskField12())
	if bf.Bits(Red).Equals(NewFieldBitsWithM([2]uint64{4, 0})) == false {
		bf.ShowDebug()
		t.Fatal("red must be 0,1")
	}
	if bf.Bits(Green).Equals(NewFieldBitsWithM([2]uint64{2, 0})) == false {
		bf.ShowDebug()
		t.Fatal("green must be 0,2")
	}

	bf = NewBitField()
	bf.placePuyo(PuyoSet{Red, Green}, [2]int{1, 3})
	bf.ShowDebug()
	bf.Drop(bf.Bits(Empty).MaskField12())
	if bf.Bits(Red).Equals(NewFieldBitsWithM([2]uint64{2 << 16, 0})) == false {
		bf.ShowDebug()
		t.Fatal("red must be 1,1")
	}
	if bf.Bits(Green).Equals(NewFieldBitsWithM([2]uint64{2, 0})) == false {
		bf.ShowDebug()
		t.Fatal("green must be 0,1")
	}
}

func TestSearchPosition(t *testing.T) {
	solved := false
	hands := []Hand{}
	bf := NewBitFieldWithMattulwan("a62gacbagecb2ae2g3")
	bf.SearchPosition(PuyoSet{Red, Blue}, hands, func(sr *SearchResult) bool {
		if sr.RensaResult.Chains == 0 {
			sr.RensaResult.BitField.SearchPosition(PuyoSet{Green, Blue}, sr.Hands, func(sr2 *SearchResult) bool {
				if sr2.RensaResult.Chains == 3 {
					sr2.BeforeSimulate.ShowDebug()
					solved = true
					fmt.Println(sr2.BeforeSimulate.MattulwanEditorUrl())
				}
				return true
			})
		}
		return true
	})
	if solved == false {
		t.Fatal("can not solve a62gacbagecb2ae2g3")
	}

	bf = NewBitField()
	hands = []Hand{}
	callCount := 0
	bf.SearchPosition(PuyoSet{Red, Blue}, hands, func(sr *SearchResult) bool {
		sr.BeforeSimulate.ShowDebug()
		callCount++
		if callCount == 3 {
			return false
		}
		return true
	})
	if callCount != 3 {
		t.Fatalf("call count must be 3 but %d", callCount)
	}
}

func TestSearchWithPuyoSetsStop(t *testing.T) {
	bf := NewBitField()
	puyoSets := []PuyoSet{
		{Red, Blue},
		{Green, Blue},
	}
	hands := []Hand{}

	callCount := 0
	bf.SearchWithPuyoSets(puyoSets, hands, func(sr *SearchResult) bool {
		if sr.Depth == 1 && callCount == 3 {
			return false
		}
		callCount++
		return true
	}, 0)

	if callCount != 3 {
		t.Fatalf("SearchWithPuyoSets callCount must be 3 but %d", callCount)
	}
}

func TestSearchWithPuyoSets(t *testing.T) {
	solved := false
	bf := NewBitFieldWithMattulwan("a62gacbagecb2ae2g3")
	puyoSets := []PuyoSet{
		{Red, Blue},
		{Green, Blue},
	}
	hands := []Hand{}
	solved = false
	bf.SearchWithPuyoSets(puyoSets, hands, func(sr *SearchResult) bool {
		if sr.Depth != len(puyoSets) {
			if sr.RensaResult.Chains != 0 {
				return false
			}
		}
		if sr.Depth == len(puyoSets) && sr.RensaResult.Chains == 3 {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
		return true
	}, 1)
	if solved == false {
		t.Fatal("can not solve a62gacbagecb2ae2g3")
	}

	bf = NewBitFieldWithMattulwan("a46ea5ea5ea5ga5ea4eba")
	puyoSets = []PuyoSet{
		{Green, Red},
		{Green, Red},
		{Green, Red},
	}
	hands = []Hand{}
	solved = false
	bf.SearchWithPuyoSets(puyoSets, hands, func(sr *SearchResult) bool {
		if sr.Depth != len(puyoSets) {
			if sr.RensaResult.Chains != 0 {
				return true
			}
		}
		if sr.Depth == len(puyoSets) && sr.RensaResult.Chains == 3 {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
		return true
	}, 1)
	if solved == false {
		t.Fatal("can not solve a46ea5ea5ea5ga5ea4eba")
	}

	bf = NewBitFieldWithMattulwan("a52ca2gbc2a2c2g2a2cgbga2g2b2a")
	puyoSets = []PuyoSet{
		{Blue, Red},
		{Red, Red},
		{Blue, Red},
	}
	hands = []Hand{}
	solved = false
	bf.SearchWithPuyoSets(puyoSets, hands, func(sr *SearchResult) bool {
		if sr.Depth != len(puyoSets) {
			if sr.RensaResult.Chains != 0 {
				return true
			}
		}
		if sr.Depth == len(puyoSets) && sr.RensaResult.Chains == 4 {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
		return true
	}, 1)
	if solved == false {
		t.Fatal("can not solve a52ca2gbc2a2c2g2a2cgbga2g2b2a")
	}

	bf = NewBitFieldWithMattulwan("a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16")
	puyoSets = []PuyoSet{
		{Red, Red},
		{Red, Red},
		{Blue, Blue},
		{Blue, Blue},
	}
	hands = []Hand{}
	solved = false
	bf.SearchWithPuyoSets(puyoSets, hands, func(sr *SearchResult) bool {
		if sr.RensaResult.BitField.Bits(Blue).IsEmpty() {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
		return true
	}, 1)
	if solved == false {
		t.Fatal("can not solve a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16")
	}

	// fmt.Println("a60cabacabadadacacaba")
	// bf = NewBitFieldWithMattulwan("a60cabacabadadacacaba")
	// puyoSets = [][2]Color{
	// 	{Blue, Red},
	// 	{Blue, Red},
	// 	{Yellow, Red},
	// 	{Blue, Red},
	// 	{Blue, Red},
	// 	{Yellow, Red},
	// }
	// bf.SearchWithPuyoSets(puyoSets, func(sr *SearchResult) bool {
	// 	if sr.Depth != len(puyoSets)-1 {
	// 		if sr.RensaResult.Chains != 0 {
	// 			return false
	// 		}
	// 	}
	// 	if sr.Depth == len(puyoSets)-1 && sr.RensaResult.Chains == 5 {
	// 		fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
	// 	}
	// 	return true
	// }, 0)

	// fmt.Println("== a7baca3baba3daea3eada3cada3bada3eaca3daca3dada3daca3eaca3daba2 ==")
	// bf = NewBitFieldWithMattulwan("a7baca3baba3daea3eada3cada3bada3eaca3daca3dada3daca3eaca3daba2")
	// puyoSets = [][2]Color{
	// 	{Green, Blue},
	// 	{Red, Green},
	// 	{Red, Red},
	// 	{Yellow, Yellow},
	// 	{Green, Green},
	// 	{Blue, Yellow},
	// }
	// solved = false
	// bf.SearchWithPuyoSets(puyoSets, func(sr *SearchResult) bool {
	// 	if sr.Depth != len(puyoSets)-1 {
	// 		if sr.RensaResult.Chains != 0 {
	// 			return false
	// 		}
	// 	}
	// 	if sr.Depth == len(puyoSets)-1 && sr.RensaResult.Chains == 9 {
	// 		solved = true
	// 		fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
	// 	}
	// 	return true
	// }, 0)
	// if solved == false {
	// 	t.Fatal("can not solve a7baca3baba3daea3eada3cada3bada3eaca3daca3dada3daca3eaca3daba2")
	// }
}
