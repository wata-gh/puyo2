use std::{
    cell::{Cell, RefCell},
    collections::HashSet,
    env, fs,
    path::PathBuf,
    rc::Rc,
    time::{Duration, Instant},
};

use chrono::Utc;
use serde::Serialize;

use puyo2::{
    BitField, Color, DedupMode, PuyoSet, SETUP_POSITIONS, SearchCondition, SearchStateKey,
    SimulatePolicy,
};

#[derive(Clone, Copy, Debug, Default)]
struct BenchMetrics {
    frames: u64,
    leaves: u64,
    nodes: u64,
    unique_states: u64,
}

impl BenchMetrics {
    fn add_assign(&mut self, other: Self) {
        self.frames += other.frames;
        self.leaves += other.leaves;
        self.nodes += other.nodes;
        self.unique_states += other.unique_states;
    }
}

#[derive(Debug, Serialize)]
struct BenchMetric {
    name: String,
    value: f64,
}

#[derive(Debug, Serialize)]
struct BenchCaseResult {
    name: String,
    iterations: u64,
    elapsed_ns: u128,
    metrics: Vec<BenchMetric>,
}

#[derive(Debug, Serialize)]
struct BenchReport {
    benchmark: String,
    generated_at: String,
    profile: String,
    min_sample_ms: u64,
    cases: Vec<BenchCaseResult>,
}

#[derive(Clone, Copy)]
struct SearchBenchCase {
    name: &'static str,
    param: &'static str,
    puyo_sets: &'static [PuyoSet],
}

fn main() {
    let args = parse_args(env::args().skip(1).collect());
    if args.help {
        print_usage();
        return;
    }

    let cases = benchmark_cases(Duration::from_millis(args.min_ms));
    let report = BenchReport {
        benchmark: "search_bench".to_string(),
        generated_at: Utc::now().to_rfc3339(),
        profile: "release".to_string(),
        min_sample_ms: args.min_ms,
        cases,
    };

    if let Some(parent) = args.out.parent() {
        fs::create_dir_all(parent).unwrap_or_else(|err| panic!("create_dir_all failed: {err}"));
    }
    let json = serde_json::to_string_pretty(&report)
        .unwrap_or_else(|err| panic!("json encode failed: {err}"));
    fs::write(&args.out, json).unwrap_or_else(|err| panic!("write failed: {err}"));

    println!("wrote {}", args.out.display());
    for case in &report.cases {
        println!(
            "{} iterations={} ns/op={:.1}",
            case.name,
            case.iterations,
            metric_value(&case.metrics, "ns/op")
        );
    }
}

struct Args {
    out: PathBuf,
    min_ms: u64,
    help: bool,
}

fn parse_args(args: Vec<String>) -> Args {
    let mut out = PathBuf::from("bench-results/search-bench.json");
    let mut min_ms = 250u64;
    let mut help = false;

    let mut iter = args.into_iter();
    while let Some(arg) = iter.next() {
        match arg.as_str() {
            "-h" | "--help" => {
                help = true;
            }
            "--out" => {
                out = PathBuf::from(iter.next().unwrap_or_else(|| {
                    panic!("--out requires a value");
                }));
            }
            "--min-ms" => {
                min_ms = iter
                    .next()
                    .unwrap_or_else(|| panic!("--min-ms requires a value"))
                    .parse::<u64>()
                    .unwrap_or_else(|err| panic!("invalid --min-ms: {err}"));
                if min_ms == 0 {
                    panic!("--min-ms must be greater than 0");
                }
            }
            _ => panic!("unknown argument: {arg}"),
        }
    }

    Args { out, min_ms, help }
}

fn print_usage() {
    println!(
        "Usage: cargo run -p puyo2 --example search_bench --release -- [--out PATH] [--min-ms N]"
    );
}

fn benchmark_cases(min_sample: Duration) -> Vec<BenchCaseResult> {
    let mut results = Vec::new();

    results.push(run_bench_case(
        "BenchmarkSearchPlacementForPos/EmptyField",
        min_sample,
        {
            let puyo_set = PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            };
            let field = BitField::new();
            let mut index = 0usize;
            move || {
                let pos = SETUP_POSITIONS[index % SETUP_POSITIONS.len()];
                index += 1;
                let frames = field
                    .search_placement_for_pos(&puyo_set, pos)
                    .map(|placement| placement.frames as u64)
                    .unwrap_or(0);
                BenchMetrics {
                    frames,
                    ..BenchMetrics::default()
                }
            }
        },
    ));
    results.push(run_bench_case(
        "BenchmarkSearchPlacementForPos/DenseField",
        min_sample,
        {
            let puyo_set = PuyoSet {
                axis: Color::Red,
                child: Color::Blue,
            };
            let field = BitField::from_mattulwan("a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16");
            let mut index = 0usize;
            move || {
                let pos = SETUP_POSITIONS[index % SETUP_POSITIONS.len()];
                index += 1;
                let frames = field
                    .search_placement_for_pos(&puyo_set, pos)
                    .map(|placement| placement.frames as u64)
                    .unwrap_or(0);
                BenchMetrics {
                    frames,
                    ..BenchMetrics::default()
                }
            }
        },
    ));

    results.push(run_bench_case(
        "BenchmarkSearchPositionV2",
        min_sample,
        || {
            let leaf_count = Rc::new(Cell::new(0usize));
            let leaf_count_callback = Rc::clone(&leaf_count);
            let mut condition = SearchCondition::with_bit_field_and_puyo_sets(
                BitField::from_mattulwan("a62gacbagecb2ae2g3"),
                vec![PuyoSet {
                    axis: Color::Red,
                    child: Color::Blue,
                }],
            );
            condition.last_callback = Some(Box::new(move |_| {
                leaf_count_callback.set(leaf_count_callback.get() + 1);
            }));
            condition.search_with_puyo_sets_v2();
            BenchMetrics {
                leaves: leaf_count.get() as u64,
                ..BenchMetrics::default()
            }
        },
    ));

    for (name, field, puyo_sets) in [
        (
            "BenchmarkSearchWithPuyoSetsV2/Depth2",
            "a62gacbagecb2ae2g3",
            vec![
                PuyoSet {
                    axis: Color::Red,
                    child: Color::Blue,
                },
                PuyoSet {
                    axis: Color::Green,
                    child: Color::Blue,
                },
            ],
        ),
        (
            "BenchmarkSearchWithPuyoSetsV2/Depth4",
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
        ),
    ] {
        results.push(run_bench_case(name, min_sample, move || {
            let leaf_count = Rc::new(Cell::new(0usize));
            let leaf_count_callback = Rc::clone(&leaf_count);
            let mut condition = SearchCondition::with_bit_field_and_puyo_sets(
                BitField::from_mattulwan(field),
                puyo_sets.clone(),
            );
            condition.last_callback = Some(Box::new(move |_| {
                leaf_count_callback.set(leaf_count_callback.get() + 1);
            }));
            condition.search_with_puyo_sets_v2();
            BenchMetrics {
                leaves: leaf_count.get() as u64,
                ..BenchMetrics::default()
            }
        }));
    }

    results.push(run_bench_case(
        "BenchmarkSearchWithPuyoSetsV2Pruned",
        min_sample,
        || {
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
            let max_depth = puyo_sets.len();
            let leaf_count = Rc::new(Cell::new(0usize));
            let leaf_count_callback = Rc::clone(&leaf_count);
            let mut condition = SearchCondition::with_bit_field_and_puyo_sets(
                BitField::from_mattulwan("a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16"),
                puyo_sets,
            );
            condition.each_hand_callback = Some(Box::new(move |search_result| {
                !(search_result.depth != max_depth
                    && search_result.rensa_result.as_ref().unwrap().chains != 0)
            }));
            condition.last_callback = Some(Box::new(move |_| {
                leaf_count_callback.set(leaf_count_callback.get() + 1);
            }));
            condition.search_with_puyo_sets_v2();
            BenchMetrics {
                leaves: leaf_count.get() as u64,
                ..BenchMetrics::default()
            }
        },
    ));

    const CASE_3_SAME: [PuyoSet; 3] = [
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
    const CASE_3_MIX: [PuyoSet; 3] = [
        PuyoSet {
            axis: Color::Red,
            child: Color::Blue,
        },
        PuyoSet {
            axis: Color::Green,
            child: Color::Blue,
        },
        PuyoSet {
            axis: Color::Red,
            child: Color::Yellow,
        },
    ];
    const CASE_4_SAME: [PuyoSet; 4] = [
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
        PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        },
    ];

    for bench_case in [
        SearchBenchCase {
            name: "3_same",
            param: "a78",
            puyo_sets: &CASE_3_SAME,
        },
        SearchBenchCase {
            name: "3_mix",
            param: "a78",
            puyo_sets: &CASE_3_MIX,
        },
        SearchBenchCase {
            name: "4_same",
            param: "a16ca5ga4cgca3g3a3g3a3g3cacg4ag5cg16",
            puyo_sets: &CASE_4_SAME,
        },
    ] {
        for (dedup, policy) in [
            (DedupMode::Off, SimulatePolicy::DetailAlways),
            (DedupMode::SamePairOrder, SimulatePolicy::DetailAlways),
            (DedupMode::State, SimulatePolicy::FastIntermediate),
        ] {
            let case_name = format!(
                "BenchmarkSearchWithPuyoSetsV2Modes/{}/{}/{}",
                bench_case.name, dedup, policy
            );
            results.push(run_bench_case(&case_name, min_sample, move || {
                let nodes = Rc::new(Cell::new(0usize));
                let nodes_callback = Rc::clone(&nodes);
                let unique = Rc::new(RefCell::new(HashSet::<SearchStateKey>::new()));
                let unique_callback = Rc::clone(&unique);
                let mut condition = SearchCondition::with_bit_field_and_puyo_sets(
                    BitField::from_mattulwan(bench_case.param),
                    bench_case.puyo_sets.to_vec(),
                );
                condition.dedup_mode = dedup;
                condition.simulate_policy = policy;
                condition.each_hand_callback = Some(Box::new(move |search_result| {
                    nodes_callback.set(nodes_callback.get() + 1);
                    let bit_field = search_result
                        .rensa_result
                        .as_ref()
                        .and_then(|result| result.bit_field.as_ref())
                        .unwrap();
                    unique_callback
                        .borrow_mut()
                        .insert(create_search_state_key(bit_field));
                    true
                }));
                condition.search_with_puyo_sets_v2();
                BenchMetrics {
                    nodes: nodes.get() as u64,
                    unique_states: unique.borrow().len() as u64,
                    ..BenchMetrics::default()
                }
            }));
        }
    }

    results
}

fn run_bench_case<F>(name: &str, min_sample: Duration, mut op: F) -> BenchCaseResult
where
    F: FnMut() -> BenchMetrics,
{
    let mut iterations = 1u64;
    let (elapsed, totals) = loop {
        let mut totals = BenchMetrics::default();
        let start = Instant::now();
        for _ in 0..iterations {
            totals.add_assign(op());
        }
        let elapsed = start.elapsed();
        if elapsed >= min_sample {
            break (elapsed, totals);
        }
        if elapsed.as_nanos() == 0 {
            iterations = iterations.saturating_mul(10);
            continue;
        }
        let factor = ((min_sample.as_nanos() / elapsed.as_nanos()) as u64)
            .saturating_add(1)
            .clamp(2, 10);
        iterations = iterations.saturating_mul(factor);
    };

    let mut metrics = vec![BenchMetric {
        name: "ns/op".to_string(),
        value: elapsed.as_secs_f64() * 1_000_000_000.0 / iterations as f64,
    }];
    if totals.frames > 0 {
        metrics.push(BenchMetric {
            name: "frames/op".to_string(),
            value: totals.frames as f64 / iterations as f64,
        });
    }
    if totals.leaves > 0 {
        metrics.push(BenchMetric {
            name: "leaves/op".to_string(),
            value: totals.leaves as f64 / iterations as f64,
        });
    }
    if totals.nodes > 0 {
        metrics.push(BenchMetric {
            name: "nodes/op".to_string(),
            value: totals.nodes as f64 / iterations as f64,
        });
    }
    if totals.unique_states > 0 {
        metrics.push(BenchMetric {
            name: "unique_states/op".to_string(),
            value: totals.unique_states as f64 / iterations as f64,
        });
    }

    BenchCaseResult {
        name: name.to_string(),
        iterations,
        elapsed_ns: elapsed.as_nanos(),
        metrics,
    }
}

fn create_search_state_key(bit_field: &BitField) -> SearchStateKey {
    SearchStateKey {
        m: bit_field.m,
        table_sig: color_table_signature(bit_field),
    }
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
        signature |= ((bit_field.table[color.idx()] as u32) & 0xf) << (index * 4);
    }
    signature
}

fn metric_value(metrics: &[BenchMetric], name: &str) -> f64 {
    metrics
        .iter()
        .find(|metric| metric.name == name)
        .map(|metric| metric.value)
        .unwrap_or(0.0)
}
