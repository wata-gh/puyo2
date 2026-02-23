package puyo2

import "fmt"

type searchStateKey struct {
	M        [3][2]uint64
	TableSig uint32
}

type SearchCondition struct {
	DisableChigiri   bool
	ChigiriableCount int
	Chigiris         int
	SetFrames        int
	DedupMode        DedupMode
	SimulatePolicy   SimulatePolicy
	StopOnChain      bool
	PuyoSets         []PuyoSet
	BitField         *BitField
	LastCallback     func(sr *SearchResult)
	EachHandCallback func(sr *SearchResult) bool
	visitedStates    map[int]map[searchStateKey]struct{}
}

func NewSearchCondition() *SearchCondition {
	cond := new(SearchCondition)
	cond.DisableChigiri = false
	cond.DedupMode = DedupModeOff
	cond.SimulatePolicy = SimulatePolicyDetailAlways
	cond.StopOnChain = false
	return cond
}

func NewSearchConditionWithBFAndPuyoSets(bf *BitField, puyoSets []PuyoSet) *SearchCondition {
	cond := NewSearchCondition()
	cond.BitField = bf
	cond.PuyoSets = puyoSets
	return cond
}

func (cond *SearchCondition) SearchWithPuyoSetsV2() {
	cond.visitedStates = nil
	cond.ensureRuntimeState()
	cond.searchWithPuyoSetsV2([]Hand{}, nil, 0)
}

func (cond *SearchCondition) searchWithPuyoSetsV2(hands []Hand, searchResult *SearchResult, posOffset int) {
	if len(cond.PuyoSets) == 0 {
		if cond.LastCallback != nil {
			cond.LastCallback(searchResult)
		}
		return
	}
	cond.SearchPositionV2(hands, posOffset)
}

func (cond *SearchCondition) SearchPositionV2(hands []Hand, posOffset int) {
	puyoSet := cond.PuyoSets[0]
	// axis=x child=y
	// 0:  1:    2:  3:
	// y   x y   x   y x
	// x         y
	positions := SetupPositions
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
	heights := cond.BitField.createHeightsArray()
	checkPosOffset := cond.useSamePairOrderDedup() && posOffset > 0 && posOffset < len(positions)
	isTerminalDepth := len(cond.PuyoSets) == 1

	for i, pos := range positions {
		if checkPosOffset && overlap(pos, positions[posOffset]) == false && i < posOffset {
			continue
		}
		placement := cond.BitField.searchPlacementForPosWithHeights(&puyoSet, pos, heights)
		if placement == nil || (cond.DisableChigiri && placement.Chigiri) {
			continue
		}
		bf := cond.BitField.Clone()
		if bf.PlacePuyoWithPlacement(placement) == false {
			panic(fmt.Sprintf("should be able to place. %+v\n", placement))
		}
		newHands := make([]Hand, len(hands))
		copy(newHands, hands)
		hand := new(Hand)
		hand.Position = pos
		hand.PuyoSet = puyoSet
		newHands = append(newHands, *hand)

		result := cond.simulateForNode(bf, isTerminalDepth)
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

		remaining := len(cond.PuyoSets) - 1
		if cond.useStateDedup() && remaining > 0 && cond.markVisited(remaining, result.BitField) {
			continue
		}
		if cond.EachHandCallback != nil && cond.EachHandCallback(searchResult) == false {
			continue
		}
		if cond.StopOnChain && searchResult.RensaResult.Chains > 0 {
			continue
		}

		nextPosOffset := 0
		if cond.useSamePairOrderDedup() && len(cond.PuyoSets) > 1 && samePuyoSet(cond.PuyoSets[0], cond.PuyoSets[1]) {
			nextPosOffset = i
		}
		newCond := new(SearchCondition)
		newCond.PuyoSets = cond.PuyoSets[1:]
		newCond.DisableChigiri = cond.DisableChigiri
		newCond.ChigiriableCount = cond.ChigiriableCount
		newCond.Chigiris = searchResult.RensaResult.Chigiris
		newCond.SetFrames = searchResult.RensaResult.SetFrames
		newCond.DedupMode = cond.DedupMode
		newCond.SimulatePolicy = cond.SimulatePolicy
		newCond.StopOnChain = cond.StopOnChain
		newCond.BitField = result.BitField
		newCond.LastCallback = cond.LastCallback
		newCond.EachHandCallback = cond.EachHandCallback
		newCond.visitedStates = cond.visitedStates
		newCond.searchWithPuyoSetsV2(newHands, searchResult, nextPosOffset)
	}
}

func (cond *SearchCondition) ensureRuntimeState() {
	if cond.useStateDedup() && cond.visitedStates == nil {
		cond.visitedStates = map[int]map[searchStateKey]struct{}{}
	}
}

func (cond *SearchCondition) useSamePairOrderDedup() bool {
	return cond.DedupMode == DedupModeSamePairOrder && cond.StopOnChain
}

func (cond *SearchCondition) useStateDedup() bool {
	return cond.DedupMode == DedupModeState || cond.DedupMode == DedupModeStateMirror
}

func (cond *SearchCondition) useMirrorStateDedup() bool {
	return cond.DedupMode == DedupModeStateMirror
}

func (cond *SearchCondition) markVisited(remaining int, bf *BitField) bool {
	if cond.useStateDedup() == false {
		return false
	}
	cond.ensureRuntimeState()
	visited := cond.visitedStates[remaining]
	if visited == nil {
		visited = map[searchStateKey]struct{}{}
		cond.visitedStates[remaining] = visited
	}
	key := createSearchStateKey(bf, cond.useMirrorStateDedup())
	if _, found := visited[key]; found {
		return true
	}
	visited[key] = struct{}{}
	return false
}

func createSearchStateKey(bf *BitField, withMirror bool) searchStateKey {
	m := bf.M
	if withMirror {
		flip := mirrorBitFieldMatrix(m)
		if lessBitFieldMatrix(flip, m) {
			m = flip
		}
	}
	return searchStateKey{
		M:        m,
		TableSig: colorTableSignature(bf.Table),
	}
}

func lessBitFieldMatrix(a [3][2]uint64, b [3][2]uint64) bool {
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(a[i]); j++ {
			if a[i][j] == b[i][j] {
				continue
			}
			return a[i][j] < b[i][j]
		}
	}
	return false
}

func mirrorBitFieldMatrix(m [3][2]uint64) [3][2]uint64 {
	flip := [3][2]uint64{{0, 0}, {0, 0}, {0, 0}}
	for i := 0; i < 3; i++ {
		flip[i][1] = (m[i][0] & 0xffff) << 16
		flip[i][1] |= (m[i][0] & 0xffff0000) >> 16
		flip[i][0] = (m[i][0] & 0xffff00000000) << 16
		flip[i][0] |= (m[i][0] & 0xffff000000000000) >> 16
		flip[i][0] |= (m[i][1] & 0xffff) << 16
		flip[i][0] |= (m[i][1] & 0xffff0000) >> 16
	}
	return flip
}

func colorTableSignature(table map[Color]Color) uint32 {
	if table == nil {
		return 0
	}
	sig := uint32(0)
	colors := [...]Color{Red, Blue, Yellow, Green, Purple}
	for i, c := range colors {
		sig |= uint32(table[c]&0xf) << (i * 4)
	}
	return sig
}

func (cond *SearchCondition) simulateForNode(bf *BitField, terminal bool) *RensaResult {
	needsDetail := terminal || cond.EachHandCallback != nil

	switch cond.SimulatePolicy {
	case SimulatePolicyFastAlways:
		if needsDetail {
			return bf.CloneForSimulation().SimulateDetail()
		}
		return bf.CloneForSimulation().Simulate()
	case SimulatePolicyFastIntermediate:
		if needsDetail {
			return bf.CloneForSimulation().SimulateDetail()
		}
		return bf.CloneForSimulation().Simulate()
	default:
		return bf.CloneForSimulation().SimulateDetail()
	}
}
