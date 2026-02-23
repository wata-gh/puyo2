package main

import (
	"sort"

	"github.com/wata-gh/puyo2"
)

const maxEquivalentCandidatesPerBase = 5000

type runRange struct {
	start int
	end   int
}

type runPermutations struct {
	run   runRange
	perms [][][2]int
}

func expandEquivalentHandsSolutions(initial *puyo2.BitField, condition puyo2.IPSNazoCondition, puyoSets []puyo2.PuyoSet, policy puyo2.SimulatePolicy, baseSolutions []solveSolutionJSON) []solveSolutionJSON {
	runs := buildEquivalentRuns(puyoSets)
	solutionMap := make(map[string]solveSolutionJSON, len(baseSolutions))
	for _, s := range baseSolutions {
		solutionMap[s.Hands] = s
	}

	if len(runs) > 0 {
		for _, base := range baseSolutions {
			hands := puyo2.ParseSimpleHands(base.Hands)
			if len(hands) != len(puyoSets) {
				continue
			}
			expandEquivalentForBase(initial, condition, policy, hands, runs, solutionMap)
		}
	}

	return sortedSolutions(solutionMap)
}

func sortedSolutions(solutionMap map[string]solveSolutionJSON) []solveSolutionJSON {
	solutions := make([]solveSolutionJSON, 0, len(solutionMap))
	for _, s := range solutionMap {
		solutions = append(solutions, s)
	}
	sort.Slice(solutions, func(i, j int) bool {
		return solutions[i].Hands < solutions[j].Hands
	})
	return solutions
}

func buildEquivalentRuns(puyoSets []puyo2.PuyoSet) []runRange {
	runs := make([]runRange, 0)
	for i := 0; i < len(puyoSets)-1; {
		j := i
		for j+1 < len(puyoSets) && samePair(puyoSets[j], puyoSets[j+1]) {
			j++
		}
		if j > i {
			runs = append(runs, runRange{start: i, end: j + 1})
		}
		i = j + 1
	}
	return runs
}

func samePair(a, b puyo2.PuyoSet) bool {
	if a.Axis == b.Axis && a.Child == b.Child {
		return true
	}
	return a.Axis == b.Child && a.Child == b.Axis
}

func expandEquivalentForBase(initial *puyo2.BitField, condition puyo2.IPSNazoCondition, policy puyo2.SimulatePolicy, hands []puyo2.Hand, runs []runRange, solutionMap map[string]solveSolutionJSON) {
	basePositions := make([][2]int, len(hands))
	for i, hand := range hands {
		basePositions[i] = hand.Position
	}

	runPerms := make([]runPermutations, 0, len(runs))
	for _, run := range runs {
		perms := generateUniquePositionPermutations(basePositions[run.start:run.end])
		runPerms = append(runPerms, runPermutations{run: run, perms: perms})
	}

	candidatePositions := make([][2]int, len(basePositions))
	copy(candidatePositions, basePositions)
	tested := 0
	stop := false

	var dfs func(idx int)
	dfs = func(idx int) {
		if stop {
			return
		}
		if idx == len(runPerms) {
			if samePositions(candidatePositions, basePositions) {
				return
			}
			if tested >= maxEquivalentCandidatesPerBase {
				stop = true
				return
			}
			tested++

			candidateHands := make([]puyo2.Hand, len(hands))
			copy(candidateHands, hands)
			for i := range candidateHands {
				candidateHands[i].Position = candidatePositions[i]
			}

			handsStr := puyo2.ToSimpleHands(candidateHands)
			if _, found := solutionMap[handsStr]; found {
				return
			}

			solution, ok := replayHandsAndBuildSolution(initial, condition, policy, candidateHands)
			if ok {
				solutionMap[handsStr] = solution
			}
			return
		}

		rp := runPerms[idx]
		backup := make([][2]int, rp.run.end-rp.run.start)
		copy(backup, candidatePositions[rp.run.start:rp.run.end])

		for _, perm := range rp.perms {
			copy(candidatePositions[rp.run.start:rp.run.end], perm)
			dfs(idx + 1)
			if stop {
				return
			}
		}
		copy(candidatePositions[rp.run.start:rp.run.end], backup)
	}

	dfs(0)
}

func samePositions(a [][2]int, b [][2]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generateUniquePositionPermutations(positions [][2]int) [][][2]int {
	n := len(positions)
	if n <= 1 {
		one := make([][2]int, n)
		copy(one, positions)
		return [][][2]int{one}
	}

	used := make([]bool, n)
	current := make([][2]int, 0, n)
	out := make([][][2]int, 0)

	var dfs func(depth int)
	dfs = func(depth int) {
		if depth == n {
			perm := make([][2]int, n)
			copy(perm, current)
			out = append(out, perm)
			return
		}

		seen := map[[2]int]struct{}{}
		for i := 0; i < n; i++ {
			if used[i] {
				continue
			}
			p := positions[i]
			if _, exists := seen[p]; exists {
				continue
			}
			seen[p] = struct{}{}
			used[i] = true
			current = append(current, p)
			dfs(depth + 1)
			current = current[:len(current)-1]
			used[i] = false
		}
	}

	dfs(0)
	return out
}

func replayHandsAndBuildSolution(initial *puyo2.BitField, condition puyo2.IPSNazoCondition, policy puyo2.SimulatePolicy, hands []puyo2.Hand) (solveSolutionJSON, bool) {
	current := initial.Clone()
	var beforeTerminal *puyo2.BitField
	var terminalResult *puyo2.RensaResult

	for i, hand := range hands {
		placed, _ := current.PlacePuyo(hand.PuyoSet, hand.Position)
		if !placed {
			return solveSolutionJSON{}, false
		}
		terminal := i == len(hands)-1
		beforeTerminal = current
		terminalResult = simulateWithPolicy(current, policy, terminal)
		if !terminal {
			current = terminalResult.BitField
		}
	}

	if beforeTerminal == nil || terminalResult == nil {
		return solveSolutionJSON{}, false
	}

	matched, _ := puyo2.EvaluateIPSNazoCondition(beforeTerminal, condition)
	if !matched {
		return solveSolutionJSON{}, false
	}

	initialField := beforeTerminal.MattulwanEditorParam()
	finalField := terminalResult.BitField.MattulwanEditorParam()
	clear := terminalResult.BitField.IsEmpty()
	return newSolution(puyo2.ToSimpleHands(hands), terminalResult.Chains, terminalResult.Score, initialField, finalField, clear), true
}

func simulateWithPolicy(bf *puyo2.BitField, policy puyo2.SimulatePolicy, terminal bool) *puyo2.RensaResult {
	switch policy {
	case puyo2.SimulatePolicyFastAlways:
		return bf.CloneForSimulation().Simulate()
	case puyo2.SimulatePolicyFastIntermediate:
		if terminal {
			return bf.CloneForSimulation().SimulateDetail()
		}
		return bf.CloneForSimulation().Simulate()
	default:
		return bf.CloneForSimulation().SimulateDetail()
	}
}
