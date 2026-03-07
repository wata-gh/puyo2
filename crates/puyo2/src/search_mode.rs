use std::{fmt, str::FromStr};

use thiserror::Error;

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash)]
pub enum DedupMode {
    #[default]
    Off,
    SamePairOrder,
    State,
    StateMirror,
}

impl fmt::Display for DedupMode {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let text = match self {
            Self::Off => "off",
            Self::SamePairOrder => "same_pair_order",
            Self::State => "state",
            Self::StateMirror => "state_mirror",
        };
        f.write_str(text)
    }
}

#[derive(Clone, Debug, Error, PartialEq, Eq)]
#[error("unknown dedup mode {0:?}")]
pub struct ParseDedupModeError(pub String);

impl FromStr for DedupMode {
    type Err = ParseDedupModeError;

    fn from_str(value: &str) -> Result<Self, Self::Err> {
        match value.trim().to_ascii_lowercase().as_str() {
            "" | "off" => Ok(Self::Off),
            "same_pair_order" => Ok(Self::SamePairOrder),
            "state" => Ok(Self::State),
            "state_mirror" => Ok(Self::StateMirror),
            _ => Err(ParseDedupModeError(value.to_string())),
        }
    }
}

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash)]
pub enum SimulatePolicy {
    #[default]
    DetailAlways,
    FastIntermediate,
    FastAlways,
}

impl fmt::Display for SimulatePolicy {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let text = match self {
            Self::DetailAlways => "detail_always",
            Self::FastIntermediate => "fast_intermediate",
            Self::FastAlways => "fast_always",
        };
        f.write_str(text)
    }
}

#[derive(Clone, Debug, Error, PartialEq, Eq)]
#[error("unknown simulate policy {0:?}")]
pub struct ParseSimulatePolicyError(pub String);

impl FromStr for SimulatePolicy {
    type Err = ParseSimulatePolicyError;

    fn from_str(value: &str) -> Result<Self, Self::Err> {
        match value.trim().to_ascii_lowercase().as_str() {
            "" | "detail_always" => Ok(Self::DetailAlways),
            "fast_intermediate" => Ok(Self::FastIntermediate),
            "fast_always" => Ok(Self::FastAlways),
            _ => Err(ParseSimulatePolicyError(value.to_string())),
        }
    }
}
