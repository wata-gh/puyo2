use puyo2::{BitField, Color, PuyoSet, parse_ips_nazo_url, parse_simple_hands, to_simple_hands};

fn fill_column_height(field: &mut BitField, x: usize, height: usize) {
    for y in 1..=height {
        field.set_color(Color::Red, x, y);
    }
}

#[test]
fn search_placement_for_pos_matches_go() {
    let field = BitField::new();
    let placement = field
        .search_placement_for_pos(
            &PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            },
            [0, 0],
        )
        .unwrap();
    assert_eq!(placement.frames, 54);

    let field = BitField::new();
    let placement = field
        .search_placement_for_pos(
            &PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            },
            [0, 2],
        )
        .unwrap();
    assert_eq!(placement.frames, 52);

    let field = BitField::new();
    let placement = field
        .search_placement_for_pos(
            &PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            },
            [0, 1],
        )
        .unwrap();
    assert_eq!(placement.frames, 54);

    let field = BitField::from_mattulwan(
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaa",
    );
    let placement = field
        .search_placement_for_pos(
            &PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            },
            [0, 1],
        )
        .unwrap();
    assert_eq!(placement.frames, 52 + 19);

    let field = BitField::from_mattulwan(
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaa",
    );
    let placement = field
        .search_placement_for_pos(
            &PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            },
            [0, 1],
        )
        .unwrap();
    assert_eq!(placement.frames, 52 + 19);
}

#[test]
fn child_can_be_on_row14_and_axis_cannot() {
    let mut field = BitField::new();
    for y in 1..=13 {
        field.set_color(Color::Green, 3, y);
    }

    let placement = field
        .search_placement_for_pos(
            &PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            },
            [2, 1],
        )
        .unwrap();
    assert_eq!(placement.axis_y, 1);
    assert_eq!(placement.child_y, 14);
    assert_eq!(placement.frames, 78);

    assert!(field.place_puyo_with_placement(&placement));
    assert_eq!(field.color(2, 1), Color::Red);
    assert_eq!(field.color(3, 14), Color::Empty);

    let mut blocked = BitField::new();
    for y in 1..=13 {
        blocked.set_color(Color::Green, 2, y);
    }
    assert!(
        blocked
            .search_placement_for_pos(
                &PuyoSet {
                    axis: Color::Red,
                    child: Color::Blue,
                },
                [2, 0],
            )
            .is_none()
    );
}

#[test]
fn consecutive_12_walls_mawashi_matches_go() {
    let mut field = BitField::new();
    fill_column_height(&mut field, 0, 5);
    fill_column_height(&mut field, 1, 11);
    fill_column_height(&mut field, 2, 10);
    fill_column_height(&mut field, 3, 12);
    fill_column_height(&mut field, 4, 12);
    fill_column_height(&mut field, 5, 5);

    let puyo_set = PuyoSet {
        axis: Color::Green,
        child: Color::Green,
    };
    assert!(field.search_placement_for_pos(&puyo_set, [5, 0]).is_some());
    assert_eq!(field.place_puyo(puyo_set, [5, 0]), (true, false));

    let mut blocked = BitField::new();
    fill_column_height(&mut blocked, 0, 5);
    fill_column_height(&mut blocked, 1, 10);
    fill_column_height(&mut blocked, 2, 10);
    fill_column_height(&mut blocked, 3, 12);
    fill_column_height(&mut blocked, 4, 12);
    fill_column_height(&mut blocked, 5, 5);

    assert!(
        blocked
            .search_placement_for_pos(&puyo_set, [5, 0])
            .is_none()
    );
    assert_eq!(blocked.place_puyo(puyo_set, [5, 0]), (false, false));
}

#[test]
fn place_puyo_wall_step6_sequence_is_placeable() {
    let decoded = parse_ips_nazo_url("qg0uswiugPAgPPAOjOSQySSSSSSSSS_q1C1u1q1u1u1__u09").unwrap();
    let mut field = BitField::from_mattulwan(&decoded.initial_field);
    let hands = parse_simple_hands("gb01gy10yb52gb43yb32yb53").unwrap();

    for (index, hand) in hands.iter().enumerate() {
        let placement = field.search_placement_for_pos(&hand.puyo_set, hand.position);
        assert!(
            placement.is_some(),
            "hand[{index}] should be placeable: hand={} heights={:?}",
            to_simple_hands(std::slice::from_ref(hand)).unwrap(),
            field.create_heights()
        );
        assert!(
            field.place_puyo(hand.puyo_set, hand.position).0,
            "hand[{index}] place failed: hand={} heights={:?}",
            to_simple_hands(std::slice::from_ref(hand)).unwrap(),
            field.create_heights()
        );
    }
}

#[test]
fn place_puyo_returns_chigiri_matches_go() {
    let mut field = BitField::from_mattulwan(
        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaa",
    );
    assert_eq!(
        field.place_puyo(
            PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            },
            [0, 1]
        ),
        (true, true)
    );
}
