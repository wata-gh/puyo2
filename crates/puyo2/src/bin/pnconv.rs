#[path = "../cli_common.rs"]
mod cli_common;

use std::process;

use puyo2::parse_ips_nazo_url;

use crate::cli_common::{collect_args, parse_flag_value, read_non_empty_inputs_from_stdin};

fn print_decoded(decoded: &puyo2::IPSNazoDecoded) {
    println!("Initial Field: {}", decoded.initial_field);
    println!("Haipuyo: {}", decoded.haipuyo);
    println!("Condition: {}", decoded.condition.text);
    println!(
        "ConditionCode: q0={} q1={} q2={}",
        decoded.condition_code[0], decoded.condition_code[1], decoded.condition_code[2]
    );
}

fn main() {
    let mut url_input = String::new();
    let mut param_input = String::new();

    let mut args = collect_args().into_iter();
    while let Some(arg) = args.next() {
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-url", "--url").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            url_input = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-param", "--param").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            param_input = value;
            continue;
        }
        eprintln!("unknown argument: {arg}");
        process::exit(1);
    }

    let inputs = if !url_input.trim().is_empty() {
        vec![url_input]
    } else if !param_input.trim().is_empty() {
        vec![param_input]
    } else {
        match read_non_empty_inputs_from_stdin() {
            Ok(values) => values,
            Err(err) => {
                eprintln!("stdin read error: {err}");
                process::exit(1);
            }
        }
    };

    if inputs.is_empty() {
        eprintln!("no input. use -url, -param, or stdin.");
        process::exit(1);
    }

    let mut failed = false;
    let mut printed = 0usize;
    for input in inputs {
        match parse_ips_nazo_url(&input) {
            Ok(decoded) => {
                if printed > 0 {
                    println!();
                }
                print_decoded(&decoded);
                printed += 1;
            }
            Err(err) => {
                eprintln!("parse error: {input}: {err}");
                failed = true;
            }
        }
    }

    if failed {
        process::exit(1);
    }
}
