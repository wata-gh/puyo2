mod common;

use serde::Deserialize;

use puyo2::parse_ips_nazo_url;

use crate::common::{output_strings, run_bin, run_bin_with_env};

#[derive(Clone, Debug, Deserialize)]
struct SolveConditionJson {
    q0: usize,
    q1: usize,
    q2: usize,
    text: String,
}

#[derive(Clone, Debug, Deserialize)]
struct SolveSolutionJson {
    hands: String,
    chains: usize,
    score: usize,
    clear: bool,
    #[serde(rename = "initialField")]
    initial_field: String,
    #[serde(rename = "finalField")]
    final_field: String,
}

#[derive(Clone, Debug, Deserialize)]
struct SolveOutputJson {
    input: String,
    #[serde(rename = "initialField")]
    initial_field: String,
    haipuyo: String,
    status: String,
    #[serde(default)]
    error: String,
    condition: SolveConditionJson,
    searched: usize,
    matched: usize,
    solutions: Vec<SolveSolutionJson>,
}

fn decode_output(stdout: &str) -> SolveOutputJson {
    serde_json::from_str(stdout).unwrap_or_else(|err| panic!("invalid json: {err}\n{stdout}"))
}

fn is_all_clear_field(field: &str) -> bool {
    field.chars().all(|ch| ch == 'a')
}

#[test]
fn pnsolve_pretty_json_matches_go_contract() {
    let param = "800F08J08A0EB_8161__270";
    let decoded = parse_ips_nazo_url(param).unwrap();
    let output = run_bin("pnsolve", &["-param", param], None);
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());

    let out = decode_output(&stdout);
    assert_eq!(out.initial_field, decoded.initial_field);
    assert_eq!(out.haipuyo, decoded.haipuyo);
    assert_eq!(out.condition.q0, decoded.condition.q0);
    assert_eq!(out.condition.q1, decoded.condition.q1);
    assert_eq!(out.condition.q2, decoded.condition.q2);
    assert_eq!(out.condition.text, decoded.condition.text);
    assert_eq!(out.matched, out.solutions.len());
    assert_eq!(out.status, "ok");
    assert!(out.searched >= out.matched);
    assert!(stdout.contains("\n  \"status\": \"ok\""));

    for solution in out.solutions {
        assert_eq!(solution.initial_field.len(), 78);
        assert_eq!(solution.final_field.len(), 78);
        assert_eq!(solution.clear, is_all_clear_field(&solution.final_field));
    }
}

#[test]
fn pnsolve_compact_json_matches_go_contract() {
    let param = "80080080oM0oM098_4141__u03";
    let output = run_bin("pnsolve", &["-param", param, "-pretty=false"], None);
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());
    assert!(!stdout.contains("\n  \""));

    let out = decode_output(&stdout);
    assert_eq!(out.input, param);
    assert!(matches!(out.status.as_str(), "ok" | "no_solution"));
}

#[test]
fn pnsolve_default_matches_explicit_optimization_flags() {
    let param = "800F08J08A0EB_8161__270";
    let default_output = run_bin("pnsolve", &["-param", param, "-pretty=false"], None);
    let explicit_output = run_bin(
        "pnsolve",
        &[
            "-param",
            param,
            "-pretty=false",
            "-dedup",
            "same_pair_order",
            "-simulate",
            "fast_intermediate",
        ],
        None,
    );

    assert_eq!(output_strings(&default_output).1, "");
    assert_eq!(output_strings(&explicit_output).1, "");
    assert_eq!(
        output_strings(&default_output).0,
        output_strings(&explicit_output).0
    );
}

#[test]
fn pnsolve_same_pair_order_is_noop_without_stop_on_chain() {
    let param = "jjgqqqqqqqqq_q1q1q1__u06";
    let same_pair_output = run_bin(
        "pnsolve",
        &[
            "-param",
            param,
            "-pretty=false",
            "-dedup",
            "same_pair_order",
            "-simulate",
            "detail_always",
        ],
        None,
    );
    let off_output = run_bin(
        "pnsolve",
        &[
            "-param",
            param,
            "-pretty=false",
            "-dedup",
            "off",
            "-simulate",
            "detail_always",
        ],
        None,
    );

    assert_eq!(output_strings(&same_pair_output).1, "");
    assert_eq!(output_strings(&off_output).1, "");
    assert_eq!(
        output_strings(&same_pair_output).0,
        output_strings(&off_output).0
    );

    let out = decode_output(&output_strings(&same_pair_output).0);
    assert_eq!(out.matched, out.solutions.len());
    assert_eq!(out.matched, 4);
}

#[test]
fn pnsolve_invalid_flags_and_input_match_go_failures() {
    let invalid_dedup = run_bin(
        "pnsolve",
        &["-param", "800F08J08A0EB_8161__270", "-dedup", "x"],
        None,
    );
    let invalid_simulate = run_bin(
        "pnsolve",
        &["-param", "800F08J08A0EB_8161__270", "-simulate", "x"],
        None,
    );
    let invalid_input = run_bin("pnsolve", &["-param", "!"], None);

    assert!(!invalid_dedup.status.success());
    assert!(
        output_strings(&invalid_dedup)
            .1
            .contains("unknown dedup mode")
    );

    assert!(!invalid_simulate.status.success());
    assert!(
        output_strings(&invalid_simulate)
            .1
            .contains("unknown simulate policy")
    );

    assert!(!invalid_input.status.success());
    assert!(output_strings(&invalid_input).1.contains("parse error"));
}

#[test]
fn pnsolve_search_failed_json_matches_go_contract() {
    let param = "o00800c00b00j00z35xx4yxiqr9aticBIbrA_G1A1__u0b";
    let output = run_bin("pnsolve", &["-param", param, "-pretty=false"], None);
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());

    let out = decode_output(&stdout);
    assert_eq!(out.status, "search_failed");
    assert!(!out.error.is_empty());
    assert_eq!(out.matched, 0);
    assert!(out.solutions.is_empty());
}

#[test]
fn pnsolve_no_index_13_panic_regression() {
    let param = "4r06P06904S04y03903N03Q02Q02k_A101o1E1__u07";
    let output = run_bin("pnsolve", &["-param", param, "-pretty=false"], None);
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());

    let out = decode_output(&stdout);
    assert_ne!(out.status, "search_failed");
    assert!(!out.error.contains("index out of range [13]"));
}

#[test]
fn pnsolve_non_all_clear_no_hands_regression_matches_go() {
    let param =
        "~000000000000000000000000000000000000000000000000000000000000000000000000111101___a01";
    let output = run_bin("pnsolve", &["-param", param, "-pretty=false"], None);
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());

    let out = decode_output(&stdout);
    assert_eq!(out.haipuyo, "");
    assert_eq!(out.searched, 1);
    assert_eq!(out.matched, 1);
    assert_eq!(out.solutions.len(), 1);
    assert!(!out.solutions[0].clear);
    assert!(!is_all_clear_field(&out.solutions[0].final_field));
    assert_eq!(out.status, "ok");
    assert_eq!(out.solutions[0].hands, "");
    assert!(out.solutions[0].chains > 0);
    assert!(out.solutions[0].score > 0);
}

#[test]
fn pnsolve_parallel_matches_single_worker_output() {
    let param = "jjgqqqqqqqqq_q1q1q1__u06";
    let single = run_bin_with_env(
        "pnsolve",
        &["-param", param, "-pretty=false"],
        None,
        &[("PUYO2_PNSOLVE_JOBS", "1")],
    );
    let parallel = run_bin_with_env(
        "pnsolve",
        &["-param", param, "-pretty=false"],
        None,
        &[("PUYO2_PNSOLVE_JOBS", "4")],
    );

    assert_eq!(output_strings(&single).1, "");
    assert_eq!(output_strings(&parallel).1, "");
    assert_eq!(output_strings(&single).0, output_strings(&parallel).0);
}

#[test]
fn pnsolve_parallel_solution_order_is_stable() {
    let param = "jjgqqqqqqqqq_q1q1q1__u06";
    let jobs_2 = run_bin_with_env(
        "pnsolve",
        &["-param", param, "-pretty=false"],
        None,
        &[("PUYO2_PNSOLVE_JOBS", "2")],
    );
    let jobs_4 = run_bin_with_env(
        "pnsolve",
        &["-param", param, "-pretty=false"],
        None,
        &[("PUYO2_PNSOLVE_JOBS", "4")],
    );
    let jobs_8 = run_bin_with_env(
        "pnsolve",
        &["-param", param, "-pretty=false"],
        None,
        &[("PUYO2_PNSOLVE_JOBS", "8")],
    );

    let out_2 = decode_output(&output_strings(&jobs_2).0);
    let out_4 = decode_output(&output_strings(&jobs_4).0);
    let out_8 = decode_output(&output_strings(&jobs_8).0);

    let hands_2 = out_2
        .solutions
        .iter()
        .map(|solution| solution.hands.clone())
        .collect::<Vec<_>>();
    let hands_4 = out_4
        .solutions
        .iter()
        .map(|solution| solution.hands.clone())
        .collect::<Vec<_>>();
    let hands_8 = out_8
        .solutions
        .iter()
        .map(|solution| solution.hands.clone())
        .collect::<Vec<_>>();

    assert_eq!(hands_2, hands_4);
    assert_eq!(hands_2, hands_8);
    assert_eq!(out_2.searched, out_4.searched);
    assert_eq!(out_2.searched, out_8.searched);
}

#[test]
fn pnsolve_parallel_search_failed_json_matches_go_contract() {
    let param = "o00800c00b00j00z35xx4yxiqr9aticBIbrA_G1A1__u0b";
    let output = run_bin_with_env(
        "pnsolve",
        &["-param", param, "-pretty=false"],
        None,
        &[("PUYO2_PNSOLVE_JOBS", "4")],
    );
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());

    let out = decode_output(&stdout);
    assert_eq!(out.status, "search_failed");
    assert!(!out.error.is_empty());
    assert_eq!(out.matched, 0);
    assert!(out.solutions.is_empty());
}
