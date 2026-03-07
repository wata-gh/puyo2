mod common;

use std::{
    collections::HashMap,
    io::{BufRead, BufReader, Read},
    process::Child,
    sync::mpsc::{self, Receiver},
    thread,
    time::{Duration, Instant},
};

use crate::common::{output_strings, run_bin, spawn_bin_with_env};

fn spawn_line_reader<R>(reader: R) -> Receiver<String>
where
    R: Read + Send + 'static,
{
    let (sender, receiver) = mpsc::channel();
    thread::spawn(move || {
        for line in BufReader::new(reader).lines() {
            match line {
                Ok(line) => {
                    if sender.send(line).is_err() {
                        break;
                    }
                }
                Err(_) => break,
            }
        }
    });
    receiver
}

fn wait_for_line<F>(
    receiver: &Receiver<String>,
    seen: &mut Vec<String>,
    timeout: Duration,
    predicate: F,
) -> Option<String>
where
    F: Fn(&str) -> bool,
{
    let deadline = Instant::now() + timeout;
    loop {
        let remaining = deadline.checked_duration_since(Instant::now())?;
        let line = receiver.recv_timeout(remaining).ok()?;
        seen.push(line.clone());
        if predicate(&line) {
            return Some(line);
        }
    }
}

fn finish_child(child: &mut Child) {
    let _ = child.kill();
    let _ = child.wait();
}

fn lines(text: &str) -> Vec<&str> {
    text.lines().filter(|line| !line.is_empty()).collect()
}

#[test]
fn nazo_help_prints_banner_and_usage() {
    let output = run_bin("nazo", &["-h"], None);
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stdout.contains("trap goroutine start."));
    assert!(stdout.contains("send SIGHUP to show progress."));
    assert!(stderr.contains("Usage of "));
    assert!(stderr.contains("-stop-on-chain"));
    assert!(!stderr.contains("elapsed:"));
}

#[test]
fn nazo_multi_hand_fixture_matches_expected_text_contract() {
    let output = run_bin(
        "nazo",
        &[
            "-param",
            "a62gacbagecb2ae2g3",
            "-hands",
            "rbgb",
            "-chains",
            "3",
        ],
        None,
    );
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stdout.contains("AllClear:false"));
    assert!(stdout.contains("Hands:rbgb"));
    assert!(stdout.contains("cpus: "));
    assert!(stdout.contains("14: ......"));
    assert!(stdout.contains("01: .GGOOO"));
    assert!(stdout.contains("rb33gb01"));
    assert!(stdout.contains("Chains:3"));
    assert!(stdout.contains("Score:1000"));
    assert!(stdout.contains("Quick:true"));
    assert!(stdout.contains("BitField:0x"));
    assert!(stdout.contains("NthResults:[0x"));
    assert!(stdout.contains(
        "https://pndsng.com/puyo/index.html?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaacaaaacgbcbagecbbeeeggg"
    ));
    assert!(stderr.contains("elapsed: "));
}

#[test]
fn nazo_fast_intermediate_matches_detail_scalar_output() {
    let detail = run_bin(
        "nazo",
        &[
            "-param",
            "a62gacbagecb2ae2g3",
            "-hands",
            "rbgb",
            "-chains",
            "3",
        ],
        None,
    );
    let fast = run_bin(
        "nazo",
        &[
            "-param",
            "a62gacbagecb2ae2g3",
            "-hands",
            "rbgb",
            "-chains",
            "3",
            "-simulate",
            "fast_intermediate",
        ],
        None,
    );
    let detail_stdout = output_strings(&detail).0;
    let fast_stdout = output_strings(&fast).0;

    for needle in [
        "rb33gb01",
        "Chains:3",
        "Score:1000",
        "Quick:true",
        "SetFrames:74",
    ] {
        assert!(detail_stdout.contains(needle), "{needle}");
        assert!(fast_stdout.contains(needle), "{needle}");
    }
}

#[test]
fn nazo_filters_cover_expected_cases() {
    let chains_gt = run_bin(
        "nazo",
        &[
            "-param",
            "a62gacbagecb2ae2g3",
            "-hands",
            "rbgb",
            "-chains",
            "3+",
        ],
        None,
    );
    assert!(output_strings(&chains_gt).0.contains("rb33gb01"));

    let clear_blue = run_bin(
        "nazo",
        &[
            "-param",
            "a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16",
            "-hands",
            "rrrrbbbb",
            "-clear",
            "b",
        ],
        None,
    );
    let clear_stdout = output_strings(&clear_blue).0;
    assert!(clear_stdout.contains("rr30rr30bb10bb30"));
    assert!(clear_stdout.contains("Chains:1"));

    let color_count = run_bin(
        "nazo",
        &[
            "-param",
            "a62gacbagecb2ae2g3",
            "-hands",
            "rbgb",
            "-color",
            "r",
            "-n",
            "4",
        ],
        None,
    );
    let color_stdout = output_strings(&color_count).0;
    assert!(color_stdout.contains("rb33gb01"));
    assert!(color_stdout.contains("Chains:3"));

    let stop_on_chain = run_bin(
        "nazo",
        &[
            "-param",
            "a62gacbagecb2ae2g3",
            "-hands",
            "rbgb",
            "-chains",
            "3",
            "-stop-on-chain",
        ],
        None,
    );
    assert!(
        output_strings(&stop_on_chain)
            .0
            .contains("StopOnChain:true")
    );
    assert!(output_strings(&stop_on_chain).0.contains("rb33gb01"));

    let disable_chigiri = run_bin(
        "nazo",
        &[
            "-param",
            "a62gacbagecb2ae2g3",
            "-hands",
            "rbgb",
            "-chains",
            "3",
            "-disablechigiri",
        ],
        None,
    );
    let disable_stdout = output_strings(&disable_chigiri).0;
    assert!(disable_stdout.contains("DisableChigiri:true"));
    assert!(!disable_stdout.contains("rb33gb01"));
}

#[test]
fn nazo_duplicate_filters_repeat_identical_lines() {
    let output = run_bin(
        "nazo",
        &[
            "-param",
            "a78",
            "-hands",
            "rrrr",
            "-allclear",
            "-clear",
            "r",
        ],
        None,
    );
    let stdout = output_strings(&output).0;
    let mut counts = HashMap::new();
    for line in lines(&stdout)
        .into_iter()
        .filter(|line| line.starts_with("https://pndsng.com/puyo/index.html?"))
    {
        *counts.entry(line.to_string()).or_insert(0usize) += 1;
    }
    assert!(counts.values().any(|count| *count >= 2), "{stdout}");
}

#[test]
fn nazo_one_hand_path_skips_board_debug() {
    let output = run_bin(
        "nazo",
        &["-param", "a78", "-hands", "rr", "-chains", "9"],
        None,
    );
    let stdout = output_strings(&output).0;

    assert!(stdout.contains("Hands:rr"));
    assert!(stdout.contains("cpus: "));
    assert!(!stdout.contains("14: ......"));
}

#[test]
fn nazo_invalid_dedup_simulate_and_chains_return_success() {
    let invalid_dedup = run_bin(
        "nazo",
        &["-param", "a78", "-hands", "rg", "-dedup", "x"],
        None,
    );
    let invalid_simulate = run_bin(
        "nazo",
        &["-param", "a78", "-hands", "rg", "-simulate", "x"],
        None,
    );
    let invalid_chains = run_bin(
        "nazo",
        &["-param", "a78", "-hands", "rg", "-chains", "x"],
        None,
    );

    assert!(invalid_dedup.status.success());
    assert!(
        output_strings(&invalid_dedup)
            .1
            .contains("unknown dedup mode")
    );

    assert!(invalid_simulate.status.success());
    assert!(
        output_strings(&invalid_simulate)
            .1
            .contains("unknown simulate policy")
    );

    assert!(invalid_chains.status.success());
    assert!(
        output_strings(&invalid_chains)
            .1
            .contains("invalid digit found in string")
    );
}

#[test]
fn nazo_invalid_clear_color_and_missing_hands_fail() {
    let invalid_clear = run_bin(
        "nazo",
        &["-param", "a78", "-hands", "rg", "-clear", "z"],
        None,
    );
    let invalid_color = run_bin(
        "nazo",
        &["-param", "a78", "-hands", "rg", "-color", "z"],
        None,
    );
    let missing_hands = run_bin("nazo", &["-param", "a78"], None);

    assert!(!invalid_clear.status.success());
    assert!(output_strings(&invalid_clear).1.contains("panic"));

    assert!(!invalid_color.status.success());
    assert!(output_strings(&invalid_color).1.contains("panic"));

    assert!(!missing_hands.status.success());
    assert!(output_strings(&missing_hands).1.contains("panic"));
}

#[cfg(unix)]
#[test]
fn nazo_sighup_prints_progress_line() {
    let mut child = spawn_bin_with_env(
        "nazo",
        &["-param", "a78", "-hands", "rrrrrrrrrrrr", "-dedup", "off"],
        &[],
    );
    let stdout = child.stdout.take().expect("child stdout must be piped");
    let stderr = child.stderr.take().expect("child stderr must be piped");
    let stdout_lines = spawn_line_reader(stdout);
    let stderr_lines = spawn_line_reader(stderr);
    let mut seen_stdout = Vec::new();
    let mut seen_stderr = Vec::new();

    let ready = wait_for_line(
        &stdout_lines,
        &mut seen_stdout,
        Duration::from_secs(5),
        |line| line.starts_with("cpus: "),
    );
    assert!(ready.is_some(), "stdout={}", seen_stdout.join("\n"),);

    let pid = child.id().to_string();
    let status = std::process::Command::new("kill")
        .args(["-HUP", &pid])
        .status()
        .expect("failed to send sighup");
    assert!(status.success());

    let progress = wait_for_line(
        &stderr_lines,
        &mut seen_stderr,
        Duration::from_secs(5),
        |line| line.starts_with('[') && line.contains('/') && line.contains('%'),
    );
    finish_child(&mut child);

    assert!(progress.is_some(), "stderr={}", seen_stderr.join("\n"));
}
