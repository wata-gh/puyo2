mod common;

use image::Rgba;

use crate::common::{create_asset_fixture, output_strings, read_png, run_bin_with_env};

#[test]
fn dpuyo_renders_canvas_with_requested_size() {
    let asset_dir = create_asset_fixture("dpuyo");
    let out_path = asset_dir.join("canvas.png");

    let output = run_bin_with_env(
        "dpuyo",
        &["-width", "3", "-height", "4", out_path.to_str().unwrap()],
        Some("rg.\nbyp\n"),
        &[("PUYO2_CONFIG", asset_dir.to_str().unwrap())],
    );
    let (stdout, stderr) = output_strings(&output);

    assert!(output.status.success(), "stderr={stderr}");
    assert!(stdout.is_empty());
    assert!(stderr.is_empty());
    assert!(out_path.exists());

    let image = read_png(&out_path);
    assert_eq!(image.dimensions(), (32 * 3, 32 * 4));
    assert_eq!(*image.get_pixel(1, 1), Rgba([3, 5, 220, 255]));
    assert_eq!(*image.get_pixel(33, 1), Rgba([14, 5, 220, 255]));
    assert_eq!(*image.get_pixel(1, 33), Rgba([25, 5, 220, 255]));
    assert_eq!(*image.get_pixel(33, 33), Rgba([36, 5, 220, 255]));
    assert_eq!(*image.get_pixel(65, 33), Rgba([47, 5, 220, 255]));
}
