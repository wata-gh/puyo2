package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wata-gh/puyo2"
)

type options struct {
	Hands    string
	Chains   int
	AllClear bool
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
		cond.EachHandCallback = func(sr *puyo2.SearchResult) bool {
			searchedCount[num]++
			// if sr.Hands[0].Position[0] == 4 && sr.Hands[0].Position[1] == 0 {
			// 	if sr.Depth > 1 {
			// 		if sr.Hands[1].Position[0] == 1 && sr.Hands[1].Position[1] == 1 {
			// 			if sr.Depth > 2 {
			// 				if sr.Hands[2].Position[0] == 1 && sr.Hands[2].Position[1] == 1 {
			// 					if sr.Depth > 3 {
			// 						sr.RensaResult.BitField.ShowDebug()
			// 						if sr.Hands[2].Position[0] == 2 && sr.Hands[2].Position[1] == 2 {
			// 							return false
			// 						}
			// 					}
			// 				}
			// 			}
			// 		}
			// 	}
			// 	if sr.Depth > 4 {
			// 		return false
			// 	}
			// 	return true
			// }
			// return false
			if opt.AllClear && sr.RensaResult.BitField.IsEmpty() {
				fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl(), params[1]+hands2Str(sr.Hands))
			}
			if opt.Chains > 0 && sr.RensaResult.Chains >= opt.Chains {
				fmt.Println(sr.BeforeSimulate.MattulwanEditorUrl(), params[1]+hands2Str(sr.Hands))
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
	for i := 0; i < cpus; i++ {
		searchedCount = append(searchedCount, 0)
		wg.Add(1)
		go search(i+1, field, puyoSets[1:], wg)
	}

	cond := puyo2.NewSearchConditionWithBFAndPuyoSets(puyo2.NewBitFieldWithMattulwanC(*param), puyoSets[0:1])
	cond.BitField.ShowDebug()
	fmt.Printf("%+v\n", puyoSets)

	cond.LastCallback = func(sr *puyo2.SearchResult) {
		searchedCount[0]++
		param := fmt.Sprintf("%s %s", sr.RensaResult.BitField.MattulwanEditorParam(), hands2Str(sr.Hands))
		field <- param
	}
	cond.SearchWithPuyoSetsV2()

	for i := 0; i < cpus; i++ {
		field <- ""
	}
	wg.Wait()
}

func main() {
	now := time.Now()
	trapSignal()
	var chainStr string
	var clearColor string
	param := flag.String("param", "a78", "puyofu")
	flag.StringVar(&opt.Hands, "hands", "", "hands")
	flag.StringVar(&chainStr, "chains", "", "chains")
	flag.StringVar(&clearColor, "clear", "", "clear color(r,g,b,y,p)")
	flag.BoolVar(&opt.AllClear, "allclear", false, "allclear")
	flag.BoolVar(&opt.DisableChigiri, "disablechigiri", false, "disable chigiri")
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
	switch clearColor {
	case "":
		opt.ClearColor = puyo2.Empty
	case "r":
		opt.ClearColor = puyo2.Red
	case "g":
		opt.ClearColor = puyo2.Green
	case "b":
		opt.ClearColor = puyo2.Blue
	case "y":
		opt.ClearColor = puyo2.Yellow
	case "p":
		opt.ClearColor = puyo2.Purple
	case "o":
		opt.ClearColor = puyo2.Ojama
	default:
		panic("clear color must be one of r,g,b,y,p,o")
	}

	run(param, &opt)
	p := message.NewPrinter(language.Japanese)
	p.Fprintf(os.Stderr, "elapsed: %vms\n", time.Since(now).Milliseconds())
}

func trapSignal() {
	c := make(chan os.Signal)
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
