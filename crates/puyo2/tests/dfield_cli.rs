mod common;

use image::Rgba;
use puyo2::{BitField, Color, FieldBits, ShapeBitField};

use crate::common::{create_asset_fixture, output_strings, read_png, run_bin_with_env};

fn env_pair(asset_dir: &std::path::Path) -> [(&str, &str); 1] {
    [("PUYO2_CONFIG", asset_dir.to_str().unwrap())]
}

#[test]
fn dfield_regular_export_writes_png_and_stdout_url() {
    let asset_dir = create_asset_fixture("dfield-regular");
    let mut field = BitField::new();
    field.set_color(Color::Red, 0, 1);
    let param = field.mattulwan_editor_param();
    let out_path = asset_dir.join("field.png");

    let output = run_bin_with_env(
        "dfield",
        &["-param", &param, "-out", out_path.to_str().unwrap()],
        None,
        &env_pair(&asset_dir),
    );
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());
    assert_eq!(stdout.trim(), field.mattulwan_editor_url());
    assert!(out_path.exists());
}

#[test]
fn dfield_stdin_multiple_lines_and_nobg_match_contract() {
    let asset_dir = create_asset_fixture("dfield-stdin");
    let output_dir = asset_dir.join("out");
    let input = "ba77\nca77\n";

    let output = run_bin_with_env(
        "dfield",
        &["-dir", output_dir.to_str().unwrap(), "-nobg"],
        Some(input),
        &env_pair(&asset_dir),
    );
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stderr.is_empty());
    assert_eq!(stdout.lines().count(), 2);
    assert!(output_dir.join("ba77.png").exists());
    assert!(output_dir.join("ca77.png").exists());

    let image = read_png(&output_dir.join("ba77.png"));
    assert_eq!(*image.get_pixel(1, 1), Rgba([0, 0, 0, 0]));
}

#[test]
fn dfield_simulate_and_hands_write_frame_directories() {
    let asset_dir = create_asset_fixture("dfield-frames");

    let mut chain_field = BitField::new();
    for &(x, y) in &[(0, 1), (0, 2), (1, 1), (1, 2)] {
        chain_field.set_color(Color::Red, x, y);
    }
    let simulate_dir = asset_dir.join("simulate");
    let simulate_output = run_bin_with_env(
        "dfield",
        &[
            "-param",
            &chain_field.mattulwan_editor_param(),
            "-simulate",
            "-dir",
            simulate_dir.to_str().unwrap(),
        ],
        None,
        &env_pair(&asset_dir),
    );
    assert!(
        simulate_output.status.success(),
        "{}",
        output_strings(&simulate_output).1
    );
    for frame in 1..=3 {
        assert!(simulate_dir.join(format!("{frame}.png")).exists());
    }

    let mut hand_field = BitField::new();
    hand_field.set_color(Color::Blue, 5, 1);
    let hands_dir = asset_dir.join("hands");
    let hands_output = run_bin_with_env(
        "dfield",
        &[
            "-param",
            &hand_field.mattulwan_editor_param(),
            "-hands",
            "rr10",
            "-dir",
            hands_dir.to_str().unwrap(),
        ],
        None,
        &env_pair(&asset_dir),
    );
    assert!(
        hands_output.status.success(),
        "{}",
        output_strings(&hands_output).1
    );
    for frame in 1..=3 {
        assert!(hands_dir.join(format!("{frame}.png")).exists());
    }
}

#[test]
fn dfield_shape_only_writes_shape_png() {
    let asset_dir = create_asset_fixture("dfield-shape");
    let out_path = asset_dir.join("shape.png");
    let mut shape_bit_field = ShapeBitField::new();
    let mut square = FieldBits::new();
    for &(x, y) in &[(0, 1), (0, 2), (1, 1), (1, 2)] {
        square.set_onebit(x, y);
    }
    shape_bit_field.add_shape(square);
    let param = shape_bit_field.field_string();

    let output = run_bin_with_env(
        "dfield",
        &[
            "-param",
            &param,
            "-shape-only",
            "-out",
            out_path.to_str().unwrap(),
        ],
        None,
        &env_pair(&asset_dir),
    );
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stdout.is_empty());
    assert!(stderr.is_empty());
    assert!(out_path.exists());

    let image = read_png(&out_path);
    assert_eq!(*image.get_pixel(33, 385), Rgba([7, 180, 40, 255]));
}
