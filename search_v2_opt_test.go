package puyo2

import (
	"fmt"
	"testing"
)

func countSearchNodes(param string, puyoSets []PuyoSet, dedup DedupMode, policy SimulatePolicy, stopOnChain bool) int {
	count := 0
	cond := NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan(param), puyoSets)
	cond.DedupMode = dedup
	cond.SimulatePolicy = policy
	cond.StopOnChain = stopOnChain
	cond.EachHandCallback = func(sr *SearchResult) bool {
		count++
		return true
	}
	cond.SearchWithPuyoSetsV2()
	return count
}

func collectTerminalResultSet(param string, puyoSets []PuyoSet, dedup DedupMode) map[string]struct{} {
	set := map[string]struct{}{}
	cond := NewSearchConditionWithBFAndPuyoSets(NewBitFieldWithMattulwan(param), puyoSets)
	cond.DedupMode = dedup
	cond.LastCallback = func(sr *SearchResult) {
		if sr == nil || sr.RensaResult == nil || sr.RensaResult.BitField == nil {
			return
		}
		rr := sr.RensaResult
		m := rr.BitField.M
		key := fmt.Sprintf(
			"%x:%x:%x:%x:%x:%x|c=%d|s=%d|e=%d|q=%t",
			m[0][0], m[0][1], m[1][0], m[1][1], m[2][0], m[2][1],
			rr.Chains, rr.Score, rr.Erased, rr.Quick,
		)
		set[key] = struct{}{}
	}
	cond.SearchWithPuyoSetsV2()
	return set
}

func TestSearchConditionDefaults(t *testing.T) {
	cond := NewSearchCondition()
	if cond.DedupMode != DedupModeOff {
		t.Fatalf("DedupMode default must be off but %s", cond.DedupMode.String())
	}
	if cond.SimulatePolicy != SimulatePolicyDetailAlways {
		t.Fatalf("SimulatePolicy default must be detail_always but %s", cond.SimulatePolicy.String())
	}
	if cond.StopOnChain {
		t.Fatalf("StopOnChain default must be false")
	}
}

func TestSearchWithPuyoSetsV2SamePairOrderReduction(t *testing.T) {
	tests := []struct {
		name  string
		param string
		sets  []PuyoSet
	}{
		{
			name:  "empty_3_same_rr",
			param: "a78",
			sets:  []PuyoSet{{Red, Red}, {Red, Red}, {Red, Red}},
		},
		{
			name:  "fixture_3_same_gr",
			param: "a46ea5ea5ea5ga5ea4eba",
			sets:  []PuyoSet{{Green, Red}, {Green, Red}, {Green, Red}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			off := countSearchNodes(tc.param, tc.sets, DedupModeOff, SimulatePolicyDetailAlways, true)
			samePair := countSearchNodes(tc.param, tc.sets, DedupModeSamePairOrder, SimulatePolicyDetailAlways, true)
			if off == 0 {
				t.Fatalf("off count must not be zero")
			}
			if float64(samePair) > float64(off)*0.5 {
				t.Fatalf("same_pair_order must reduce >=50%% with stop-on-chain=true. off=%d same_pair=%d", off, samePair)
			}
		})
	}
}

func TestSearchWithPuyoSetsV2SamePairOrderDisabledWithoutStopOnChain(t *testing.T) {
	tests := []struct {
		name  string
		param string
		sets  []PuyoSet
	}{
		{
			name:  "empty_3_same_rr",
			param: "a78",
			sets:  []PuyoSet{{Red, Red}, {Red, Red}, {Red, Red}},
		},
		{
			name:  "fixture_3_same_gr",
			param: "a46ea5ea5ea5ga5ea4eba",
			sets:  []PuyoSet{{Green, Red}, {Green, Red}, {Green, Red}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			off := countSearchNodes(tc.param, tc.sets, DedupModeOff, SimulatePolicyDetailAlways, false)
			samePair := countSearchNodes(tc.param, tc.sets, DedupModeSamePairOrder, SimulatePolicyDetailAlways, false)
			if off != samePair {
				t.Fatalf("same_pair_order must be disabled when stop-on-chain=false. off=%d same_pair=%d", off, samePair)
			}
		})
	}
}

func TestSearchWithPuyoSetsV2StateResultSetEqual(t *testing.T) {
	param := "a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16"
	puyoSets := []PuyoSet{
		{Red, Red},
		{Red, Red},
		{Blue, Blue},
		{Blue, Blue},
	}
	offSet := collectTerminalResultSet(param, puyoSets, DedupModeOff)
	stateSet := collectTerminalResultSet(param, puyoSets, DedupModeState)
	if len(offSet) != len(stateSet) {
		t.Fatalf("terminal result set size mismatch. off=%d state=%d", len(offSet), len(stateSet))
	}
	for key := range offSet {
		if _, found := stateSet[key]; found == false {
			t.Fatalf("state dedup missing terminal result key=%s", key)
		}
	}
}
