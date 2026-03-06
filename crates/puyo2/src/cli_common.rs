#![allow(dead_code)]

use std::{
    any::Any,
    env,
    io::{self, Read},
    panic::{self, AssertUnwindSafe},
};

pub fn parse_flag_value<I>(
    current: &str,
    args: &mut I,
    short: &str,
    long: &str,
) -> Result<Option<String>, String>
where
    I: Iterator<Item = String>,
{
    if current == short || current == long {
        let value = args
            .next()
            .ok_or_else(|| format!("missing value for {short}"))?;
        return Ok(Some(value));
    }

    if let Some(value) = current.strip_prefix(&format!("{short}=")) {
        return Ok(Some(value.to_string()));
    }
    if let Some(value) = current.strip_prefix(&format!("{long}=")) {
        return Ok(Some(value.to_string()));
    }
    Ok(None)
}

pub fn matches_switch(current: &str, short: &str, long: &str) -> bool {
    current == short || current == long
}

pub fn parse_bool_value(value: &str, flag: &str) -> Result<bool, String> {
    match value.trim().to_ascii_lowercase().as_str() {
        "true" | "t" | "1" => Ok(true),
        "false" | "f" | "0" => Ok(false),
        _ => Err(format!("invalid boolean for {flag}: {value}")),
    }
}

pub fn read_single_non_empty_input_from_stdin() -> Result<String, String> {
    let inputs = read_non_empty_inputs_from_stdin()?;
    match inputs.len() {
        0 => Err("no input".to_string()),
        1 => Ok(inputs.into_iter().next().unwrap()),
        n => Err(format!(
            "multiple inputs are not supported: got {n} non-empty lines"
        )),
    }
}

pub fn read_non_empty_inputs_from_stdin() -> Result<Vec<String>, String> {
    let mut input = String::new();
    io::stdin()
        .read_to_string(&mut input)
        .map_err(|err| err.to_string())?;
    Ok(input
        .lines()
        .map(str::trim)
        .filter(|line| !line.is_empty())
        .map(ToString::to_string)
        .collect())
}

pub fn collect_args() -> Vec<String> {
    env::args().skip(1).collect()
}

pub fn catch_unwind_silent<F, T>(f: F) -> Result<T, String>
where
    F: FnOnce() -> T,
{
    let hook = panic::take_hook();
    panic::set_hook(Box::new(|_| {}));
    let result = panic::catch_unwind(AssertUnwindSafe(f));
    panic::set_hook(hook);
    result.map_err(panic_payload_to_string)
}

fn panic_payload_to_string(payload: Box<dyn Any + Send>) -> String {
    if let Some(message) = payload.downcast_ref::<String>() {
        return message.clone();
    }
    if let Some(message) = payload.downcast_ref::<&str>() {
        return (*message).to_string();
    }
    "panic during search".to_string()
}
