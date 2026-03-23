use thiserror::Error;

use crate::{Color, ColorParseError, Hand, PuyoSet};

#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum HandParseError {
    #[error("haipuyo length must be even, got {0}")]
    OddHaipuyoLength(usize),
    #[error("hand string length must be a multiple of 4, got {0}")]
    InvalidHandStringLength(usize),
    #[error("haipuyo normalization only supports up to four colors, got {0}")]
    TooManyColorsToNormalize(usize),
    #[error("invalid digit {value:?} at index {index}")]
    InvalidDigit { value: char, index: usize },
    #[error(transparent)]
    InvalidColor(#[from] ColorParseError),
}

pub fn expand_mattulwan_param(field: &str) -> String {
    if field.chars().count() == 78 && !field.chars().any(|ch| ch.is_ascii_digit()) {
        return field.to_string();
    }

    let mut expanded = String::new();
    let mut chars = field.chars().peekable();
    while let Some(ch) = chars.next() {
        if ch.is_ascii_digit() {
            expanded.push(ch);
            continue;
        }

        let mut digits = String::new();
        while let Some(next) = chars.peek() {
            if !next.is_ascii_digit() {
                break;
            }
            digits.push(*next);
            chars.next();
        }

        if digits.is_empty() {
            expanded.push(ch);
            continue;
        }

        let count = digits.parse::<usize>().unwrap_or_default();
        expanded.extend(std::iter::repeat_n(ch, count));
    }

    expanded
}

pub fn haipuyo_to_puyo_sets(haipuyo: &str) -> Result<Vec<PuyoSet>, HandParseError> {
    let chars: Vec<char> = haipuyo.chars().collect();
    if !chars.len().is_multiple_of(2) {
        return Err(HandParseError::OddHaipuyoLength(chars.len()));
    }

    let mut puyo_sets = Vec::with_capacity(chars.len() / 2);
    for chunk in chars.chunks_exact(2) {
        puyo_sets.push(PuyoSet {
            axis: Color::from_hand_char(chunk[0])?,
            child: Color::from_hand_char(chunk[1])?,
        });
    }
    Ok(puyo_sets)
}

/// Normalizes a haipuyo string into a canonical digit string.
///
/// Colors are canonicalized up to global renaming, and each pair is treated as
/// unordered so that `12` and `21` normalize identically. `p` is treated as a
/// placeholder for the first unused basic color in `[r, b, y, g]`. The return
/// value is a two-digits-per-pair string and is not accepted by
/// [`haipuyo_to_puyo_sets`].
pub fn normalize_haipuyo(haipuyo: &str) -> Result<String, HandParseError> {
    let mut puyo_sets = haipuyo_to_puyo_sets(haipuyo)?;
    resolve_purple_placeholder(&mut puyo_sets)?;

    let used_colors = collect_used_colors(&puyo_sets);
    if used_colors.len() > 4 {
        return Err(HandParseError::TooManyColorsToNormalize(used_colors.len()));
    }
    if used_colors.is_empty() {
        return Ok(String::new());
    }

    let mut best: Option<String> = None;
    for_each_color_permutation(&used_colors, &mut |permutation| {
        let candidate = normalize_for_permutation(&puyo_sets, permutation);
        if best.as_ref().is_none_or(|current| candidate < *current) {
            best = Some(candidate);
        }
    });

    Ok(best.unwrap_or_default())
}

fn resolve_purple_placeholder(puyo_sets: &mut [PuyoSet]) -> Result<(), HandParseError> {
    if !puyo_sets
        .iter()
        .any(|puyo_set| puyo_set.axis == Color::Purple || puyo_set.child == Color::Purple)
    {
        return Ok(());
    }

    let mut used_without_purple = Vec::with_capacity(4);
    for puyo_set in puyo_sets.iter() {
        for color in [puyo_set.axis, puyo_set.child] {
            if color == Color::Purple || used_without_purple.contains(&color) {
                continue;
            }
            used_without_purple.push(color);
        }
    }

    let replacement = [Color::Red, Color::Blue, Color::Yellow, Color::Green]
        .into_iter()
        .find(|color| !used_without_purple.contains(color))
        .ok_or_else(|| {
            HandParseError::TooManyColorsToNormalize(count_distinct_colors(puyo_sets))
        })?;

    for puyo_set in puyo_sets {
        if puyo_set.axis == Color::Purple {
            puyo_set.axis = replacement;
        }
        if puyo_set.child == Color::Purple {
            puyo_set.child = replacement;
        }
    }

    Ok(())
}

fn count_distinct_colors(puyo_sets: &[PuyoSet]) -> usize {
    collect_used_colors(puyo_sets).len()
}

fn collect_used_colors(puyo_sets: &[PuyoSet]) -> Vec<Color> {
    let mut used = Vec::with_capacity(4);
    for puyo_set in puyo_sets {
        for color in [puyo_set.axis, puyo_set.child] {
            if !used.contains(&color) {
                used.push(color);
            }
        }
    }
    used
}

fn for_each_color_permutation<F>(colors: &[Color], f: &mut F)
where
    F: FnMut(&[Color]),
{
    let mut permutation = colors.to_vec();
    permute_recursive(0, &mut permutation, f);
}

fn permute_recursive<F>(start: usize, permutation: &mut [Color], f: &mut F)
where
    F: FnMut(&[Color]),
{
    if start == permutation.len() {
        f(permutation);
        return;
    }

    for index in start..permutation.len() {
        permutation.swap(start, index);
        permute_recursive(start + 1, permutation, f);
        permutation.swap(start, index);
    }
}

fn normalize_for_permutation(puyo_sets: &[PuyoSet], permutation: &[Color]) -> String {
    let mut mapped_order = [u8::MAX; Color::COUNT];
    for (index, color) in permutation.iter().copied().enumerate() {
        mapped_order[color.idx()] = index as u8;
    }

    let mut relabel = [0u8; 4];
    let mut next_label = 1u8;
    let mut normalized = String::with_capacity(puyo_sets.len() * 2);

    for puyo_set in puyo_sets {
        let mut axis = mapped_order[puyo_set.axis.idx()];
        let mut child = mapped_order[puyo_set.child.idx()];
        debug_assert_ne!(axis, u8::MAX);
        debug_assert_ne!(child, u8::MAX);
        if axis > child {
            std::mem::swap(&mut axis, &mut child);
        }

        for value in [axis, child] {
            let slot = &mut relabel[value as usize];
            if *slot == 0 {
                *slot = next_label;
                next_label += 1;
            }
            normalized.push(char::from(b'0' + *slot));
        }
    }

    normalized
}

pub fn to_simple_hands(hands: &[Hand]) -> Result<String, HandParseError> {
    let mut encoded = String::with_capacity(hands.len() * 4);
    for hand in hands {
        encoded.push(hand.puyo_set.axis.to_hand_char()?);
        encoded.push(hand.puyo_set.child.to_hand_char()?);
        encoded.push(
            char::from_digit(
                u32::try_from(hand.position[0]).map_err(|_| HandParseError::InvalidDigit {
                    value: '?',
                    index: 0,
                })?,
                10,
            )
            .ok_or(HandParseError::InvalidDigit {
                value: '?',
                index: 0,
            })?,
        );
        encoded.push(
            char::from_digit(
                u32::try_from(hand.position[1]).map_err(|_| HandParseError::InvalidDigit {
                    value: '?',
                    index: 1,
                })?,
                10,
            )
            .ok_or(HandParseError::InvalidDigit {
                value: '?',
                index: 1,
            })?,
        );
    }
    Ok(encoded)
}

pub fn parse_simple_hands(hands: &str) -> Result<Vec<Hand>, HandParseError> {
    let chars: Vec<char> = hands.chars().collect();
    if !chars.len().is_multiple_of(4) {
        return Err(HandParseError::InvalidHandStringLength(chars.len()));
    }

    let mut parsed = Vec::with_capacity(chars.len() / 4);
    for (chunk_index, chunk) in chars.chunks_exact(4).enumerate() {
        let base = chunk_index * 4;
        let axis = Color::from_hand_char(chunk[0])?;
        let child = Color::from_hand_char(chunk[1])?;
        let row = chunk[2].to_digit(10).ok_or(HandParseError::InvalidDigit {
            value: chunk[2],
            index: base + 2,
        })? as usize;
        let dir = chunk[3].to_digit(10).ok_or(HandParseError::InvalidDigit {
            value: chunk[3],
            index: base + 3,
        })? as usize;
        parsed.push(Hand {
            puyo_set: PuyoSet { axis, child },
            position: [row, dir],
        });
    }

    Ok(parsed)
}
