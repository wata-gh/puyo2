package puyo2

import (
	"fmt"
	"testing"
)

func TestSearchPositionV2(t *testing.T) {
	solved := false
	cond := SearchCondition{
		BitField:       NewBitFieldWithMattulwan("a62gacbagecb2ae2g3"),
		PuyoSets:       []PuyoSet{{Red, Blue}, {Green, Blue}},
		DisableChigiri: false,
		LastCallback: func(sr *SearchResult) {
			if sr.RensaResult.Chains == 3 {
				solved = true
			}
		},
	}
	cond.SearchWithPuyoSetsV2()
	if solved == false {
		t.Fatal("can not solve a62gacbagecb2ae2g3")
	}
}

func TestSearchWithPuyoSets(t *testing.T) {
	solved := false
	puyoSets := []PuyoSet{
		{Red, Blue},
		{Green, Blue},
	}
	cond := NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan("a62gacbagecb2ae2g3"), puyoSets)
	cond.EachHandCallback = func(sr *SearchResult) bool {
		if sr.Depth != len(puyoSets) {
			if sr.RensaResult.Chains != 0 {
				return false
			}
		}
		return true
	}
	cond.LastCallback = func(sr *SearchResult) {
		if sr.RensaResult.Chains == 3 {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
	}
	cond.SearchWithPuyoSetsV2()
	if solved == false {
		t.Fatal("can not solve a62gacbagecb2ae2g3")
	}

	solved = false
	puyoSets = []PuyoSet{
		{Green, Red},
		{Green, Red},
		{Green, Red},
	}
	cond = NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan("a46ea5ea5ea5ga5ea4eba"), puyoSets)
	cond.EachHandCallback = func(sr *SearchResult) bool {
		if sr.Depth != len(puyoSets) {
			if sr.RensaResult.Chains != 0 {
				return false
			}
		}
		return true
	}
	cond.LastCallback = func(sr *SearchResult) {
		if sr.RensaResult.Chains == 3 {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
	}
	cond.SearchWithPuyoSetsV2()
	if solved == false {
		t.Fatal("can not solve a46ea5ea5ea5ga5ea4eba")
	}

	solved = false
	puyoSets = []PuyoSet{
		{Blue, Red},
		{Red, Red},
		{Blue, Red},
	}
	cond = NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan("a52ca2gbc2a2c2g2a2cgbga2g2b2a"), puyoSets)
	cond.EachHandCallback = func(sr *SearchResult) bool {
		if sr.Depth != len(puyoSets) {
			if sr.RensaResult.Chains != 0 {
				return false
			}
		}
		return true
	}
	cond.LastCallback = func(sr *SearchResult) {
		if sr.RensaResult.Chains == 4 {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
	}
	cond.SearchWithPuyoSetsV2()
	if solved == false {
		t.Fatal("can not solve a52ca2gbc2a2c2g2a2cgbga2g2b2a")
	}

	solved = false
	puyoSets = []PuyoSet{
		{Red, Red},
		{Red, Red},
		{Blue, Blue},
		{Blue, Blue},
	}
	cond = NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan("a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16"), puyoSets)
	cond.LastCallback = func(sr *SearchResult) {
		if sr.RensaResult.BitField.Bits(Blue).IsEmpty() {
			solved = true
			fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
		}
	}
	cond.SearchWithPuyoSetsV2()
	if solved == false {
		t.Fatal("can not solve a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16")
	}

	// solved = false
	// puyoSets = []PuyoSet{
	// 	{Blue, Red},
	// 	{Blue, Red},
	// 	{Yellow, Red},
	// 	{Blue, Red},
	// 	{Blue, Red},
	// 	{Yellow, Red},
	// }
	// cond = NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan("a60cabacabadadacacaba"), puyoSets)
	// cond.EachHandCallback = func(sr *SearchResult) bool {
	// 	if sr.Depth != len(puyoSets) {
	// 		if sr.RensaResult.Chains != 0 {
	// 			return false
	// 		}
	// 	}
	// 	return true
	// }
	// cond.LastCallback = func(sr *SearchResult) {
	// 	if sr.RensaResult.Chains == 5 {
	// 		solved = true
	// 		fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
	// 	}
	// }
	// cond.SearchWithPuyoSetsV2()
	// if solved == false {
	// 	t.Fatal("can not solve a60cabacabadadacacaba")
	// }

	// solved = false
	// puyoSets = []PuyoSet{
	// 	{Green, Blue},
	// 	{Red, Green},
	// 	{Red, Red},
	// 	{Yellow, Yellow},
	// 	{Green, Green},
	// 	{Blue, Yellow},
	// }
	// cond = NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan("a7baca3baba3daea3eada3cada3bada3eaca3daca3dada3daca3eaca3daba2"), puyoSets)
	// cond.EachHandCallback = func(sr *SearchResult) bool {
	// 	if sr.Depth != len(puyoSets) {
	// 		if sr.RensaResult.Chains != 0 {
	// 			return false
	// 		}
	// 	}
	// 	return true
	// }
	// cond.LastCallback = func(sr *SearchResult) {
	// 	if sr.RensaResult.Chains == 9 {
	// 		solved = true
	// 		fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl())
	// 	}
	// }
	// cond.SearchWithPuyoSetsV2()
	// if solved == false {
	// 	t.Fatal("can not solve a7baca3baba3daea3eada3cada3bada3eaca3daca3dada3daca3eaca3daba2")
	// }
}
