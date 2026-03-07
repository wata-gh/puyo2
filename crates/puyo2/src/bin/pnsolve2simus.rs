#[path = "../cli_common.rs"]
mod cli_common;

use std::{fs, io::Read, process};

use serde_json::Value;
use url::form_urlencoded::Serializer;

use crate::cli_common::{collect_args, matches_switch, parse_flag_value};

const DEFAULT_BASE_URL: &str = "https://puyo-rsrch.com";
const LOCAL_BASE_URL: &str = "http://localhost:3000";
const USAGE: &str = "Usage: pnsolve2simus [--local|-l] [JSON_FILE]";

fn fail_with(message: &str) -> ! {
    eprintln!("{message}");
    process::exit(1);
}

fn main() {
    let mut base_url = DEFAULT_BASE_URL.to_string();
    let mut paths = Vec::new();

    let mut args = collect_args().into_iter();
    while let Some(arg) = args.next() {
        if matches_switch(&arg, "-l", "--local") {
            base_url = LOCAL_BASE_URL.to_string();
            continue;
        }
        if parse_flag_value(&arg, &mut args, "-l", "--local")
            .unwrap_or_else(|err| fail_with(&format!("option error: {err}")))
            .is_some()
        {
            fail_with("option error: --local does not take a value");
        }
        if arg.starts_with('-') {
            fail_with(&format!("option error: unknown option: {arg}"));
        }
        paths.push(arg);
    }

    if paths.len() > 1 {
        fail_with(&format!("too many arguments\n{USAGE}"));
    }

    let input = if let Some(path) = paths.first() {
        match fs::read_to_string(path) {
            Ok(content) => content,
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => {
                fail_with(&format!("file not found: {path}"))
            }
            Err(err) if err.kind() == std::io::ErrorKind::PermissionDenied => {
                fail_with(&format!("permission denied: {path}"))
            }
            Err(err) => fail_with(&err.to_string()),
        }
    } else {
        let mut input = String::new();
        std::io::stdin()
            .read_to_string(&mut input)
            .unwrap_or_else(|err| fail_with(&err.to_string()));
        input
    };

    let payload: Value = serde_json::from_str(&input)
        .unwrap_or_else(|err| fail_with(&format!("invalid JSON: {err}")));
    let payload = payload
        .as_object()
        .unwrap_or_else(|| fail_with("top-level JSON must be an object"));

    let root_initial_field = payload
        .get("initialField")
        .and_then(Value::as_str)
        .filter(|value| !value.is_empty())
        .unwrap_or_else(|| fail_with("initialField is missing or empty"));

    let solutions = payload
        .get("solutions")
        .and_then(Value::as_array)
        .unwrap_or_else(|| fail_with("solutions must be an array"));

    for (index, solution) in solutions.iter().enumerate() {
        let solution = solution
            .as_object()
            .unwrap_or_else(|| fail_with(&format!("solutions[{index}] must be an object")));
        let hands = solution
            .get("hands")
            .and_then(Value::as_str)
            .filter(|value| !value.is_empty())
            .unwrap_or_else(|| fail_with(&format!("solutions[{index}].hands is missing or empty")));

        let query = Serializer::new(String::new())
            .append_pair("fs", root_initial_field)
            .append_pair("h", hands)
            .finish();
        println!("{base_url}/simus?{query}");
    }
}
