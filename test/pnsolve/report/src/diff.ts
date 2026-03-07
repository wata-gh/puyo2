import type { JsonValue } from "./artifact";

export interface SolutionDiff {
  added: JsonValue[];
  removed: JsonValue[];
}

export interface BasicDiffRow {
  path: string;
  before: JsonValue | null;
  after: JsonValue | null;
  changed: boolean;
}

function sortObjectKeys(value: JsonValue): JsonValue {
  if (Array.isArray(value)) {
    return value.map(sortObjectKeys);
  }
  if (value && typeof value === "object") {
    const sortedEntries = Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, nested]) => [key, sortObjectKeys(nested as JsonValue)]);
    return Object.fromEntries(sortedEntries) as JsonValue;
  }
  return value;
}

export function stableStringify(value: JsonValue): string {
  return JSON.stringify(sortObjectKeys(value));
}

function getSolutions(payload: JsonValue | null): JsonValue[] {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return [];
  }
  const solutions = (payload as Record<string, JsonValue>).solutions;
  return Array.isArray(solutions) ? solutions : [];
}

export function diffSolutions(beforePayload: JsonValue | null, afterPayload: JsonValue | null): SolutionDiff {
  const beforeMap = new Map(getSolutions(beforePayload).map((solution) => [stableStringify(solution), solution]));
  const afterMap = new Map(getSolutions(afterPayload).map((solution) => [stableStringify(solution), solution]));

  const removed: JsonValue[] = [];
  for (const [key, value] of beforeMap.entries()) {
    if (!afterMap.has(key)) {
      removed.push(value);
    }
  }

  const added: JsonValue[] = [];
  for (const [key, value] of afterMap.entries()) {
    if (!beforeMap.has(key)) {
      added.push(value);
    }
  }

  return { added, removed };
}

function getValueAtPath(payload: JsonValue | null, path: string[]): JsonValue | null {
  let current: JsonValue | null = payload;
  for (const segment of path) {
    if (!current || typeof current !== "object" || Array.isArray(current)) {
      return null;
    }
    current = ((current as Record<string, JsonValue>)[segment] ?? null) as JsonValue | null;
  }
  return current;
}

export function buildBasicDiffRows(beforePayload: JsonValue | null, afterPayload: JsonValue | null): BasicDiffRow[] {
  const paths = [
    ["input"],
    ["initialField"],
    ["haipuyo"],
    ["condition", "q0"],
    ["condition", "q1"],
    ["condition", "q2"],
    ["condition", "text"],
    ["status"],
    ["error"],
    ["matched"],
  ];

  return paths.map((path) => {
    const before = getValueAtPath(beforePayload, path);
    const after = getValueAtPath(afterPayload, path);
    return {
      path: path.join("."),
      before,
      after,
      changed: stableStringify(before) !== stableStringify(after),
    };
  });
}

export function asPrettyJson(value: JsonValue | null): string {
  if (value === null) {
    return "null";
  }
  return JSON.stringify(value, null, 2);
}
