#[path = "../cli_common.rs"]
mod cli_common;

use std::{fs, io::Read, process};

use puyo2::{BitField, Color, parse_simple_hands};

use crate::cli_common::{collect_args, matches_switch, parse_flag_value};

#[derive(Clone, Debug)]
struct Options {
    trans: String,
    out: String,
    dir: String,
    hands: String,
    no_bg: bool,
    simulate: bool,
    shape_only: bool,
    table: String,
}

impl Default for Options {
    fn default() -> Self {
        Self {
            trans: "a78".to_string(),
            out: String::new(),
            dir: String::new(),
            hands: String::new(),
            no_bg: false,
            simulate: false,
            shape_only: false,
            table: String::new(),
        }
    }
}

fn exp_hands(param: &str, hands_str: &str, path: &str) {
    let mut bit_field = BitField::from_mattulwan(param);
    let hands = parse_simple_hands(hands_str).unwrap_or_else(|err| panic!("{err}"));
    for hand in &hands {
        let mut converted = bit_field.table[hand.puyo_set.axis.idx()];
        if converted == Color::Empty {
            bit_field.table[hand.puyo_set.axis.idx()] = hand.puyo_set.axis;
        }
        converted = bit_field.table[hand.puyo_set.child.idx()];
        if converted == Color::Empty {
            bit_field.table[hand.puyo_set.child.idx()] = hand.puyo_set.child;
        }
    }
    if bit_field.table[Color::Purple.idx()] == Color::Purple {
        for color in [Color::Red, Color::Blue, Color::Yellow, Color::Green] {
            if bit_field.table[color.idx()] == Color::Empty {
                bit_field.table[Color::Purple.idx()] = color;
            }
        }
    }
    bit_field
        .export_hands_simulate_image(&hands, path)
        .unwrap_or_else(|err| panic!("{err}"));
}

fn exp_simulate(param: &str, path: &str) {
    let bit_field = BitField::from_mattulwan(param);
    bit_field
        .export_simulate_image(path)
        .unwrap_or_else(|err| panic!("{err}"));
}

fn exp(param: &str, trans: &str, out: &str, no_bg: bool) {
    let bit_field = BitField::from_mattulwan(param);
    let trans_field = BitField::from_mattulwan(trans).bits(Color::Red);
    if no_bg {
        bit_field
            .export_only_puyo_image(out)
            .unwrap_or_else(|err| panic!("{err}"));
    } else {
        bit_field
            .export_image_with_transparent(out, &trans_field)
            .unwrap_or_else(|err| panic!("{err}"));
    }
    println!("{}", bit_field.mattulwan_editor_url());
}

fn exp_shape(param: &str, out: &str) {
    let shape_bit_field = puyo2::ShapeBitField::from_field_string(param);
    shape_bit_field
        .export_image(out)
        .unwrap_or_else(|err| panic!("{err}"));
}

fn exp_shape_simulate(param: &str, out: &str) {
    let shape_bit_field = puyo2::ShapeBitField::from_field_string(param);
    shape_bit_field
        .export_chain_image(out)
        .unwrap_or_else(|err| panic!("{err}"));
}

fn run(param: &str, options: &Options) {
    if !options.dir.is_empty() {
        fs::create_dir_all(&options.dir).unwrap_or_else(|err| panic!("{err}"));
    }
    if options.simulate {
        if options.shape_only {
            exp_shape_simulate(param, &options.dir);
        } else {
            exp_simulate(param, &options.dir);
        }
    } else if !options.hands.is_empty() {
        exp_hands(param, &options.hands, &options.dir);
    } else {
        let mut path = if options.out.is_empty() {
            format!("{param}.png")
        } else {
            options.out.clone()
        };
        if !options.dir.is_empty() {
            path = format!("{}/{}", options.dir, path);
        }
        if options.shape_only {
            exp_shape(param, &path);
        } else {
            exp(param, &options.trans, &path, options.no_bg);
        }
    }
}

fn main() {
    let mut param = "a78".to_string();
    let mut options = Options::default();

    let mut args = collect_args().into_iter();
    while let Some(arg) = args.next() {
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
            parse_flag_value(&arg, &mut args, "-trans", "--trans").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            options.trans = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-out", "--out").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            options.out = value;
            continue;
        }
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-dir", "--dir").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            options.dir = value;
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
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-table", "--table").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            options.table = value;
            continue;
        }
        if matches_switch(&arg, "-nobg", "--nobg") {
            options.no_bg = true;
            continue;
        }
        if matches_switch(&arg, "-simulate", "--simulate") {
            options.simulate = true;
            continue;
        }
        if matches_switch(&arg, "-shape-only", "--shape-only") {
            options.shape_only = true;
            continue;
        }
        eprintln!("unknown argument: {arg}");
        process::exit(1);
    }

    if param != "a78" {
        run(&param, &options);
    } else {
        let mut input = String::new();
        std::io::stdin().read_to_string(&mut input).unwrap();
        for line in input.lines() {
            run(line, &options);
        }
    }
}
