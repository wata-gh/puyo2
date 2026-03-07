use puyo2::{BitField, Color, FieldBits, PuyoSet, PuyoSetPlacement};

fn assert_bits_eq(actual: FieldBits, expected: FieldBits) {
    assert_eq!(actual.to_int_array(), expected.to_int_array());
}

#[test]
fn to_chain_shapes_matches_go() {
    let field = BitField::from_mattulwan("a54ea3eaebdece3bd2eb2dc3");
    let expected = vec![
        [262172, 0],
        [17180262402, 0],
        [1125925676646400, 4],
        [562949953421312, 131078],
        [562949953421312, 393218],
    ];
    assert_eq!(field.to_chain_shapes_u64_array(), expected);
}

#[test]
fn drop_vanished_matches_go() {
    let mut field = BitField::new();
    field.set_color(Color::Red, 0, 13);
    let mut vanished = FieldBits::new();
    vanished.set_onebit(0, 12);
    field.drop_vanished(vanished);
    assert_eq!(field.color(0, 12), Color::Red);
}

#[test]
fn mask_field_matches_go() {
    let field = BitField::from_mattulwan(
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaeaaaaacccaaaeeeaaabcdaeabbcddaccdeee",
    );
    let mask_field = BitField::from_mattulwan(
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdaeabbcddaccdeee",
    );
    let masked = field.mask_field(&mask_field.overall_shape());
    assert_eq!(
        masked.mattulwan_editor_param(),
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdaeabbcddaccdeee"
    );
}

#[test]
fn equal_chain_matches_go() {
    let field = BitField::from_mattulwan("a54ea3eaebdece3bd2eb2dc3");
    assert!(field.equal_chain(&field));

    let diff_colors = BitField::from_mattulwan("a54ba3babcebdb3ce2bc2ed3");
    assert!(field.equal_chain(&diff_colors));

    let diff_shape = BitField::from_mattulwan("a54ba3b3cebdb3ce3c2ed3");
    assert!(!field.equal_chain(&diff_shape));
}

#[test]
fn rensa_will_occur_matches_go() {
    let chainable = BitField::from_mattulwan("a54ea3eaebdece3bd2eb2dc3");
    assert!(chainable.rensa_will_occur());

    let not_chainable = BitField::from_mattulwan("a78");
    assert!(!not_chainable.rensa_will_occur());
}

#[test]
fn from_mattulwan_and_haipuyo_keeps_purple_mapping() {
    let mut field = BitField::from_mattulwan_and_haipuyo("ba77", "pprr").unwrap();
    assert_eq!(field.color(0, 13), Color::Red);
    field.set_color(Color::Purple, 0, 1);
    assert_eq!(field.color(0, 1), Color::Purple);
}

#[test]
fn set_mattulwan_matches_go() {
    let field = "a54ea3eaebdece3bd2eb2dc3";
    let mut bit_field = BitField::new();
    bit_field.set_mattulwan_param(field);
    let expected = BitField::from_mattulwan(field);
    assert_eq!(
        bit_field.mattulwan_editor_param(),
        expected.mattulwan_editor_param()
    );

    let field = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaacaaaaabfaaaadbdaabbffaa";
    let mut bit_field =
        BitField::with_colors(vec![Color::Red, Color::Blue, Color::Yellow, Color::Purple]);
    bit_field.set_mattulwan_param(field);
    let expected = BitField::from_mattulwan(field);
    assert_eq!(
        bit_field.mattulwan_editor_param(),
        expected.mattulwan_editor_param()
    );
}

#[test]
fn simulate_detail_matches_go() {
    let cases = [
        ("a54ea3eaebdece3bd2eb2dc3", 5, 4_840, true),
        (
            "a2bdeb2c2bdcecbde2bcbdecedcedcdcedbcedcedbedcedbeb2cbcbcdec2ebcdebebcdebebcdeb",
            19,
            175_080,
            true,
        ),
        ("a34ca5dca4dca4dca4bda4eba4eba3e2b", 2, 1_720, true),
        ("a34ba5dba4dba4dba4bda4eba4eba3e2b", 2, 1_360, true),
        (
            "ga2g2c2a2g2dca2dgegae2bcgae2dcgaed2egcbeb2gdbedbcgbde2dg2c2edbgedcdcgd2c3ge2c",
            9,
            42_540,
            false,
        ),
    ];

    for (param, chains, score, expect_empty) in cases {
        let mut field = BitField::from_mattulwan(param);
        let result = field.simulate_detail();
        assert_eq!(result.chains, chains);
        assert_eq!(result.score, score);
        assert_eq!(result.bit_field.as_ref().unwrap().is_empty(), expect_empty);
    }
}

#[test]
fn simulate_matches_go() {
    let cases = [
        ("a54ea3eaebdece3bd2eb2dc3", 5, true),
        (
            "a2bdeb2c2bdcecbde2bcbdecedcedcdcedbcedcedbedcedbeb2cbcbcdec2ebcdebebcdebebcdeb",
            19,
            true,
        ),
        (
            "ga2g2c2a2g2dca2dgegae2bcgae2dcgaed2egcbeb2gdbedbcgbde2dg2c2edbgedcdcgd2c3ge2c",
            9,
            false,
        ),
    ];

    for (param, chains, expect_empty) in cases {
        let mut field = BitField::from_mattulwan(param);
        let result = field.simulate();
        assert_eq!(result.chains, chains);
        assert_eq!(result.bit_field.as_ref().unwrap().is_empty(), expect_empty);
    }
}

#[test]
fn from_mattulwan_reads_purple() {
    let field = BitField::from_mattulwan("a54ba5bedafab2ed2ae2df3");
    assert_eq!(field.color(3, 1), Color::Purple);
}

#[test]
fn register_color_keeps_original_copy_untouched() {
    let original = BitField::from_mattulwan("ba77");
    let mut cloned = original;

    cloned.register_color(Color::Blue);

    assert_eq!(original.color_table()[Color::Blue.idx()], Color::Empty);
    assert!(!original.colors().contains(&Color::Blue));
    assert_eq!(cloned.color_table()[Color::Blue.idx()], Color::Blue);
    assert!(cloned.colors().contains(&Color::Blue));
}

#[test]
fn colors_accessor_keeps_purple_mapping_order() {
    let field = BitField::from_mattulwan(
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaacaaaaabfaaaadbdaabbffaa",
    );
    assert_eq!(
        field.colors(),
        &[Color::Red, Color::Blue, Color::Yellow, Color::Purple]
    );
}

#[test]
fn place_puyo_with_placement_validation_matches_go() {
    let mut field = BitField::from_mattulwan("ba77");
    assert!(!field.place_puyo_with_placement(&PuyoSetPlacement {
        puyo_set: None,
        pos: [0, 0],
        axis_x: 0,
        axis_y: 0,
        child_x: 0,
        child_y: 0,
        chigiri: false,
        frames: 0,
    }));

    assert!(!field.place_puyo_with_placement(&PuyoSetPlacement {
        puyo_set: Some(PuyoSet {
            axis: Color::Red,
            child: Color::Blue,
        }),
        pos: [0, 0],
        axis_x: 0,
        axis_y: 13,
        child_x: 1,
        child_y: 13,
        chigiri: false,
        frames: 0,
    }));

    assert!(!field.place_puyo_with_placement(&PuyoSetPlacement {
        puyo_set: Some(PuyoSet {
            axis: Color::Red,
            child: Color::Blue,
        }),
        pos: [0, 0],
        axis_x: -1,
        axis_y: 1,
        child_x: 0,
        child_y: 1,
        chigiri: false,
        frames: 0,
    }));
}

#[test]
fn set_color_with_field_bits_purple_matches_go() {
    let mut field =
        BitField::with_colors(vec![Color::Red, Color::Blue, Color::Yellow, Color::Purple]);
    let mut bits = FieldBits::new();
    bits.set_onebit(0, 1);
    field.set_color_with_field_bits(Color::Purple, bits);
    assert_eq!(field.color(0, 1), Color::Purple);
}

#[test]
fn to_chain_shapes_empty_when_no_chain() {
    let mut field = BitField::new();
    field.set_color(Color::Red, 0, 1);
    assert!(field.to_chain_shapes().is_empty());
}

#[test]
fn place_puyo_drop_results_match_go() {
    let mut field = BitField::new();
    field.place_puyo(
        PuyoSet {
            axis: Color::Red,
            child: Color::Green,
        },
        [0, 0],
    );
    field.drop_vanished(field.bits(Color::Empty).mask_field12());
    assert_bits_eq(field.bits(Color::Red), FieldBits::with_matrix([2, 0]));
    assert_bits_eq(field.bits(Color::Green), FieldBits::with_matrix([4, 0]));

    let mut field = BitField::new();
    field.place_puyo(
        PuyoSet {
            axis: Color::Red,
            child: Color::Green,
        },
        [0, 1],
    );
    field.drop_vanished(field.bits(Color::Empty).mask_field12());
    assert_bits_eq(field.bits(Color::Red), FieldBits::with_matrix([2, 0]));
    assert_bits_eq(
        field.bits(Color::Green),
        FieldBits::with_matrix([2 << 16, 0]),
    );

    let mut field = BitField::new();
    field.place_puyo(
        PuyoSet {
            axis: Color::Red,
            child: Color::Green,
        },
        [0, 2],
    );
    field.drop_vanished(field.bits(Color::Empty).mask_field12());
    assert_bits_eq(field.bits(Color::Red), FieldBits::with_matrix([4, 0]));
    assert_bits_eq(field.bits(Color::Green), FieldBits::with_matrix([2, 0]));

    let mut field = BitField::new();
    field.place_puyo(
        PuyoSet {
            axis: Color::Red,
            child: Color::Green,
        },
        [1, 3],
    );
    field.drop_vanished(field.bits(Color::Empty).mask_field12());
    assert_bits_eq(field.bits(Color::Red), FieldBits::with_matrix([2 << 16, 0]));
    assert_bits_eq(field.bits(Color::Green), FieldBits::with_matrix([2, 0]));
}

#[test]
fn drop_vanished_matches_reference_on_random_fields() {
    let mut seed = 0x6c8e_9cf5_7093_2bd5u64;
    for _ in 0..512 {
        let field = random_field(&mut seed);
        let vanished = field.find_vanishing_bits();

        let mut actual = field;
        actual.drop_vanished(vanished);

        let expected = drop_vanished_reference(*field.matrix(), vanished);

        assert_eq!(
            actual.matrix(),
            &expected,
            "matrix mismatch: field={} vanished={:?}",
            field.mattulwan_editor_param(),
            vanished.to_int_array()
        );
        assert_eq!(actual.color_table(), field.color_table());
        assert_eq!(actual.colors(), field.colors());
    }
}

fn random_field(seed: &mut u64) -> BitField {
    let palette = if next_u64(seed) & 1 == 0 {
        [Color::Red, Color::Blue, Color::Yellow, Color::Green]
    } else {
        [Color::Red, Color::Blue, Color::Yellow, Color::Purple]
    };
    let mut field = BitField::from_mattulwan("a78");
    for y in 1..=13 {
        for x in 0..6 {
            if next_u64(seed) % 10 >= 4 {
                continue;
            }
            let color = if next_u64(seed) % 8 == 0 {
                Color::Ojama
            } else {
                palette[(next_u64(seed) as usize) % palette.len()]
            };
            field.set_color(color, x, y);
        }
    }
    field
}

fn next_u64(seed: &mut u64) -> u64 {
    *seed = seed.wrapping_mul(6364136223846793005).wrapping_add(1);
    *seed
}

fn drop_vanished_reference(mut matrix: [[u64; 2]; 3], vanished: FieldBits) -> [[u64; 2]; 3] {
    let vanished_matrix = vanished.to_int_array();
    let mut dropmask = [0u64; 2];
    for x in 0..6 {
        let idx = x >> 2;
        let vc = vanished.col_bits(x).count_ones();
        let rotated = (((1u64 << vc) - 1).rotate_left(14u32 - vc)) << ((x & 3) * 16);
        dropmask[idx] |= rotated;
    }
    for plane in 0..matrix.len() {
        let r0 = extract_reference(matrix[plane][0], !vanished_matrix[0]);
        let r1 = extract_reference(matrix[plane][1], !vanished_matrix[1]);
        matrix[plane][0] = deposit_reference(r0, !dropmask[0]);
        matrix[plane][1] = deposit_reference(r1, !dropmask[1]);
    }
    matrix
}

fn extract_reference(x: u64, mut mask: u64) -> u64 {
    let mut result = 0u64;
    let mut next = 1u64;
    loop {
        let lsb = mask & mask.wrapping_neg();
        if lsb == 0 {
            return result;
        }
        mask ^= lsb;
        if x & lsb != 0 {
            result |= next;
        }
        next <<= 1;
    }
}

fn deposit_reference(x: u64, mut mask: u64) -> u64 {
    let mut result = 0u64;
    let mut next = 1u64;
    loop {
        let lsb = mask & mask.wrapping_neg();
        if lsb == 0 {
            return result;
        }
        mask ^= lsb;
        if x & next != 0 {
            result |= lsb;
        }
        next <<= 1;
    }
}
