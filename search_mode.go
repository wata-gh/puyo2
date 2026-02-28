package puyo2

import (
	"fmt"
	"strings"
)

type DedupMode int

const (
	DedupModeOff DedupMode = iota
	DedupModeSamePairOrder
	DedupModeState
	DedupModeStateMirror
)

func (m DedupMode) String() string {
	switch m {
	case DedupModeOff:
		return "off"
	case DedupModeSamePairOrder:
		return "same_pair_order"
	case DedupModeState:
		return "state"
	case DedupModeStateMirror:
		return "state_mirror"
	default:
		return fmt.Sprintf("unknown(%d)", int(m))
	}
}

func ParseDedupMode(v string) (DedupMode, error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "", "off":
		return DedupModeOff, nil
	case "same_pair_order":
		return DedupModeSamePairOrder, nil
	case "state":
		return DedupModeState, nil
	case "state_mirror":
		return DedupModeStateMirror, nil
	default:
		return DedupModeOff, fmt.Errorf("unknown dedup mode %q", v)
	}
}

type SimulatePolicy int

const (
	SimulatePolicyDetailAlways SimulatePolicy = iota
	SimulatePolicyFastIntermediate
	SimulatePolicyFastAlways
)

func (p SimulatePolicy) String() string {
	switch p {
	case SimulatePolicyDetailAlways:
		return "detail_always"
	case SimulatePolicyFastIntermediate:
		return "fast_intermediate"
	case SimulatePolicyFastAlways:
		return "fast_always"
	default:
		return fmt.Sprintf("unknown(%d)", int(p))
	}
}

func ParseSimulatePolicy(v string) (SimulatePolicy, error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "", "detail_always":
		return SimulatePolicyDetailAlways, nil
	case "fast_intermediate":
		return SimulatePolicyFastIntermediate, nil
	case "fast_always":
		return SimulatePolicyFastAlways, nil
	default:
		return SimulatePolicyDetailAlways, fmt.Errorf("unknown simulate policy %q", v)
	}
}
