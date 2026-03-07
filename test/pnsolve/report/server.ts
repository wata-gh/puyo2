import path from "node:path";
import { loadArtifactReport } from "./src/artifact";
import { renderDashboardPage, renderErrorPage, renderLevelPage, renderNotFoundPage } from "./src/render";
import { buildDashboardView, getLevelView, parseFilter } from "./src/summary";

export interface ServerConfig {
  artifactRoot: string;
  port: number;
}

function parseArgs(argv: string[]): ServerConfig {
  const repoRoot = path.resolve(import.meta.dir, "../../..");
  let artifactRoot = path.resolve(repoRoot, "test/pnsolve/artifacts/latest");
  let port = 8787;

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--artifact") {
      const value = argv[i + 1];
      if (!value) {
        throw new Error("--artifact requires a value");
      }
      artifactRoot = path.resolve(process.cwd(), value);
      i += 1;
      continue;
    }
    if (arg === "--port") {
      const value = argv[i + 1];
      if (!value) {
        throw new Error("--port requires a value");
      }
      const parsed = Number.parseInt(value, 10);
      if (!Number.isInteger(parsed) || parsed <= 0) {
        throw new Error(`invalid port: ${value}`);
      }
      port = parsed;
      i += 1;
      continue;
    }
    throw new Error(`unknown argument: ${arg}`);
  }

  return { artifactRoot, port };
}

export function createFetchHandler(config: ServerConfig) {
  return async function fetchHandler(request: Request): Promise<Response> {
    const url = new URL(request.url);

    try {
      if (url.pathname === "/static/styles.css") {
        const cssPath = path.join(import.meta.dir, "static/styles.css");
        return new Response(Bun.file(cssPath), {
          headers: { "content-type": "text/css; charset=utf-8" },
        });
      }

      const report = await loadArtifactReport(config.artifactRoot);
      const dashboard = buildDashboardView(report);

      if (url.pathname === "/") {
        const html = renderDashboardPage(dashboard);
        return new Response(html, {
          headers: { "content-type": "text/html; charset=utf-8" },
        });
      }

      const levelMatch = url.pathname.match(/^\/levels\/([^/]+)$/);
      if (levelMatch) {
        const levelName = decodeURIComponent(levelMatch[1] ?? "");
        const filter = parseFilter(url.searchParams.get("filter"));
        const level = getLevelView(dashboard, levelName);
        if (!level) {
          const html = renderNotFoundPage(dashboard, `level not found: ${levelName}`);
          return new Response(html, {
            status: 404,
            headers: { "content-type": "text/html; charset=utf-8" },
          });
        }
        const html = renderLevelPage(dashboard, level, filter);
        return new Response(html, {
          headers: { "content-type": "text/html; charset=utf-8" },
        });
      }

      const html = renderNotFoundPage(dashboard, `path not found: ${url.pathname}`);
      return new Response(html, {
        status: 404,
        headers: { "content-type": "text/html; charset=utf-8" },
      });
    } catch (error) {
      const html = renderErrorPage(error);
      return new Response(html, {
        status: 500,
        headers: { "content-type": "text/html; charset=utf-8" },
      });
    }
  };
}

export function startServer(config: ServerConfig) {
  const server = Bun.serve({
    port: config.port,
    hostname: "127.0.0.1",
    fetch: createFetchHandler(config),
  });

  console.log(`pnsolve report: http://127.0.0.1:${server.port}/`);
  console.log(`artifact root: ${config.artifactRoot}`);
  return server;
}

if (import.meta.main) {
  const config = parseArgs(Bun.argv.slice(2));
  startServer(config);
}
