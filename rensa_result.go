package puyo2

import "fmt"

type NthResult struct {
	Nth         int
	ErasedPuyos []SingleResult
}

type SingleResult struct {
	Color     Color
	Connected int
}

type RensaResult struct {
	Chains     int
	Score      int
	Frames     int
	Erased     int
	Quick      bool
	BitField   *BitField
	NthResults []*NthResult
}

func NewNthResult(nth int) *NthResult {
	r := new(NthResult)
	r.Nth = nth
	return r
}

func NewRensaResult() *RensaResult {
	return new(RensaResult)
}

func NewRensaResultWithResult(chains int, score int, frames int, quick bool, bf *BitField) *RensaResult {
	r := new(RensaResult)
	r.Chains = chains
	r.Score = score
	r.Frames = frames
	r.Quick = quick
	r.BitField = bf
	return r
}

func (rr *RensaResult) AddChain() {
	rr.Chains += 1
}

func (rr *RensaResult) AddErased(erased int) {
	rr.Erased += erased
}

func (rr *RensaResult) AddScore(score int) {
	rr.Score += score
}

func (rr *RensaResult) SetBitField(bf *BitField) {
	rr.BitField = bf
}

func (rr *RensaResult) SetQuick(quick bool) {
	rr.Quick = quick
}

func (rr *RensaResult) NthResult(n int) *NthResult {
	for _, nth := range rr.NthResults {
		if nth.Nth == n {
			return nth
		}
	}
	panic(fmt.Sprintf("no %d th result. ", +n))
}
