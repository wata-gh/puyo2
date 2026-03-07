use thiserror::Error;

use crate::{Color, ColorParseError, Hand, PuyoSet};

#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum HandParseError {
    #[error("haipuyo length must be even, got {0}")]
    OddHaipuyoLength(usize),
    #[error("hand string length must be a multiple of 4, got {0}")]
    InvalidHandStringLength(usize),
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
