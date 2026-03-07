#[path = "../cli_common.rs"]
mod cli_common;

use std::{
    cell::RefCell,
    process,
    rc::Rc,
    str::FromStr,
    sync::{
        Arc, Mutex,
        atomic::{AtomicUsize, Ordering},
        mpsc,
    },
    thread,
};

use serde::Serialize;

use puyo2::{
    BitField, DedupMode, Hand, IPSNazoCondition, PuyoSet, SearchCondition, SearchResult,
    SimulatePolicy, evaluate_ips_nazo_condition, haipuyo_to_puyo_sets, parse_ips_nazo_url,
    to_simple_hands,
};

use crate::cli_common::{
    catch_unwind_silent, collect_args, parse_bool_value, parse_flag_value,
    read_single_non_empty_input_from_stdin,
};

#[derive(Clone, Debug, Default, Serialize)]
struct SolveConditionJson {
    q0: usize,
    q1: usize,
    q2: usize,
    text: String,
}

#[derive(Clone, Debug, Default, Serialize)]
struct SolveSolutionJson {
    hands: String,
    chains: usize,
    score: usize,
    clear: bool,
    #[serde(rename = "initialField")]
    initial_field: String,
    #[serde(rename = "finalField")]
    final_field: String,
}

#[derive(Clone, Debug, Default, Serialize)]
struct SolveOutputJson {
    input: String,
    #[serde(rename = "initialField")]
    initial_field: String,
    haipuyo: String,
    condition: SolveConditionJson,
    status: String,
    #[serde(skip_serializing_if = "String::is_empty")]
    error: String,
    searched: usize,
    matched: usize,
    solutions: Vec<SolveSolutionJson>,
}

#[derive(Clone, Debug, Default)]
struct SearchAccum {
    searched: usize,
    solutions: Vec<SolveSolutionJson>,
}

#[derive(Clone, Debug)]
struct SearchOptions {
    disable_chigiri: bool,
    dedup_mode: DedupMode,
    simulate_policy: SimulatePolicy,
    expand_equivalent_hands: bool,
}

#[derive(Clone, Debug)]
struct Job {
    root_order: usize,
    bit_field: BitField,
    hands_prefix: Vec<Hand>,
}

#[derive(Clone, Debug)]
struct SolutionRecord {
    root_order: usize,
    local_order: usize,
    solution: SolveSolutionJson,
}

#[derive(Clone, Debug, Default)]
struct WorkerOutput {
    searched: usize,
    solutions: Vec<SolutionRecord>,
}

#[derive(Debug)]
enum WorkItem {
    Job(Job),
    Shutdown,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum SearchMode {
    Normal,
    Expansion,
}

fn new_solution(
    hands: String,
    chains: usize,
    score: usize,
    initial_field: String,
    final_field: String,
    clear: bool,
) -> SolveSolutionJson {
    SolveSolutionJson {
        hands,
        chains,
        score,
        clear,
        initial_field,
        final_field,
    }
}

fn condition_json(condition: &IPSNazoCondition) -> SolveConditionJson {
    SolveConditionJson {
        q0: condition.q0,
        q1: condition.q1,
        q2: condition.q2,
        text: condition.text.clone(),
    }
}

fn resolve_parallel_jobs() -> usize {
    std::env::var("PUYO2_PNSOLVE_JOBS")
        .ok()
        .and_then(|value| value.parse::<usize>().ok())
        .filter(|value| *value > 0)
        .unwrap_or_else(|| std::thread::available_parallelism().map_or(1, |count| count.get()))
}

fn record_solution(
    records: &mut Vec<SolutionRecord>,
    root_order: usize,
    local_order: usize,
    prefix: &str,
    sr: &SearchResult,
    condition: &IPSNazoCondition,
) {
    let before_simulate = sr
        .before_simulate
        .as_ref()
        .expect("search result must have before_simulate");
    let (matched, _) = evaluate_ips_nazo_condition(before_simulate, condition);
    if !matched {
        return;
    }

    let result = sr
        .rensa_result
        .as_ref()
        .expect("search result must have rensa_result");
    let final_field = result.bit_field.as_ref().expect("bit field");
    records.push(SolutionRecord {
        root_order,
        local_order,
        solution: new_solution(
            format!(
                "{prefix}{}",
                to_simple_hands(&sr.hands).expect("hands must be serializable")
            ),
            result.chains,
            result.score,
            before_simulate.mattulwan_editor_param(),
            final_field.mattulwan_editor_param(),
            final_field.is_empty(),
        ),
    });
}

fn search_job_terminal(
    job: Job,
    puyo_sets: Vec<PuyoSet>,
    condition: IPSNazoCondition,
    options: SearchOptions,
) -> Result<WorkerOutput, String> {
    let prefix = to_simple_hands(&job.hands_prefix).expect("hands must be serializable");
    let output = Rc::new(RefCell::new(WorkerOutput::default()));
    let mut local_order = 0usize;
    let mut condition_search =
        SearchCondition::with_bit_field_and_puyo_sets(job.bit_field, puyo_sets);
    condition_search.disable_chigiri = options.disable_chigiri;
    condition_search.dedup_mode = options.dedup_mode;
    condition_search.simulate_policy = options.simulate_policy;
    condition_search.stop_on_chain = false;
    let output_callback = Rc::clone(&output);
    condition_search.last_callback = Some(Box::new(move |sr: &SearchResult| {
        let mut output = output_callback.borrow_mut();
        output.searched += 1;
        record_solution(
            &mut output.solutions,
            job.root_order,
            local_order,
            &prefix,
            sr,
            &condition,
        );
        local_order += 1;
    }));
    catch_unwind_silent(move || condition_search.search_with_puyo_sets_v2())?;
    Ok(output.borrow().clone())
}

fn search_job_expansion(
    job: Job,
    puyo_sets: Vec<PuyoSet>,
    condition: IPSNazoCondition,
    options: SearchOptions,
) -> Result<WorkerOutput, String> {
    let prefix = to_simple_hands(&job.hands_prefix).expect("hands must be serializable");
    let output = Rc::new(RefCell::new(WorkerOutput::default()));
    let mut local_order = 0usize;
    let mut condition_search =
        SearchCondition::with_bit_field_and_puyo_sets(job.bit_field, puyo_sets);
    condition_search.disable_chigiri = options.disable_chigiri;
    condition_search.dedup_mode = DedupMode::Off;
    condition_search.simulate_policy = options.simulate_policy;
    condition_search.stop_on_chain = true;

    let output_each = Rc::clone(&output);
    let condition_each = condition.clone();
    let prefix_each = prefix.clone();
    condition_search.each_hand_callback = Some(Box::new(move |sr: &SearchResult| {
        if sr
            .rensa_result
            .as_ref()
            .expect("search result must have rensa_result")
            .chains
            > 0
        {
            let mut output = output_each.borrow_mut();
            output.searched += 1;
            record_solution(
                &mut output.solutions,
                job.root_order,
                local_order,
                &prefix_each,
                sr,
                &condition_each,
            );
            local_order += 1;
        }
        true
    }));

    let output_last = Rc::clone(&output);
    condition_search.last_callback = Some(Box::new(move |sr: &SearchResult| {
        let mut output = output_last.borrow_mut();
        output.searched += 1;
        record_solution(
            &mut output.solutions,
            job.root_order,
            local_order,
            &prefix,
            sr,
            &condition,
        );
        local_order += 1;
    }));
    catch_unwind_silent(move || condition_search.search_with_puyo_sets_v2())?;
    Ok(output.borrow().clone())
}

fn search_job(
    job: Job,
    puyo_sets: Vec<PuyoSet>,
    condition: IPSNazoCondition,
    options: SearchOptions,
    mode: SearchMode,
) -> Result<WorkerOutput, String> {
    match mode {
        SearchMode::Normal => search_job_terminal(job, puyo_sets, condition, options),
        SearchMode::Expansion => search_job_expansion(job, puyo_sets, condition, options),
    }
}

fn run_search(
    initial: BitField,
    puyo_sets: Vec<PuyoSet>,
    condition: IPSNazoCondition,
    options: SearchOptions,
    jobs: usize,
) -> Result<SearchAccum, String> {
    let mode = if options.expand_equivalent_hands {
        SearchMode::Expansion
    } else {
        SearchMode::Normal
    };

    if puyo_sets.len() == 1 || jobs <= 1 {
        let output = search_job(
            Job {
                root_order: 0,
                bit_field: initial,
                hands_prefix: Vec::new(),
            },
            puyo_sets,
            condition,
            options,
            mode,
        )?;
        let mut records = output.solutions;
        records.sort_by_key(|record| (record.root_order, record.local_order));
        return Ok(SearchAccum {
            searched: output.searched,
            solutions: records.into_iter().map(|record| record.solution).collect(),
        });
    }

    thread::scope(|scope| {
        let (sender, receiver) = mpsc::channel::<WorkItem>();
        let receiver = Arc::new(Mutex::new(receiver));
        let searched = Arc::new(AtomicUsize::new(0));
        let mut handles = Vec::with_capacity(jobs);
        let remaining = puyo_sets[1..].to_vec();

        for _ in 0..jobs {
            let receiver = Arc::clone(&receiver);
            let condition = condition.clone();
            let options = options.clone();
            let searched = Arc::clone(&searched);
            let remaining = remaining.clone();
            handles.push(scope.spawn(move || -> Result<WorkerOutput, String> {
                let mut worker_output = WorkerOutput::default();
                loop {
                    let item = {
                        let guard = receiver
                            .lock()
                            .unwrap_or_else(|poisoned| poisoned.into_inner());
                        guard.recv()
                    };
                    match item {
                        Ok(WorkItem::Job(job)) => {
                            let output = search_job(
                                job,
                                remaining.clone(),
                                condition.clone(),
                                options.clone(),
                                mode,
                            )?;
                            searched.fetch_add(output.searched, Ordering::Relaxed);
                            worker_output.searched += output.searched;
                            worker_output.solutions.extend(output.solutions);
                        }
                        Ok(WorkItem::Shutdown) | Err(_) => break,
                    }
                }
                Ok(worker_output)
            }));
        }

        let first = vec![puyo_sets[0]];
        let root_records = Rc::new(RefCell::new(Vec::<SolutionRecord>::new()));
        let root_terminal_count = Rc::new(RefCell::new(0usize));
        let root_order = Rc::new(RefCell::new(0usize));
        let sender_for_root = sender.clone();
        let mut root_condition = SearchCondition::with_bit_field_and_puyo_sets(initial, first);
        root_condition.disable_chigiri = options.disable_chigiri;
        root_condition.dedup_mode = match mode {
            SearchMode::Normal => options.dedup_mode,
            SearchMode::Expansion => DedupMode::Off,
        };
        root_condition.simulate_policy = options.simulate_policy;
        root_condition.stop_on_chain = mode == SearchMode::Expansion;

        if mode == SearchMode::Expansion {
            let root_records = Rc::clone(&root_records);
            let root_terminal_count = Rc::clone(&root_terminal_count);
            let root_order_each = Rc::clone(&root_order);
            let condition_each = condition.clone();
            root_condition.each_hand_callback = Some(Box::new(move |sr: &SearchResult| {
                if sr
                    .rensa_result
                    .as_ref()
                    .expect("search result must have rensa_result")
                    .chains
                    > 0
                {
                    *root_terminal_count.borrow_mut() += 1;
                    let current_root_order = {
                        let mut order = root_order_each.borrow_mut();
                        let current = *order;
                        *order += 1;
                        current
                    };
                    record_solution(
                        &mut root_records.borrow_mut(),
                        current_root_order,
                        0,
                        "",
                        sr,
                        &condition_each,
                    );
                }
                true
            }));
        }

        let root_order_last = Rc::clone(&root_order);
        root_condition.last_callback = Some(Box::new(move |sr: &SearchResult| {
            let current_root_order = {
                let mut order = root_order_last.borrow_mut();
                let current = *order;
                *order += 1;
                current
            };
            let final_field = sr
                .rensa_result
                .as_ref()
                .and_then(|result| result.bit_field.as_ref())
                .cloned()
                .expect("root search result must have final bit field");
            sender_for_root
                .send(WorkItem::Job(Job {
                    root_order: current_root_order,
                    bit_field: final_field,
                    hands_prefix: sr.hands.clone(),
                }))
                .expect("worker queue must be available");
        }));
        let root_result = catch_unwind_silent(move || root_condition.search_with_puyo_sets_v2());

        for _ in 0..jobs {
            let _ = sender.send(WorkItem::Shutdown);
        }
        drop(sender);

        let mut worker_outputs = Vec::with_capacity(jobs);
        let mut worker_error = None;
        for handle in handles {
            match handle.join() {
                Ok(Ok(output)) => worker_outputs.push(output),
                Ok(Err(err)) => {
                    if worker_error.is_none() {
                        worker_error = Some(err);
                    }
                }
                Err(_) => {
                    if worker_error.is_none() {
                        worker_error = Some("panic during search".to_string());
                    }
                }
            }
        }

        if let Err(err) = root_result {
            return Err(err);
        }
        if let Some(err) = worker_error {
            return Err(err);
        }

        let root_searched = *root_terminal_count.borrow();
        let mut records = root_records.take();
        records.extend(
            worker_outputs
                .into_iter()
                .flat_map(|output| output.solutions)
                .collect::<Vec<_>>(),
        );
        records.sort_by_key(|record| (record.root_order, record.local_order));
        Ok(SearchAccum {
            searched: root_searched + searched.load(Ordering::Relaxed),
            solutions: records.into_iter().map(|record| record.solution).collect(),
        })
    })
}

fn main() {
    let mut url_input = String::new();
    let mut param_input = String::new();
    let mut disable_chigiri = false;
    let mut pretty = true;
    let mut dedup_mode = DedupMode::SamePairOrder.to_string();
    let mut simulate_policy = SimulatePolicy::FastIntermediate.to_string();
    let mut expand_equivalent_hands = false;

    let mut args = collect_args().into_iter();
    while let Some(arg) = args.next() {
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-url", "--url").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            url_input = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-param", "--param").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            param_input = value;
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
        if let Some(value) = parse_flag_value(&arg, &mut args, "-pretty", "--pretty")
            .unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            pretty = parse_bool_value(&value, "-pretty").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            });
            continue;
        }
        if arg == "-disablechigiri" || arg == "--disablechigiri" {
            disable_chigiri = true;
            continue;
        }
        if arg == "-expand-equivalent-hands" || arg == "--expand-equivalent-hands" {
            expand_equivalent_hands = true;
            continue;
        }
        eprintln!("unknown argument: {arg}");
        process::exit(1);
    }

    let input = if !url_input.trim().is_empty() {
        url_input.trim().to_string()
    } else if !param_input.trim().is_empty() {
        param_input.trim().to_string()
    } else {
        match read_single_non_empty_input_from_stdin() {
            Ok(value) => value,
            Err(err) => {
                eprintln!("stdin error: {err}");
                process::exit(1);
            }
        }
    };

    let decoded = match parse_ips_nazo_url(&input) {
        Ok(decoded) => decoded,
        Err(err) => {
            eprintln!("parse error: {err}");
            process::exit(1);
        }
    };
    let parsed_dedup_mode = DedupMode::from_str(&dedup_mode).unwrap_or_else(|err| {
        eprintln!("{err}");
        process::exit(1);
    });
    let parsed_simulate_policy = SimulatePolicy::from_str(&simulate_policy).unwrap_or_else(|err| {
        eprintln!("{err}");
        process::exit(1);
    });

    let mut out = SolveOutputJson {
        input,
        initial_field: decoded.initial_field.clone(),
        haipuyo: decoded.haipuyo.clone(),
        condition: condition_json(&decoded.condition),
        status: "ok".to_string(),
        error: String::new(),
        searched: 0,
        matched: 0,
        solutions: Vec::new(),
    };

    let puyo_sets = haipuyo_to_puyo_sets(&decoded.haipuyo).unwrap_or_else(|err| {
        eprintln!("parse error: {err}");
        process::exit(1);
    });
    let jobs = resolve_parallel_jobs();
    let initial = match catch_unwind_silent(|| {
        if decoded.haipuyo.is_empty() {
            BitField::from_mattulwan(&decoded.initial_field)
        } else {
            BitField::from_mattulwan_and_haipuyo(&decoded.initial_field, &decoded.haipuyo)
                .unwrap_or_else(|err| panic!("{err}"))
        }
    }) {
        Ok(initial) => initial,
        Err(err) => {
            out.status = "search_failed".to_string();
            out.error = err;
            let json = if pretty {
                serde_json::to_string_pretty(&out)
            } else {
                serde_json::to_string(&out)
            }
            .unwrap_or_else(|encode_err| {
                eprintln!("json encode error: {encode_err}");
                process::exit(1);
            });
            println!("{json}");
            return;
        }
    };

    if puyo_sets.is_empty() {
        out.searched = 1;
        let (matched, _) = evaluate_ips_nazo_condition(&initial, &decoded.condition);
        if matched {
            let mut simulation = initial.clone_for_simulation();
            let result = simulation.simulate_detail();
            let final_field = result.bit_field.as_ref().unwrap().mattulwan_editor_param();
            let clear = result.bit_field.as_ref().unwrap().is_empty();
            out.solutions.push(new_solution(
                String::new(),
                result.chains,
                result.score,
                initial.mattulwan_editor_param(),
                final_field,
                clear,
            ));
            out.matched = 1;
        }
    } else {
        let search_result = run_search(
            initial.clone(),
            puyo_sets,
            decoded.condition.clone(),
            SearchOptions {
                disable_chigiri,
                dedup_mode: parsed_dedup_mode,
                simulate_policy: parsed_simulate_policy,
                expand_equivalent_hands,
            },
            jobs,
        );

        match search_result {
            Err(err) => {
                out.status = "search_failed".to_string();
                out.error = err;
                out.solutions.clear();
                out.matched = 0;
            }
            Ok(accum) => {
                out.searched = accum.searched;
                out.solutions = accum.solutions;
                out.matched = out.solutions.len();
            }
        }
    }

    if out.status == "ok" && out.matched == 0 {
        out.status = "no_solution".to_string();
    }

    let json_result = if pretty {
        serde_json::to_string_pretty(&out)
    } else {
        serde_json::to_string(&out)
    };
    match json_result {
        Ok(json) => println!("{json}"),
        Err(err) => {
            eprintln!("json encode error: {err}");
            process::exit(1);
        }
    }
}
