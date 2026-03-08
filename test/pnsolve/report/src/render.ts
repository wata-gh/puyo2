import { asPrettyJson } from "./diff";
import type { JsonValue } from "./artifact";
import { filterCases, type CaseFilter, type CaseView, type DashboardView, type LevelTone, type LevelView } from "./summary";

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function renderLayout(title: string, body: string): string {
  return `<!doctype html>
<html lang="ja">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="theme-color" content="#16181d" />
    <title>${escapeHtml(title)}</title>
    <link rel="stylesheet" href="/static/styles.css" />
  </head>
  <body>
    ${body}
  </body>
</html>`;
}

function statusLabel(status: string | null): string {
  if (!status) {
    return "なし";
  }
  return status;
}

function checkStatusLabel(status: CaseView["checkStatus"]): string {
  switch (status) {
    case "MATCH":
      return "一致";
    case "DIFF":
      return "差分あり";
    case "RUN_ERROR":
      return "実行エラー";
    case "JD_ERROR":
      return "比較エラー";
    case "BEFORE_MISSING":
      return "基準なし";
    case "UNEXECUTED":
      return "未実行";
  }
}

function toneLabel(tone: LevelTone): string {
  switch (tone) {
    case "clean":
      return "差分なし";
    case "diff":
      return "差分あり";
    case "error":
      return "エラーあり";
    case "unexecuted":
      return "未実行";
  }
}

function renderMetric(label: string, value: string, emphasis = false): string {
  return `<div class="metric${emphasis ? " metric-emphasis" : ""}">
    <div class="metric-label">${escapeHtml(label)}</div>
    <div class="metric-value">${escapeHtml(value)}</div>
  </div>`;
}

function renderHeader(dashboard: DashboardView, title: string, subtitle: string, backHref?: string): string {
  const backLink = backHref
    ? `<a class="back-link" href="${escapeHtml(backHref)}">← ダッシュボードへ戻る</a>`
    : "";
  return `<header class="page-header">
    <div class="page-header-inner">
      ${backLink}
      <div class="eyebrow">pnsolve check report</div>
      <h1>${escapeHtml(title)}</h1>
      <p class="page-subtitle">${escapeHtml(subtitle)}</p>
      <div class="meta-row">
        <span>generated: ${escapeHtml(dashboard.generatedAt)}</span>
        <span>artifact: ${escapeHtml(dashboard.artifactRoot)}</span>
      </div>
    </div>
  </header>`;
}

function renderDashboardCard(level: LevelView): string {
  const href = `/levels/${encodeURIComponent(level.name)}`;
  return `<a class="level-card tone-${level.tone}" href="${href}">
    <div class="level-band" aria-hidden="true"></div>
    <div class="level-card-body">
      <div class="level-card-top">
        <div>
          <div class="level-name">${escapeHtml(level.name)}</div>
          <div class="badge-row">
            <span class="badge">${level.executed ? "実行済み" : "未実行"}</span>
            <span class="badge">${escapeHtml(toneLabel(level.tone))}</span>
            <span class="badge badge-muted">${escapeHtml(`${level.checkedCases} / ${level.totalCases}`)}</span>
          </div>
        </div>
        <div class="diff-hero">
          <div class="diff-hero-value">${escapeHtml(`${level.diffCases}`)}</div>
          <div class="diff-hero-label">差分 case</div>
        </div>
      </div>
      <div class="metric-grid">
        ${renderMetric("error", String(level.errorCases))}
        ${renderMetric("solution +", String(level.solutionAdded))}
        ${renderMetric("solution -", String(level.solutionRemoved))}
        ${renderMetric("matched Δ", String(level.matchedDeltaSum))}
      </div>
    </div>
  </a>`;
}

function formatValue(value: JsonValue | null): string {
  if (typeof value === "string") {
    return value;
  }
  if (value === null) {
    return "null";
  }
  return JSON.stringify(value);
}

function renderDiffRows(caseView: CaseView): string {
  const rows = caseView.basicDiffRows
    .map((row) => `<tr class="${row.changed ? "row-changed" : "row-same"}">
      <th>${escapeHtml(row.path)}</th>
      <td class="diff-before">${escapeHtml(formatValue(row.before))}</td>
      <td class="diff-after">${escapeHtml(formatValue(row.after))}</td>
    </tr>`)
    .join("");

  return `<section class="panel">
    <h3>基本情報 diff</h3>
    <table class="diff-table">
      <thead>
        <tr>
          <th>path</th>
          <th>baseline</th>
          <th>current</th>
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
  </section>`;
}

function renderSolutionList(items: JsonValue[], variant: "added" | "removed"): string {
  if (items.length === 0) {
    return `<div class="empty-state">なし</div>`;
  }
  return items
    .map((item) => {
      const obj = item && typeof item === "object" && !Array.isArray(item) ? (item as Record<string, JsonValue>) : null;
      const hands = obj && typeof obj.hands === "string" ? obj.hands : null;
      const initialField = obj && typeof obj.initialField === "string" ? obj.initialField : null;
      const link =
        hands && initialField
          ? `<a href="http://localhost:3000/simus?${new URLSearchParams({ fs: initialField, h: hands }).toString()}" target="_blank" rel="noreferrer">simus</a>`
          : "";
      return `<div class="solution-card ${variant}">
        <div class="solution-card-header">
          <span>${hands ? escapeHtml(hands) : "solution"}</span>
          ${link}
        </div>
        <pre>${escapeHtml(asPrettyJson(item))}</pre>
      </div>`;
    })
    .join("");
}

function renderSolutionDiff(caseView: CaseView): string {
  return `<section class="panel">
    <h3>solutions の追加 / 削除</h3>
    <div class="split-grid">
      <div class="column column-removed">
        <div class="column-title">baseline のみ</div>
        ${renderSolutionList(caseView.removedSolutions, "removed")}
      </div>
      <div class="column column-added">
        <div class="column-title">current のみ</div>
        ${renderSolutionList(caseView.addedSolutions, "added")}
      </div>
    </div>
  </section>`;
}

function renderRawJson(caseView: CaseView): string {
  return `<section class="panel">
    <h3>生 JSON</h3>
    <div class="split-grid">
      <div class="column">
        <div class="column-title">before.raw.json</div>
        <pre>${escapeHtml(asPrettyJson(caseView.beforeRaw))}</pre>
      </div>
      <div class="column">
        <div class="column-title">after.raw.json</div>
        <pre>${escapeHtml(asPrettyJson(caseView.afterRaw))}</pre>
      </div>
    </div>
  </section>`;
}

function renderSolutionLinks(caseView: CaseView): string {
  if (caseView.solutionLinks.length === 0) {
    return `<div class="link-list empty-state">current solutions なし</div>`;
  }
  return `<div class="link-list">
    ${caseView.solutionLinks
      .map(
        (link, index) =>
          `<a href="${escapeHtml(link.url)}" target="_blank" rel="noreferrer">solution ${index + 1}: ${escapeHtml(link.hands || "(empty)")}</a>`,
      )
      .join("")}
  </div>`;
}

function renderCaseDetails(caseView: CaseView): string {
  const errorBlock = caseView.stderrText
    ? `<section class="panel">
        <h3>stderr</h3>
        <pre>${escapeHtml(caseView.stderrText)}</pre>
      </section>`
    : "";
  const jdBlock = caseView.diffText
    ? `<section class="panel">
        <h3>jd diff</h3>
        <pre>${escapeHtml(caseView.diffText)}</pre>
      </section>`
    : "";
  const compactMatch = caseView.checkStatus === "MATCH"
    ? `<section class="panel">
        <h3>差分</h3>
        <div class="empty-state">差分なし</div>
      </section>`
    : "";

  return `<div class="case-details">
    <section class="panel">
      <h3>リンク</h3>
      <div class="link-list">
        <a href="${escapeHtml(caseView.originalUrl)}" target="_blank" rel="noreferrer">オリジナルページ</a>
      </div>
      ${renderSolutionLinks(caseView)}
    </section>
    ${compactMatch}
    ${renderDiffRows(caseView)}
    ${renderSolutionDiff(caseView)}
    ${renderRawJson(caseView)}
    ${jdBlock}
    ${errorBlock}
  </div>`;
}

function renderCaseCard(caseView: CaseView): string {
  const matchedDelta = caseView.matchedDelta === null ? "-" : String(caseView.matchedDelta);
  return `<details class="case-card case-${caseView.checkStatus.toLowerCase()}">
    <summary>
      <div class="case-summary">
        <div class="case-main">
          <div class="case-title-row">
            <span class="case-param">${escapeHtml(caseView.param)}</span>
            <span class="case-check-status badge">${escapeHtml(checkStatusLabel(caseView.checkStatus))}</span>
          </div>
          <div class="case-condition">${escapeHtml(caseView.conditionText ?? "条件テキストなし")}</div>
        </div>
        <div class="case-stats">
          <div><span>baseline</span><strong>${escapeHtml(statusLabel(caseView.baselineStatus))}</strong></div>
          <div><span>current</span><strong>${escapeHtml(statusLabel(caseView.currentStatus))}</strong></div>
          <div><span>matched Δ</span><strong>${escapeHtml(matchedDelta)}</strong></div>
          <div><span>solutions</span><strong>${escapeHtml(String(caseView.currentSolutionCount))}</strong></div>
        </div>
      </div>
    </summary>
    ${renderCaseDetails(caseView)}
  </details>`;
}

function renderFilterLink(level: LevelView, current: CaseFilter, target: CaseFilter, label: string): string {
  const href = target === "all"
    ? `/levels/${encodeURIComponent(level.name)}`
    : `/levels/${encodeURIComponent(level.name)}?filter=${target}`;
  const className = current === target ? "filter-pill current" : "filter-pill";
  return `<a class="${className}" href="${href}">${escapeHtml(label)}</a>`;
}

export function renderDashboardPage(dashboard: DashboardView): string {
  const body = `
    ${renderHeader(dashboard, "pnsolve check ダッシュボード", "各 level の実行状況と差分量を一覧します。")}
    <main class="page-shell">
      <section class="panel intro-panel">
        <h2>Summary</h2>
        <div class="metric-grid wide">
          ${renderMetric("levels", String(dashboard.levels.length))}
          ${renderMetric("executed", String(dashboard.levels.filter((level) => level.executed).length))}
          ${renderMetric("diff levels", String(dashboard.levels.filter((level) => level.diffCases > 0).length))}
          ${renderMetric("error levels", String(dashboard.levels.filter((level) => level.errorCases > 0).length))}
        </div>
      </section>
      <section class="level-grid">
        ${dashboard.levels.map(renderDashboardCard).join("")}
      </section>
    </main>
  `;
  return renderLayout("pnsolve check report", body);
}

export function renderLevelPage(dashboard: DashboardView, level: LevelView, filter: CaseFilter): string {
  const filteredCases = filterCases(level, filter);
  const body = `
    ${renderHeader(dashboard, `${level.name} 詳細`, `${level.checkedCases} / ${level.totalCases} cases checked`, "/")}
    <main class="page-shell">
      <section class="panel">
        <div class="level-summary-top">
          <div>
            <div class="badge-row">
              <span class="badge">${level.executed ? "実行済み" : "未実行"}</span>
              <span class="badge badge-muted">${escapeHtml(toneLabel(level.tone))}</span>
            </div>
            <h2>${escapeHtml(level.name)}</h2>
          </div>
          <div class="filter-row">
            ${renderFilterLink(level, filter, "all", "すべて")}
            ${renderFilterLink(level, filter, "diff", "差分あり")}
            ${renderFilterLink(level, filter, "error", "エラー")}
            ${renderFilterLink(level, filter, "match", "一致")}
          </div>
        </div>
        <div class="metric-grid wide">
          ${renderMetric("checked / total", `${level.checkedCases} / ${level.totalCases}`, true)}
          ${renderMetric("diff", String(level.diffCases))}
          ${renderMetric("error", String(level.errorCases))}
          ${renderMetric("solution +", String(level.solutionAdded))}
          ${renderMetric("solution -", String(level.solutionRemoved))}
          ${renderMetric("matched Δ", String(level.matchedDeltaSum))}
        </div>
      </section>
      <section class="case-list">
        ${filteredCases.length > 0 ? filteredCases.map(renderCaseCard).join("") : `<div class="panel empty-state">該当する問題はありません。</div>`}
      </section>
    </main>
  `;
  return renderLayout(`${level.name} | pnsolve check report`, body);
}

export function renderNotFoundPage(dashboard: DashboardView, message: string): string {
  return renderLayout(
    "Not Found",
    `
      ${renderHeader(dashboard, "ページが見つかりません", message, "/")}
      <main class="page-shell">
        <section class="panel empty-state">${escapeHtml(message)}</section>
      </main>
    `,
  );
}

export function renderErrorPage(error: unknown): string {
  const message = error instanceof Error ? error.message : String(error);
  return renderLayout(
    "Error",
    `
      <main class="page-shell standalone">
        <section class="panel error-panel">
          <div class="eyebrow">pnsolve check report</div>
          <h1>レポートを読み込めませんでした</h1>
          <pre>${escapeHtml(message)}</pre>
        </section>
      </main>
    `,
  );
}
