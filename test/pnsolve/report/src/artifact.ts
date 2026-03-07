import fs from "node:fs/promises";
import path from "node:path";

export type JsonValue =
  | null
  | boolean
  | number
  | string
  | JsonValue[]
  | { [key: string]: JsonValue };

export type CheckStatus = "MATCH" | "DIFF" | "BEFORE_MISSING" | "RUN_ERROR" | "JD_ERROR";

export interface ArtifactManifest {
  formatVersion: number;
  createdAt: string;
  repoRoot: string;
  normalizedFilter: string;
  executedLevels: string[];
  singleMode: boolean;
  binPath: string;
  summary?: {
    checked: number;
    diff: number;
    missingBefore: number;
    runError: number;
    jdError: number;
  };
}

export interface CaseMeta {
  level: string;
  param: string;
  status: CheckStatus;
  beforePath: string | null;
  afterPath: string | null;
  beforeNormalizedPath: string | null;
  afterNormalizedPath: string | null;
  diffPath: string | null;
  stderrPath: string | null;
}

export interface CaseArtifacts {
  meta: CaseMeta | null;
  beforeRaw: JsonValue | null;
  afterRaw: JsonValue | null;
  beforeNormalized: JsonValue | null;
  afterNormalized: JsonValue | null;
  diffText: string | null;
  stderrText: string | null;
}

export interface LevelArtifacts {
  name: string;
  params: string[];
  cases: Map<string, CaseArtifacts>;
}

export interface ArtifactReport {
  artifactRoot: string;
  manifest: ArtifactManifest;
  repoRoot: string;
  levels: LevelArtifacts[];
}

async function readJsonFile<T>(filePath: string): Promise<T> {
  const text = await fs.readFile(filePath, "utf8");
  return JSON.parse(text) as T;
}

async function readTextIfExists(filePath: string | null): Promise<string | null> {
  if (!filePath) {
    return null;
  }
  try {
    return await fs.readFile(filePath, "utf8");
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") {
      return null;
    }
    throw error;
  }
}

async function readJsonIfExists(filePath: string | null): Promise<JsonValue | null> {
  if (!filePath) {
    return null;
  }
  try {
    return await readJsonFile<JsonValue>(filePath);
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") {
      return null;
    }
    throw error;
  }
}

async function listLevelNames(repoRoot: string): Promise<string[]> {
  const baseDir = path.join(repoRoot, "test/pnsolve");
  const entries = await fs.readdir(baseDir, { withFileTypes: true });
  const levelNames = entries
    .filter((entry) => entry.isDirectory() && /^level\d+$/.test(entry.name))
    .map((entry) => entry.name)
    .sort((left, right) => left.localeCompare(right, undefined, { numeric: true }));
  return levelNames;
}

async function readLevelParams(repoRoot: string, level: string): Promise<string[]> {
  const listPath = path.join(repoRoot, "test/pnsolve", level, "list");
  try {
    const content = await fs.readFile(listPath, "utf8");
    return content
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter((line) => line.length > 0);
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") {
      return [];
    }
    throw error;
  }
}

async function loadCaseArtifacts(artifactRoot: string, level: string, param: string): Promise<CaseArtifacts | null> {
  const caseDir = path.join(artifactRoot, "levels", level, "cases", param);
  const metaPath = path.join(caseDir, "meta.json");
  const meta = await readJsonIfExists(metaPath);
  if (!meta || typeof meta !== "object" || Array.isArray(meta)) {
    return null;
  }

  const parsedMeta = meta as unknown as CaseMeta;
  const resolve = (relativePath: string | null) =>
    relativePath ? path.join(artifactRoot, relativePath) : null;

  return {
    meta: parsedMeta,
    beforeRaw: await readJsonIfExists(resolve(parsedMeta.beforePath)),
    afterRaw: await readJsonIfExists(resolve(parsedMeta.afterPath)),
    beforeNormalized: await readJsonIfExists(resolve(parsedMeta.beforeNormalizedPath)),
    afterNormalized: await readJsonIfExists(resolve(parsedMeta.afterNormalizedPath)),
    diffText: await readTextIfExists(resolve(parsedMeta.diffPath)),
    stderrText: await readTextIfExists(resolve(parsedMeta.stderrPath)),
  };
}

async function listArtifactCaseParams(artifactRoot: string, level: string): Promise<string[]> {
  const casesDir = path.join(artifactRoot, "levels", level, "cases");
  try {
    const entries = await fs.readdir(casesDir, { withFileTypes: true });
    return entries
      .filter((entry) => entry.isDirectory())
      .map((entry) => entry.name)
      .sort((left, right) => left.localeCompare(right));
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") {
      return [];
    }
    throw error;
  }
}

export async function loadArtifactReport(artifactRootInput: string): Promise<ArtifactReport> {
  const artifactRoot = path.resolve(artifactRootInput);
  const manifestPath = path.join(artifactRoot, "manifest.json");
  const manifest = await readJsonFile<ArtifactManifest>(manifestPath);
  const repoRoot = path.resolve(manifest.repoRoot);

  const allRepoLevels = await listLevelNames(repoRoot);
  const levelNames = Array.from(new Set([...allRepoLevels, ...manifest.executedLevels])).sort((left, right) =>
    left.localeCompare(right, undefined, { numeric: true }),
  );

  const levels: LevelArtifacts[] = [];
  for (const level of levelNames) {
    const levelParams = await readLevelParams(repoRoot, level);
    const artifactParams = await listArtifactCaseParams(artifactRoot, level);
    const orderedParams = Array.from(new Set([...levelParams, ...artifactParams]));
    const cases = new Map<string, CaseArtifacts>();

    for (const param of orderedParams) {
      const caseArtifacts = await loadCaseArtifacts(artifactRoot, level, param);
      if (caseArtifacts) {
        cases.set(param, caseArtifacts);
      }
    }

    levels.push({
      name: level,
      params: orderedParams,
      cases,
    });
  }

  return {
    artifactRoot,
    manifest,
    repoRoot,
    levels,
  };
}
