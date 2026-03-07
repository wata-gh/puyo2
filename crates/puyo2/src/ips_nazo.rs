use percent_encoding::percent_decode_str;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use url::Url;

use crate::{BitField, Color, FieldBits};

const IPS_NAZO_ENCODE_CHARS: &str =
    "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-?";
const IPS_NAZO_DATA_MODE_CHARS: &str =
    "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.";

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct IPSNazoCondition {
    pub q0: usize,
    pub q1: usize,
    pub q2: usize,
    pub template: String,
    pub color: String,
    pub text: String,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct IPSNazoDecoded {
    #[serde(rename = "initialField")]
    pub initial_field: String,
    pub haipuyo: String,
    pub condition: IPSNazoCondition,
    #[serde(rename = "conditionCode")]
    pub condition_code: [usize; 3],
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct IPSNazoJudgeMetrics {
    pub remaining: [usize; 8],
    pub erased: [usize; 8],
    #[serde(rename = "chainCount")]
    pub chain_count: usize,
    #[serde(rename = "matchedInSimul")]
    pub matched_in_simul: bool,
    #[serde(rename = "matchedByConnect")]
    pub matched_by_connect: bool,
}

#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum IpsNazoError {
    #[error("input is empty")]
    EmptyInput,
    #[error("query is empty")]
    EmptyQuery,
    #[error("invalid url: {0}")]
    InvalidUrl(String),
    #[error("invalid query encoding: {0}")]
    InvalidQueryEncoding(String),
    #[error("{0}")]
    Invalid(String),
}

pub fn parse_ips_nazo_url(input: &str) -> Result<IPSNazoDecoded, IpsNazoError> {
    let query = normalize_ips_nazo_query(input)?;
    let segments = split_ips_nazo_segments(&query);
    let initial_field = decode_ips_nazo_field(&segments[0])?;
    let haipuyo = decode_ips_nazo_haipuyo(&segments[1])?;
    let condition = decode_ips_nazo_condition(&segments[3])?;
    Ok(IPSNazoDecoded {
        initial_field,
        haipuyo,
        condition_code: [condition.q0, condition.q1, condition.q2],
        condition,
    })
}

pub fn evaluate_ips_nazo_condition(
    before: &BitField,
    cond: &IPSNazoCondition,
) -> (bool, IPSNazoJudgeMetrics) {
    let mut metrics = IPSNazoJudgeMetrics::default();
    let mut bit_field = before.clone();

    loop {
        let mut vanished = FieldBits::new();
        let mut chain_erased = [0usize; 8];
        let mut chain_places = [0usize; 8];
        let mut any_erased = false;

        for color in bit_field.colors.clone() {
            let Some(color_index) = color_to_ips_nazo_index(color) else {
                continue;
            };
            let vb = bit_field.bits(color).mask_field12().find_vanishing_bits();
            if vb.is_empty() {
                continue;
            }
            any_erased = true;
            vanished = vanished.or(vb);
            vb.iterate_bit_with_masking(|seed| {
                let group = seed.expand(vb);
                let size = group.popcount();
                metrics.erased[0] += size;
                metrics.erased[7] += size;
                metrics.erased[color_index] += size;
                chain_erased[0] += size;
                chain_erased[7] += size;
                chain_erased[color_index] += size;
                chain_places[0] += 1;
                chain_places[7] += 1;
                chain_places[color_index] += 1;
                if cond.q0 == 52 || cond.q0 == 53 {
                    if match_connect_target(cond.q1, color_index)
                        && ((cond.q0 == 52 && size == cond.q2)
                            || (cond.q0 == 53 && size >= cond.q2))
                    {
                        metrics.matched_by_connect = true;
                    }
                }
                group
            });
        }

        if !any_erased {
            break;
        }
        metrics.chain_count += 1;

        let ojama = vanished.expand1(bit_field.bits(Color::Ojama));
        let ojama_count = ojama.popcount();
        if ojama_count > 0 {
            vanished = vanished.or(ojama);
            metrics.erased[0] += ojama_count;
            metrics.erased[6] += ojama_count;
            chain_erased[0] += ojama_count;
            chain_erased[6] += ojama_count;
        }

        match cond.q0 {
            40 if count_erased_colors(chain_erased) == cond.q2 => metrics.matched_in_simul = true,
            41 if count_erased_colors(chain_erased) >= cond.q2 => metrics.matched_in_simul = true,
            42 if valid_ips_nazo_index(cond.q1) && chain_erased[cond.q1] == cond.q2 => {
                metrics.matched_in_simul = true
            }
            43 if valid_ips_nazo_index(cond.q1) && chain_erased[cond.q1] >= cond.q2 => {
                metrics.matched_in_simul = true
            }
            44 if valid_ips_nazo_index(cond.q1) && chain_places[cond.q1] == cond.q2 => {
                metrics.matched_in_simul = true
            }
            45 if valid_ips_nazo_index(cond.q1) && chain_places[cond.q1] >= cond.q2 => {
                metrics.matched_in_simul = true
            }
            _ => {}
        }

        bit_field.drop_vanished(vanished);
    }

    metrics.remaining = count_remaining_pieces(&bit_field);
    let matched = match cond.q0 {
        2 => valid_ips_nazo_index(cond.q1) && metrics.remaining[cond.q1] == 0,
        10 => count_erased_colors(metrics.erased) == cond.q2,
        11 => count_erased_colors(metrics.erased) >= cond.q2,
        12 => valid_ips_nazo_index(cond.q1) && metrics.erased[cond.q1] == cond.q2,
        13 => valid_ips_nazo_index(cond.q1) && metrics.erased[cond.q1] >= cond.q2,
        30 => metrics.chain_count == cond.q2,
        31 => metrics.chain_count >= cond.q2,
        32 => {
            valid_ips_nazo_index(cond.q1)
                && metrics.chain_count == cond.q2
                && metrics.remaining[cond.q1] == 0
        }
        33 => {
            valid_ips_nazo_index(cond.q1)
                && metrics.chain_count >= cond.q2
                && metrics.remaining[cond.q1] == 0
        }
        40..=45 => metrics.matched_in_simul,
        52 | 53 => metrics.matched_by_connect,
        _ => false,
    };
    (matched, metrics)
}

fn normalize_ips_nazo_query(input: &str) -> Result<String, IpsNazoError> {
    let mut query = input.trim().to_string();
    if query.is_empty() {
        return Err(IpsNazoError::EmptyInput);
    }

    if query.contains("://") {
        let url = Url::parse(&query).map_err(|err| IpsNazoError::InvalidUrl(err.to_string()))?;
        query = url.query().ok_or(IpsNazoError::EmptyQuery)?.to_string();
    } else {
        let lower = query.to_ascii_lowercase();
        if let Some(idx) = lower.rfind("pn.html?") {
            query = query[idx + "pn.html?".len()..].to_string();
        }
    }

    if let Some(stripped) = query.strip_prefix('?') {
        query = stripped.to_string();
    }
    if let Some(idx) = query.find('#') {
        query.truncate(idx);
    }
    if query.contains('%') {
        query = percent_decode_str(&query)
            .decode_utf8()
            .map_err(|err| IpsNazoError::InvalidQueryEncoding(err.to_string()))?
            .into_owned();
    }
    if query.is_empty() {
        return Err(IpsNazoError::EmptyQuery);
    }

    Ok(query)
}

fn split_ips_nazo_segments(query: &str) -> [String; 4] {
    let mut segments = std::array::from_fn(|_| String::new());
    for (idx, segment) in query.split('_').take(4).enumerate() {
        segments[idx] = segment.to_string();
    }
    segments
}

fn decode_ips_nazo_field(segment: &str) -> Result<String, IpsNazoError> {
    let values = decode_ips_nazo_field_values(segment)?;
    let mut decoded = String::with_capacity(values.len());
    for (idx, value) in values.into_iter().enumerate() {
        match value {
            0 => decoded.push('a'),
            1 => decoded.push('b'),
            2 => decoded.push('e'),
            3 => decoded.push('c'),
            4 => decoded.push('d'),
            5 => decoded.push('f'),
            6 => decoded.push('g'),
            7..=9 => {
                return Err(IpsNazoError::Invalid(format!(
                    "field contains unsupported cell value {value} at index {idx}"
                )));
            }
            _ => {
                return Err(IpsNazoError::Invalid(format!(
                    "field contains unknown cell value {value} at index {idx}"
                )));
            }
        }
    }
    Ok(decoded)
}

fn decode_ips_nazo_field_values(segment: &str) -> Result<Vec<usize>, IpsNazoError> {
    if segment.starts_with('~') || segment.starts_with('‾') {
        return decode_ips_nazo_field_data_mode(segment);
    }

    validate_chars(segment, IPS_NAZO_ENCODE_CHARS)
        .map_err(|detail| IpsNazoError::Invalid(format!("invalid field segment: {detail}")))?;

    let mut padded = segment.to_string();
    while padded.len() < 39 {
        padded.insert(0, '0');
    }

    let bytes = padded.as_bytes();
    let mut values = vec![0usize; 13 * 6];
    for y in 0..=12 {
        for x in 1..=3 {
            let data_index = y * 3 + (x - 1);
            let ch = bytes.get(data_index).copied();
            let Some(ch) = ch else {
                continue;
            };
            let color_index = encode_char_index(char::from(ch)).unwrap_or_default();
            if color_index >= 64 {
                continue;
            }
            let base = y * 6 + (x - 1) * 2;
            values[base] = color_index / 8;
            values[base + 1] = color_index % 8;
        }
    }

    Ok(values)
}

fn decode_ips_nazo_field_data_mode(segment: &str) -> Result<Vec<usize>, IpsNazoError> {
    let data = segment
        .strip_prefix('~')
        .or_else(|| segment.strip_prefix('‾'))
        .unwrap_or(segment);

    validate_chars(data, IPS_NAZO_DATA_MODE_CHARS)
        .map_err(|detail| IpsNazoError::Invalid(format!("invalid field segment: {detail}")))?;

    let mut body = String::new();
    for item in data.split('.') {
        if item.is_empty() {
            body.push_str("000000");
            continue;
        }
        body.push_str(item);
        let rem = item.len() % 6;
        if rem > 0 {
            body.push_str(&"0".repeat(6 - rem));
        }
    }

    while body.len() < 78 {
        body.insert(0, '0');
    }

    let mut values = vec![0usize; 13 * 6];
    for y in 0..=12 {
        for x in 1..=6 {
            let data_index = y * 6 + (x - 1);
            values[data_index] = decode_ips_nazo_data_mode_cell(
                body.as_bytes()
                    .get(data_index)
                    .copied()
                    .map(char::from)
                    .unwrap_or('0'),
            );
        }
    }

    Ok(values)
}

fn decode_ips_nazo_data_mode_cell(ch: char) -> usize {
    match ch {
        '1' | 'r' | 'R' => 1,
        '2' | 'g' | 'G' => 2,
        '3' | 'b' | 'B' => 3,
        '4' | 'y' | 'Y' => 4,
        '5' | 'p' | 'P' => 5,
        '6' | 'o' | 'O' | 'j' | 'J' => 6,
        '7' | 'w' | 'W' => 7,
        '8' | 't' | 'T' | 'i' | 'I' => 8,
        '9' | 'k' | 'K' => 9,
        _ => 0,
    }
}

fn decode_ips_nazo_haipuyo(segment: &str) -> Result<String, IpsNazoError> {
    validate_chars(segment, IPS_NAZO_ENCODE_CHARS)
        .map_err(|detail| IpsNazoError::Invalid(format!("invalid operation segment: {detail}")))?;

    let routes = decode_ips_nazo_operation_routes(segment);
    if routes.first().is_none_or(String::is_empty) {
        return Ok(String::new());
    }

    let mut haipuyo = String::new();
    for op in routes[0].split(',') {
        let bytes = op.as_bytes();
        if bytes.len() < 2 {
            continue;
        }
        let axis = encode_char_index(char::from(bytes[0])).unwrap_or_default();
        let child = encode_char_index(char::from(bytes[1])).unwrap_or_default();
        let Some(axis_letter) = ips_nazo_color_index_to_letter(axis) else {
            continue;
        };
        let Some(child_letter) = ips_nazo_color_index_to_letter(child) else {
            continue;
        };
        haipuyo.push(axis_letter);
        haipuyo.push(child_letter);
    }

    Ok(haipuyo)
}

fn decode_ips_nazo_operation_routes(segment: &str) -> Vec<String> {
    let bytes = segment.as_bytes();
    let mut i = 1usize;
    let mut routes = vec![String::new()];
    let mut route_index: isize = -1;
    let mut route_class: isize = -1;
    let mut op = String::new();

    while (i * 2) <= bytes.len() {
        let mut data_index = (i - 1) * 2;
        let ci1 = encode_char_index(char::from(bytes[data_index])).unwrap_or_default() as isize;
        data_index += 1;
        let ci2 = encode_char_index(char::from(bytes[data_index])).unwrap_or_default() as isize;

        let mut dt = ci1 % 2;
        if dt != route_class {
            if route_index >= 0 {
                routes[route_index as usize].push_str(&op);
            }
            route_class = dt;
            route_index += 1;
            if route_index as usize >= routes.len() {
                routes.push(String::new());
            }
            op.clear();
        } else if !op.is_empty() {
            routes[route_index as usize].push_str(&op);
            routes[route_index as usize].push(',');
            op.clear();
        }

        dt = ci1 / 2;
        if dt == 31 {
            // noop
        } else if dt == 30 {
            push_encode_char(&mut op, 8);
            push_encode_char(&mut op, ci2 as usize);
            push_encode_char(&mut op, 64);
            push_encode_char(&mut op, 64);
        } else {
            let mut idx = dt % 6 + 1;
            if idx == 6 {
                push_encode_char(&mut op, idx as usize);
                push_encode_char(&mut op, (dt / 6) as usize);
                push_encode_char(&mut op, ci2 as usize);
                push_encode_char(&mut op, 64);
            } else {
                push_encode_char(&mut op, idx as usize);
                push_encode_char(&mut op, (dt / 6 + 1) as usize);
                idx = ci2 % 2;
                dt = ci2 / 2;
                idx = if idx == 0 { dt % 6 + 1 } else { 7 };
                push_encode_char(&mut op, idx as usize);
                push_encode_char(&mut op, (dt / 6) as usize);
            }
        }

        i += 1;
    }

    if route_index >= 0 {
        routes[route_index as usize].push_str(&op);
    }

    routes
}

fn decode_ips_nazo_condition(segment: &str) -> Result<IPSNazoCondition, IpsNazoError> {
    validate_chars(segment, IPS_NAZO_ENCODE_CHARS)
        .map_err(|detail| IpsNazoError::Invalid(format!("invalid condition segment: {detail}")))?;

    let mut codes = [0usize; 3];
    for (idx, ch) in segment.chars().take(3).enumerate() {
        codes[idx] = encode_char_index(ch).unwrap_or_default();
    }

    let template = condition_template(codes[0]).to_string();
    let color = condition_color(codes[1]).to_string();
    let mut text = template.clone();
    if !text.is_empty() {
        text = text.replacen('c', &color, 1);
        text = text.replacen('n', &codes[2].to_string(), 1);
    }

    Ok(IPSNazoCondition {
        q0: codes[0],
        q1: codes[1],
        q2: codes[2],
        template,
        color,
        text,
    })
}

fn validate_chars(value: &str, allowed: &str) -> Result<(), String> {
    for (idx, ch) in value.char_indices() {
        if !allowed.contains(ch) {
            return Err(format!(
                "contains unsupported character {ch:?} at index {idx}"
            ));
        }
    }
    Ok(())
}

fn encode_char_index(ch: char) -> Option<usize> {
    IPS_NAZO_ENCODE_CHARS.find(ch)
}

fn push_encode_char(output: &mut String, idx: usize) {
    if let Some(ch) = IPS_NAZO_ENCODE_CHARS.as_bytes().get(idx) {
        output.push(char::from(*ch));
    }
}

fn ips_nazo_color_index_to_letter(idx: usize) -> Option<char> {
    match idx {
        1 => Some('r'),
        2 => Some('g'),
        3 => Some('b'),
        4 => Some('y'),
        5 => Some('p'),
        _ => None,
    }
}

fn count_remaining_pieces(bit_field: &BitField) -> [usize; 8] {
    let mut remaining = [0usize; 8];
    for y in 1..=13 {
        for x in 0..6 {
            match bit_field.color(x, y) {
                Color::Red => {
                    remaining[1] += 1;
                    remaining[7] += 1;
                }
                Color::Green => {
                    remaining[2] += 1;
                    remaining[7] += 1;
                }
                Color::Blue => {
                    remaining[3] += 1;
                    remaining[7] += 1;
                }
                Color::Yellow => {
                    remaining[4] += 1;
                    remaining[7] += 1;
                }
                Color::Purple => {
                    remaining[5] += 1;
                    remaining[7] += 1;
                }
                Color::Ojama => remaining[6] += 1,
                _ => {}
            }
        }
    }
    remaining[0] = remaining[7] + remaining[6];
    remaining
}

fn count_erased_colors(counts: [usize; 8]) -> usize {
    counts[1..=5].iter().filter(|count| **count > 0).count()
}

fn color_to_ips_nazo_index(color: Color) -> Option<usize> {
    match color {
        Color::Red => Some(1),
        Color::Green => Some(2),
        Color::Blue => Some(3),
        Color::Yellow => Some(4),
        Color::Purple => Some(5),
        _ => None,
    }
}

fn match_connect_target(target: usize, color: usize) -> bool {
    target == 0 || target == 7 || target == color
}

fn valid_ips_nazo_index(idx: usize) -> bool {
    idx <= 7
}

fn condition_template(code: usize) -> &'static str {
    match code {
        0 => "条件を選択してください",
        2 => "cぷよ全て消す",
        10 => "n色消す",
        11 => "n色以上消す",
        12 => "cぷよn個消す",
        13 => "cぷよn個以上消す",
        30 => "n連鎖する",
        31 => "n連鎖以上する",
        32 => "n連鎖＆cぷよ全て消す",
        33 => "n連鎖以上＆cぷよ全て消す",
        40 => "n色同時に消す",
        41 => "n色以上同時に消す",
        42 => "cぷよn個同時に消す",
        43 => "cぷよn個以上同時に消す",
        44 => "cぷよn箇所同時に消す",
        45 => "cぷよn箇所以上同時に消す",
        52 => "cぷよn連結で消す",
        53 => "cぷよn連結以上で消す",
        _ => "",
    }
}

fn condition_color(code: usize) -> &'static str {
    match code {
        1 => "赤",
        2 => "緑",
        3 => "青",
        4 => "黄",
        5 => "紫",
        6 => "おじゃま",
        7 => "色",
        _ => "",
    }
}
