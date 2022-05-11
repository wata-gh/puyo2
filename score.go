package puyo2

func LongBonus(numErased int) int {
	longBonusTable := []int{
		0, 0, 0, 0, 0, 2, 3, 4, 5, 6, 7, 10,
	}
	if numErased > 11 {
		numErased = 11
	}
	return longBonusTable[numErased]
}

func ColorBonus(numColors int) int {
	colorBonusTable := []int{
		0, 0, 3, 6, 12, 24,
	}
	return colorBonusTable[numColors]
}

func RensaBonus(nthRensa int) int {
	rensaBonusTable := []int{
		0, 0, 8, 16, 32, 64, 96, 128, 160, 192,
		224, 256, 288, 320, 352, 384, 416, 448, 480, 512,
	}
	return rensaBonusTable[nthRensa]
}

func CalcRensaBonusCoef(rensaBonusCoef int, longBonusCoef int, colorBonusCoef int) int {
	coef := rensaBonusCoef + longBonusCoef + colorBonusCoef
	if coef == 0 {
		return 1
	}
	if coef > 999 {
		return 999
	}
	return coef
}
