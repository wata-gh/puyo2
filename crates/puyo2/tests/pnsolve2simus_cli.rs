mod common;

use std::fs;

use crate::common::{output_strings, run_bin, write_temp_file};

#[test]
fn pnsolve2simus_stdin_and_file_input_match_contract() {
    let payload = r#"{
  "initialField": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "solutions": [
    { "hands": "rr00" },
    { "hands": "gb10" }
  ]
}"#;
    let stdin_output = run_bin("pnsolve2simus", &[], Some(payload));
    let (stdin_stdout, stdin_stderr) = output_strings(&stdin_output);

    assert!(stdin_output.status.success(), "stderr={stdin_stderr}");
    assert!(stdin_stderr.is_empty());
    assert_eq!(
        stdin_stdout,
        "https://puyo-rsrch.com/simus?fs=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&h=rr00\nhttps://puyo-rsrch.com/simus?fs=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&h=gb10\n"
    );

    let path = write_temp_file("pnsolve2simus", payload);
    let file_output = run_bin("pnsolve2simus", &["--local", path.to_str().unwrap()], None);
    let (file_stdout, file_stderr) = output_strings(&file_output);
    let _ = fs::remove_file(path);

    assert!(file_output.status.success(), "stderr={file_stderr}");
    assert!(file_stderr.is_empty());
    assert_eq!(
        file_stdout,
        "http://localhost:3000/simus?fs=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&h=rr00\nhttp://localhost:3000/simus?fs=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&h=gb10\n"
    );
}

#[test]
fn pnsolve2simus_invalid_json_and_missing_fields_match_contract() {
    let invalid_json = run_bin("pnsolve2simus", &[], Some("{"));
    let missing_initial_field = run_bin(
        "pnsolve2simus",
        &[],
        Some(r#"{"solutions":[{"hands":"rr00"}]}"#),
    );
    let missing_solutions = run_bin("pnsolve2simus", &[], Some(r#"{"initialField":"a"}"#));
    let missing_hands = run_bin(
        "pnsolve2simus",
        &[],
        Some(r#"{"initialField":"a","solutions":[{}]}"#),
    );
    let top_level_array = run_bin("pnsolve2simus", &[], Some("[]"));

    assert!(!invalid_json.status.success());
    assert!(output_strings(&invalid_json).1.contains("invalid JSON:"));

    assert!(!missing_initial_field.status.success());
    assert!(
        output_strings(&missing_initial_field)
            .1
            .contains("initialField is missing or empty")
    );

    assert!(!missing_solutions.status.success());
    assert!(
        output_strings(&missing_solutions)
            .1
            .contains("solutions must be an array")
    );

    assert!(!missing_hands.status.success());
    assert!(
        output_strings(&missing_hands)
            .1
            .contains("solutions[0].hands is missing or empty")
    );

    assert!(!top_level_array.status.success());
    assert!(
        output_strings(&top_level_array)
            .1
            .contains("top-level JSON must be an object")
    );
}

#[test]
fn pnsolve2simus_file_not_found_and_too_many_args_fail() {
    let missing_file = run_bin("pnsolve2simus", &["/no/such/file.json"], None);
    let too_many_args = run_bin("pnsolve2simus", &["a.json", "b.json"], None);

    assert!(!missing_file.status.success());
    assert!(
        output_strings(&missing_file)
            .1
            .contains("file not found: /no/such/file.json")
    );

    assert!(!too_many_args.status.success());
    assert!(
        output_strings(&too_many_args)
            .1
            .contains("too many arguments")
    );
    assert!(
        output_strings(&too_many_args)
            .1
            .contains("Usage: pnsolve2simus")
    );
}
