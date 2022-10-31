package puyo2

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func color2rbgyp(c Color) string {
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
	s := ""
	r := regexp.MustCompile(`(([a-z])([0-9]*))?`)
	for _, v := range r.FindAllStringSubmatch(field, -1) {
		if v[3] == "" {
			s += v[0]
		} else {
			len, _ := strconv.Atoi(v[3])
			for i := 0; i < len; i++ {
				s += v[2]
			}
		}
	}
	return s
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
		fmt.Fprintf(&b, "%s%s%d%d", color2rbgyp(hand.PuyoSet.Axis), color2rbgyp(hand.PuyoSet.Child), hand.Position[0], hand.Position[1])
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
