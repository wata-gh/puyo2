use puyo2::{BitField, Color, FieldBits};

fn assert_bits_eq(actual: FieldBits, expected: FieldBits) {
    assert_eq!(actual.to_int_array(), expected.to_int_array());
}

#[test]
fn fast_lift_matches_go() {
    let mut bits = FieldBits::new();
    bits.set_onebit(0, 1);
    bits.set_onebit(0, 2);
    bits.set_onebit(1, 1);
    bits.set_onebit(2, 1);

    let lifted1 = bits.fast_lift(1);
    let mut expect1 = FieldBits::new();
    expect1.set_onebit(0, 2);
    expect1.set_onebit(0, 3);
    expect1.set_onebit(1, 2);
    expect1.set_onebit(2, 2);
    assert_bits_eq(lifted1, expect1);

    let lifted2 = bits.fast_lift(18);
    let mut expect2 = FieldBits::new();
    expect2.set_onebit(1, 3);
    expect2.set_onebit(1, 4);
    expect2.set_onebit(2, 3);
    expect2.set_onebit(3, 3);
    assert_bits_eq(lifted2, expect2);
}

#[test]
fn not_matches_go() {
    let mut bits = FieldBits::new();
    bits.set_onebit(0, 1);
    let inverted = bits.not();
    assert_eq!(
        inverted.to_int_array(),
        [18446744073709551613, 18446744073709551615]
    );
}

#[test]
fn and_and_not_match_go() {
    let mut bits = FieldBits::new();
    bits.set_onebit(0, 4);
    bits.set_onebit(0, 3);
    bits.set_onebit(0, 2);
    bits.set_onebit(1, 2);

    let mut other = FieldBits::new();
    other.set_onebit(0, 4);
    other.set_onebit(0, 2);
    other.set_onebit(2, 2);

    let result = bits.and(other);
    let mut expected = FieldBits::new();
    expected.set_onebit(0, 4);
    expected.set_onebit(0, 2);
    assert_bits_eq(result, expected);

    let mut upper = FieldBits::new();
    upper.set_onebit(4, 4);
    upper.set_onebit(4, 3);
    upper.set_onebit(4, 2);
    upper.set_onebit(5, 2);

    let mut merged = bits.or(upper);
    merged.and_not_mut(bits);
    assert_bits_eq(merged, upper);
}

#[test]
fn equals_matches_go() {
    let mut bits = FieldBits::new();
    bits.set_onebit(0, 1);

    let mut same = FieldBits::new();
    same.set_onebit(0, 1);
    assert!(bits.equals(&same));

    let mut different = FieldBits::new();
    different.set_onebit(1, 1);
    assert!(!bits.equals(&different));
}

#[test]
fn find_vanishing_bits_matches_go() {
    let mut field = BitField::from_mattulwan("a54ea3eaebdece3bd2eb2dc3");
    let vanished = field.bits(Color::Green).find_vanishing_bits();
    let mut expect = FieldBits::new();
    expect.set_onebit(0, 4);
    expect.set_onebit(0, 3);
    expect.set_onebit(0, 2);
    expect.set_onebit(1, 2);
    assert_bits_eq(vanished, expect);

    assert!(field.simulate1());
    let vanished = field.bits(Color::Red).find_vanishing_bits();
    let mut expect = FieldBits::new();
    expect.set_onebit(0, 1);
    expect.set_onebit(1, 1);
    expect.set_onebit(1, 2);
    expect.set_onebit(2, 2);
    assert_bits_eq(vanished, expect);

    assert!(field.simulate1());
    let vanished = field.bits(Color::Yellow).find_vanishing_bits();
    let mut expect = FieldBits::new();
    expect.set_onebit(2, 1);
    expect.set_onebit(2, 2);
    expect.set_onebit(3, 2);
    expect.set_onebit(4, 2);
    assert_bits_eq(vanished, expect);

    assert!(field.simulate1());
    let vanished = field.bits(Color::Blue).find_vanishing_bits();
    let mut expect = FieldBits::new();
    expect.set_onebit(3, 1);
    expect.set_onebit(4, 1);
    expect.set_onebit(4, 2);
    expect.set_onebit(5, 1);
    assert_bits_eq(vanished, expect);

    assert!(field.simulate1());
    let vanished = field.bits(Color::Green).find_vanishing_bits();
    let mut expect = FieldBits::new();
    expect.set_onebit(3, 1);
    expect.set_onebit(4, 1);
    expect.set_onebit(5, 1);
    expect.set_onebit(5, 2);
    assert_bits_eq(vanished, expect);
}

#[test]
fn iterate_bit_with_masking_matches_go() {
    let mut bits = FieldBits::new();
    bits.set_onebit(0, 4);
    bits.set_onebit(0, 3);
    bits.set_onebit(0, 2);
    bits.set_onebit(1, 2);
    bits.set_onebit(4, 4);
    bits.set_onebit(4, 3);
    bits.set_onebit(4, 2);
    bits.set_onebit(5, 2);

    let mut seen = Vec::new();
    bits.iterate_bit_with_masking(|candidate| {
        let expanded = candidate.expand(bits);
        seen.push(expanded);
        expanded
    });

    assert_eq!(seen.len(), 2);

    let mut first = FieldBits::new();
    first.set_onebit(0, 4);
    first.set_onebit(0, 3);
    first.set_onebit(0, 2);
    first.set_onebit(1, 2);
    assert_bits_eq(seen[0], first);

    let mut second = FieldBits::new();
    second.set_onebit(4, 4);
    second.set_onebit(4, 3);
    second.set_onebit(4, 2);
    second.set_onebit(5, 2);
    assert_bits_eq(seen[1], second);
}
