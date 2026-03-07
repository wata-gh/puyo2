import { buildBasicDiffRows, diffSolutions } from "./diff";
import type { ArtifactReport, CheckStatus, JsonValue, LevelArtifacts } from "./artifact";

export type CaseFilter = "all" | "diff" | "error" | "match";
export type LevelTone = "unexecuted" | "clean" | "diff" | "error";
export type CaseDisplayStatus = CheckStatus | "UNEXECUTED";

export interface SolutionLink {
  hands: string;
  url: string;
}

export interface CaseView {
  level: string;
  param: string;
  order: number;
  checkStatus: CaseDisplayStatus;
  conditionText: string | null;
  originalUrl: string;
  baselineStatus: string | null;
  currentStatus: string | null;
  matchedDelta: number | null;
  currentSolutionCount: number;
  solutionDelta: {
    added: number;
    removed: number;
  };
  solutionLinks: SolutionLink[];
  basicDiffRows: ReturnType<typeof buildBasicDiffRows>;
  addedSolutions: JsonValue[];
  removedSolutions: JsonValue[];
  beforeRaw: JsonValue | null;
  afterRaw: JsonValue | null;
  beforeNormalized: JsonValue | null;
  afterNormalized: JsonValue | null;
  diffText: string | null;
  stderrText: string | null;
}

export interface LevelView {
  name: string;
  totalCases: number;
  checkedCases: number;
  diffCases: number;
  errorCases: number;
  solutionAdded: number;
  solutionRemoved: number;
  matchedDeltaSum: number;
  tone: LevelTone;
  executed: boolean;
  cases: CaseView[];
}

export interface DashboardView {
  generatedAt: string;
  artifactRoot: string;
  repoRoot: string;
  levels: LevelView[];
}

function getObjectValue(payload: JsonValue | null, key: string): JsonValue | null {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return null;
  }
  return ((payload as Record<string, JsonValue>)[key] ?? null) as JsonValue | null;
}

function getStringValue(payload: JsonValue | null, key: string): string | null {
  const value = getObjectValue(payload, key);
  return typeof value === "string" ? value : null;
}

function getNumberValue(payload: JsonValue | null, key: string): number | null {
  const value = getObjectValue(payload, key);
  return typeof value === "number" ? value : null;
}

function getConditionText(beforeRaw: JsonValue | null, afterRaw: JsonValue | null): string | null {
  for (const candidate of [beforeRaw, afterRaw]) {
    const condition = getObjectValue(candidate, "condition");
    if (condition && typeof condition === "object" && !Array.isArray(condition)) {
      const text = (condition as Record<string, JsonValue>).text;
      if (typeof text === "string") {
        return text;
      }
    }
  }
  return null;
}

function getInput(beforeRaw: JsonValue | null, afterRaw: JsonValue | null, fallback: string): string {
  return getStringValue(beforeRaw, "input") ?? getStringValue(afterRaw, "input") ?? fallback;
}

function getSolutions(payload: JsonValue | null): JsonValue[] {
  const solutions = getObjectValue(payload, "solutions");
  return Array.isArray(solutions) ? solutions : [];
}

function buildSolutionLinks(afterRaw: JsonValue | null): SolutionLink[] {
  const initialField = getStringValue(afterRaw, "initialField");
  if (!initialField) {
    return [];
  }

  return getSolutions(afterRaw)
    .map((solution) => {
      if (!solution || typeof solution !== "object" || Array.isArray(solution)) {
        return null;
      }
      const hands = (solution as Record<string, JsonValue>).hands;
      if (typeof hands !== "string") {
        return null;
      }
      const params = new URLSearchParams({ fs: initialField, h: hands });
      return {
        hands,
        url: `http://localhost:3000/simus?${params.toString()}`,
      } satisfies SolutionLink;
    })
    .filter((value): value is SolutionLink => value !== null);
}

function classifyLevelTone(level: Omit<LevelView, "tone">): LevelTone {
  if (!level.executed) {
    return "unexecuted";
  }
  if (level.errorCases > 0) {
    return "error";
  }
  if (level.diffCases > 0) {
    return "diff";
  }
  return "clean";
}

function isErrorStatus(status: CaseDisplayStatus): boolean {
  return status === "RUN_ERROR" || status === "JD_ERROR" || status === "BEFORE_MISSING";
}

function shouldIncludeCase(filter: CaseFilter, caseView: CaseView): boolean {
  if (filter === "all") {
    return true;
  }
  if (filter === "diff") {
    return caseView.checkStatus === "DIFF";
  }
  if (filter === "error") {
    return isErrorStatus(caseView.checkStatus);
  }
  return caseView.checkStatus === "MATCH";
}

function buildCaseView(level: LevelArtifacts, param: string, order: number): CaseView {
  const artifacts = level.cases.get(param) ?? null;
  const checkStatus = artifacts?.meta?.status ?? "UNEXECUTED";
  const beforeRaw = artifacts?.beforeRaw ?? null;
  const afterRaw = artifacts?.afterRaw ?? null;
  const beforeNormalized = artifacts?.beforeNormalized ?? beforeRaw;
  const afterNormalized = artifacts?.afterNormalized ?? afterRaw;
  const solutionDiff = diffSolutions(beforeNormalized, afterNormalized);
  const beforeMatched = getNumberValue(beforeRaw, "matched");
  const afterMatched = getNumberValue(afterRaw, "matched");
  const input = getInput(beforeRaw, afterRaw, param);

  return {
    level: level.name,
    param,
    order,
    checkStatus,
    conditionText: getConditionText(beforeRaw, afterRaw),
    originalUrl: `https://ishikawapuyo.net/simu/pn.html?${encodeURIComponent(input)}`,
    baselineStatus: getStringValue(beforeRaw, "status"),
    currentStatus: getStringValue(afterRaw, "status"),
    matchedDelta: beforeMatched !== null && afterMatched !== null ? afterMatched - beforeMatched : null,
    currentSolutionCount: getSolutions(afterRaw).length,
    solutionDelta: {
      added: solutionDiff.added.length,
      removed: solutionDiff.removed.length,
    },
    solutionLinks: buildSolutionLinks(afterRaw),
    basicDiffRows: buildBasicDiffRows(beforeNormalized, afterNormalized),
    addedSolutions: solutionDiff.added,
    removedSolutions: solutionDiff.removed,
    beforeRaw,
    afterRaw,
    beforeNormalized,
    afterNormalized,
    diffText: artifacts?.diffText ?? null,
    stderrText: artifacts?.stderrText ?? null,
  };
}

function buildLevelView(level: LevelArtifacts, executedLevels: Set<string>): LevelView {
  const cases = level.params.map((param, index) => buildCaseView(level, param, index));
  const checkedCases = cases.filter((caseView) => caseView.checkStatus !== "UNEXECUTED").length;
  const diffCases = cases.filter((caseView) => caseView.checkStatus === "DIFF").length;
  const errorCases = cases.filter((caseView) => isErrorStatus(caseView.checkStatus)).length;
  const solutionAdded = cases.reduce((sum, caseView) => sum + caseView.solutionDelta.added, 0);
  const solutionRemoved = cases.reduce((sum, caseView) => sum + caseView.solutionDelta.removed, 0);
  const matchedDeltaSum = cases.reduce((sum, caseView) => sum + (caseView.matchedDelta ?? 0), 0);
  const executed = executedLevels.has(level.name) || checkedCases > 0;

  const partial: Omit<LevelView, "tone"> = {
    name: level.name,
    totalCases: level.params.length,
    checkedCases,
    diffCases,
    errorCases,
    solutionAdded,
    solutionRemoved,
    matchedDeltaSum,
    executed,
    cases,
  };

  return {
    ...partial,
    tone: classifyLevelTone(partial),
  };
}

export function buildDashboardView(report: ArtifactReport): DashboardView {
  const executedLevels = new Set(report.manifest.executedLevels);
  const levels = report.levels.map((level) => buildLevelView(level, executedLevels));
  return {
    generatedAt: report.manifest.createdAt,
    artifactRoot: report.artifactRoot,
    repoRoot: report.repoRoot,
    levels,
  };
}

export function parseFilter(input: string | null): CaseFilter {
  if (input === "diff" || input === "error" || input === "match") {
    return input;
  }
  return "all";
}

export function getLevelView(dashboard: DashboardView, levelName: string): LevelView | null {
  return dashboard.levels.find((level) => level.name === levelName) ?? null;
}

export function filterCases(level: LevelView, filter: CaseFilter): CaseView[] {
  return level.cases.filter((caseView) => shouldIncludeCase(filter, caseView));
}
