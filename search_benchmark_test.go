package puyo2

import "testing"

var (
	benchmarkSearchFramesSink int
	benchmarkSearchLeafSink   int
)

func BenchmarkSearchPlacementForPos(b *testing.B) {
	puyoSet := PuyoSet{Axis: Red, Child: Blue}
	cases := []struct {
		name string
		bf   *BitField
	}{
		{
			name: "EmptyField",
			bf:   NewBitField(),
		},
		{
			name: "DenseField",
			bf:   NewBitFieldWithMattulwan("a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16"),
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			sumFrames := 0
			positionsLen := len(SetupPositions)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pos := SetupPositions[i%positionsLen]
				placement := tc.bf.SearchPlacementForPos(&puyoSet, pos)
				if placement != nil {
					sumFrames += placement.Frames
				}
			}
			benchmarkSearchFramesSink = sumFrames
		})
	}
}

func BenchmarkSearchPositionV2(b *testing.B) {
	cond := NewSearchConditionWithBFAndPuyoSets(
		NewBitFieldWithMattulwan("a62gacbagecb2ae2g3"),
		[]PuyoSet{{Axis: Red, Child: Blue}},
	)
	leafCount := 0
	totalLeaves := 0
	cond.LastCallback = func(sr *SearchResult) {
		leafCount++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		leafCount = 0
		cond.SearchPositionV2([]Hand{})
		totalLeaves += leafCount
	}
	benchmarkSearchLeafSink = totalLeaves
}

func BenchmarkSearchWithPuyoSetsV2(b *testing.B) {
	cases := []struct {
		name     string
		field    string
		puyoSets []PuyoSet
	}{
		{
			name:  "Depth2",
			field: "a62gacbagecb2ae2g3",
			puyoSets: []PuyoSet{
				{Axis: Red, Child: Blue},
				{Axis: Green, Child: Blue},
			},
		},
		{
			name:  "Depth4",
			field: "a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16",
			puyoSets: []PuyoSet{
				{Axis: Red, Child: Red},
				{Axis: Red, Child: Red},
				{Axis: Blue, Child: Blue},
				{Axis: Blue, Child: Blue},
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			cond := NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan(tc.field), tc.puyoSets)
			leafCount := 0
			totalLeaves := 0
			cond.LastCallback = func(sr *SearchResult) {
				leafCount++
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				leafCount = 0
				cond.SearchWithPuyoSetsV2()
				totalLeaves += leafCount
			}
			benchmarkSearchLeafSink = totalLeaves
		})
	}
}

func BenchmarkSearchWithPuyoSetsV2Pruned(b *testing.B) {
	puyoSets := []PuyoSet{
		{Axis: Red, Child: Red},
		{Axis: Red, Child: Red},
		{Axis: Blue, Child: Blue},
		{Axis: Blue, Child: Blue},
	}
	maxDepth := len(puyoSets)
	cond := NewSearchConditionWithBFAndPuyoSets(
		NewBitFieldWithMattulwan("a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16"),
		puyoSets,
	)
	cond.EachHandCallback = func(sr *SearchResult) bool {
		if sr.Depth != maxDepth && sr.RensaResult.Chains != 0 {
			return false
		}
		return true
	}

	leafCount := 0
	totalLeaves := 0
	cond.LastCallback = func(sr *SearchResult) {
		leafCount++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		leafCount = 0
		cond.SearchWithPuyoSetsV2()
		totalLeaves += leafCount
	}
	benchmarkSearchLeafSink = totalLeaves
}
