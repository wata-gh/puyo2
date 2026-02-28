package puyo2

import "testing"

type searchBenchCase struct {
	name    string
	param   string
	puyoSet []PuyoSet
}

type searchBenchStateKey struct {
	Depth int
	Key   searchStateKey
}

func runSearchBench(b *testing.B, benchCase searchBenchCase, dedup DedupMode, policy SimulatePolicy) {
	var totalNodes int
	var totalUnique int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodes := 0
		unique := map[searchBenchStateKey]struct{}{}
		cond := NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan(benchCase.param), benchCase.puyoSet)
		cond.DedupMode = dedup
		cond.SimulatePolicy = policy
		cond.EachHandCallback = func(sr *SearchResult) bool {
			nodes++
			key := searchBenchStateKey{
				Depth: sr.Depth,
				Key:   createSearchStateKey(sr.RensaResult.BitField, false),
			}
			unique[key] = struct{}{}
			return true
		}
		cond.SearchWithPuyoSetsV2()
		totalNodes += nodes
		totalUnique += len(unique)
	}
	b.StopTimer()
	b.ReportMetric(float64(totalNodes)/float64(b.N), "nodes/op")
	b.ReportMetric(float64(totalUnique)/float64(b.N), "unique_states/op")
}

func BenchmarkSearchWithPuyoSetsV2Modes(b *testing.B) {
	benchCases := []searchBenchCase{
		{
			name:    "3_same",
			param:   "a78",
			puyoSet: []PuyoSet{{Red, Red}, {Red, Red}, {Red, Red}},
		},
		{
			name:    "3_mix",
			param:   "a78",
			puyoSet: []PuyoSet{{Red, Blue}, {Green, Blue}, {Red, Yellow}},
		},
		{
			name:    "4_same",
			param:   "a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16",
			puyoSet: []PuyoSet{{Red, Red}, {Red, Red}, {Red, Red}, {Red, Red}},
		},
	}
	for _, tc := range benchCases {
		tc := tc
		b.Run(tc.name+"/off/detail_always", func(b *testing.B) {
			runSearchBench(b, tc, DedupModeOff, SimulatePolicyDetailAlways)
		})
		b.Run(tc.name+"/same_pair_order/detail_always", func(b *testing.B) {
			runSearchBench(b, tc, DedupModeSamePairOrder, SimulatePolicyDetailAlways)
		})
		b.Run(tc.name+"/state/fast_intermediate", func(b *testing.B) {
			runSearchBench(b, tc, DedupModeState, SimulatePolicyFastIntermediate)
		})
	}
}
