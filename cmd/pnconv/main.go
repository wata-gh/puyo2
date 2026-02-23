package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wata-gh/puyo2"
)

func printDecoded(decoded *puyo2.IPSNazoDecoded) {
	fmt.Printf("Initial Field: %s\n", decoded.InitialField)
	fmt.Printf("Haipuyo: %s\n", decoded.Haipuyo)
	fmt.Printf("Condition: %s\n", decoded.Condition.Text)
	fmt.Printf("ConditionCode: q0=%d q1=%d q2=%d\n", decoded.ConditionCode[0], decoded.ConditionCode[1], decoded.ConditionCode[2])
}

func main() {
	var urlInput string
	var paramInput string
	flag.StringVar(&urlInput, "url", "", "IPSなぞぷよのURL")
	flag.StringVar(&paramInput, "param", "", "IPSなぞぷよのクエリ文字列")
	flag.Parse()

	inputs := []string{}
	if strings.TrimSpace(urlInput) != "" {
		inputs = append(inputs, urlInput)
	} else if strings.TrimSpace(paramInput) != "" {
		inputs = append(inputs, paramInput)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			inputs = append(inputs, line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "stdin read error: %v\n", err)
			os.Exit(1)
		}
	}

	if len(inputs) == 0 {
		fmt.Fprintln(os.Stderr, "no input. use -url, -param, or stdin.")
		os.Exit(1)
	}

	failed := false
	printed := 0
	for _, input := range inputs {
		decoded, err := puyo2.ParseIPSNazoURL(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %s: %v\n", input, err)
			failed = true
			continue
		}
		if printed > 0 {
			fmt.Println()
		}
		printDecoded(decoded)
		printed++
	}

	if failed {
		os.Exit(1)
	}
}
