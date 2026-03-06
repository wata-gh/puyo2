mod common;

use puyo2::parse_ips_nazo_url;

use crate::common::{output_strings, run_bin};

#[test]
fn pnconv_single_input_matches_go_text_output() {
    let input = "800F08J08A0EB_8161__270";
    let decoded = parse_ips_nazo_url(input).unwrap();
    let output = run_bin("pnconv", &["-param", input], None);
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());
    let expected = format!(
        "Initial Field: {}\nHaipuyo: {}\nCondition: {}\nConditionCode: q0={} q1={} q2={}\n",
        decoded.initial_field,
        decoded.haipuyo,
        decoded.condition.text,
        decoded.condition_code[0],
        decoded.condition_code[1],
        decoded.condition_code[2],
    );
    assert_eq!(stdout, expected);
}

#[test]
fn pnconv_raw_query_and_path_style_match() {
    let raw = "800F08J08A0EB_8161__270";
    let path_style = "pn.html?800F08J08A0EB_8161__270";
    let url = "https://ips.karou.jp/simu/pn.html?800F08J08A0EB_8161__270";

    let raw_output = run_bin("pnconv", &["-param", raw], None);
    let path_output = run_bin("pnconv", &["-param", path_style], None);
    let url_output = run_bin("pnconv", &["-url", url], None);

    assert_eq!(
        output_strings(&raw_output).0,
        output_strings(&path_output).0
    );
    assert_eq!(output_strings(&raw_output).0, output_strings(&url_output).0);
}

#[test]
fn pnconv_stdin_multiple_inputs_preserves_success_output_and_exits_nonzero_on_error() {
    let stdin = "800F08J08A0EB_8161__270\n!\n";
    let output = run_bin("pnconv", &[], Some(stdin));
    let (stdout, stderr) = output_strings(&output);

    assert!(!output.status.success());
    assert!(stdout.contains("Initial Field: "));
    assert!(stdout.contains("Haipuyo: "));
    assert!(stderr.contains("parse error: !:"));
}
