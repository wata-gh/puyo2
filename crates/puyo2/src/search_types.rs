use std::collections::{HashMap, HashSet};

use serde::{Deserialize, Serialize};

use crate::{BitField, DedupMode, Hand, PuyoSet, RensaResult, SearchStateKey, SimulatePolicy};

pub type LastCallback = Box<dyn FnMut(&SearchResult)>;
pub type EachHandCallback = Box<dyn FnMut(&SearchResult) -> bool>;

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct SearchResult {
    #[serde(rename = "rensaResult")]
    pub rensa_result: Option<RensaResult>,
    #[serde(rename = "beforeSimulate")]
    pub before_simulate: Option<BitField>,
    pub depth: usize,
    pub position: [usize; 2],
    #[serde(rename = "positionNum")]
    pub position_num: usize,
    pub hands: Vec<Hand>,
}

pub struct SearchCondition {
    pub disable_chigiri: bool,
    pub chigiriable_count: usize,
    pub chigiris: usize,
    pub set_frames: usize,
    pub dedup_mode: DedupMode,
    pub simulate_policy: SimulatePolicy,
    pub stop_on_chain: bool,
    pub puyo_sets: Vec<PuyoSet>,
    pub bit_field: Option<BitField>,
    pub last_callback: Option<LastCallback>,
    pub each_hand_callback: Option<EachHandCallback>,
    pub(crate) visited_states: Option<HashMap<usize, HashSet<SearchStateKey>>>,
}

impl Default for SearchCondition {
    fn default() -> Self {
        Self {
            disable_chigiri: false,
            chigiriable_count: 0,
            chigiris: 0,
            set_frames: 0,
            dedup_mode: DedupMode::Off,
            simulate_policy: SimulatePolicy::DetailAlways,
            stop_on_chain: false,
            puyo_sets: Vec::new(),
            bit_field: None,
            last_callback: None,
            each_hand_callback: None,
            visited_states: None,
        }
    }
}

impl SearchCondition {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_bit_field_and_puyo_sets(bit_field: BitField, puyo_sets: Vec<PuyoSet>) -> Self {
        Self {
            bit_field: Some(bit_field),
            puyo_sets,
            ..Self::default()
        }
    }
}
