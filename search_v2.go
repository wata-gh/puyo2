package puyo2

import "fmt"

type SearchCondition struct {
	DisableChigiri   bool
	ChigiriableCount int
	Chigiris         int
	SetFrames        int
	PuyoSets         []PuyoSet
	BitField         *BitField
	LastCallback     func(sr *SearchResult)
	EachHandCallback func(sr *SearchResult) bool
}

func NewSearchCondition() *SearchCondition {
	cond := new(SearchCondition)
	cond.DisableChigiri = false
	return cond
}

func NewSearchConditionWithBFAndPuyoSets(bf *BitField, puyoSets []PuyoSet) *SearchCondition {
	cond := new(SearchCondition)
	cond.DisableChigiri = false
	cond.BitField = bf
	cond.PuyoSets = puyoSets
	return cond
}

func (cond *SearchCondition) SearchWithPuyoSetsV2() {
	cond.searchWithPuyoSetsV2([]Hand{}, nil)
}

func (cond *SearchCondition) searchWithPuyoSetsV2(hands []Hand, searchResult *SearchResult) {
	if len(cond.PuyoSets) == 0 {
		if cond.LastCallback != nil {
			cond.LastCallback(searchResult)
		}
		return
	}
	cond.SearchPositionV2(hands)
}

func (cond *SearchCondition) SearchPositionV2(hands []Hand) {
	puyoSet := cond.PuyoSets[0]
	// axis=x child=y
	// 0:  1:    2:  3:
	// y   x y   x   y x
	// x         y
	positions := [][2]int{
		{0, 0}, {0, 1}, {0, 2},
		{1, 0}, {1, 1}, {1, 2}, {1, 3},
		{2, 0}, {2, 1}, {2, 2}, {2, 3},
		{3, 0}, {3, 1}, {3, 2}, {3, 3},
		{4, 0}, {4, 1}, {4, 2}, {4, 3},
		{5, 0}, {5, 2}, {5, 3},
	}
	if puyoSet.Axis == puyoSet.Child {
		positions = [][2]int{
			{0, 0}, {0, 1},
			{1, 0}, {1, 1},
			{2, 0}, {2, 1},
			{3, 0}, {3, 1},
			{4, 0}, {4, 1},
			{5, 0},
		}
	}

	// if there is a puyo at the x, you can't place puyos.
	if cond.BitField.Color(2, 12) != Empty {
		return
	}

	for i, pos := range positions {
		bf := cond.BitField.Clone()
		placement := bf.SearchPlacementForPos(&puyoSet, pos)
		if placement == nil || (cond.DisableChigiri && placement.Chigiri) {
			continue
		}
		if bf.PlacePuyoWithPlacement(placement) == false {
			panic(fmt.Sprintf("should be able to place. %+v\n", placement))
		}
		newHands := make([]Hand, len(hands))
		copy(newHands, hands)
		hand := new(Hand)
		hand.Position = pos
		hand.PuyoSet = puyoSet
		newHands = append(newHands, *hand)

		result := bf.Clone().SimulateDetail() // for erased puyo count
		searchResult := new(SearchResult)
		searchResult.BeforeSimulate = bf
		searchResult.Position = pos
		searchResult.PositionNum = i
		searchResult.RensaResult = result
		searchResult.Hands = newHands
		searchResult.Depth = len(newHands)
		searchResult.RensaResult.SetFrames = cond.SetFrames + placement.Frames
		if placement.Chigiri {
			searchResult.RensaResult.Chigiris = cond.Chigiris + 1
		} else {
			searchResult.RensaResult.Chigiris = cond.Chigiris
		}

		if cond.EachHandCallback == nil || cond.EachHandCallback(searchResult) {
			newCond := new(SearchCondition)
			newCond.PuyoSets = cond.PuyoSets[1:]
			newCond.DisableChigiri = cond.DisableChigiri
			newCond.ChigiriableCount = cond.ChigiriableCount
			newCond.BitField = result.BitField
			newCond.LastCallback = cond.LastCallback
			newCond.EachHandCallback = cond.EachHandCallback
			newCond.Chigiris = searchResult.RensaResult.Chigiris
			newCond.SetFrames = searchResult.RensaResult.SetFrames
			newCond.searchWithPuyoSetsV2(newHands, searchResult)
		}
	}
}
