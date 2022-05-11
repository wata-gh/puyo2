package puyo2

type RensaResult struct {
	Chains   int
	Score    int
	Frames   int
	Quick    bool
	BitField *BitField
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

func (rr *RensaResult) AddScore(score int) {
	rr.Score += score
}

func (rr *RensaResult) SetBitField(bf *BitField) {
	rr.BitField = bf
}
func (rr *RensaResult) SetQuick(quick bool) {
	rr.Quick = quick
}
