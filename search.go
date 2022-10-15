package puyo2

import (
	"math/bits"
)

const ()

type SearchResult struct {
	RensaResult    *RensaResult
	BeforeSimulate *BitField
	Depth          int
	Position       [2]int
	PositionNum    int
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

func (bf *BitField) placePuyo(puyoSet PuyoSet, pos [2]int) (bool, bool) {
	// TODO check invalid placement and return false.
	ax := pos[0]
	cx := pos[0]

	empty := bf.Bits(Empty)

	heights := map[int]int{}
	for i := 0; i < 6; i++ {
		emptyBits := empty.ColBits(i)
		if i < 4 {
			emptyBits >>= 16 * i
		} else {
			emptyBits >>= (i - 4) * 16
		}
		emptyBits |= 0xC000
		heights[i] = 16 - bits.OnesCount64(emptyBits)
	}

	y := heights[ax] + 1
	ay := y
	cy := y + 1
	switch pos[1] {
	case 0:
	case 1:
		cx += 1
		cy = heights[cx] + 1
	case 2:
		ay = cy
		cy = y
	case 3:
		cx -= 1
		cy = heights[cx] + 1
	}

	x := 0
	if ax != 2 || cx != 2 {
		if ax > 2 || cx > 2 {
			if ax > cx {
				x = ax
			} else {
				x = cx
			}
		} else if ax < 2 || cx < 2 {
			if ax < cx {
				x = ax
			} else {
				x = cx
			}
		}
	}
	if x != 2 {
		// right side
		if x > 2 {
			for i := 3; i < x; i++ {
				// height 13 wall
				if heights[i] >= 13 {
					return false, false
				}
				if heights[i] == 12 {
					hasStep := heights[1] == 12 && heights[3] == 12 // hasama-chomu
					if hasStep == false {
						for j := i - 1; j >= 0; j-- {
							if heights[j] >= 12 {
								break
							}
							if heights[j] == 11 {
								hasStep = true
							}
						}
					}
					if hasStep == false {
						return false, false
					}
				}
			}
		} else { // left side
			for i := 1; i >= x; i-- {
				// height 13 wall
				if heights[i] >= 13 {
					return false, false
				}
				if heights[i] == 12 {
					hasStep := heights[1] == 12 && heights[3] == 12 // hasama-chomu
					if hasStep == false {
						for j := i + 1; j < len(heights); j++ {
							if heights[j] >= 12 {
								break
							}
							if heights[j] == 11 {
								hasStep = true
								break
							}
						}
					}
					if hasStep == false {
						return false, false
					} else {
						break
					}
				}
			}
		}
	}

	if ay > 13 {
		return false, false
	}
	// can't place puyo if 14th is used.
	if cy == 14 {
		if bf.Color(cx, cy) != Empty {
			return false, false
		}
	}
	bf.SetColor(puyoSet.Axis, ax, ay)
	bf.SetColor(puyoSet.Child, cx, cy)
	return true, ax != cx && ay != cy
}

func samePuyoSet(puyoSet1 PuyoSet, puyoSet2 PuyoSet) bool {
	if puyoSet1.Axis == puyoSet2.Axis && puyoSet1.Child == puyoSet2.Child {
		return true
	}
	if puyoSet1.Axis == puyoSet2.Child && puyoSet1.Child == puyoSet2.Axis {
		return true
	}
	return false
}

func (obf *BitField) SearchWithPuyoSets(puyoSets []PuyoSet, hands []Hand, callback func(sr *SearchResult) bool, depth int, posOffset int) bool {
	if len(puyoSets) == 0 {
		return true
	}
	return obf.SearchPosition(puyoSets[0], hands, func(sr *SearchResult) bool {
		sr.Depth = depth
		if sr.RensaResult.Chains != 0 {
			callback(sr)
			return true
		}
		posOffset := 0
		if len(puyoSets) > 1 && samePuyoSet(puyoSets[0], puyoSets[1]) {
			posOffset = sr.PositionNum
		}
		return sr.RensaResult.BitField.SearchWithPuyoSets(puyoSets[1:], sr.Hands, callback, depth+1, posOffset)
	}, posOffset)
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
			}, 0)
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
		obf.SearchPosition(puyoSet, hands, callback, 0)
	}
}

func overlap(pos1 [2]int, pos2 [2]int) bool {
	if pos1[0] == pos2[0] {
		return true
	}
	if (pos1[1] == 1) && ((pos2[0] == (pos1[0] + 1)) || (pos2[1] == 3 && pos2[0] == (pos1[0]+2))) {
		return true
	}
	if (pos1[1] == 3) && ((pos2[0] == (pos1[0] - 1)) || (pos2[1] == 1 && pos2[0] == (pos1[0]-2))) {
		return true
	}

	if (pos2[1] == 1) && ((pos1[0] == (pos2[0] + 1)) || (pos1[1] == 3 && pos1[0] == (pos2[0]+2))) {
		return true
	}
	if (pos2[1] == 3) && ((pos1[0] == (pos2[0] - 1)) || (pos1[1] == 1 && pos1[0] == (pos2[0]-2))) {
		return true
	}
	return false
}

func (obf *BitField) SearchPosition(puyoSet PuyoSet, hands []Hand, callback func(sr *SearchResult) bool, posOffset int) bool {
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
	if obf.Color(2, 12) != Empty {
		return true
	}

	for i, pos := range positions {
		if overlap(pos, positions[posOffset]) == false && i < posOffset {
			continue
		}
		bf := obf.Clone()
		placed, chigiri := bf.placePuyo(puyoSet, pos)
		if placed == false || chigiri {
			// fmt.Fprintf(os.Stderr, "placed: %v, chigiri: %v\n", placed, chigiri)
			continue
		}
		newHands := make([]Hand, len(hands))
		copy(newHands, hands)
		hand := new(Hand)
		hand.Position = pos
		hand.PuyoSet = puyoSet
		newHands = append(newHands, *hand)

		// result := bf.SimulateWithNewBitField()
		result := bf.Clone().SimulateDetail() // for erased puyo count
		searchResult := new(SearchResult)
		searchResult.BeforeSimulate = bf
		searchResult.Position = pos
		searchResult.PositionNum = i
		searchResult.RensaResult = result
		searchResult.Hands = newHands
		if callback(searchResult) == false {
			return false
		}
	}
	return true
}
