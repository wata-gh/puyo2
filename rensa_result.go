package puyo2

import "fmt"

type NthResult struct {
	Nth         int            `json:"nth"`
	ErasedPuyos []SingleResult `json:"erasedPuyos"`
}

type SingleResult struct {
	Color     Color `json:"color"`
	Connected int   `json:"connected"`
}

type RensaResult struct {
	Chains      int          `json:"chains"`
	Chigiris    int          `json:"chigiris"`
	Score       int          `json:"score"`
	RensaFrames int          `json:"rensaFrames"`
	SetFrames   int          `json:"setFrames"`
	Erased      int          `json:"erased"`
	Quick       bool         `json:"quick"`
	BitField    *BitField    `json:"bitFeild"`
	NthResults  []*NthResult `json:"nthResults"`
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
	r.RensaFrames = frames
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
