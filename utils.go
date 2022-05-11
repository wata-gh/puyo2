package puyo2

import (
	"regexp"
	"strconv"
)

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
