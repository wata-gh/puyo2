package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wata-gh/puyo2"
)

type solveConditionJSON struct {
	Q0   int    `json:"q0"`
	Q1   int    `json:"q1"`
	Q2   int    `json:"q2"`
	Text string `json:"text"`
}

type solveSolutionJSON struct {
	Hands        string `json:"hands"`
	Chains       int    `json:"chains"`
	Score        int    `json:"score"`
	Clear        bool   `json:"clear"`
	InitialField string `json:"initialField"`
	FinalField   string `json:"finalField"`
}

type solveOutputJSON struct {
	Input        string              `json:"input"`
	InitialField string              `json:"initialField"`
	Haipuyo      string              `json:"haipuyo"`
	Condition    solveConditionJSON  `json:"condition"`
	Status       string              `json:"status"`
	Error        string              `json:"error,omitempty"`
	Searched     int                 `json:"searched"`
	Matched      int                 `json:"matched"`
	Solutions    []solveSolutionJSON `json:"solutions"`
}

func newSolution(hands string, chains int, score int, initialField string, finalField string, clear bool) solveSolutionJSON {
	return solveSolutionJSON{
		Hands:        hands,
		Chains:       chains,
		Score:        score,
		Clear:        clear,
		InitialField: initialField,
		FinalField:   finalField,
	}
}

func readSingleInputFromStdin() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	lines := make([]string, 0, 2)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) > 1 {
			return "", fmt.Errorf("multiple inputs are not supported: got %d non-empty lines", len(lines))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if len(lines) == 0 {
		return "", fmt.Errorf("no input")
	}
	return lines[0], nil
}

func toPuyoSets(haipuyo string) []puyo2.PuyoSet {
	ptrs := puyo2.Haipuyo2PuyoSets(haipuyo)
	sets := make([]puyo2.PuyoSet, 0, len(ptrs))
	for _, p := range ptrs {
		if p == nil {
			continue
		}
		sets = append(sets, *p)
	}
	return sets
}

func main() {
	var urlInput string
	var paramInput string
	var disableChigiri bool
	var pretty bool

	flag.StringVar(&urlInput, "url", "", "IPSなぞぷよのURL")
	flag.StringVar(&paramInput, "param", "", "IPSなぞぷよのクエリ文字列")
	flag.BoolVar(&disableChigiri, "disablechigiri", false, "disable chigiri placements")
	flag.BoolVar(&pretty, "pretty", true, "pretty print JSON")
	flag.Parse()

	input := strings.TrimSpace(urlInput)
	if input == "" {
		input = strings.TrimSpace(paramInput)
	}
	if input == "" {
		var err error
		input, err = readSingleInputFromStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "stdin error: %v\n", err)
			os.Exit(1)
		}
	}

	decoded, err := puyo2.ParseIPSNazoURL(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	out := solveOutputJSON{
		Input:        input,
		InitialField: decoded.InitialField,
		Haipuyo:      decoded.Haipuyo,
		Status:       "ok",
		Condition: solveConditionJSON{
			Q0:   decoded.Condition.Q0,
			Q1:   decoded.Condition.Q1,
			Q2:   decoded.Condition.Q2,
			Text: decoded.Condition.Text,
		},
		Solutions: []solveSolutionJSON{},
	}

	initial := puyo2.NewBitFieldWithMattulwanC(decoded.InitialField)
	puyoSets := toPuyoSets(decoded.Haipuyo)
	if len(puyoSets) == 0 {
		out.Searched = 1
		matched, _ := puyo2.EvaluateIPSNazoCondition(initial, decoded.Condition)
		if matched {
			result := initial.Clone().SimulateDetail()
			initialField := initial.MattulwanEditorParam()
			finalField := result.BitField.MattulwanEditorParam()
			clear := result.BitField.IsEmpty()
			out.Solutions = append(out.Solutions, newSolution("", result.Chains, result.Score, initialField, finalField, clear))
			out.Matched = 1
		}
	} else {
		cond := puyo2.NewSearchConditionWithBFAndPuyoSets(initial, puyoSets)
		cond.DisableChigiri = disableChigiri
		cond.LastCallback = func(sr *puyo2.SearchResult) {
			out.Searched++
			if sr == nil {
				return
			}
			matched, _ := puyo2.EvaluateIPSNazoCondition(sr.BeforeSimulate, decoded.Condition)
			if !matched {
				return
			}
			initialField := sr.BeforeSimulate.MattulwanEditorParam()
			finalField := sr.RensaResult.BitField.MattulwanEditorParam()
			clear := sr.RensaResult.BitField.IsEmpty()
			out.Solutions = append(out.Solutions, newSolution(puyo2.ToSimpleHands(sr.Hands), sr.RensaResult.Chains, sr.RensaResult.Score, initialField, finalField, clear))
		}
		var recovered any
		func() {
			defer func() {
				recovered = recover()
			}()
			cond.SearchWithPuyoSetsV2()
		}()
		if recovered != nil {
			out.Status = "search_failed"
			out.Error = fmt.Sprint(recovered)
			out.Solutions = []solveSolutionJSON{}
			out.Matched = 0
		} else {
			out.Matched = len(out.Solutions)
		}
	}
	if out.Status == "ok" && out.Matched == 0 {
		out.Status = "no_solution"
	}

	enc := json.NewEncoder(os.Stdout)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "json encode error: %v\n", err)
		os.Exit(1)
	}
}
