package puyo2

type SearchResult struct {
	RensaResult    *RensaResult
	BeforeSimulate *BitField
	Depth          int
	Position       [2]int
}

func (bf *BitField) placePuyo(puyoSet [2]Color, pos [2]int) bool {
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
	bf.SetColor(puyoSet[0], ax, ay)
	bf.SetColor(puyoSet[1], cx, cy)
	return true
}

func (obf *BitField) SearchWithPuyoSets(puyoSets [][2]Color, callback func(rr *SearchResult) bool, depth int) {
	if len(puyoSets) == 0 {
		return
	}
	obf.SearchPosition(puyoSets[0], func(sr *SearchResult) bool {
		sr.Depth = depth
		if callback(sr) {
			sr.RensaResult.BitField.SearchWithPuyoSets(puyoSets[1:], callback, depth+1)
		}
		return true
	})
}

func (obf *BitField) SearchAllWithHandDepth(maxDepth int, callback func(rr *SearchResult) bool) {
	puyoSets := [10][2]Color{
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
	for depth := 0; depth < maxDepth; depth++ {
		for _, puyoSet := range puyoSets {
			obf.SearchPosition(puyoSet, func(sr *SearchResult) bool {
				sr.Depth = depth
				return callback(sr)
			})
		}
	}
}

func (obf *BitField) SearchAll1(callback func(rr *SearchResult) bool) {
	puyoSets := [10][2]Color{
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
	for _, puyoSet := range puyoSets {
		obf.SearchPosition(puyoSet, callback)
	}
}

func (obf *BitField) SearchPosition(puyoSet [2]Color, callback func(rr *SearchResult) bool) {
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
		bf.placePuyo(puyoSet, pos)
		v := bf.Bits(Empty).MaskField12()
		bf.Drop(v)
		result := bf.SimulateWithNewBitField()
		searchResult := new(SearchResult)
		searchResult.BeforeSimulate = bf
		searchResult.Position = pos
		searchResult.RensaResult = result
		callback(searchResult)
	}
}
