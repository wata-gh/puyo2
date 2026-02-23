package puyo2

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const ipsNazoEncodeChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-?"

const ipsNazoDataModeChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ."

var ipsNazoConditionTemplates = map[int]string{
	0:  "条件を選択してください",
	2:  "cぷよ全て消す",
	10: "n色消す",
	11: "n色以上消す",
	12: "cぷよn個消す",
	13: "cぷよn個以上消す",
	30: "n連鎖する",
	31: "n連鎖以上する",
	32: "n連鎖＆cぷよ全て消す",
	33: "n連鎖以上＆cぷよ全て消す",
	40: "n色同時に消す",
	41: "n色以上同時に消す",
	42: "cぷよn個同時に消す",
	43: "cぷよn個以上同時に消す",
	44: "cぷよn箇所同時に消す",
	45: "cぷよn箇所以上同時に消す",
	52: "cぷよn連結で消す",
	53: "cぷよn連結以上で消す",
}

var ipsNazoConditionColors = map[int]string{
	1: "赤",
	2: "緑",
	3: "青",
	4: "黄",
	5: "紫",
	6: "おじゃま",
	7: "色",
}

type IPSNazoCondition struct {
	Q0       int
	Q1       int
	Q2       int
	Template string
	Color    string
	Text     string
}

type IPSNazoDecoded struct {
	InitialField  string
	Haipuyo       string
	Condition     IPSNazoCondition
	ConditionCode [3]int
}

func ParseIPSNazoURL(input string) (*IPSNazoDecoded, error) {
	query, err := normalizeIPSNazoQuery(input)
	if err != nil {
		return nil, err
	}
	segments := splitIPSNazoSegments(query)
	initialField, err := decodeIPSNazoField(segments[0])
	if err != nil {
		return nil, err
	}
	haipuyo, err := decodeIPSNazoHaipuyo(segments[1])
	if err != nil {
		return nil, err
	}
	cond, err := decodeIPSNazoCondition(segments[3])
	if err != nil {
		return nil, err
	}
	return &IPSNazoDecoded{
		InitialField:  initialField,
		Haipuyo:       haipuyo,
		Condition:     cond,
		ConditionCode: [3]int{cond.Q0, cond.Q1, cond.Q2},
	}, nil
}

func normalizeIPSNazoQuery(input string) (string, error) {
	query := strings.TrimSpace(input)
	if query == "" {
		return "", fmt.Errorf("input is empty")
	}
	if strings.Contains(query, "://") {
		u, err := url.Parse(query)
		if err != nil {
			return "", fmt.Errorf("invalid url: %w", err)
		}
		if u.RawQuery != "" {
			query = u.RawQuery
		}
	}
	if strings.HasPrefix(query, "?") {
		query = query[1:]
	}
	if idx := strings.Index(query, "?"); idx >= 0 {
		query = query[idx+1:]
	}
	if idx := strings.Index(query, "#"); idx >= 0 {
		query = query[:idx]
	}
	if strings.Contains(query, "%") {
		decoded, err := url.QueryUnescape(query)
		if err != nil {
			return "", fmt.Errorf("invalid query encoding: %w", err)
		}
		query = decoded
	}
	if query == "" {
		return "", fmt.Errorf("query is empty")
	}
	return query, nil
}

func splitIPSNazoSegments(query string) [4]string {
	var segments [4]string
	split := strings.Split(query, "_")
	for i := 0; i < len(segments) && i < len(split); i++ {
		segments[i] = split[i]
	}
	return segments
}

func decodeIPSNazoField(segment string) (string, error) {
	values, err := decodeIPSNazoFieldValues(segment)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i, v := range values {
		switch v {
		case 0:
			b.WriteByte('a')
		case 1:
			b.WriteByte('b')
		case 2:
			b.WriteByte('e')
		case 3:
			b.WriteByte('c')
		case 4:
			b.WriteByte('d')
		case 5:
			b.WriteByte('f')
		case 6:
			b.WriteByte('g')
		case 7, 8, 9:
			return "", fmt.Errorf("field contains unsupported cell value %d at index %d", v, i)
		default:
			return "", fmt.Errorf("field contains unknown cell value %d at index %d", v, i)
		}
	}
	return b.String(), nil
}

func decodeIPSNazoFieldValues(segment string) ([]int, error) {
	if strings.HasPrefix(segment, "~") || strings.HasPrefix(segment, "‾") {
		return decodeIPSNazoFieldDataMode(segment)
	}
	if err := validateChars(segment, ipsNazoEncodeChars); err != nil {
		return nil, fmt.Errorf("invalid field segment: %w", err)
	}
	for len(segment) < 39 {
		segment = "0" + segment
	}
	values := make([]int, 13*6)
	for y := 0; y <= 12; y++ {
		for x := 1; x <= 3; x++ {
			di := y*3 + (x - 1)
			if di > len(segment)-1 {
				continue
			}
			ci := strings.IndexByte(ipsNazoEncodeChars, segment[di])
			if ci < 0 || ci >= 64 {
				continue
			}
			base := y*6 + (x-1)*2
			values[base] = ci / 8
			values[base+1] = ci % 8
		}
	}
	return values, nil
}

func decodeIPSNazoFieldDataMode(segment string) ([]int, error) {
	data := segment
	if strings.HasPrefix(data, "~") {
		data = strings.TrimPrefix(data, "~")
	} else {
		data = strings.TrimPrefix(data, "‾")
	}
	if err := validateChars(data, ipsNazoDataModeChars); err != nil {
		return nil, fmt.Errorf("invalid field segment: %w", err)
	}
	var dt strings.Builder
	for _, item := range strings.Split(data, ".") {
		if item == "" {
			dt.WriteString("000000")
			continue
		}
		dt.WriteString(item)
		if rem := len(item) % 6; rem > 0 {
			dt.WriteString(strings.Repeat("0", 6-rem))
		}
	}
	body := dt.String()
	for len(body) < 78 {
		body = "0" + body
	}
	values := make([]int, 13*6)
	for y := 0; y <= 12; y++ {
		for x := 1; x <= 6; x++ {
			di := y*6 + (x - 1)
			values[di] = decodeIPSNazoDataModeCell(body[di])
		}
	}
	return values, nil
}

func decodeIPSNazoDataModeCell(c byte) int {
	switch c {
	case '1', 'r', 'R':
		return 1
	case '2', 'g', 'G':
		return 2
	case '3', 'b', 'B':
		return 3
	case '4', 'y', 'Y':
		return 4
	case '5', 'p', 'P':
		return 5
	case '6', 'o', 'O', 'j', 'J':
		return 6
	case '7', 'w', 'W':
		return 7
	case '8', 't', 'T', 'i', 'I':
		return 8
	case '9', 'k', 'K':
		return 9
	default:
		return 0
	}
}

func decodeIPSNazoHaipuyo(segment string) (string, error) {
	if err := validateChars(segment, ipsNazoEncodeChars); err != nil {
		return "", fmt.Errorf("invalid operation segment: %w", err)
	}
	routes := decodeIPSNazoOperationRoutes(segment)
	if len(routes) == 0 || routes[0] == "" {
		return "", nil
	}
	var b strings.Builder
	ops := strings.Split(routes[0], ",")
	for _, op := range ops {
		if len(op) < 2 {
			continue
		}
		axis := strings.IndexByte(ipsNazoEncodeChars, op[0])
		child := strings.IndexByte(ipsNazoEncodeChars, op[1])
		axisLetter, ok := ipsNazoColorIndexToLetter(axis)
		if !ok {
			continue
		}
		childLetter, ok := ipsNazoColorIndexToLetter(child)
		if !ok {
			continue
		}
		b.WriteString(axisLetter)
		b.WriteString(childLetter)
	}
	return b.String(), nil
}

func decodeIPSNazoOperationRoutes(segment string) []string {
	i := 1
	ol := []string{""}
	oli := -1
	olc := -1
	o := ""
	for (i * 2) <= len(segment) {
		di := (i - 1) * 2
		ci1 := strings.IndexByte(ipsNazoEncodeChars, segment[di])
		di++
		ci2 := strings.IndexByte(ipsNazoEncodeChars, segment[di])
		dt := ci1 % 2
		if dt != olc {
			if oli >= 0 {
				ol[oli] = ol[oli] + o
			}
			olc = dt
			oli++
			if oli >= len(ol) {
				ol = append(ol, "")
			}
			o = ""
		} else {
			if o != "" {
				ol[oli] = ol[oli] + o + ","
				o = ""
			}
		}
		dt = ci1 / 2
		if dt == 31 {
		} else if dt == 30 {
			o += charAtIPSNazoEncode(8)
			o += charAtIPSNazoEncode(ci2)
			o += charAtIPSNazoEncode(64)
			o += charAtIPSNazoEncode(64)
		} else {
			idx := dt%6 + 1
			if idx == 6 {
				o += charAtIPSNazoEncode(idx)
				o += charAtIPSNazoEncode(dt / 6)
				o += charAtIPSNazoEncode(ci2)
				o += charAtIPSNazoEncode(64)
			} else {
				o += charAtIPSNazoEncode(idx)
				o += charAtIPSNazoEncode(dt/6 + 1)
				idx = ci2 % 2
				dt = ci2 / 2
				if idx == 0 {
					idx = dt%6 + 1
				} else {
					idx = 7
				}
				o += charAtIPSNazoEncode(idx)
				o += charAtIPSNazoEncode(dt / 6)
			}
		}
		i++
	}
	if oli >= 0 {
		ol[oli] = ol[oli] + o
	}
	return ol
}

func charAtIPSNazoEncode(idx int) string {
	if idx < 0 || idx >= len(ipsNazoEncodeChars) {
		return ""
	}
	return string(ipsNazoEncodeChars[idx])
}

func ipsNazoColorIndexToLetter(idx int) (string, bool) {
	switch idx {
	case 1:
		return "r", true
	case 2:
		return "g", true
	case 3:
		return "b", true
	case 4:
		return "y", true
	case 5:
		return "p", true
	default:
		return "", false
	}
}

func decodeIPSNazoCondition(segment string) (IPSNazoCondition, error) {
	if err := validateChars(segment, ipsNazoEncodeChars); err != nil {
		return IPSNazoCondition{}, fmt.Errorf("invalid condition segment: %w", err)
	}
	codes := [3]int{}
	for i := 0; i < len(codes) && i < len(segment); i++ {
		ci := strings.IndexByte(ipsNazoEncodeChars, segment[i])
		if ci < 0 {
			ci = 0
		}
		codes[i] = ci
	}
	template := ipsNazoConditionTemplates[codes[0]]
	color := ipsNazoConditionColors[codes[1]]
	text := template
	if text != "" {
		text = strings.Replace(text, "c", color, 1)
		text = strings.Replace(text, "n", strconv.Itoa(codes[2]), 1)
	}
	return IPSNazoCondition{
		Q0:       codes[0],
		Q1:       codes[1],
		Q2:       codes[2],
		Template: template,
		Color:    color,
		Text:     text,
	}, nil
}

func validateChars(value string, allowed string) error {
	for i, r := range value {
		if strings.IndexRune(allowed, r) < 0 {
			return fmt.Errorf("contains unsupported character %q at index %d", r, i)
		}
	}
	return nil
}
