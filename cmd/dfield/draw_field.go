package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wata-gh/puyo2"
)

type options struct {
	Trans     string
	Out       string
	Dir       string
	Hands     string
	NoBG      bool
	Simulate  bool
	ShapeOnly bool
	Table     string
}

func letter2Color(puyo string) puyo2.Color {
	switch puyo {
	case "r":
		return puyo2.Red
	case "g":
		return puyo2.Green
	case "y":
		return puyo2.Yellow
	case "b":
		return puyo2.Blue
	case "p":
		return puyo2.Purple
	}
	panic("letter must be one of r,g,y,b,p but " + puyo)
}

func parseSimpleHands(handsStr string) []puyo2.Hand {
	var hands []puyo2.Hand
	data := strings.Split(handsStr, "")
	for i := 0; i < len(data); i += 4 {
		axis := letter2Color(data[i])
		child := letter2Color(data[i+1])
		row, err := strconv.Atoi(data[i+2])
		if err != nil {
			panic(err)
		}
		dir, err := strconv.Atoi(data[i+3])
		if err != nil {
			panic(err)
		}
		hands = append(hands, puyo2.Hand{PuyoSet: puyo2.PuyoSet{Axis: axis, Child: child}, Position: [2]int{row, dir}})
	}
	return hands
}

func expHands(param string, handsStr string, path string) {
	bf := puyo2.NewBitFieldWithMattulwanC(param)
	hands := parseSimpleHands(handsStr)
	for _, hand := range hands {
		c := bf.Table[hand.PuyoSet.Axis]
		if c == puyo2.Empty {
			bf.Table[hand.PuyoSet.Axis] = hand.PuyoSet.Axis
		}
		c = bf.Table[hand.PuyoSet.Child]
		if c == puyo2.Empty {
			bf.Table[hand.PuyoSet.Child] = hand.PuyoSet.Child
		}
	}
	if bf.Table[puyo2.Purple] == puyo2.Purple {
		for _, c := range []puyo2.Color{puyo2.Red, puyo2.Blue, puyo2.Yellow, puyo2.Green} {
			if bf.Table[c] == puyo2.Empty {
				bf.Table[puyo2.Purple] = c
			}
		}
	}
	bf.ExportHandsSimulateImage(hands, path)
}

func expSimulate(param string, path string) {
	bf := puyo2.NewBitFieldWithMattulwanC(param)
	bf.ExportSimulateImage(path)
}

func exp(param string, trans string, out string, nobg bool) {
	bf := puyo2.NewBitFieldWithMattulwanC(param)
	bf.ShowDebug()
	tbf := puyo2.NewBitFieldWithMattulwan(trans).Bits(puyo2.Red)
	if nobg {
		bf.ExportOnlyPuyoImage(out)
	} else {
		bf.ExportImageWithTransparent(out, tbf)
	}
	fmt.Println(bf.MattulwanEditorUrl())
}

func expShape(param string, out string) {
	sbf := puyo2.NewShapeBitFieldWithFieldString(param)
	sbf.ShowDebug()
	sbf.ExportImage(out)
}

func expShapeSimulate(param string, out string) {
	sbf := puyo2.NewShapeBitFieldWithFieldString(param)
	sbf.ShowDebug()
	sbf.ExportChainImage(out)
}

func run(param *string, opt *options) {
	if opt.Dir != "" {
		os.Mkdir(opt.Dir, 0755)
	}
	if opt.Simulate {
		if opt.ShapeOnly {
			expShapeSimulate(*param, opt.Dir)
		} else {
			expSimulate(*param, opt.Dir)
		}
	} else if opt.Hands != "" {
		expHands(*param, opt.Hands, opt.Dir)
	} else {
		path := opt.Out
		if opt.Out == "" {
			path = *param + ".png"
		}
		if opt.Dir != "" {
			path = opt.Dir + "/" + path
		}
		if opt.ShapeOnly {
			expShape(*param, path)
		} else {
			exp(*param, opt.Trans, path, opt.NoBG)
		}
	}
}

func main() {
	param := flag.String("param", "a78", "puyofu")
	var opt options
	flag.StringVar(&opt.Trans, "trans", "a78", "transparent field")
	flag.StringVar(&opt.Out, "out", "", "output file path")
	flag.StringVar(&opt.Dir, "dir", "", "output directory path")
	flag.StringVar(&opt.Hands, "hands", "", "hands")
	flag.StringVar(&opt.Table, "table", "", "table")
	flag.BoolVar(&opt.NoBG, "nobg", false, "don't draw background")
	flag.BoolVar(&opt.Simulate, "simulate", false, "simulate")
	flag.BoolVar(&opt.ShapeOnly, "shape-only", false, "use shape only")
	flag.Parse()

	if *param != "a78" {
		run(param, &opt)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			param := strings.TrimRight(scanner.Text(), "\n")
			run(&param, &opt)
		}
	}
}
