package puyo2

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

func newIPSCond(q0 int, q1 int, q2 int) IPSNazoCondition {
	return IPSNazoCondition{Q0: q0, Q1: q1, Q2: q2}
}

func createRed4Field(withOjama bool) *BitField {
	bf := NewBitField()
	for x := 0; x < 4; x++ {
		bf.SetColor(Red, x, 1)
	}
	if withOjama {
		bf.SetColor(Ojama, 4, 1)
	}
	return bf
}

func TestEvaluateIPSNazoConditionBasic(t *testing.T) {
	bf := createRed4Field(false)
	match, metrics := EvaluateIPSNazoCondition(bf, newIPSCond(30, 0, 1))
	if !match {
		t.Fatal("30連鎖条件が一致しない")
	}
	if metrics.ChainCount != 1 {
		t.Fatalf("ChainCount=%d want 1", metrics.ChainCount)
	}
	if metrics.Erased[1] != 4 {
		t.Fatalf("Erased red=%d want 4", metrics.Erased[1])
	}
	if metrics.Remaining[1] != 0 {
		t.Fatalf("Remaining red=%d want 0", metrics.Remaining[1])
	}

	cases := []struct {
		name  string
		cond  IPSNazoCondition
		match bool
	}{
		{"31 false", newIPSCond(31, 0, 2), false},
		{"2 red", newIPSCond(2, 1, 0), true},
		{"2 color", newIPSCond(2, 7, 0), true},
		{"10 true", newIPSCond(10, 0, 1), true},
		{"11 false", newIPSCond(11, 0, 2), false},
		{"12 red", newIPSCond(12, 1, 4), true},
		{"13 red false", newIPSCond(13, 1, 5), false},
		{"12 color", newIPSCond(12, 7, 4), true},
		{"32 true", newIPSCond(32, 1, 1), true},
		{"33 false", newIPSCond(33, 1, 2), false},
		{"40 true", newIPSCond(40, 0, 1), true},
		{"41 false", newIPSCond(41, 0, 2), false},
		{"42 true", newIPSCond(42, 1, 4), true},
		{"43 false", newIPSCond(43, 1, 5), false},
		{"44 true", newIPSCond(44, 1, 1), true},
		{"45 false", newIPSCond(45, 1, 2), false},
		{"52 true", newIPSCond(52, 1, 4), true},
		{"53 false", newIPSCond(53, 1, 5), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := EvaluateIPSNazoCondition(bf, tc.cond)
			if got != tc.match {
				t.Fatalf("match=%v want %v", got, tc.match)
			}
		})
	}
}

func TestEvaluateIPSNazoConditionOjamaAndColorIndex(t *testing.T) {
	bf := createRed4Field(true)
	cases := []struct {
		name  string
		cond  IPSNazoCondition
		match bool
	}{
		{"2 ojama", newIPSCond(2, 6, 0), true},
		{"2 color", newIPSCond(2, 7, 0), true},
		{"12 ojama", newIPSCond(12, 6, 1), true},
		{"42 ojama", newIPSCond(42, 6, 1), true},
		{"44 ojama false", newIPSCond(44, 6, 1), false},
		{"44 ojama zero", newIPSCond(44, 6, 0), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, metrics := EvaluateIPSNazoCondition(bf, tc.cond)
			if got != tc.match {
				t.Fatalf("match=%v want %v metrics=%+v", got, tc.match, metrics)
			}
		})
	}
}

type pnsolveOutput struct {
	Input        string `json:"input"`
	InitialField string `json:"initialField"`
	Haipuyo      string `json:"haipuyo"`
	Status       string `json:"status"`
	Error        string `json:"error"`
	Condition    struct {
		Q0   int    `json:"q0"`
		Q1   int    `json:"q1"`
		Q2   int    `json:"q2"`
		Text string `json:"text"`
	} `json:"condition"`
	Searched  int `json:"searched"`
	Matched   int `json:"matched"`
	Solutions []struct {
		Hands        string `json:"hands"`
		Chains       int    `json:"chains"`
		Score        int    `json:"score"`
		Clear        bool   `json:"clear"`
		InitialField string `json:"initialField"`
		FinalField   string `json:"finalField"`
	} `json:"solutions"`
}

func isAllClearField(field string) bool {
	if len(field) == 0 {
		return true
	}
	for _, r := range field {
		if r != 'a' {
			return false
		}
	}
	return true
}

func runPNSolve(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "./cmd/pnsolve"}, args...)...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func decodePNSolveOutput(t *testing.T, stdout string) pnsolveOutput {
	t.Helper()
	var out pnsolveOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}
	return out
}

func handsSet(out pnsolveOutput) map[string]struct{} {
	set := map[string]struct{}{}
	for _, s := range out.Solutions {
		set[s.Hands] = struct{}{}
	}
	return set
}

func sortedHands(out pnsolveOutput) []string {
	hands := make([]string, 0, len(out.Solutions))
	for _, s := range out.Solutions {
		hands = append(hands, s.Hands)
	}
	sort.Strings(hands)
	return hands
}

func orderedHands(out pnsolveOutput) []string {
	hands := make([]string, 0, len(out.Solutions))
	for _, s := range out.Solutions {
		hands = append(hands, s.Hands)
	}
	return hands
}

func TestPNSolveCommandJSON(t *testing.T) {
	param := "800F08J08A0EB_8161__270"
	stdout, stderr, err := runPNSolve(t, "-param", param)
	if err != nil {
		t.Fatalf("pnsolve error=%v stderr=%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}

	var out pnsolveOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}

	decoded, err := ParseIPSNazoURL(param)
	if err != nil {
		t.Fatalf("ParseIPSNazoURL error=%v", err)
	}
	if out.InitialField != decoded.InitialField {
		t.Fatalf("initialField=%s want %s", out.InitialField, decoded.InitialField)
	}
	if out.Haipuyo != decoded.Haipuyo {
		t.Fatalf("haipuyo=%s want %s", out.Haipuyo, decoded.Haipuyo)
	}
	if out.Condition.Q0 != decoded.Condition.Q0 || out.Condition.Q1 != decoded.Condition.Q1 || out.Condition.Q2 != decoded.Condition.Q2 {
		t.Fatalf("condition code mismatch: %+v", out.Condition)
	}
	if out.Matched != len(out.Solutions) {
		t.Fatalf("matched=%d solutions=%d", out.Matched, len(out.Solutions))
	}
	if out.Status != "ok" {
		t.Fatalf("status=%s want ok", out.Status)
	}
	if out.Searched < out.Matched {
		t.Fatalf("searched=%d matched=%d", out.Searched, out.Matched)
	}
	for _, s := range out.Solutions {
		if len(s.InitialField) != 78 || len(s.FinalField) != 78 {
			t.Fatalf("field length invalid: %+v", s)
		}
		if s.Clear != isAllClearField(s.FinalField) {
			t.Fatalf("clear must match finalField state: %+v", s)
		}
	}
}

func TestPNSolveCommandCompactJSON(t *testing.T) {
	param := "80080080oM0oM098_4141__u03"
	stdout, stderr, err := runPNSolve(t, "-param", param, "-pretty=false")
	if err != nil {
		t.Fatalf("pnsolve error=%v stderr=%s", err, stderr)
	}
	if strings.Contains(stdout, "\n  \"") {
		t.Fatalf("compact json must not contain indentation: %q", stdout)
	}

	var out pnsolveOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}
	if out.Input != param {
		t.Fatalf("input=%s want %s", out.Input, param)
	}
	if out.Status != "ok" && out.Status != "no_solution" {
		t.Fatalf("status=%s is invalid", out.Status)
	}
}

func TestPNSolveCommandDefaultMatchesExplicitOptimizationFlags(t *testing.T) {
	param := "800F08J08A0EB_8161__270"
	stdoutDefault, stderrDefault, err := runPNSolve(t, "-param", param, "-pretty=false")
	if err != nil {
		t.Fatalf("default pnsolve error=%v stderr=%s", err, stderrDefault)
	}
	if stderrDefault != "" {
		t.Fatalf("unexpected default stderr: %s", stderrDefault)
	}

	stdoutExplicit, stderrExplicit, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-dedup", "same_pair_order",
		"-simulate", "fast_intermediate",
	)
	if err != nil {
		t.Fatalf("explicit pnsolve error=%v stderr=%s", err, stderrExplicit)
	}
	if stderrExplicit != "" {
		t.Fatalf("unexpected explicit stderr: %s", stderrExplicit)
	}
	if stdoutDefault != stdoutExplicit {
		t.Fatalf("default output must equal explicit optimization output.\ndefault=%s\nexplicit=%s", stdoutDefault, stdoutExplicit)
	}
}

func TestPNSolveCommandCompatibilityModeRuns(t *testing.T) {
	param := "800F08J08A0EB_8161__270"
	stdout, stderr, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-dedup", "off",
		"-simulate", "detail_always",
	)
	if err != nil {
		t.Fatalf("compatibility mode pnsolve error=%v stderr=%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}

	out := decodePNSolveOutput(t, stdout)
	if out.Status != "ok" && out.Status != "no_solution" {
		t.Fatalf("status=%s is invalid", out.Status)
	}
	if out.Matched != len(out.Solutions) {
		t.Fatalf("matched=%d solutions=%d", out.Matched, len(out.Solutions))
	}
}

func TestPNSolveCommandSamePairOrderDisabledWithoutStopOnChain(t *testing.T) {
	param := "jjgqqqqqqqqq_q1q1q1__u06"
	stdoutSamePair, stderrSamePair, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-dedup", "same_pair_order",
		"-simulate", "detail_always",
	)
	if err != nil {
		t.Fatalf("same_pair_order pnsolve error=%v stderr=%s", err, stderrSamePair)
	}
	if stderrSamePair != "" {
		t.Fatalf("unexpected same_pair_order stderr: %s", stderrSamePair)
	}

	stdoutOff, stderrOff, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-dedup", "off",
		"-simulate", "detail_always",
	)
	if err != nil {
		t.Fatalf("dedup off pnsolve error=%v stderr=%s", err, stderrOff)
	}
	if stderrOff != "" {
		t.Fatalf("unexpected dedup off stderr: %s", stderrOff)
	}
	if stdoutSamePair != stdoutOff {
		t.Fatalf("same_pair_order must behave as off when stop-on-chain=false.\nsame_pair=%s\noff=%s", stdoutSamePair, stdoutOff)
	}
	outSamePair := decodePNSolveOutput(t, stdoutSamePair)
	if outSamePair.Matched != len(outSamePair.Solutions) {
		t.Fatalf("same_pair_order matched=%d solutions=%d", outSamePair.Matched, len(outSamePair.Solutions))
	}
	if outSamePair.Matched != 4 {
		t.Fatalf("same_pair_order matched=%d want 4", outSamePair.Matched)
	}

	stdoutExpanded, stderrExpanded, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-dedup", "same_pair_order",
		"-simulate", "detail_always",
		"-expand-equivalent-hands",
	)
	if err != nil {
		t.Fatalf("expanded pnsolve error=%v stderr=%s", err, stderrExpanded)
	}
	if stderrExpanded != "" {
		t.Fatalf("unexpected expanded stderr: %s", stderrExpanded)
	}
	if stdoutExpanded != stdoutSamePair {
		t.Fatalf("expand-equivalent-hands must be no-op when same_pair_order is inactive.\nbase=%s\nexpanded=%s", stdoutSamePair, stdoutExpanded)
	}
}

func TestPNSolveCommandExpandEquivalentHandsNoOpForDedupOff(t *testing.T) {
	param := "800F08J08A0EB_8161__270"
	stdoutBase, stderrBase, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-dedup", "off",
		"-simulate", "detail_always",
	)
	if err != nil {
		t.Fatalf("base pnsolve error=%v stderr=%s", err, stderrBase)
	}
	if stderrBase != "" {
		t.Fatalf("unexpected base stderr: %s", stderrBase)
	}

	stdoutExpanded, stderrExpanded, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-dedup", "off",
		"-simulate", "detail_always",
		"-expand-equivalent-hands",
	)
	if err != nil {
		t.Fatalf("expanded pnsolve error=%v stderr=%s", err, stderrExpanded)
	}
	if stderrExpanded != "" {
		t.Fatalf("unexpected expanded stderr: %s", stderrExpanded)
	}
	if stdoutBase != stdoutExpanded {
		t.Fatalf("expand-equivalent-hands must be no-op for dedup off.\nbase=%s\nexpanded=%s", stdoutBase, stdoutExpanded)
	}
}

func TestPNSolveCommandExpandEquivalentHandsSearchFailedNoOp(t *testing.T) {
	param := "o00800c00b00j00z35xx4yxiqr9aticBIbrA_G1A1__u0b"
	stdoutBase, stderrBase, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
	)
	if err != nil {
		t.Fatalf("base pnsolve error=%v stderr=%s", err, stderrBase)
	}
	if stderrBase != "" {
		t.Fatalf("unexpected base stderr: %s", stderrBase)
	}

	stdoutExpanded, stderrExpanded, err := runPNSolve(
		t,
		"-param", param,
		"-pretty=false",
		"-expand-equivalent-hands",
	)
	if err != nil {
		t.Fatalf("expanded pnsolve error=%v stderr=%s", err, stderrExpanded)
	}
	if stderrExpanded != "" {
		t.Fatalf("unexpected expanded stderr: %s", stderrExpanded)
	}
	base := decodePNSolveOutput(t, stdoutBase)
	expanded := decodePNSolveOutput(t, stdoutExpanded)
	if base.Status != "search_failed" || expanded.Status != "search_failed" {
		t.Fatalf("status must be search_failed. base=%s expanded=%s", base.Status, expanded.Status)
	}
	if base.Matched != 0 || expanded.Matched != 0 || len(base.Solutions) != 0 || len(expanded.Solutions) != 0 {
		t.Fatalf("search_failed output must have no solutions. base=%+v expanded=%+v", base, expanded)
	}
	if !strings.Contains(base.Error, "should be able to place.") || !strings.Contains(expanded.Error, "should be able to place.") {
		t.Fatalf("search_failed error message changed unexpectedly. base=%q expanded=%q", base.Error, expanded.Error)
	}
}

func TestPNSolveCommandInvalidDedupFlag(t *testing.T) {
	_, stderr, err := runPNSolve(t, "-param", "800F08J08A0EB_8161__270", "-dedup", "x")
	if err == nil {
		t.Fatal("pnsolve must fail for invalid dedup mode")
	}
	if !strings.Contains(stderr, "unknown dedup mode") {
		t.Fatalf("stderr must include unknown dedup mode: %s", stderr)
	}
}

func TestPNSolveCommandInvalidSimulateFlag(t *testing.T) {
	_, stderr, err := runPNSolve(t, "-param", "800F08J08A0EB_8161__270", "-simulate", "x")
	if err == nil {
		t.Fatal("pnsolve must fail for invalid simulate policy")
	}
	if !strings.Contains(stderr, "unknown simulate policy") {
		t.Fatalf("stderr must include unknown simulate policy: %s", stderr)
	}
}

func TestPNSolveCommandInvalidInput(t *testing.T) {
	_, stderr, err := runPNSolve(t, "-param", "!")
	if err == nil {
		t.Fatal("pnsolve must fail for invalid input")
	}
	if !strings.Contains(stderr, "parse error") {
		t.Fatalf("stderr must include parse error: %s", stderr)
	}
}

func TestPNSolveCommandSearchFailedJSON(t *testing.T) {
	param := "o00800c00b00j00z35xx4yxiqr9aticBIbrA_G1A1__u0b"
	stdout, stderr, err := runPNSolve(t, "-param", param, "-pretty=false")
	if err != nil {
		t.Fatalf("pnsolve error=%v stderr=%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}

	var out pnsolveOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}
	if out.Status != "search_failed" {
		t.Fatalf("status=%s want search_failed", out.Status)
	}
	if out.Error == "" {
		t.Fatal("search_failed must include error")
	}
	if out.Matched != 0 || len(out.Solutions) != 0 {
		t.Fatalf("matched=%d solutions=%d want 0", out.Matched, len(out.Solutions))
	}
}

func TestPNSolveCommandNoIndex13Panic(t *testing.T) {
	param := "4r06P06904S04y03903N03Q02Q02k_A101o1E1__u07"
	stdout, stderr, err := runPNSolve(t, "-param", param, "-pretty=false")
	if err != nil {
		t.Fatalf("pnsolve error=%v stderr=%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}

	var out pnsolveOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}
	if out.Status == "search_failed" {
		t.Fatalf("status must not be search_failed: %+v", out)
	}
	if strings.Contains(out.Error, "index out of range [13]") {
		t.Fatalf("error must not include index out of range [13]: %q", out.Error)
	}
}

func TestPNSolveCommandNonAllClearRegression(t *testing.T) {
	param := "~000000000000000000000000000000000000000000000000000000000000000000000000111101___a01"
	stdout, stderr, err := runPNSolve(t, "-param", param, "-pretty=false")
	if err != nil {
		t.Fatalf("pnsolve error=%v stderr=%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}

	var out pnsolveOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}
	if out.Matched != 1 || len(out.Solutions) != 1 {
		t.Fatalf("matched=%d solutions=%d want 1", out.Matched, len(out.Solutions))
	}
	if out.Solutions[0].Clear {
		t.Fatalf("clear must be false for non-all-clear final field: %+v", out.Solutions[0])
	}
	if isAllClearField(out.Solutions[0].FinalField) {
		t.Fatalf("finalField must be non-empty: %s", out.Solutions[0].FinalField)
	}
	if out.Status != "ok" {
		t.Fatalf("status=%s want ok", out.Status)
	}
}

func TestManualHandsSolveOjamaAllClear(t *testing.T) {
	param := "M00M0MM6MM6SM4So6sMy9jCsPz9zPaCPiC_G1u1s1e1i1u1__260"
	handsStr := "yy30yb30bb10gg50yg41yb00"

	decoded, err := ParseIPSNazoURL(param)
	if err != nil {
		t.Fatalf("ParseIPSNazoURL error=%v", err)
	}
	hands := ParseSimpleHands(handsStr)
	if len(hands) == 0 {
		t.Fatal("hands must not be empty")
	}

	bf := NewBitFieldWithMattulwanC(decoded.InitialField)
	for i, hand := range hands {
		placed, _ := bf.PlacePuyo(hand.PuyoSet, hand.Position)
		if !placed {
			t.Fatalf("hand[%d] place failed: hand=%s heights=%v", i, ToSimpleHands([]Hand{hand}), bf.CreateHeights())
		}
	}

	matched, metrics := EvaluateIPSNazoCondition(bf, decoded.Condition)
	if !matched {
		t.Fatalf("manual hands must satisfy condition. metrics=%+v", metrics)
	}
	if metrics.Remaining[6] != 0 {
		t.Fatalf("remaining ojama must be 0 but %d", metrics.Remaining[6])
	}
}
