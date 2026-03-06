use puyo2::{
    BitField, Color, IPSNazoCondition, evaluate_ips_nazo_condition, parse_ips_nazo_url,
    parse_simple_hands,
};

fn new_condition(q0: usize, q1: usize, q2: usize) -> IPSNazoCondition {
    IPSNazoCondition {
        q0,
        q1,
        q2,
        ..IPSNazoCondition::default()
    }
}

fn create_red4_field(with_ojama: bool) -> BitField {
    let mut bit_field = BitField::new();
    for x in 0..4 {
        bit_field.set_color(Color::Red, x, 1);
    }
    if with_ojama {
        bit_field.set_color(Color::Ojama, 4, 1);
    }
    bit_field
}

#[test]
fn evaluate_ips_nazo_condition_basic_matches_go() {
    let bit_field = create_red4_field(false);
    let (matched, metrics) = evaluate_ips_nazo_condition(&bit_field, &new_condition(30, 0, 1));
    assert!(matched);
    assert_eq!(metrics.chain_count, 1);
    assert_eq!(metrics.erased[1], 4);
    assert_eq!(metrics.remaining[1], 0);

    let cases = [
        ("31 false", new_condition(31, 0, 2), false),
        ("2 red", new_condition(2, 1, 0), true),
        ("2 color", new_condition(2, 7, 0), true),
        ("10 true", new_condition(10, 0, 1), true),
        ("11 false", new_condition(11, 0, 2), false),
        ("12 red", new_condition(12, 1, 4), true),
        ("13 red false", new_condition(13, 1, 5), false),
        ("12 color", new_condition(12, 7, 4), true),
        ("32 true", new_condition(32, 1, 1), true),
        ("33 false", new_condition(33, 1, 2), false),
        ("40 true", new_condition(40, 0, 1), true),
        ("41 false", new_condition(41, 0, 2), false),
        ("42 true", new_condition(42, 1, 4), true),
        ("43 false", new_condition(43, 1, 5), false),
        ("44 true", new_condition(44, 1, 1), true),
        ("45 false", new_condition(45, 1, 2), false),
        ("52 true", new_condition(52, 1, 4), true),
        ("53 false", new_condition(53, 1, 5), false),
    ];

    for (name, condition, expected) in cases {
        let (got, _) = evaluate_ips_nazo_condition(&bit_field, &condition);
        assert_eq!(got, expected, "{name}");
    }
}

#[test]
fn evaluate_ips_nazo_condition_ojama_and_color_index_match_go() {
    let bit_field = create_red4_field(true);
    let cases = [
        ("2 ojama", new_condition(2, 6, 0), true),
        ("2 color", new_condition(2, 7, 0), true),
        ("12 ojama", new_condition(12, 6, 1), true),
        ("42 ojama", new_condition(42, 6, 1), true),
        ("44 ojama false", new_condition(44, 6, 1), false),
        ("44 ojama zero", new_condition(44, 6, 0), true),
    ];

    for (name, condition, expected) in cases {
        let (got, metrics) = evaluate_ips_nazo_condition(&bit_field, &condition);
        assert_eq!(got, expected, "{name}: {metrics:?}");
    }
}

#[test]
fn evaluate_ips_nazo_condition_manual_hands_regression_matches_go() {
    let param = "M00M0MM6MM6SM4So6sMy9jCsPz9zPaCPiC_G1u1s1e1i1u1__260";
    let hands = "yy30yb30bb10gg50yg41yb00";

    let decoded = parse_ips_nazo_url(param).unwrap();
    let parsed_hands = parse_simple_hands(hands).unwrap();
    let mut bit_field = BitField::from_mattulwan(&decoded.initial_field);

    for (index, hand) in parsed_hands.iter().enumerate() {
        let (placed, _) = bit_field.place_puyo(hand.puyo_set, hand.position);
        assert!(
            placed,
            "hand[{index}] place failed: heights={:?}",
            bit_field.create_heights()
        );
    }

    let (matched, metrics) = evaluate_ips_nazo_condition(&bit_field, &decoded.condition);
    assert!(matched, "metrics={metrics:?}");
    assert_eq!(metrics.remaining[6], 0);
}
