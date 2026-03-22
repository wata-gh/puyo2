use std::str::FromStr;

use puyo2::{
    Color, ColorParseError, DedupMode, Hand, HandParseError, SearchCondition, SimulatePolicy,
    expand_mattulwan_param, haipuyo_to_puyo_sets, normalize_haipuyo, parse_simple_hands,
    to_simple_hands,
};

#[test]
fn expand_mattulwan_param_matches_go() {
    let expanded = expand_mattulwan_param("a58babcdbeb3cd2bc2de3");
    assert_eq!(
        expanded,
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaababcdbebbbcddbccdeee"
    );
}

#[test]
fn simple_hands_round_trip() {
    let hands = parse_simple_hands("gb01gy10yb52gb43yb32yb53").unwrap();
    assert_eq!(to_simple_hands(&hands).unwrap(), "gb01gy10yb52gb43yb32yb53");
}

#[test]
fn haipuyo_to_puyo_sets_matches_go() {
    let puyo_sets = haipuyo_to_puyo_sets("pryr").unwrap();
    assert_eq!(puyo_sets.len(), 2);
    assert_eq!(puyo_sets[0].axis, Color::Purple);
    assert_eq!(puyo_sets[0].child, Color::Red);
    assert_eq!(puyo_sets[1].axis, Color::Yellow);
    assert_eq!(puyo_sets[1].child, Color::Red);
}

#[test]
fn normalize_haipuyo_empty_string_is_empty() {
    assert_eq!(normalize_haipuyo("").unwrap(), "");
}

#[test]
fn normalize_haipuyo_collapses_same_pair_order() {
    assert_eq!(normalize_haipuyo("rrrb").unwrap(), "1112");
    assert_eq!(
        normalize_haipuyo("rrrb").unwrap(),
        normalize_haipuyo("rrbr").unwrap()
    );
}

#[test]
fn normalize_haipuyo_collapses_cross_pair_color_symmetry() {
    assert_eq!(normalize_haipuyo("rbrg").unwrap(), "1213");
    assert_eq!(
        normalize_haipuyo("rbrg").unwrap(),
        normalize_haipuyo("rbbg").unwrap()
    );
}

#[test]
fn normalize_haipuyo_collapses_global_color_renaming() {
    assert_eq!(normalize_haipuyo("rbgy").unwrap(), "1234");
    assert_eq!(
        normalize_haipuyo("rbgy").unwrap(),
        normalize_haipuyo("bgyr").unwrap()
    );
}

#[test]
fn normalize_haipuyo_resolves_purple_placeholder() {
    assert_eq!(normalize_haipuyo("pprr").unwrap(), "1122");
    assert_eq!(
        normalize_haipuyo("pprr").unwrap(),
        normalize_haipuyo("bbrr").unwrap()
    );
}

#[test]
fn normalize_haipuyo_rejects_five_color_inputs() {
    assert!(matches!(
        normalize_haipuyo("prbgyrby"),
        Err(HandParseError::TooManyColorsToNormalize(5))
    ));
}

#[test]
fn normalize_haipuyo_preserves_existing_parse_failures() {
    assert!(matches!(
        normalize_haipuyo("rgb"),
        Err(HandParseError::OddHaipuyoLength(3))
    ));
    assert!(matches!(
        normalize_haipuyo("rx"),
        Err(HandParseError::InvalidColor(ColorParseError::InvalidLetter(letter)))
            if letter == "x"
    ));
}

#[test]
fn parse_modes_match_go_strings() {
    assert_eq!(DedupMode::default().to_string(), "off");
    assert_eq!(SimulatePolicy::default().to_string(), "detail_always");
    assert_eq!(
        DedupMode::from_str("same_pair_order").unwrap(),
        DedupMode::SamePairOrder
    );
    assert_eq!(
        SimulatePolicy::from_str("fast_intermediate").unwrap(),
        SimulatePolicy::FastIntermediate
    );
}

#[test]
fn search_condition_defaults_match_go() {
    let cond = SearchCondition::new();
    assert_eq!(cond.dedup_mode, DedupMode::Off);
    assert_eq!(cond.simulate_policy, SimulatePolicy::DetailAlways);
    assert!(!cond.stop_on_chain);
}

#[test]
fn to_simple_hands_rejects_non_hand_colors() {
    let hands = vec![Hand {
        puyo_set: puyo2::PuyoSet {
            axis: Color::Ojama,
            child: Color::Red,
        },
        position: [0, 0],
    }];
    assert!(to_simple_hands(&hands).is_err());
}
