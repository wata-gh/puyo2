use std::{
    cell::{Cell, RefCell},
    rc::Rc,
};

use puyo2::{BitField, Color, PuyoSet, SearchCondition, SearchResult, SimulatePolicy};

fn assert_callback_contract(sr: &SearchResult) {
    assert!(sr.before_simulate.is_some(), "before_simulate must be set");
    assert!(sr.rensa_result.is_some(), "rensa_result must be set");
}

fn red_blue() -> PuyoSet {
    PuyoSet {
        axis: Color::Red,
        child: Color::Blue,
    }
}

#[test]
fn search_with_puyo_sets_v2_finds_known_solutions() {
    #[derive(Clone, Copy)]
    enum Expectation {
        Chains(usize),
        BlueEmpty,
    }

    let fixtures = [
        (
            "a62gacbagecb2ae2g3",
            vec![
                red_blue(),
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Blue,
                },
            ],
            Expectation::Chains(3),
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
            Expectation::Chains(3),
        ),
        (
            "a52ca2gbc2a2c2g2a2cgbga2g2b2a",
            vec![
                PuyoSet {
                    axis: Color::Blue,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Red,
                },
                PuyoSet {
                    axis: Color::Blue,
                    child: Color::Red,
                },
            ],
            Expectation::Chains(4),
        ),
        (
            "a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16",
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
                    axis: Color::Blue,
                    child: Color::Blue,
                },
                PuyoSet {
                    axis: Color::Blue,
                    child: Color::Blue,
                },
            ],
            Expectation::BlueEmpty,
        ),
    ];

    for (param, puyo_sets, expectation) in fixtures {
        let solved = Rc::new(Cell::new(false));
        let solved_handle = Rc::clone(&solved);
        let mut cond = SearchCondition::with_bit_field_and_puyo_sets(
            BitField::from_mattulwan(param),
            puyo_sets,
        );
        cond.last_callback = Some(Box::new(move |sr| {
            assert_callback_contract(sr);
            let result = sr.rensa_result.as_ref().unwrap();
            match expectation {
                Expectation::Chains(chains) if result.chains == chains => {
                    solved_handle.set(true);
                }
                Expectation::BlueEmpty
                    if result
                        .bit_field
                        .as_ref()
                        .unwrap()
                        .bits(Color::Blue)
                        .is_empty() =>
                {
                    solved_handle.set(true);
                }
                _ => {}
            }
        }));
        cond.search_with_puyo_sets_v2();
        assert!(solved.get(), "fixture {param} must be solvable");
    }
}

#[test]
fn search_with_puyo_sets_v2_fast_intermediate_callback_can_read_nth_result() {
    let saw_callback = Rc::new(Cell::new(false));
    let saw_callback_handle = Rc::clone(&saw_callback);
    let mut cond = SearchCondition::with_bit_field_and_puyo_sets(
        BitField::from_mattulwan("a78"),
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
    );
    cond.simulate_policy = SimulatePolicy::FastIntermediate;
    cond.each_hand_callback = Some(Box::new(move |sr| {
        assert_callback_contract(sr);
        saw_callback_handle.set(true);
        let result = sr.rensa_result.as_ref().unwrap();
        for nth in 1..=result.chains {
            let _ = result.nth_result(nth);
        }
        true
    }));

    cond.search_with_puyo_sets_v2();
    assert!(saw_callback.get(), "each_hand_callback must be called");
}

#[test]
fn search_with_puyo_sets_v2_fast_always_terminal_uses_detail_simulation() {
    let callback_count = Rc::new(Cell::new(0usize));
    let callback_count_handle = Rc::clone(&callback_count);
    let mismatches = Rc::new(RefCell::new(Vec::new()));
    let mismatches_handle = Rc::clone(&mismatches);
    let mut cond = SearchCondition::with_bit_field_and_puyo_sets(
        BitField::from_mattulwan("a54ea3eaebdece3bd2eb2dc3"),
        vec![PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        }],
    );
    cond.simulate_policy = SimulatePolicy::FastAlways;
    cond.last_callback = Some(Box::new(move |sr| {
        assert_callback_contract(sr);
        callback_count_handle.set(callback_count_handle.get() + 1);
        let before_simulate = sr.before_simulate.as_ref().unwrap();
        let got = sr.rensa_result.as_ref().unwrap();
        let want = before_simulate.clone_for_simulation().simulate_detail();
        if got.chains != want.chains
            || got.score != want.score
            || got.erased != want.erased
            || got.quick != want.quick
            || got.bit_field != want.bit_field
            || got.nth_results != want.nth_results
        {
            mismatches_handle
                .borrow_mut()
                .push((sr.position, got.clone(), want));
        }
    }));

    cond.search_with_puyo_sets_v2();

    assert!(
        callback_count.get() > 0,
        "last_callback must be called for terminal nodes"
    );
    assert!(
        mismatches.borrow().is_empty(),
        "terminal fast_always results must match simulate_detail: {:?}",
        mismatches.borrow()
    );
}
