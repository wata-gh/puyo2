package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
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

func parseHands(handsStr string) []puyo2.Hand {
	var hands []puyo2.Hand
	hands = append(hands, puyo2.Hand{PuyoSet: puyo2.PuyoSet{Axis: puyo2.Green, Child: puyo2.Green}, Position: [2]int{1, 0}})
	hands = append(hands, puyo2.Hand{PuyoSet: puyo2.PuyoSet{Axis: puyo2.Red, Child: puyo2.Red}, Position: [2]int{1, 0}})
	return hands
}

func exp_hands(param string, handsStr string, path string) {
	bf := puyo2.NewBitFieldWithMattulwan(param)
	hands := parseHands(handsStr)
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
