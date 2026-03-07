#[path = "../cli_common.rs"]
mod cli_common;

use std::{
    process,
    str::FromStr,
    sync::{
        Arc, Mutex,
        atomic::{AtomicUsize, Ordering},
        mpsc,
    },
    thread,
    time::Instant,
};

use chrono::Local;
use signal_hook::{consts::signal::SIGHUP, iterator::Signals};

use puyo2::{BitField, Color, DedupMode, Hand, SearchCondition, SearchResult, SimulatePolicy};

use crate::cli_common::{collect_args, parse_flag_value};

#[derive(Clone, Debug)]
struct Options {
    all_clear: bool,
    chains_eq: usize,
    chains_gt: usize,
    clear_color: Color,
    color: Color,
    n: usize,
    disable_chigiri: bool,
    stop_on_chain: bool,
    dedup_mode: DedupMode,
    simulate_policy: SimulatePolicy,
    hands: String,
}

impl Default for Options {
    fn default() -> Self {
        Self {
            all_clear: false,
            chains_eq: 0,
            chains_gt: 0,
            clear_color: Color::Empty,
            color: Color::Empty,
            n: 0,
            disable_chigiri: false,
            stop_on_chain: false,
            dedup_mode: DedupMode::Off,
            simulate_policy: SimulatePolicy::DetailAlways,
            hands: String::new(),
        }
    }
}

#[derive(Clone, Debug)]
struct Job {
    bit_field: BitField,
    hands_prefix: Vec<Hand>,
}

#[derive(Debug)]
enum WorkItem {
    Job(Job),
    Shutdown,
}

fn main() {
    let started_at = Instant::now();
    print_trap_banner();

    let program = std::env::args()
        .next()
        .unwrap_or_else(|| "nazo".to_string());
    let mut param = "a78".to_string();
    let mut chain_str = String::new();
    let mut clear_color = String::new();
    let mut color = String::new();
    let mut dedup_mode = DedupMode::Off.to_string();
    let mut simulate_policy = SimulatePolicy::DetailAlways.to_string();
    let mut options = Options::default();
    let mut show_help = false;

    let mut args = collect_args().into_iter();
    while let Some(arg) = args.next() {
        if arg == "-h" || arg == "--help" {
            show_help = true;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-param", "--param").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            param = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-hands", "--hands").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            options.hands = value;
            continue;
        }
        if let Some(value) = parse_flag_value(&arg, &mut args, "-chains", "--chains")
            .unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            chain_str = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-clear", "--clear").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            clear_color = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-color", "--color").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            color = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-dedup", "--dedup").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            dedup_mode = value;
            continue;
        }
        if let Some(value) = parse_flag_value(&arg, &mut args, "-simulate", "--simulate")
            .unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            simulate_policy = value;
            continue;
        }
        if let Some(value) = parse_flag_value(&arg, &mut args, "-n", "--n").unwrap_or_else(|err| {
            eprintln!("{err}");
            process::exit(1);
        }) {
            options.n = value.parse::<usize>().unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            });
            continue;
        }
        match arg.as_str() {
            "-allclear" | "--allclear" => options.all_clear = true,
            "-disablechigiri" | "--disablechigiri" => options.disable_chigiri = true,
            "-stop-on-chain" | "--stop-on-chain" => options.stop_on_chain = true,
            _ => {
                eprintln!("unknown argument: {arg}");
                process::exit(1);
            }
        }
    }

    if show_help {
        print_usage(&program);
        return;
    }

    if chain_str.contains('+') {
        let chains = chain_str.replace('+', "");
        match chains.parse::<usize>() {
            Ok(value) => options.chains_gt = value,
            Err(err) => {
                eprintln!("{chain_str} {err}");
                return;
            }
        }
    } else if !chain_str.is_empty() {
        match chain_str.parse::<usize>() {
            Ok(value) => options.chains_eq = value,
            Err(err) => {
                eprintln!("{chain_str} {err}");
                return;
            }
        }
    }

    options.clear_color = parse_clear_color_or_panic(&clear_color);
    options.color = parse_color_or_panic(&color);
    options.dedup_mode = match DedupMode::from_str(&dedup_mode) {
        Ok(mode) => mode,
        Err(err) => {
            eprintln!("{err}");
            return;
        }
    };
    options.simulate_policy = match SimulatePolicy::from_str(&simulate_policy) {
        Ok(policy) => policy,
        Err(err) => {
            eprintln!("{err}");
            return;
        }
    };

    run(&param, &options);
    eprintln!("elapsed: {}ms", started_at.elapsed().as_millis());
}

fn run(param: &str, options: &Options) {
    let puyo_sets = parse_puyo_sets_or_panic(&options.hands);
    let cpus = std::thread::available_parallelism().map_or(1, |count| count.get());
    let searched_counts = Arc::new(
        (0..if puyo_sets.len() == 1 { 1 } else { cpus + 1 })
            .map(|_| AtomicUsize::new(0))
            .collect::<Vec<_>>(),
    );
    let all_search_count = 22_f64.powi(puyo_sets.len() as i32);
    let _signals = install_progress_handler(Arc::clone(&searched_counts), all_search_count);

    println!("{}", format_options(options));
    println!("cpus: {cpus}");

    if puyo_sets.len() == 1 {
        let job = Job {
            bit_field: BitField::from_mattulwan(param),
            hands_prefix: Vec::new(),
        };
        thread::scope(|scope| {
            let options = options.clone();
            let counts = Arc::clone(&searched_counts);
            let puyo_sets = puyo_sets.clone();
            scope.spawn(move || {
                search_job(0, job, puyo_sets, options, counts);
            });
        });
        return;
    }

    let initial = BitField::from_mattulwan(param);
    print!("{initial}");

    thread::scope(|scope| {
        let (sender, receiver) = mpsc::channel::<WorkItem>();
        let receiver = Arc::new(Mutex::new(receiver));
        let remaining_puyo_sets = puyo_sets[1..].to_vec();

        for worker_idx in 0..cpus {
            let receiver = Arc::clone(&receiver);
            let counts = Arc::clone(&searched_counts);
            let options = options.clone();
            let remaining_puyo_sets = remaining_puyo_sets.clone();
            scope.spawn(move || {
                loop {
                    let item = {
                        let guard = receiver
                            .lock()
                            .unwrap_or_else(|poisoned| poisoned.into_inner());
                        guard.recv()
                    };
                    match item {
                        Ok(WorkItem::Job(job)) => search_job(
                            worker_idx + 1,
                            job,
                            remaining_puyo_sets.clone(),
                            options.clone(),
                            Arc::clone(&counts),
                        ),
                        Ok(WorkItem::Shutdown) | Err(_) => break,
                    }
                }
            });
        }

        let counts = Arc::clone(&searched_counts);
        let options = options.clone();
        let sender_for_root = sender.clone();
        let first = vec![puyo_sets[0]];
        let mut condition =
            SearchCondition::with_bit_field_and_puyo_sets(BitField::from_mattulwan(param), first);
        condition.disable_chigiri = options.disable_chigiri;
        condition.dedup_mode = options.dedup_mode;
        condition.simulate_policy = options.simulate_policy;
        condition.stop_on_chain = options.stop_on_chain;
        condition.last_callback = Some(Box::new(move |sr: &SearchResult| {
            counts[0].fetch_add(1, Ordering::Relaxed);
            let final_field = sr
                .rensa_result
                .as_ref()
                .and_then(|result| result.bit_field.as_ref())
                .cloned()
                .expect("root search result must have final bit field");
            sender_for_root
                .send(WorkItem::Job(Job {
                    bit_field: final_field,
                    hands_prefix: sr.hands.clone(),
                }))
                .expect("worker queue must be available");
        }));
        condition.search_with_puyo_sets_v2();

        for _ in 0..cpus {
            let _ = sender.send(WorkItem::Shutdown);
        }
    });
}

fn search_job(
    counter_index: usize,
    job: Job,
    puyo_sets: Vec<puyo2::PuyoSet>,
    options: Options,
    searched_counts: Arc<Vec<AtomicUsize>>,
) {
    let prefix = to_simple_hands_or_panic(&job.hands_prefix);
    let clear_color = options.clear_color;
    let color = options.color;
    let n = options.n;
    let mut condition = SearchCondition::with_bit_field_and_puyo_sets(job.bit_field, puyo_sets);
    condition.disable_chigiri = options.disable_chigiri;
    condition.dedup_mode = options.dedup_mode;
    condition.simulate_policy = options.simulate_policy;
    condition.stop_on_chain = options.stop_on_chain;
    condition.each_hand_callback = Some(Box::new(move |sr: &SearchResult| {
        searched_counts[counter_index].fetch_add(1, Ordering::Relaxed);
        let result = sr
            .rensa_result
            .as_ref()
            .expect("search result must have rensa_result");
        let before_simulate = sr
            .before_simulate
            .as_ref()
            .expect("search result must have before_simulate");
        let hands = format!("{prefix}{}", to_simple_hands_or_panic(&sr.hands));
        let line = format_result_line(before_simulate, &hands, result);

        if options.all_clear && result.bit_field.as_ref().expect("bit field").is_empty() {
            println!("{line}");
        }
        if clear_color != Color::Empty
            && result
                .bit_field
                .as_ref()
                .expect("bit field")
                .bits(clear_color)
                .is_empty()
        {
            println!("{line}");
        }
        if options.chains_eq > 0 && result.chains == options.chains_eq {
            println!("{line}");
        }
        if options.chains_gt > 0 && result.chains >= options.chains_gt {
            println!("{line}");
        }
        if n > 0 && color != Color::Empty && result.chains > 0 {
            for nth in 1..=result.chains {
                let erased_count = result
                    .nth_result(nth)
                    .map(|entry| {
                        entry
                            .erased_puyos
                            .iter()
                            .filter(|erased| erased.color == color)
                            .map(|erased| erased.connected)
                            .sum::<usize>()
                    })
                    .unwrap_or_default();
                if erased_count == n {
                    println!("{line}");
                }
            }
        }
        true
    }));
    condition.search_with_puyo_sets_v2();
}

fn format_options(options: &Options) -> String {
    format!(
        "&{{AllClear:{} ChainsEQ:{} ChainsGT:{} ClearColor:{} Color:{} N:{} DisableChigiri:{} StopOnChain:{} DedupMode:{} SimulatePolicy:{} Hands:{}}}",
        options.all_clear,
        options.chains_eq,
        options.chains_gt,
        options.clear_color as u8,
        options.color as u8,
        options.n,
        options.disable_chigiri,
        options.stop_on_chain,
        options.dedup_mode,
        options.simulate_policy,
        options.hands
    )
}

fn format_result_line(
    before_simulate: &BitField,
    hands: &str,
    result: &puyo2::RensaResult,
) -> String {
    let bit_field = result.bit_field.as_ref().expect("bit field");
    let nth_results = if result.nth_results.is_empty() {
        String::new()
    } else {
        result
            .nth_results
            .iter()
            .map(|nth| format!("{:p}", nth))
            .collect::<Vec<_>>()
            .join(" ")
    };
    format!(
        "{} {} &{{Chains:{} Chigiris:{} Score:{} RensaFrames:{} SetFrames:{} Erased:{} Quick:{} BitField:{:p} NthResults:[{}]}}",
        before_simulate.mattulwan_editor_url(),
        hands,
        result.chains,
        result.chigiris,
        result.score,
        result.rensa_frames,
        result.set_frames,
        result.erased,
        result.quick,
        bit_field,
        nth_results
    )
}

fn print_trap_banner() {
    println!("trap goroutine start. {}", process::id());
    println!("send SIGHUP to show progress.");
}

fn print_usage(program: &str) {
    eprintln!("Usage of {program}:");
    eprintln!("  -allclear");
    eprintln!("\tallclear");
    eprintln!("  -chains string");
    eprintln!("\tchains");
    eprintln!("  -clear string");
    eprintln!("\tclear color(r,g,b,y,p)");
    eprintln!("  -color string");
    eprintln!("\tcolor(r,g,b,y,p)");
    eprintln!("  -dedup string");
    eprintln!(
        "\tdedup mode(off,same_pair_order,state,state_mirror). same_pair_order is effective only with -stop-on-chain (default \"off\")"
    );
    eprintln!("  -disablechigiri");
    eprintln!("\tdisable chigiri");
    eprintln!("  -hands string");
    eprintln!("\thands");
    eprintln!("  -n int");
    eprintln!("\tpuyo count");
    eprintln!("  -param string");
    eprintln!("\tpuyofu (default \"a78\")");
    eprintln!("  -simulate string");
    eprintln!(
        "\tsimulate policy(detail_always,fast_intermediate,fast_always) (default \"detail_always\")"
    );
    eprintln!("  -stop-on-chain");
    eprintln!("\tstop branch when chain occurs");
}

fn parse_clear_color_or_panic(color: &str) -> Color {
    match color {
        "" => Color::Empty,
        "r" => Color::Red,
        "g" => Color::Green,
        "b" => Color::Blue,
        "y" => Color::Yellow,
        "p" => Color::Purple,
        "o" => Color::Ojama,
        _ => panic!("clear color must be one of r,g,b,y,p,o"),
    }
}

fn parse_color_or_panic(color: &str) -> Color {
    match color {
        "" => Color::Empty,
        "r" => Color::Red,
        "g" => Color::Green,
        "b" => Color::Blue,
        "y" => Color::Yellow,
        "p" => Color::Purple,
        _ => panic!("clear color must be one of r,g,b,y,p,o"),
    }
}

fn parse_puyo_sets_or_panic(hands: &str) -> Vec<puyo2::PuyoSet> {
    if hands.is_empty() {
        panic!("hands required");
    }
    let chars: Vec<char> = hands.chars().collect();
    if !chars.len().is_multiple_of(2) {
        panic!("hands length must be even");
    }
    chars
        .chunks_exact(2)
        .map(|chunk| puyo2::PuyoSet {
            axis: parse_hand_color_or_panic(chunk[0]),
            child: parse_hand_color_or_panic(chunk[1]),
        })
        .collect()
}

fn parse_hand_color_or_panic(ch: char) -> Color {
    Color::from_hand_char(ch).unwrap_or_else(|_| panic!("letter must be one of r,g,y,b,p"))
}

fn to_simple_hands_or_panic(hands: &[Hand]) -> String {
    puyo2::to_simple_hands(hands).unwrap_or_else(|err| panic!("{err}"))
}

fn install_progress_handler(
    searched_counts: Arc<Vec<AtomicUsize>>,
    all_search_count: f64,
) -> Option<thread::JoinHandle<()>> {
    let mut signals = Signals::new([SIGHUP]).ok()?;
    Some(thread::spawn(move || {
        for _ in signals.forever() {
            let total = searched_counts
                .iter()
                .map(|count| count.load(Ordering::Relaxed))
                .sum::<usize>();
            let now = Local::now().format("%Y-%m-%d %H:%M:%S");
            eprintln!(
                "[{}] {}/{:.6}({:.6}%)",
                now,
                total,
                all_search_count,
                total as f64 * 100.0 / all_search_count
            );
        }
    }))
}
