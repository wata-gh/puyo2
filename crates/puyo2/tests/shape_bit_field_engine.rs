use puyo2::{BitField, FieldBits, ShapeBitField};

fn shape(points: &[(usize, usize)]) -> FieldBits {
    let mut bits = FieldBits::new();
    for &(x, y) in points {
        bits.set_onebit(x, y);
    }
    bits
}

#[test]
fn fill_chainable_color_matches_go() {
    let cases = [
        (
            "........................23....11....13....12....52....42....33....445...455...",
            Some("aaaaaaaaaaaaaaaaaaaaaaaacdaaaabbaaaabdaaaabcaaaaecaaaabcaaaaddaaaabbeaaabeeaaa"),
        ),
        (
            "..............................5.....32....12....11....212...334...355...4454..",
            Some("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaacaaaaadcaaaabcaaaabbaaaacbcaaaddbaaadccaaabbcbaa"),
        ),
        (
            ".............................................................2....513....4....",
            None,
        ),
    ];

    for (field_string, expected) in cases {
        let shape_bit_field = ShapeBitField::from_field_string(field_string);
        let actual = shape_bit_field
            .fill_chainable_color()
            .map(|bit_field| bit_field.mattulwan_editor_param());
        assert_eq!(actual.as_deref(), expected);
    }
}

#[test]
fn insert_shape_matches_go() {
    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(shape(&[(0, 1), (0, 2), (0, 3), (1, 1)]));
    shape_bit_field.insert_shape(shape(&[(0, 2), (0, 3), (1, 2), (2, 2)]));
    assert_eq!(shape_bit_field.shapes[0].to_int_array(), [131122, 0]);

    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(shape(&[(0, 1), (0, 2), (1, 1), (2, 1)]));
    shape_bit_field.insert_shape(shape(&[(0, 1), (0, 2), (1, 1), (2, 1)]));
    assert_eq!(shape_bit_field.shapes[0].to_int_array(), [17180131352, 0]);

    shape_bit_field.insert_shape(shape(&[(0, 1), (0, 2), (1, 1), (2, 1)]));
    assert_eq!(shape_bit_field.shapes[0].to_int_array(), [34360262752, 0]);

    shape_bit_field.insert_shape(shape(&[(0, 1), (1, 1), (1, 2), (2, 1)]));
    assert_eq!(shape_bit_field.shapes[0].to_int_array(), [68721574080, 0]);

    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(shape(&[(3, 1), (3, 2), (4, 1), (5, 1)]));
    shape_bit_field.insert_shape(shape(&[(3, 1), (3, 2), (4, 1), (5, 1)]));
    assert_eq!(
        shape_bit_field.shapes[0].to_int_array(),
        [6755399441055744, 262148]
    );

    shape_bit_field.insert_shape(shape(&[(3, 1), (3, 2), (4, 1), (5, 1)]));
    assert_eq!(
        shape_bit_field.shapes[0].to_int_array(),
        [27021597764222976, 524296]
    );

    shape_bit_field.insert_shape(shape(&[(3, 1), (4, 1), (4, 2), (5, 1)]));
    assert_eq!(
        shape_bit_field.shapes[0].to_int_array(),
        [54043195528445952, 1048608]
    );
}

#[test]
fn simulate_matches_go() {
    let mut two_chain = ShapeBitField::new();
    two_chain.add_shape(shape(&[(0, 1), (0, 2), (1, 1), (1, 3)]));
    two_chain.add_shape(shape(&[(1, 2), (2, 1), (2, 2), (3, 1)]));
    let result = two_chain.simulate();
    assert_eq!(result.chains, 2);
    assert!(result.shape_bit_field.is_some());

    let mut five_chain = ShapeBitField::new();
    five_chain.add_shape(shape(&[(0, 2), (0, 3), (0, 4), (1, 2)]));
    five_chain.add_shape(shape(&[(0, 1), (1, 1), (1, 3), (2, 2)]));
    five_chain.add_shape(shape(&[(2, 1), (2, 3), (3, 2), (4, 2), (5, 2)]));
    five_chain.add_shape(shape(&[(3, 1), (4, 1), (4, 3), (5, 1)]));
    five_chain.add_shape(shape(&[(3, 3), (4, 4), (5, 3), (5, 4)]));
    let result = five_chain.simulate();
    assert_eq!(result.chains, 5);
    assert!(result.shape_bit_field.is_some());
}

#[test]
fn field_string_matches_go() {
    let field_string =
        "..........................................................5.123545112335223444";
    let shape_bit_field = ShapeBitField::from_field_string(field_string);
    assert_eq!(shape_bit_field.field_string(), field_string);

    let mut rebuilt = ShapeBitField::new();
    for shape in &shape_bit_field.shapes {
        rebuilt.add_shape(*shape);
    }
    assert_eq!(rebuilt.field_string(), field_string);
}

#[test]
fn key_string_matches_known_fixture() {
    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(shape(&[(0, 1), (0, 2), (1, 1), (1, 3)]));
    shape_bit_field.add_shape(shape(&[(1, 2), (2, 1), (2, 2), (3, 1)]));
    assert_eq!(shape_bit_field.key_string(), "_a0006:0_2000600040000:0");
}

#[test]
fn chain_ordered_field_string_matches_go() {
    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(shape(&[(0, 1), (0, 2), (1, 1), (1, 3)]));
    shape_bit_field.add_shape(shape(&[(1, 2), (2, 1), (2, 2), (3, 1)]));
    let result = shape_bit_field.simulate();
    assert_eq!(result.chains, 2);
    assert_eq!(
        shape_bit_field.chain_ordered_field_string(),
        ".............................................................2....211...2211.."
    );
}

#[test]
fn expand_3_puyo_shapes_matches_regression() {
    let shape_bit_field = ShapeBitField::from_field_string(
        ".............................................................2....513....4....",
    );
    let expanded = shape_bit_field.expand_3_puyo_shapes();
    assert_eq!(
        expanded.field_string(),
        ".............................................................2....513....4...."
    );
}

#[test]
fn bit_field_to_shape_bit_field_matches_go() {
    let bit_field = BitField::from_mattulwan("a54ea3eaebdece3bd2eb2dc3");
    let shape_bit_field = bit_field.to_shape_bit_field();
    assert_eq!(
        shape_bit_field.field_string(),
        "......................................................1.....1.....112335223444"
    );
    assert_eq!(
        shape_bit_field
            .shapes
            .iter()
            .map(|shape| shape.to_int_array())
            .collect::<Vec<_>>(),
        bit_field.to_chain_shapes_u64_array()
    );
}

#[test]
fn shape_num_reports_missing_and_present_shapes() {
    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(shape(&[(0, 1), (0, 2), (1, 1), (1, 2)]));
    assert_eq!(shape_bit_field.shape_num(0, 1), 0);
    assert_eq!(shape_bit_field.shape_num(5, 13), -1);
}

#[test]
fn original_overall_shape_stays_stable_after_simulation() {
    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(shape(&[(0, 1), (0, 2), (1, 1), (1, 2)]));
    let original = shape_bit_field.original_overall_shape();
    let result = shape_bit_field.simulate();
    assert_eq!(result.chains, 1);
    assert!(result.shape_bit_field.is_some());
    assert_eq!(shape_bit_field.original_overall_shape(), original);
    assert_eq!(shape_bit_field.overall_shape(), FieldBits::new());
    assert_eq!(shape_bit_field.shape_count(), 1);
    assert_eq!(shape_bit_field.original_shapes.len(), 1);
    assert_eq!(shape_bit_field.original_shapes[0], original);
    assert_eq!(shape_bit_field.shapes[0], FieldBits::new());
    assert_eq!(shape_bit_field.chain_ordered_shapes.len(), 1);
    assert_eq!(shape_bit_field.chain_ordered_shapes[0].len(), 1);
    assert_eq!(shape_bit_field.chain_ordered_shapes[0][0], original);
    assert_eq!(shape_bit_field.shape_num(0, 1), -1);
    assert!(shape_bit_field.is_empty());
    assert_eq!(shape_bit_field.original_shapes[0].popcount(), 4);
    assert_eq!(shape_bit_field.chain_ordered_shapes[0][0].popcount(), 4);
    assert_eq!(
        shape_bit_field.original_shapes[0],
        shape(&[(0, 1), (0, 2), (1, 1), (1, 2)])
    );
    assert_eq!(
        shape_bit_field.chain_ordered_shapes[0][0],
        shape(&[(0, 1), (0, 2), (1, 1), (1, 2)])
    );
}

#[test]
fn drop_matches_reference_on_random_shapes() {
    let mut seed = 0x7f4a_7c15_d2b8_9e31u64;
    for _ in 0..128 {
        let mut actual = ShapeBitField::new();
        let shape_count = (next_u64(&mut seed) % 8 + 1) as usize;
        for _ in 0..shape_count {
            actual.add_shape(random_bits(&mut seed));
        }
        actual.key_string = Some("cached".to_string());
        let vanished = random_bits(&mut seed);
        let expected = drop_shapes_reference(&actual.shapes, vanished);

        actual.drop(vanished);

        assert_eq!(
            actual
                .shapes
                .iter()
                .map(FieldBits::to_int_array)
                .collect::<Vec<_>>(),
            expected
        );
        assert_eq!(actual.key_string, None);
    }
}

fn random_bits(seed: &mut u64) -> FieldBits {
    let mut bits = FieldBits::new();
    for y in 1..=13 {
        for x in 0..6 {
            if next_u64(seed) & 3 == 0 {
                bits.set_onebit(x, y);
            }
        }
    }
    bits
}

fn drop_shapes_reference(shapes: &[FieldBits], vanished: FieldBits) -> Vec<[u64; 2]> {
    shapes
        .iter()
        .map(|shape| drop_shape_reference(shape.to_int_array(), vanished))
        .collect()
}

fn drop_shape_reference(mut matrix: [u64; 2], vanished: FieldBits) -> [u64; 2] {
    let vanished_matrix = vanished.to_int_array();
    let mut dropmask = [0u64; 2];
    for x in 0..6 {
        let idx = x >> 2;
        let vc = vanished.col_bits(x).count_ones();
        let rotated = (((1u64 << vc) - 1).rotate_left(14u32 - vc)) << ((x & 3) * 16);
        dropmask[idx] |= rotated;
    }

    let r0 = extract_reference(matrix[0], !vanished_matrix[0]);
    let r1 = extract_reference(matrix[1], !vanished_matrix[1]);
    matrix[0] = deposit_reference(r0, !dropmask[0]);
    matrix[1] = deposit_reference(r1, !dropmask[1]);
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

fn next_u64(seed: &mut u64) -> u64 {
    *seed = seed.wrapping_mul(6364136223846793005).wrapping_add(1);
    *seed
}
