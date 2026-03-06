use serde::{Deserialize, Serialize};

use crate::{BitField, Color, ShapeBitField};

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct PuyoSet {
    pub axis: Color,
    pub child: Color,
}

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct Hand {
    #[serde(rename = "puyoSet")]
    pub puyo_set: PuyoSet,
    pub position: [usize; 2],
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct PuyoSetPlacement {
    #[serde(skip_serializing, skip_deserializing, default)]
    pub puyo_set: Option<PuyoSet>,
    #[serde(rename = "pos")]
    pub pos: [usize; 2],
    #[serde(rename = "axisX")]
    pub axis_x: isize,
    #[serde(rename = "axisY")]
    pub axis_y: isize,
    #[serde(rename = "childX")]
    pub child_x: isize,
    #[serde(rename = "childY")]
    pub child_y: isize,
    pub chigiri: bool,
    pub frames: usize,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct SearchStateKey {
    pub m: [[u64; 2]; 3],
    #[serde(rename = "tableSig")]
    pub table_sig: u32,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct NthResult {
    pub nth: usize,
    #[serde(rename = "erasedPuyos")]
    pub erased_puyos: Vec<SingleResult>,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct SingleResult {
    pub color: Color,
    pub connected: usize,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct RensaResult {
    pub chains: usize,
    pub chigiris: usize,
    pub score: usize,
    #[serde(rename = "rensaFrames")]
    pub rensa_frames: usize,
    #[serde(rename = "setFrames")]
    pub set_frames: usize,
    pub erased: usize,
    pub quick: bool,
    #[serde(rename = "bitFeild")]
    pub bit_field: Option<BitField>,
    #[serde(rename = "nthResults")]
    pub nth_results: Vec<NthResult>,
}

impl RensaResult {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_result(
        chains: usize,
        score: usize,
        frames: usize,
        quick: bool,
        bit_field: BitField,
    ) -> Self {
        Self {
            chains,
            score,
            rensa_frames: frames,
            quick,
            bit_field: Some(bit_field),
            ..Self::default()
        }
    }

    pub fn add_chain(&mut self) {
        self.chains += 1;
    }

    pub fn add_erased(&mut self, erased: usize) {
        self.erased += erased;
    }

    pub fn add_score(&mut self, score: usize) {
        self.score += score;
    }

    pub fn set_bit_field(&mut self, bit_field: BitField) {
        self.bit_field = Some(bit_field);
    }

    pub fn set_quick(&mut self, quick: bool) {
        self.quick = quick;
    }

    pub fn nth_result(&self, nth: usize) -> Option<&NthResult> {
        self.nth_results.iter().find(|item| item.nth == nth)
    }
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct ShapeRensaResult {
    pub chains: usize,
    pub score: usize,
    pub frames: usize,
    pub erased: usize,
    pub quick: bool,
    #[serde(rename = "shapeBitField")]
    pub shape_bit_field: Option<ShapeBitField>,
}

impl ShapeRensaResult {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_result(
        chains: usize,
        score: usize,
        frames: usize,
        quick: bool,
        shape_bit_field: ShapeBitField,
    ) -> Self {
        Self {
            chains,
            score,
            frames,
            quick,
            shape_bit_field: Some(shape_bit_field),
            ..Self::default()
        }
    }

    pub fn add_chain(&mut self) {
        self.chains += 1;
    }

    pub fn add_erased(&mut self, erased: usize) {
        self.erased += erased;
    }

    pub fn add_score(&mut self, score: usize) {
        self.score += score;
    }

    pub fn set_shape_bit_field(&mut self, shape_bit_field: ShapeBitField) {
        self.shape_bit_field = Some(shape_bit_field);
    }

    pub fn set_quick(&mut self, quick: bool) {
        self.quick = quick;
    }
}

pub const SETUP_POSITIONS: [[usize; 2]; 22] = [
    [0, 0],
    [0, 1],
    [0, 2],
    [1, 0],
    [1, 1],
    [1, 2],
    [1, 3],
    [2, 0],
    [2, 1],
    [2, 2],
    [2, 3],
    [3, 0],
    [3, 1],
    [3, 2],
    [3, 3],
    [4, 0],
    [4, 1],
    [4, 2],
    [4, 3],
    [5, 0],
    [5, 2],
    [5, 3],
];

pub const ALL_PUYO_SETS: [PuyoSet; 16] = [
    PuyoSet {
        axis: Color::Red,
        child: Color::Red,
    },
    PuyoSet {
        axis: Color::Red,
        child: Color::Blue,
    },
    PuyoSet {
        axis: Color::Red,
        child: Color::Yellow,
    },
    PuyoSet {
        axis: Color::Red,
        child: Color::Green,
    },
    PuyoSet {
        axis: Color::Blue,
        child: Color::Red,
    },
    PuyoSet {
        axis: Color::Blue,
        child: Color::Blue,
    },
    PuyoSet {
        axis: Color::Blue,
        child: Color::Yellow,
    },
    PuyoSet {
        axis: Color::Blue,
        child: Color::Green,
    },
    PuyoSet {
        axis: Color::Yellow,
        child: Color::Red,
    },
    PuyoSet {
        axis: Color::Yellow,
        child: Color::Blue,
    },
    PuyoSet {
        axis: Color::Yellow,
        child: Color::Yellow,
    },
    PuyoSet {
        axis: Color::Yellow,
        child: Color::Green,
    },
    PuyoSet {
        axis: Color::Green,
        child: Color::Red,
    },
    PuyoSet {
        axis: Color::Green,
        child: Color::Blue,
    },
    PuyoSet {
        axis: Color::Green,
        child: Color::Yellow,
    },
    PuyoSet {
        axis: Color::Green,
        child: Color::Green,
    },
];

pub const UNIQUE_PUYO_SETS: [PuyoSet; 10] = [
    PuyoSet {
        axis: Color::Red,
        child: Color::Red,
    },
    PuyoSet {
        axis: Color::Red,
        child: Color::Blue,
    },
    PuyoSet {
        axis: Color::Red,
        child: Color::Yellow,
    },
    PuyoSet {
        axis: Color::Red,
        child: Color::Green,
    },
    PuyoSet {
        axis: Color::Blue,
        child: Color::Blue,
    },
    PuyoSet {
        axis: Color::Blue,
        child: Color::Yellow,
    },
    PuyoSet {
        axis: Color::Blue,
        child: Color::Green,
    },
    PuyoSet {
        axis: Color::Yellow,
        child: Color::Yellow,
    },
    PuyoSet {
        axis: Color::Yellow,
        child: Color::Green,
    },
    PuyoSet {
        axis: Color::Green,
        child: Color::Green,
    },
];

pub const SET_FRAMES_TABLE: [[usize; 6]; 15] = [
    [0, 0, 0, 0, 0, 0],
    [54, 52, 50, 52, 54, 56],
    [52, 50, 48, 50, 52, 54],
    [50, 48, 46, 48, 50, 52],
    [48, 46, 44, 46, 48, 50],
    [46, 44, 42, 44, 46, 48],
    [44, 42, 40, 42, 44, 46],
    [42, 40, 38, 40, 42, 44],
    [40, 38, 36, 38, 40, 42],
    [38, 36, 34, 36, 38, 40],
    [36, 34, 32, 34, 36, 38],
    [34, 32, 30, 32, 34, 36],
    [32, 30, 28, 30, 32, 34],
    [30, 28, 26, 28, 30, 32],
    [30, 28, 26, 28, 30, 32],
];

pub const CHIGIRI_FRAMES_TABLE: [usize; 14] =
    [0, 19, 24, 28, 31, 34, 37, 40, 42, 44, 46, 48, 50, 52];
