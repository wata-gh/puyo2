package puyo2

type ShapeRensaResult struct {
	Chains        int
	Score         int
	Frames        int
	Erased        int
	Quick         bool
	ShapeBitField *ShapeBitField
}

func NewShapeRensaResult() *ShapeRensaResult {
	return new(ShapeRensaResult)
}

func NewShapeRensaResultWithResult(chains int, score int, frames int, quick bool, sbf *ShapeBitField) *ShapeRensaResult {
	r := new(ShapeRensaResult)
	r.Chains = chains
	r.Score = score
	r.Frames = frames
	r.Quick = quick
	r.ShapeBitField = sbf
	return r
}

func (srr *ShapeRensaResult) AddChain() {
	srr.Chains += 1
}

func (srr *ShapeRensaResult) AddErased(erased int) {
	srr.Erased += erased
}

func (srr *ShapeRensaResult) AddScore(score int) {
	srr.Score += score
}

func (srr *ShapeRensaResult) SetShapeBitField(sbf *ShapeBitField) {
	srr.ShapeBitField = sbf
}
func (srr *ShapeRensaResult) SetQuick(quick bool) {
	srr.Quick = quick
}
