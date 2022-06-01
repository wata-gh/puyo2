package puyo2

type SearchResult struct {
	RensaResult    *RensaResult
	BeforeSimulate *BitField
	Depth          int
	Position       [2]int
	Hands          []Hand
}

type PuyoSet struct {
	Axis  Color
	Child Color
}

type Hand struct {
	PuyoSet  PuyoSet
	Position [2]int
}

func (bf *BitField) placePuyo(puyoSet PuyoSet, pos [2]int) bool {
	// TODO check invalid placement and return false.
	ax := pos[0]
	cx := pos[0]
	ay := 12
	cy := 13
	switch pos[1] {
	case 0:
	case 1:
		ay = 13
		cy = 13
		cx += 1
	case 2:
		ay = 13
		cy = 12
	case 3:
		ay = 13
		cy = 13
		cx -= 1
	}
	bf.SetColor(puyoSet.Axis, ax, ay)
	bf.SetColor(puyoSet.Child, cx, cy)
	return true
}

func (obf *BitField) SearchWithPuyoSets(puyoSets []PuyoSet, hands []Hand, callback func(sr *SearchResult) bool, depth int) bool {
	if len(puyoSets) == 0 {
		return true
	}
	return obf.SearchPosition(puyoSets[0], hands, func(sr *SearchResult) bool {
		sr.Depth = depth
		if callback(sr) {
			return sr.RensaResult.BitField.SearchWithPuyoSets(puyoSets[1:], sr.Hands, callback, depth+1)
		}
		return false
	})
}

func (obf *BitField) SearchAllWithHandDepth(maxDepth int, callback func(sr *SearchResult) bool) {
	puyoSets := [10]PuyoSet{
		{Red, Red},
		{Red, Blue},
		{Red, Yellow},
		{Red, Green},
		{Blue, Blue},
		{Blue, Yellow},
		{Blue, Green},
		{Yellow, Yellow},
		{Yellow, Green},
		{Green, Green},
	}
	hands := []Hand{}
	for depth := 0; depth < maxDepth; depth++ {
		for _, puyoSet := range puyoSets {
			obf.SearchPosition(puyoSet, hands, func(sr *SearchResult) bool {
				sr.Depth = depth
				return callback(sr)
			})
		}
	}
}

func (obf *BitField) SearchAll1(callback func(sr *SearchResult) bool) {
	puyoSets := [10]PuyoSet{
		{Red, Red},
		{Red, Blue},
		{Red, Yellow},
		{Red, Green},
		{Blue, Blue},
		{Blue, Yellow},
		{Blue, Green},
		{Yellow, Yellow},
		{Yellow, Green},
		{Green, Green},
	}
	hands := []Hand{}
	for _, puyoSet := range puyoSets {
		obf.SearchPosition(puyoSet, hands, callback)
	}
}

func (obf *BitField) SearchPosition(puyoSet PuyoSet, hands []Hand, callback func(sr *SearchResult) bool) bool {
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

	for _, pos := range positions {
		bf := obf.Clone()
		if bf.placePuyo(puyoSet, pos) == false {
			continue
		}
		newHands := make([]Hand, len(hands))
		copy(newHands, hands)
		hand := new(Hand)
		hand.Position = pos
		hand.PuyoSet = puyoSet
		newHands = append(newHands, *hand)

		v := bf.Bits(Empty).MaskField12()
		bf.Drop(v)
		// result := bf.SimulateWithNewBitField()
		result := bf.Clone().SimulateDetail() // for erased puyo count
		searchResult := new(SearchResult)
		searchResult.BeforeSimulate = bf
		searchResult.Position = pos
		searchResult.RensaResult = result
		searchResult.Hands = newHands
		if callback(searchResult) == false {
			return false
		}
	}
	return true
}
