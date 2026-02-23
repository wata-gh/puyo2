package puyo2

type IPSNazoJudgeMetrics struct {
	Remaining        [8]int `json:"remaining"`
	Erased           [8]int `json:"erased"`
	ChainCount       int    `json:"chainCount"`
	MatchedInSimul   bool   `json:"matchedInSimul"`
	MatchedByConnect bool   `json:"matchedByConnect"`
}

func EvaluateIPSNazoCondition(before *BitField, cond IPSNazoCondition) (bool, IPSNazoJudgeMetrics) {
	metrics := IPSNazoJudgeMetrics{}
	if before == nil {
		return false, metrics
	}

	bf := before.Clone()
	for {
		vanished := NewFieldBits()
		chainErased := [8]int{}
		chainPlaces := [8]int{}
		anyErased := false

		for _, color := range bf.Colors {
			colorIndex, ok := colorToIPSNazoIndex(color)
			if !ok {
				continue
			}
			vb := bf.Bits(color).MaskField12().FindVanishingBits()
			if vb.IsEmpty() {
				continue
			}
			anyErased = true
			vanished = vanished.Or(vb)
			vb.IterateBitWithMasking(func(seed *FieldBits) *FieldBits {
				group := seed.Expand(vb)
				size := group.PopCount()
				metrics.Erased[0] += size
				metrics.Erased[7] += size
				metrics.Erased[colorIndex] += size
				chainErased[0] += size
				chainErased[7] += size
				chainErased[colorIndex] += size
				chainPlaces[0]++
				chainPlaces[7]++
				chainPlaces[colorIndex]++
				if cond.Q0 == 52 || cond.Q0 == 53 {
					if matchConnectTarget(cond.Q1, colorIndex) {
						if (cond.Q0 == 52 && size == cond.Q2) || (cond.Q0 == 53 && size >= cond.Q2) {
							metrics.MatchedByConnect = true
						}
					}
				}
				return group
			})
		}

		if !anyErased {
			break
		}
		metrics.ChainCount++

		ojama := vanished.Expand1(bf.Bits(Ojama))
		ojamaCount := ojama.PopCount()
		if ojamaCount > 0 {
			vanished = vanished.Or(ojama)
			metrics.Erased[0] += ojamaCount
			metrics.Erased[6] += ojamaCount
			chainErased[0] += ojamaCount
			chainErased[6] += ojamaCount
		}

		switch cond.Q0 {
		case 40:
			if countErasedColors(chainErased) == cond.Q2 {
				metrics.MatchedInSimul = true
			}
		case 41:
			if countErasedColors(chainErased) >= cond.Q2 {
				metrics.MatchedInSimul = true
			}
		case 42:
			if validIPSNazoIndex(cond.Q1) && chainErased[cond.Q1] == cond.Q2 {
				metrics.MatchedInSimul = true
			}
		case 43:
			if validIPSNazoIndex(cond.Q1) && chainErased[cond.Q1] >= cond.Q2 {
				metrics.MatchedInSimul = true
			}
		case 44:
			if validIPSNazoIndex(cond.Q1) && chainPlaces[cond.Q1] == cond.Q2 {
				metrics.MatchedInSimul = true
			}
		case 45:
			if validIPSNazoIndex(cond.Q1) && chainPlaces[cond.Q1] >= cond.Q2 {
				metrics.MatchedInSimul = true
			}
		}

		bf.Drop(vanished)
	}

	metrics.Remaining = countRemainingPieces(bf)
	matched := false
	switch cond.Q0 {
	case 2:
		if validIPSNazoIndex(cond.Q1) {
			matched = metrics.Remaining[cond.Q1] == 0
		}
	case 10:
		matched = countErasedColors(metrics.Erased) == cond.Q2
	case 11:
		matched = countErasedColors(metrics.Erased) >= cond.Q2
	case 12:
		if validIPSNazoIndex(cond.Q1) {
			matched = metrics.Erased[cond.Q1] == cond.Q2
		}
	case 13:
		if validIPSNazoIndex(cond.Q1) {
			matched = metrics.Erased[cond.Q1] >= cond.Q2
		}
	case 30:
		matched = metrics.ChainCount == cond.Q2
	case 31:
		matched = metrics.ChainCount >= cond.Q2
	case 32:
		if validIPSNazoIndex(cond.Q1) {
			matched = metrics.ChainCount == cond.Q2 && metrics.Remaining[cond.Q1] == 0
		}
	case 33:
		if validIPSNazoIndex(cond.Q1) {
			matched = metrics.ChainCount >= cond.Q2 && metrics.Remaining[cond.Q1] == 0
		}
	case 40, 41, 42, 43, 44, 45:
		matched = metrics.MatchedInSimul
	case 52, 53:
		matched = metrics.MatchedByConnect
	default:
		matched = false
	}
	return matched, metrics
}

func countRemainingPieces(bf *BitField) [8]int {
	var remaining [8]int
	for y := 1; y <= 13; y++ {
		for x := 0; x < 6; x++ {
			switch bf.Color(x, y) {
			case Red:
				remaining[1]++
				remaining[7]++
			case Green:
				remaining[2]++
				remaining[7]++
			case Blue:
				remaining[3]++
				remaining[7]++
			case Yellow:
				remaining[4]++
				remaining[7]++
			case Purple:
				remaining[5]++
				remaining[7]++
			case Ojama:
				remaining[6]++
			}
		}
	}
	remaining[0] = remaining[7] + remaining[6]
	return remaining
}

func countErasedColors(counts [8]int) int {
	n := 0
	for i := 1; i <= 5; i++ {
		if counts[i] > 0 {
			n++
		}
	}
	return n
}

func colorToIPSNazoIndex(c Color) (int, bool) {
	switch c {
	case Red:
		return 1, true
	case Green:
		return 2, true
	case Blue:
		return 3, true
	case Yellow:
		return 4, true
	case Purple:
		return 5, true
	default:
		return 0, false
	}
}

func matchConnectTarget(target int, color int) bool {
	return target == 0 || target == 7 || target == color
}

func validIPSNazoIndex(idx int) bool {
	return 0 <= idx && idx <= 7
}
