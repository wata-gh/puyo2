use std::{
    cell::{Cell, RefCell},
    collections::HashSet,
    rc::Rc,
};

use puyo2::{BitField, Color, DedupMode, PuyoSet, SearchCondition, SearchResult, SimulatePolicy};

fn assert_callback_contract(sr: &SearchResult) {
    assert!(sr.before_simulate.is_some(), "before_simulate must be set");
    assert!(sr.rensa_result.is_some(), "rensa_result must be set");
}

fn count_search_nodes(
    param: &str,
    puyo_sets: Vec<PuyoSet>,
    dedup: DedupMode,
    policy: SimulatePolicy,
    stop_on_chain: bool,
    prune_intermediate_chains: bool,
) -> usize {
    let count = Rc::new(Cell::new(0usize));
    let count_handle = Rc::clone(&count);
    let terminal_depth = puyo_sets.len();
    let mut cond =
        SearchCondition::with_bit_field_and_puyo_sets(BitField::from_mattulwan(param), puyo_sets);
    cond.dedup_mode = dedup;
    cond.simulate_policy = policy;
    cond.stop_on_chain = stop_on_chain;
    cond.each_hand_callback = Some(Box::new(move |sr| {
        assert_callback_contract(sr);
        count_handle.set(count_handle.get() + 1);
        if prune_intermediate_chains
            && sr.depth != terminal_depth
            && sr.rensa_result.as_ref().unwrap().chains != 0
        {
            return false;
        }
        true
    }));
    cond.search_with_puyo_sets_v2();
    count.get()
}

fn terminal_result_key(sr: &SearchResult) -> String {
    assert_callback_contract(sr);
    let result = sr.rensa_result.as_ref().unwrap();
    let bit_field = result.bit_field.as_ref().unwrap();
    let m = bit_field.matrix();
    format!(
        "{:x}:{:x}:{:x}:{:x}:{:x}:{:x}|c={}|s={}|e={}|q={}",
        m[0][0],
        m[0][1],
        m[1][0],
        m[1][1],
        m[2][0],
        m[2][1],
        result.chains,
        result.score,
        result.erased,
        result.quick
    )
}

fn collect_terminal_result_set(
    param: &str,
    puyo_sets: Vec<PuyoSet>,
    dedup: DedupMode,
    stop_on_chain: bool,
    prune_intermediate_chains: bool,
) -> HashSet<String> {
    let result_set = Rc::new(RefCell::new(HashSet::new()));
    let result_set_handle = Rc::clone(&result_set);
    let terminal_depth = puyo_sets.len();
    let mut cond =
        SearchCondition::with_bit_field_and_puyo_sets(BitField::from_mattulwan(param), puyo_sets);
    cond.dedup_mode = dedup;
    cond.stop_on_chain = stop_on_chain;
    if prune_intermediate_chains {
        cond.each_hand_callback = Some(Box::new(move |sr| {
            assert_callback_contract(sr);
            !(sr.depth != terminal_depth && sr.rensa_result.as_ref().unwrap().chains != 0)
        }));
    }
    cond.last_callback = Some(Box::new(move |sr| {
        result_set_handle
            .borrow_mut()
            .insert(terminal_result_key(sr));
    }));
    cond.search_with_puyo_sets_v2();
    result_set.borrow().clone()
}

#[test]
fn search_condition_defaults_match_go() {
    let cond = SearchCondition::new();
    assert_eq!(cond.dedup_mode, DedupMode::Off);
    assert_eq!(cond.simulate_policy, SimulatePolicy::DetailAlways);
    assert!(!cond.stop_on_chain);
}

#[test]
fn search_with_puyo_sets_v2_same_pair_order_reduction() {
    let tests = [
        (
            "a78",
            vec![
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Red,
                },
            ],
        ),
        (
            "a46ea5ea5ea5ga5ea4eba",
            vec![
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Red,
                },
            ],
        ),
    ];

    for (param, puyo_sets) in tests {
        let off = count_search_nodes(
            param,
            puyo_sets.clone(),
            DedupMode::Off,
            SimulatePolicy::DetailAlways,
            true,
            false,
        );
        let same_pair = count_search_nodes(
            param,
            puyo_sets,
            DedupMode::SamePairOrder,
            SimulatePolicy::DetailAlways,
            true,
            false,
        );
        assert!(off > 0, "off count must not be zero");
        assert!(
            same_pair * 2 <= off,
            "same_pair_order must reduce at least 50% when stop_on_chain=true: off={off} same_pair={same_pair}"
        );
    }
}

#[test]
fn search_with_puyo_sets_v2_same_pair_order_disabled_without_stop_on_chain() {
    let tests = [
        (
            "a78",
            vec![
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Red,
                },
            ],
        ),
        (
            "a46ea5ea5ea5ga5ea4eba",
            vec![
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Red,
                },
            ],
        ),
    ];

    for (param, puyo_sets) in tests {
        let off = count_search_nodes(
            param,
            puyo_sets.clone(),
            DedupMode::Off,
            SimulatePolicy::DetailAlways,
            false,
            false,
        );
        let same_pair = count_search_nodes(
            param,
            puyo_sets,
            DedupMode::SamePairOrder,
            SimulatePolicy::DetailAlways,
            false,
            false,
        );
        assert_eq!(
            off, same_pair,
            "same_pair_order must be disabled when stop_on_chain=false"
        );
    }
}

#[test]
fn search_with_puyo_sets_v2_state_result_set_equal() {
    let param = "a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16";
    let puyo_sets = vec![
        PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        },
        PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        },
        PuyoSet {
            axis: Color::Blue,
            child: Color::Blue,
        },
        PuyoSet {
            axis: Color::Blue,
            child: Color::Blue,
        },
    ];

    let off_set =
        collect_terminal_result_set(param, puyo_sets.clone(), DedupMode::Off, false, false);
    let state_set = collect_terminal_result_set(param, puyo_sets, DedupMode::State, false, false);
    assert_eq!(off_set.len(), state_set.len());
    for key in &off_set {
        assert!(state_set.contains(key), "state dedup missing {key}");
    }
}

#[test]
fn search_with_puyo_sets_v2_state_mirror_reduces_nodes_and_preserves_result_set() {
    let puyo_sets = vec![
        PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        },
        PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        },
        PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        },
    ];

    let state_count = count_search_nodes(
        "a78",
        puyo_sets.clone(),
        DedupMode::State,
        SimulatePolicy::DetailAlways,
        true,
        true,
    );
    let mirror_count = count_search_nodes(
        "a78",
        puyo_sets.clone(),
        DedupMode::StateMirror,
        SimulatePolicy::DetailAlways,
        true,
        true,
    );
    assert!(
        mirror_count < state_count,
        "state_mirror must reduce node count: state={state_count} mirror={mirror_count}"
    );

    let state_set =
        collect_terminal_result_set("a78", puyo_sets.clone(), DedupMode::State, true, true);
    let mirror_set =
        collect_terminal_result_set("a78", puyo_sets, DedupMode::StateMirror, true, true);
    assert_eq!(state_set, mirror_set);
}
