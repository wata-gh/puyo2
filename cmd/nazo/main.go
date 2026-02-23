package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wata-gh/puyo2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type options struct {
	AllClear       bool
	ChainsEQ       int
	ChainsGT       int
	ClearColor     puyo2.Color
	Color          puyo2.Color
	N              int
	DisableChigiri bool
	StopOnChain    bool
	DedupMode      puyo2.DedupMode
	SimulatePolicy puyo2.SimulatePolicy
	Hands          string
}

var opt options
var searchedCount []int
var allSearchCnt float64

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
	panic("letter must be one of r,g,y,b,p")
}

func color2Letter(c puyo2.Color) string {
	switch c {
	case puyo2.Red:
		return "r"
	case puyo2.Green:
		return "g"
	case puyo2.Yellow:
		return "y"
	case puyo2.Blue:
		return "b"
	case puyo2.Purple:
		return "p"
	}
	panic("color must be one of r,g,y,b,p")
}

func hand2Str(hand puyo2.Hand) string {
	return fmt.Sprintf("%s%s%d%d", color2Letter(hand.PuyoSet.Axis), color2Letter(hand.PuyoSet.Child), hand.Position[0], hand.Position[1])
}

func hands2Str(hands []puyo2.Hand) string {
	var b strings.Builder
	for _, hand := range hands {
		fmt.Fprint(&b, hand2Str(hand))
	}
	return b.String()
}

func parsePuyoSets(handsStr string) []puyo2.PuyoSet {
	var puyoSets []puyo2.PuyoSet
	data := strings.Split(handsStr, "")
	for i := 0; i < len(data); i += 2 {
		axis := letter2Color(data[i])
		child := letter2Color(data[i+1])
		puyoSets = append(puyoSets, puyo2.PuyoSet{Axis: axis, Child: child})
	}
	return puyoSets
}

func search(num int, field chan string, puyoSets []puyo2.PuyoSet, wg *sync.WaitGroup) {
	for {
		param := <-field
		if param == "" {
			break
		}
		params := strings.Split(param, " ")
		cond := puyo2.NewSearchConditionWithBFAndPuyoSets(puyo2.NewBitFieldWithMattulwanC(params[0]), puyoSets)
		cond.DisableChigiri = opt.DisableChigiri
		cond.DedupMode = opt.DedupMode
		cond.SimulatePolicy = opt.SimulatePolicy
		cond.StopOnChain = opt.StopOnChain
		cond.EachHandCallback = func(sr *puyo2.SearchResult) bool {
			searchedCount[num]++
			if opt.AllClear && sr.RensaResult.BitField.IsEmpty() {
				fmt.Printf("%s %s %+v\n", sr.BeforeSimulate.MattulwanEditorUrl(), params[1]+hands2Str(sr.Hands), sr.RensaResult)
			}
			if opt.ClearColor != puyo2.Empty && sr.RensaResult.BitField.Bits(opt.ClearColor).IsEmpty() {
				fmt.Printf("%s %s %+v\n", sr.BeforeSimulate.MattulwanEditorUrl(), params[1]+hands2Str(sr.Hands), sr.RensaResult)
			}
			if opt.ChainsEQ > 0 && sr.RensaResult.Chains == opt.ChainsEQ {
				fmt.Printf("%s %s %+v\n", sr.BeforeSimulate.MattulwanEditorUrl(), params[1]+hands2Str(sr.Hands), sr.RensaResult)
			}
			if opt.ChainsGT > 0 && sr.RensaResult.Chains >= opt.ChainsGT {
				fmt.Printf("%s %s %+v\n", sr.BeforeSimulate.MattulwanEditorUrl(), params[1]+hands2Str(sr.Hands), sr.RensaResult)
			}
			if opt.N > 0 && opt.Color != puyo2.Empty {
				if sr.RensaResult.Chains > 0 {
					for n := 1; n <= sr.RensaResult.Chains; n++ {
						nth := sr.RensaResult.NthResult(n)
						erasedCount := 0
						for _, erased := range nth.ErasedPuyos {
							if erased.Color == opt.Color {
								erasedCount += erased.Connected
							}
						}
						if erasedCount == opt.N {
							fmt.Printf("%s %s %+v\n", sr.BeforeSimulate.MattulwanEditorUrl(), params[1]+hands2Str(sr.Hands), sr.RensaResult)
						}
					}
				}
			}
			return true
		}
		cond.SearchWithPuyoSetsV2()
	}
	wg.Done()
}

func run(param *string, opt *options) {
	fmt.Printf("%+v\n", opt)
	field := make(chan string)
	cpus := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	fmt.Println("cpus:", cpus)
	puyoSets := parsePuyoSets(opt.Hands)
	allSearchCnt = math.Pow(22, float64(len(puyoSets)))
	searchedCount = append(searchedCount, 0)

	if len(puyoSets) == 1 {
		searchedCount = append(searchedCount, 0)
		wg.Add(1)
		go search(1, field, puyoSets, wg)
		field <- fmt.Sprintf("%s ", *param)
		field <- ""
	} else {
		for i := 0; i < cpus; i++ {
			searchedCount = append(searchedCount, 0)
			wg.Add(1)
			go search(i+1, field, puyoSets[1:], wg)
		}
		cond := puyo2.NewSearchConditionWithBFAndPuyoSets(puyo2.NewBitFieldWithMattulwanC(*param), puyoSets[0:1])
		cond.DisableChigiri = opt.DisableChigiri
		cond.DedupMode = opt.DedupMode
		cond.SimulatePolicy = opt.SimulatePolicy
		cond.StopOnChain = opt.StopOnChain
		cond.BitField.ShowDebug()

		cond.LastCallback = func(sr *puyo2.SearchResult) {
			searchedCount[0]++
			param := fmt.Sprintf("%s %s", sr.RensaResult.BitField.MattulwanEditorParam(), hands2Str(sr.Hands))
			field <- param
		}
		cond.SearchWithPuyoSetsV2()
		for i := 0; i < cpus; i++ {
			field <- ""
		}
	}

	wg.Wait()
}

func str2Color(color string) puyo2.Color {
	switch color {
	case "":
		return puyo2.Empty
	case "r":
		return puyo2.Red
	case "g":
		return puyo2.Green
	case "b":
		return puyo2.Blue
	case "y":
		return puyo2.Yellow
	case "p":
		return puyo2.Purple
	case "o":
		return puyo2.Ojama
	default:
		panic("clear color must be one of r,g,b,y,p,o")
	}
}

func main() {
	now := time.Now()
	trapSignal()
	var chainStr string
	var clearColor string
	var color string
	var dedupMode string
	var simulatePolicy string
	param := flag.String("param", "a78", "puyofu")
	flag.StringVar(&opt.Hands, "hands", "", "hands")
	flag.StringVar(&chainStr, "chains", "", "chains")
	flag.StringVar(&clearColor, "clear", "", "clear color(r,g,b,y,p)")
	flag.StringVar(&color, "color", "", "color(r,g,b,y,p)")
	flag.StringVar(&dedupMode, "dedup", puyo2.DedupModeOff.String(), "dedup mode(off,same_pair_order,state,state_mirror). same_pair_order is effective only with -stop-on-chain")
	flag.StringVar(&simulatePolicy, "simulate", puyo2.SimulatePolicyDetailAlways.String(), "simulate policy(detail_always,fast_intermediate,fast_always)")
	flag.IntVar(&opt.N, "n", 0, "puyo count")
	flag.BoolVar(&opt.AllClear, "allclear", false, "allclear")
	flag.BoolVar(&opt.DisableChigiri, "disablechigiri", false, "disable chigiri")
	flag.BoolVar(&opt.StopOnChain, "stop-on-chain", false, "stop branch when chain occurs")
	flag.Parse()
	if strings.Contains(chainStr, "+") {
		chains := strings.Replace(chainStr, "+", "", -1)
		n, err := strconv.Atoi(chains)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", chainStr, err)
			return
		}
		opt.ChainsGT = n
	} else if chainStr != "" {
		n, err := strconv.Atoi(chainStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", chainStr, err)
			return
		}
		opt.ChainsEQ = n
	}
	opt.ClearColor = str2Color(clearColor)
	opt.Color = str2Color(color)
	var err error
	opt.DedupMode, err = puyo2.ParseDedupMode(dedupMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	opt.SimulatePolicy, err = puyo2.ParseSimulatePolicy(simulatePolicy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	run(param, &opt)
	p := message.NewPrinter(language.Japanese)
	p.Fprintf(os.Stderr, "elapsed: %vms\n", time.Since(now).Milliseconds())
}

func trapSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		fmt.Println("trap goroutine start.", os.Getpid())
		fmt.Println("send SIGHUP to show progress.")
		for range c {
			t := time.Now()
			total := 0
			for _, cnt := range searchedCount {
				total += cnt
			}
			fmt.Fprintf(os.Stderr, "[%s] %d/%f(%f%%)\n", t.Format("2006-01-02 15:04:05"), total, allSearchCnt, float64(total)*100/allSearchCnt)
		}
	}()
}
