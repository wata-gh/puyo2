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

type Options struct {
	Trans    string
	Out      string
	Dir      string
	Hands    string
	NoBG     bool
	Simulate bool
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
	}
	panic("letter must be one of r,g,y,b")
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

func exp_hands(param string, handsStr string, path string) {
	bf := puyo2.NewBitFieldWithMattulwan(param)
	hands := parseSimpleHands(handsStr)
	bf.ExportHandsSimulateImage(hands, path)
}

func exp_simulate(param string, path string) {
	bf := puyo2.NewBitFieldWithMattulwan(param)
	bf.ExportSimulateImage(path)
}

func exp(param string, trans string, out string, nobg bool) {
	bf := puyo2.NewBitFieldWithMattulwan(param)
	bf.ShowDebug()
	tbf := puyo2.NewBitFieldWithMattulwan(trans).Bits(puyo2.Red)
	if nobg {
		bf.ExportOnlyPuyoImage(out)
	} else {
		bf.ExportImageWithTransparent(out, tbf)
	}
	fmt.Println(bf.MattulwanEditorUrl())
}

func run(param *string, opt *Options) {
	if opt.Simulate {
		exp_simulate(*param, opt.Dir)
	} else if opt.Hands != "" {
		exp_hands(*param, opt.Hands, opt.Dir)
	} else {
		path := opt.Out
		if opt.Out == "" {
			path = *param + ".png"
		}
		if opt.Dir != "" {
			path = opt.Dir + "/" + path
		}
		exp(*param, opt.Trans, path, opt.NoBG)
	}
}

func main() {
	param := flag.String("param", "a78", "puyofu")
	var opt Options
	flag.StringVar(&opt.Trans, "trans", "a78", "transparent field")
	flag.StringVar(&opt.Out, "out", "", "output file path")
	flag.StringVar(&opt.Dir, "dir", "", "output directory path")
	flag.StringVar(&opt.Hands, "hands", "", "hands")
	flag.BoolVar(&opt.NoBG, "nobg", false, "don't draw background")
	flag.BoolVar(&opt.Simulate, "simulate", false, "simulate")
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
