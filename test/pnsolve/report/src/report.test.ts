import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { loadArtifactReport } from "./artifact";
import { createFetchHandler } from "../server";
import { buildDashboardView, getLevelView } from "./summary";

const repoRoot = path.resolve(import.meta.dir, "../../../..");

async function mkdirp(dirPath: string) {
  await fs.mkdir(dirPath, { recursive: true });
}

async function writeJson(filePath: string, value: unknown) {
  await mkdirp(path.dirname(filePath));
  await fs.writeFile(filePath, `${JSON.stringify(value, null, 2)}\n`);
}

async function createSampleArtifact(): Promise<string> {
  const artifactRoot = await fs.mkdtemp(path.join(os.tmpdir(), "pnsolve-report-"));
  const level1Params = (await fs.readFile(path.join(repoRoot, "test/pnsolve/level1/list"), "utf8"))
    .split(/\r?\n/)
    .filter(Boolean);
  const level2Params = (await fs.readFile(path.join(repoRoot, "test/pnsolve/level2/list"), "utf8"))
    .split(/\r?\n/)
    .filter(Boolean);

  const matchParam = level1Params[0];
  const diffParam = level1Params[1];
  const runErrorParam = level1Params[2];
  const beforeMissingParam = level2Params[0];
  const jdErrorParam = level2Params[1];

  const beforeRaw = {
    input: matchParam,
    initialField: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    haipuyo: "rg",
    condition: { q0: 2, q1: 0, q2: 2, text: "2連鎖する" },
    status: "ok",
    searched: 10,
    matched: 1,
    solutions: [
      {
        hands: "rg12",
        chains: 2,
        score: 360,
        initialField: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        finalField: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      },
    ],
  };

  const beforeNormalized = {
    ...beforeRaw,
    searched: undefined,
    solutions: beforeRaw.solutions,
  };
  delete (beforeNormalized as Record<string, unknown>).searched;

  const afterRawMatch = structuredClone(beforeRaw);
  const afterNormalizedMatch = structuredClone(beforeNormalized);

  const afterRawDiff = {
    ...beforeRaw,
    input: diffParam,
    matched: 2,
    solutions: [
      ...beforeRaw.solutions,
      {
        hands: "by30",
        chains: 3,
        score: 720,
        initialField: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        finalField: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      },
    ],
  };
  const afterNormalizedDiff = structuredClone(afterRawDiff);
  delete (afterNormalizedDiff as Record<string, unknown>).searched;

  const manifest = {
    formatVersion: 1,
    createdAt: "2026-03-07T00:00:00Z",
    repoRoot,
    normalizedFilter: 'del(.searched) | (if has("solutions") then .solutions |= map(del(.clear)) else . end)',
    executedLevels: ["level1", "level2"],
    singleMode: false,
    binPath: path.join(repoRoot, "target/debug/pnsolve"),
    summary: {
      checked: 5,
      diff: 1,
      missingBefore: 1,
      runError: 1,
      jdError: 1,
    },
  };
  await writeJson(path.join(artifactRoot, "manifest.json"), manifest);

  const writeCase = async (
    level: string,
    param: string,
    meta: Record<string, string | null>,
    files: Record<string, string | object>,
  ) => {
    const caseDir = path.join(artifactRoot, "levels", level, "cases", param);
    await mkdirp(caseDir);
    await writeJson(path.join(caseDir, "meta.json"), {
      level,
      param,
      ...meta,
    });
    for (const [name, value] of Object.entries(files)) {
      const filePath = path.join(caseDir, name);
      await mkdirp(path.dirname(filePath));
      if (typeof value === "string") {
        await fs.writeFile(filePath, value);
      } else {
        await writeJson(filePath, value);
      }
    }
  };

  await writeCase(
    "level1",
    matchParam,
    {
      status: "MATCH",
      beforePath: `levels/level1/cases/${matchParam}/before.raw.json`,
      afterPath: `levels/level1/cases/${matchParam}/after.raw.json`,
      beforeNormalizedPath: `levels/level1/cases/${matchParam}/before.normalized.json`,
      afterNormalizedPath: `levels/level1/cases/${matchParam}/after.normalized.json`,
      diffPath: null,
      stderrPath: null,
    },
    {
      "before.raw.json": beforeRaw,
      "after.raw.json": afterRawMatch,
      "before.normalized.json": beforeNormalized,
      "after.normalized.json": afterNormalizedMatch,
    },
  );

  await writeCase(
    "level1",
    diffParam,
    {
      status: "DIFF",
      beforePath: `levels/level1/cases/${diffParam}/before.raw.json`,
      afterPath: `levels/level1/cases/${diffParam}/after.raw.json`,
      beforeNormalizedPath: `levels/level1/cases/${diffParam}/before.normalized.json`,
      afterNormalizedPath: `levels/level1/cases/${diffParam}/after.normalized.json`,
      diffPath: `levels/level1/cases/${diffParam}/diff.jd`,
      stderrPath: null,
    },
    {
      "before.raw.json": { ...beforeRaw, input: diffParam },
      "after.raw.json": afterRawDiff,
      "before.normalized.json": { ...beforeNormalized, input: diffParam },
      "after.normalized.json": afterNormalizedDiff,
      "diff.jd": '^ "SET"\n@ ["matched"]\n- 1\n+ 2\n',
    },
  );

  await writeCase(
    "level1",
    runErrorParam,
    {
      status: "RUN_ERROR",
      beforePath: `levels/level1/cases/${runErrorParam}/before.raw.json`,
      afterPath: null,
      beforeNormalizedPath: null,
      afterNormalizedPath: null,
      diffPath: null,
      stderrPath: `levels/level1/cases/${runErrorParam}/stderr.txt`,
    },
    {
      "before.raw.json": { ...beforeRaw, input: runErrorParam },
      "stderr.txt": "pnsolve failed\n",
    },
  );

  await writeCase(
    "level2",
    beforeMissingParam,
    {
      status: "BEFORE_MISSING",
      beforePath: null,
      afterPath: `levels/level2/cases/${beforeMissingParam}/after.raw.json`,
      beforeNormalizedPath: null,
      afterNormalizedPath: `levels/level2/cases/${beforeMissingParam}/after.normalized.json`,
      diffPath: null,
      stderrPath: null,
    },
    {
      "after.raw.json": { ...afterRawMatch, input: beforeMissingParam },
      "after.normalized.json": { ...afterNormalizedMatch, input: beforeMissingParam },
    },
  );

  await writeCase(
    "level2",
    jdErrorParam,
    {
      status: "JD_ERROR",
      beforePath: `levels/level2/cases/${jdErrorParam}/before.raw.json`,
      afterPath: `levels/level2/cases/${jdErrorParam}/after.raw.json`,
      beforeNormalizedPath: null,
      afterNormalizedPath: null,
      diffPath: null,
      stderrPath: `levels/level2/cases/${jdErrorParam}/stderr.txt`,
    },
    {
      "before.raw.json": { ...beforeRaw, input: jdErrorParam },
      "after.raw.json": { ...afterRawMatch, input: jdErrorParam },
      "stderr.txt": "jq normalize failed\n",
    },
  );

  return artifactRoot;
}

test("dashboard summary aggregates diff and error counts", async () => {
  const artifactRoot = await createSampleArtifact();
  const report = await loadArtifactReport(artifactRoot);
  const dashboard = buildDashboardView(report);

  const level1 = getLevelView(dashboard, "level1");
  const level2 = getLevelView(dashboard, "level2");
  const level3 = getLevelView(dashboard, "level3");

  expect(level1).not.toBeNull();
  expect(level2).not.toBeNull();
  expect(level3).not.toBeNull();

  expect(level1?.checkedCases).toBe(3);
  expect(level1?.diffCases).toBe(1);
  expect(level1?.errorCases).toBe(1);
  expect(level1?.solutionAdded).toBe(1);
  expect(level1?.solutionRemoved).toBe(1);
  expect(level1?.matchedDeltaSum).toBe(1);

  expect(level2?.errorCases).toBe(2);
  expect(level2?.tone).toBe("error");
  expect(level3?.executed).toBe(false);
  expect(level3?.tone).toBe("unexecuted");
});

test("level page renders condition text and solution links", async () => {
  const artifactRoot = await createSampleArtifact();
  const report = await loadArtifactReport(artifactRoot);
  const dashboard = buildDashboardView(report);
  const level1 = getLevelView(dashboard, "level1");

  expect(level1).not.toBeNull();
  const html = (await import("./render")).renderLevelPage(dashboard, level1!, "all");

  expect(html).toContain("2連鎖する");
  expect(html).toContain("オリジナルページ");
  expect(html).toContain("http://localhost:3000/simus?fs=");
  expect(html).toContain("solution 1");
  expect(html).toContain("差分なし");
});

test("http routes return dashboard and level detail", async () => {
  const artifactRoot = await createSampleArtifact();
  const server = Bun.serve({
    port: 0,
    hostname: "127.0.0.1",
    fetch: createFetchHandler({ artifactRoot, port: 0 }),
  });

  try {
    const dashboardResponse = await fetch(`${server.url} ` .trim());
    const levelResponse = await fetch(`${server.url}levels/level1?filter=diff`);

    expect(dashboardResponse.status).toBe(200);
    expect(levelResponse.status).toBe(200);

    const dashboardHtml = await dashboardResponse.text();
    const levelHtml = await levelResponse.text();

    expect(dashboardHtml).toContain("pnsolve check ダッシュボード");
    expect(dashboardHtml).toContain("level3");
    expect(dashboardHtml).toContain("未実行");
    expect(levelHtml).toContain("差分あり");
    expect(levelHtml).toContain("jd diff");
  } finally {
    server.stop(true);
  }
});
