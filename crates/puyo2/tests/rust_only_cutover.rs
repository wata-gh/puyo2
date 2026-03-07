use std::{
    fs,
    path::{Path, PathBuf},
};

fn repo_root() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .ancestors()
        .nth(2)
        .unwrap()
        .to_path_buf()
}

fn collect_files(root: &Path, out: &mut Vec<PathBuf>) {
    let entries = fs::read_dir(root).unwrap_or_else(|err| panic!("read_dir failed: {err}"));
    let mut paths = entries
        .map(|entry| entry.unwrap().path())
        .collect::<Vec<PathBuf>>();
    paths.sort();

    for path in paths {
        if path.is_dir() {
            if path.file_name().and_then(|name| name.to_str()) == Some("target") {
                continue;
            }
            collect_files(&path, out);
        } else {
            out.push(path);
        }
    }
}

#[test]
fn repository_contains_no_legacy_go_or_ruby_sources() {
    let root = repo_root();
    let mut files = Vec::new();
    collect_files(&root, &mut files);

    let legacy = files
        .into_iter()
        .filter(|path| {
            matches!(
                path.extension().and_then(|ext| ext.to_str()),
                Some("go") | Some("rb")
            ) || matches!(
                path.file_name().and_then(|name| name.to_str()),
                Some("go.mod") | Some("go.sum")
            )
        })
        .map(|path| path.strip_prefix(&root).unwrap().display().to_string())
        .collect::<Vec<String>>();

    assert!(
        legacy.is_empty(),
        "legacy Go/Ruby files remain: {}",
        legacy.join(", ")
    );
}

#[test]
fn docs_and_ci_are_rust_only() {
    let root = repo_root();
    let readme = fs::read_to_string(root.join("README.md")).unwrap();
    let agents = fs::read_to_string(root.join("AGENTS.md")).unwrap();
    let rust_workflow = fs::read_to_string(root.join(".github/workflows/rust.yml")).unwrap();
    let pnsolve_check = fs::read_to_string(root.join("test/pnsolve/check")).unwrap();
    let pnsolve_solve = fs::read_to_string(root.join("test/pnsolve/solve")).unwrap();

    assert!(
        !root.join(".github/workflows/go.yml").exists(),
        "legacy Go workflow must be removed"
    );

    for (name, content) in [
        ("README.md", readme.as_str()),
        ("AGENTS.md", agents.as_str()),
        ("rust.yml", rust_workflow.as_str()),
        ("test/pnsolve/check", pnsolve_check.as_str()),
        ("test/pnsolve/solve", pnsolve_solve.as_str()),
    ] {
        assert!(
            !content.contains("go build ./..."),
            "{name} still references go build"
        );
        assert!(
            !content.contains("go test ./..."),
            "{name} still references go test"
        );
        assert!(
            !content.contains("go run ./cmd/"),
            "{name} still references Go CLI paths"
        );
        assert!(
            !content.contains("go.mod"),
            "{name} still references go.mod"
        );
        assert!(
            !content.contains(".github/workflows/go.yml"),
            "{name} still references the removed Go workflow"
        );
    }

    assert!(readme.contains("cargo build --workspace --all-targets"));
    assert!(readme.contains("cargo test --workspace"));
    assert!(agents.contains("cargo build --workspace --all-targets"));
    assert!(agents.contains("cargo test --workspace"));
}
