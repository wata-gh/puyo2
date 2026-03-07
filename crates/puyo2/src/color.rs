use serde_repr::{Deserialize_repr, Serialize_repr};
use thiserror::Error;

#[derive(
    Clone,
    Copy,
    Debug,
    Default,
    PartialEq,
    Eq,
    Hash,
    PartialOrd,
    Ord,
    Serialize_repr,
    Deserialize_repr,
)]
#[repr(u8)]
pub enum Color {
    #[default]
    Empty = 0,
    Ojama = 1,
    Wall = 2,
    Iron = 3,
    Red = 4,
    Blue = 5,
    Yellow = 6,
    Green = 7,
    Purple = 8,
}

#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum ColorParseError {
    #[error("letter must be one of r,g,y,b,p but {0}")]
    InvalidLetter(String),
    #[error("color {0:?} cannot be represented as a hand letter")]
    UnsupportedColor(Color),
}

impl Color {
    pub const COUNT: usize = 9;

    pub const fn idx(self) -> usize {
        self as usize
    }

    pub const fn from_bits(bits: u8) -> Self {
        match bits {
            0 => Self::Empty,
            1 => Self::Ojama,
            2 => Self::Wall,
            3 => Self::Iron,
            4 => Self::Red,
            5 => Self::Blue,
            6 => Self::Yellow,
            7 => Self::Green,
            8 => Self::Purple,
            _ => panic!("invalid color bits"),
        }
    }

    pub const fn is_special(self) -> bool {
        matches!(self, Self::Empty | Self::Ojama | Self::Wall | Self::Iron)
    }

    pub fn from_hand_char(ch: char) -> Result<Self, ColorParseError> {
        match ch {
            'r' => Ok(Self::Red),
            'g' => Ok(Self::Green),
            'y' => Ok(Self::Yellow),
            'b' => Ok(Self::Blue),
            'p' => Ok(Self::Purple),
            _ => Err(ColorParseError::InvalidLetter(ch.to_string())),
        }
    }

    pub fn to_hand_char(self) -> Result<char, ColorParseError> {
        match self {
            Self::Red => Ok('r'),
            Self::Green => Ok('g'),
            Self::Yellow => Ok('y'),
            Self::Blue => Ok('b'),
            Self::Purple => Ok('p'),
            _ => Err(ColorParseError::UnsupportedColor(self)),
        }
    }
}
