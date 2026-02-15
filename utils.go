package puyo2

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func Color2rbgyp(c Color) string {
	switch c {
	case Red:
		return "r"
	case Blue:
		return "b"
	case Green:
		return "g"
	case Yellow:
		return "y"
	case Purple:
		return "p"
	}
	panic(fmt.Sprintf("invalid color. %d", c))
}

func Rbygp2Color(puyo string) Color {
	switch puyo {
	case "r":
		return Red
	case "g":
		return Green
	case "y":
		return Yellow
	case "b":
		return Blue
	case "p":
		return Purple
	}
	panic("letter must be one of r,g,y,b,p but " + puyo)
}

func Deposit(x uint64, mask uint64) uint64 {
	var res uint64
	var next uint64 = 1
	for {
		lsb := mask & -mask
		if lsb == 0 {
			return res
		}
		mask ^= lsb
		if x&next != 0 {
			res |= lsb
		}
		next <<= 1
	}
}

func ExpandMattulwanParam(field string) string {
	if len(field) == 78 {
		hasNumber := false
		for _, r := range field {
			if '0' <= r && r <= '9' {
				hasNumber = true
				break
			}
		}
		if hasNumber == false {
			return field
		}
	}
	var b = strings.Builder{}
	r := regexp.MustCompile(`(([a-z])([0-9]*))?`)
	for _, v := range r.FindAllStringSubmatch(field, -1) {
		if v[3] == "" {
			b.WriteString(v[0])
		} else {
			len, _ := strconv.Atoi(v[3])
			for i := 0; i < len; i++ {
				b.WriteString(v[2])
			}
		}
	}
	return b.String()
}

func Extract(x uint64, mask uint64) uint64 {
	var res uint64
	var next uint64 = 1
	for {
		lsb := mask & -mask
		if lsb == 0 {
			return res
		}
		mask ^= lsb
		if x&lsb != 0 {
			res |= next
		}
		next <<= 1
	}
}

func Haipuyo2PuyoSets(haipuyo string) []*PuyoSet {
	var puyoSet *PuyoSet
	puyoSets := []*PuyoSet{}
	for i, p := range strings.Split(haipuyo, "") {
		if i%2 == 0 {
			puyoSet = &PuyoSet{}
			puyoSet.Axis = Rbygp2Color(p)
		} else {
			puyoSet.Child = Rbygp2Color(p)
			puyoSets = append(puyoSets, puyoSet)
		}
	}
	return puyoSets
}

func ToSimpleHands(hands []Hand) string {
	var b strings.Builder
	for _, hand := range hands {
		fmt.Fprintf(&b, "%s%s%d%d", Color2rbgyp(hand.PuyoSet.Axis), Color2rbgyp(hand.PuyoSet.Child), hand.Position[0], hand.Position[1])
	}
	return b.String()
}

func ToSimpleHands2(hands []*Hand) string {
	var b strings.Builder
	for _, hand := range hands {
		fmt.Fprintf(&b, "%s%s%d%d", Color2rbgyp(hand.PuyoSet.Axis), Color2rbgyp(hand.PuyoSet.Child), hand.Position[0], hand.Position[1])
	}
	return b.String()
}

func ParseSimpleHands(handsStr string) []Hand {
	var hands []Hand
	data := strings.Split(handsStr, "")
	for i := 0; i < len(data); i += 4 {
		axis := Rbygp2Color(data[i])
		child := Rbygp2Color(data[i+1])
		row, err := strconv.Atoi(data[i+2])
		if err != nil {
			panic(err)
		}
		dir, err := strconv.Atoi(data[i+3])
		if err != nil {
			panic(err)
		}
		hands = append(hands, Hand{PuyoSet: PuyoSet{Axis: axis, Child: child}, Position: [2]int{row, dir}})
	}
	return hands
}

func ParseSimpleHands2(handsStr string) []*Hand {
	var hands []*Hand
	data := strings.Split(handsStr, "")
	for i := 0; i < len(data); i += 4 {
		axis := Rbygp2Color(data[i])
		child := Rbygp2Color(data[i+1])
		row, err := strconv.Atoi(data[i+2])
		if err != nil {
			panic(err)
		}
		dir, err := strconv.Atoi(data[i+3])
		if err != nil {
			panic(err)
		}
		hands = append(hands, &Hand{PuyoSet: PuyoSet{Axis: axis, Child: child}, Position: [2]int{row, dir}})
	}
	return hands
}

var encodeChar = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ[]"

func convertBitDataToFcode(bitData []int) string {
	var fcode = ""
	for {
		index := 0
		num := 0
		for bit := 0; bit < 6; /* FCODE_CHAR_BIT */ bit++ {
			index = len(fcode)*6 + bit
			if index < len(bitData) {
				num += (bitData[index] << bit)
			}
		}
		fcode += strings.Split(encodeChar, "")[num]
		if index >= len(bitData)-1 {
			break
		}
	}
	return fcode
}

func getNextSettingFcode(hands []*Hand) string {
	handConv := map[Color]int{
		Red:    0,
		Green:  1,
		Blue:   2,
		Yellow: 3,
		Purple: 4,
	}
	puyoConvLen := 5
	puyoHistoryLen := len(hands)
	nextSettingData := []int{}
	oneData := 0
	for i, hand := range hands {
		pair := [2]int{handConv[hand.PuyoSet.Axis], handConv[hand.PuyoSet.Child]}
		settingNum := pair[0]*puyoConvLen + pair[1]
		oneData = settingNum
		if i < puyoHistoryLen {
			axis := hand.Position[0] + 1
			dir := hand.Position[1]
			historyNum := ((axis << 2) + dir)
			oneData |= (historyNum << 7)
		}
		nextSettingData = append(nextSettingData, oneData)
	}

	bitData := []int{}
	for i := 0; i < len(nextSettingData); i++ {
		for bit := 0; bit < 12; /* FCODE_NEXT_DATA_BIT */ bit++ {
			bitData = append(bitData, (nextSettingData[i]>>bit)&1)
		}
	}
	return "_" + convertBitDataToFcode(bitData)
}

func Hands2Puyop(hands []*Hand) string {
	return getNextSettingFcode(hands)
}
