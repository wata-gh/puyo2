use std::collections::{HashMap, HashSet};

use crate::{
    BitField, Color, DedupMode, EachHandCallback, Hand, LastCallback, PuyoSet, RensaResult,
    SETUP_POSITIONS, SearchCondition, SearchResult, SearchStateKey, SimulatePolicy,
};

const SAME_COLOR_POSITIONS: [[usize; 2]; 11] = [
    [0, 0],
    [0, 1],
    [1, 0],
    [1, 1],
    [2, 0],
    [2, 1],
    [3, 0],
    [3, 1],
    [4, 0],
    [4, 1],
    [5, 0],
];

impl SearchCondition {
    pub fn search_with_puyo_sets_v2(&mut self) {
        let Some(bit_field) = self.bit_field else {
            panic!("search_with_puyo_sets_v2 requires bit_field");
        };
        if self.puyo_sets.is_empty() {
            self.visited_states = None;
            return;
        }

        self.visited_states = if uses_state_dedup(self.dedup_mode) {
            Some(HashMap::new())
        } else {
            None
        };

        let puyo_sets = self.puyo_sets.clone();
        let mut runtime = SearchRuntime {
            disable_chigiri: self.disable_chigiri,
            chigiriable_count: self.chigiriable_count,
            chigiris: self.chigiris,
            set_frames: self.set_frames,
            dedup_mode: self.dedup_mode,
            simulate_policy: self.simulate_policy,
            stop_on_chain: self.stop_on_chain,
            last_callback: &mut self.last_callback,
            each_hand_callback: &mut self.each_hand_callback,
            visited_states: &mut self.visited_states,
        };
        search_with_puyo_sets_v2(&mut runtime, &puyo_sets, &bit_field, Vec::new(), 0);
    }
}

struct SearchRuntime<'a> {
    disable_chigiri: bool,
    chigiriable_count: usize,
    chigiris: usize,
    set_frames: usize,
    dedup_mode: DedupMode,
    simulate_policy: SimulatePolicy,
    stop_on_chain: bool,
    last_callback: &'a mut Option<LastCallback>,
    each_hand_callback: &'a mut Option<EachHandCallback>,
    visited_states: &'a mut Option<HashMap<usize, HashSet<SearchStateKey>>>,
}

fn search_with_puyo_sets_v2(
    runtime: &mut SearchRuntime<'_>,
    puyo_sets: &[PuyoSet],
    bit_field: &BitField,
    hands: Vec<Hand>,
    pos_offset: usize,
) {
    if puyo_sets.is_empty() {
        return;
    }

    search_position_v2(runtime, puyo_sets, bit_field, hands, pos_offset);
}

fn search_position_v2(
    runtime: &mut SearchRuntime<'_>,
    puyo_sets: &[PuyoSet],
    bit_field: &BitField,
    hands: Vec<Hand>,
    pos_offset: usize,
) {
    let puyo_set = puyo_sets[0];
    let positions: &[[usize; 2]] = if puyo_set.axis == puyo_set.child {
        &SAME_COLOR_POSITIONS
    } else {
        &SETUP_POSITIONS
    };

    if bit_field.color(2, 12) != Color::Empty {
        return;
    }

    let heights = bit_field.create_heights();
    let check_pos_offset = use_same_pair_order_dedup(runtime.dedup_mode, runtime.stop_on_chain)
        && pos_offset > 0
        && pos_offset < positions.len();
    let is_terminal_depth = puyo_sets.len() == 1;

    for (index, pos) in positions.iter().copied().enumerate() {
        if check_pos_offset && !overlap(pos, positions[pos_offset]) && index < pos_offset {
            continue;
        }

        let Some(placement) =
            bit_field.search_placement_for_pos_with_heights(&puyo_set, pos, heights)
        else {
            continue;
        };
        if runtime.disable_chigiri && placement.chigiri {
            continue;
        }

        let mut before_simulate = *bit_field;
        if !before_simulate.place_puyo_with_placement(&placement) {
            panic!("should be able to place. {placement:?}");
        }

        if runtime.stop_on_chain && !is_terminal_depth && runtime.each_hand_callback.is_none() {
            if before_simulate.rensa_will_occur() {
                continue;
            }
        }

        let mut new_hands = hands.clone();
        new_hands.push(Hand {
            puyo_set,
            position: pos,
        });

        let mut rensa_result = simulate_for_node(
            runtime.simulate_policy,
            before_simulate,
            is_terminal_depth,
            runtime.each_hand_callback.is_some(),
        );
        rensa_result.set_frames = runtime.set_frames + placement.frames;
        rensa_result.chigiris = if placement.chigiri {
            runtime.chigiris + 1
        } else {
            runtime.chigiris
        };

        let remaining = puyo_sets.len() - 1;
        if uses_state_dedup(runtime.dedup_mode)
            && remaining > 0
            && mark_visited(
                runtime.visited_states,
                remaining,
                rensa_result
                    .bit_field
                    .as_ref()
                    .expect("simulate_for_node must set result bit_field"),
                runtime.dedup_mode,
            )
        {
            continue;
        }

        let search_result = SearchResult {
            rensa_result: Some(rensa_result),
            before_simulate: Some(before_simulate),
            depth: new_hands.len(),
            position: pos,
            position_num: index,
            hands: new_hands.clone(),
        };

        if let Some(callback) = runtime.each_hand_callback.as_mut() {
            if !callback(&search_result) {
                continue;
            }
        }
        if runtime.stop_on_chain
            && search_result
                .rensa_result
                .as_ref()
                .expect("search result must contain rensa_result")
                .chains
                > 0
        {
            continue;
        }

        if remaining == 0 {
            if let Some(callback) = runtime.last_callback.as_mut() {
                callback(&search_result);
            }
            continue;
        }

        let next_pos_offset =
            if use_same_pair_order_dedup(runtime.dedup_mode, runtime.stop_on_chain)
                && puyo_sets.len() > 1
                && same_puyo_set(puyo_sets[0], puyo_sets[1])
            {
                index
            } else {
                0
            };

        let child_runtime = SearchRuntime {
            disable_chigiri: runtime.disable_chigiri,
            chigiriable_count: runtime.chigiriable_count,
            chigiris: search_result
                .rensa_result
                .as_ref()
                .expect("search result must contain rensa_result")
                .chigiris,
            set_frames: search_result
                .rensa_result
                .as_ref()
                .expect("search result must contain rensa_result")
                .set_frames,
            dedup_mode: runtime.dedup_mode,
            simulate_policy: runtime.simulate_policy,
            stop_on_chain: runtime.stop_on_chain,
            last_callback: runtime.last_callback,
            each_hand_callback: runtime.each_hand_callback,
            visited_states: runtime.visited_states,
        };
        let mut child_runtime = child_runtime;
        search_with_puyo_sets_v2(
            &mut child_runtime,
            &puyo_sets[1..],
            search_result
                .rensa_result
                .as_ref()
                .and_then(|result| result.bit_field.as_ref())
                .expect("simulate_for_node must set result bit_field"),
            new_hands,
            next_pos_offset,
        );
    }
}

fn simulate_for_node(
    simulate_policy: SimulatePolicy,
    mut bit_field: BitField,
    terminal: bool,
    has_each_hand_callback: bool,
) -> RensaResult {
    let needs_detail = terminal || has_each_hand_callback;
    match simulate_policy {
        SimulatePolicy::FastAlways | SimulatePolicy::FastIntermediate => {
            if needs_detail {
                bit_field.simulate_detail()
            } else {
                bit_field.simulate()
            }
        }
        SimulatePolicy::DetailAlways => bit_field.simulate_detail(),
    }
}

fn use_same_pair_order_dedup(dedup_mode: DedupMode, stop_on_chain: bool) -> bool {
    dedup_mode == DedupMode::SamePairOrder && stop_on_chain
}

fn uses_state_dedup(dedup_mode: DedupMode) -> bool {
    matches!(dedup_mode, DedupMode::State | DedupMode::StateMirror)
}

fn uses_mirror_state_dedup(dedup_mode: DedupMode) -> bool {
    dedup_mode == DedupMode::StateMirror
}

fn mark_visited(
    visited_states: &mut Option<HashMap<usize, HashSet<SearchStateKey>>>,
    remaining: usize,
    bit_field: &BitField,
    dedup_mode: DedupMode,
) -> bool {
    let visited_states = visited_states
        .as_mut()
        .expect("visited_states must exist when state dedup is enabled");
    let visited = visited_states.entry(remaining).or_insert_with(HashSet::new);
    let key = create_search_state_key(bit_field, uses_mirror_state_dedup(dedup_mode));
    !visited.insert(key)
}

fn create_search_state_key(bit_field: &BitField, with_mirror: bool) -> SearchStateKey {
    let mut matrix = *bit_field.matrix();
    if with_mirror {
        let flip = mirror_bit_field_matrix(matrix);
        if less_bit_field_matrix(flip, matrix) {
            matrix = flip;
        }
    }
    SearchStateKey {
        m: matrix,
        table_sig: color_table_signature(bit_field),
    }
}

fn less_bit_field_matrix(a: [[u64; 2]; 3], b: [[u64; 2]; 3]) -> bool {
    for row in 0..a.len() {
        for col in 0..a[row].len() {
            if a[row][col] == b[row][col] {
                continue;
            }
            return a[row][col] < b[row][col];
        }
    }
    false
}

fn mirror_bit_field_matrix(m: [[u64; 2]; 3]) -> [[u64; 2]; 3] {
    let mut flip = [[0u64; 2]; 3];
    for i in 0..3 {
        flip[i][1] = (m[i][0] & 0xffff) << 16;
        flip[i][1] |= (m[i][0] & 0xffff0000) >> 16;
        flip[i][0] = (m[i][0] & 0xffff00000000) << 16;
        flip[i][0] |= (m[i][0] & 0xffff000000000000) >> 16;
        flip[i][0] |= (m[i][1] & 0xffff) << 16;
        flip[i][0] |= (m[i][1] & 0xffff0000) >> 16;
    }
    flip
}

fn color_table_signature(bit_field: &BitField) -> u32 {
    let mut signature = 0u32;
    for (index, color) in [
        Color::Red,
        Color::Blue,
        Color::Yellow,
        Color::Green,
        Color::Purple,
    ]
    .into_iter()
    .enumerate()
    {
        signature |= ((bit_field.color_table()[color.idx()] as u32) & 0xf) << (index * 4);
    }
    signature
}

fn same_puyo_set(left: PuyoSet, right: PuyoSet) -> bool {
    (left.axis == right.axis && left.child == right.child)
        || (left.axis == right.child && left.child == right.axis)
}

fn overlap(pos1: [usize; 2], pos2: [usize; 2]) -> bool {
    if pos1[0] == pos2[0] {
        return true;
    }
    if pos1[1] == 1 && (pos2[0] == pos1[0] + 1 || (pos2[1] == 3 && pos2[0] == pos1[0] + 2)) {
        return true;
    }
    if pos1[1] == 3
        && ((pos1[0] > 0 && pos2[0] == pos1[0] - 1)
            || (pos1[0] > 1 && pos2[1] == 1 && pos2[0] == pos1[0] - 2))
    {
        return true;
    }
    if pos2[1] == 1 && (pos1[0] == pos2[0] + 1 || (pos1[1] == 3 && pos1[0] == pos2[0] + 2)) {
        return true;
    }
    if pos2[1] == 3
        && ((pos2[0] > 0 && pos1[0] == pos2[0] - 1)
            || (pos2[0] > 1 && pos1[1] == 1 && pos1[0] == pos2[0] - 2))
    {
        return true;
    }
    false
}
