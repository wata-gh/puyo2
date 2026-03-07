mod common;

use image::{Rgba, RgbaImage, imageops};
use puyo2::{BitField, Color, FieldBits, Hand, PuyoSet, ShapeBitField};

use crate::common::{create_asset_fixture, read_png, with_puyo2_config};

fn normal_tile(tile_x: u8, tile_y: u8) -> Rgba<u8> {
    Rgba([
        tile_x.wrapping_mul(11).wrapping_add(3),
        tile_y.wrapping_mul(17).wrapping_add(5),
        220,
        255,
    ])
}

fn transparent_tile(tile_x: u8, tile_y: u8) -> Rgba<u8> {
    Rgba([
        tile_x.wrapping_mul(11).wrapping_add(3),
        tile_y.wrapping_mul(17).wrapping_add(5),
        120,
        96,
    ])
}

fn shape_tile(tile_x: u8, tile_y: u8) -> Rgba<u8> {
    Rgba([
        tile_x.wrapping_mul(13).wrapping_add(7),
        tile_y.wrapping_mul(19).wrapping_add(9),
        40,
        255,
    ])
}

fn composite(dst: Rgba<u8>, src: Rgba<u8>) -> Rgba<u8> {
    let mut background = RgbaImage::from_pixel(1, 1, dst);
    let foreground = RgbaImage::from_pixel(1, 1, src);
    imageops::overlay(&mut background, &foreground, 0, 0);
    *background.get_pixel(0, 0)
}

fn field_cell_pixel(x: usize, y: usize) -> (u32, u32) {
    (((x + 1) * 32 + 1) as u32, ((13 - y) * 32 + 1) as u32)
}

fn hand_field_pixel(x: usize, y: usize) -> (u32, u32) {
    let (px, py) = field_cell_pixel(x, y);
    (px, py + 3 * 32)
}

#[test]
fn bit_field_exports_match_png_contract() {
    let asset_dir = create_asset_fixture("bit-field-export");
    let _guard = with_puyo2_config(&asset_dir);

    let mut bit_field = BitField::new();
    bit_field.set_color(Color::Red, 0, 1);

    let export_path = asset_dir.join("field.png");
    bit_field.export_image(&export_path).unwrap();
    let image = read_png(&export_path);
    assert_eq!(image.dimensions(), (32 * 8, 32 * 14));
    assert_eq!(*image.get_pixel(33, 385), normal_tile(0, 0));
    assert_eq!(*image.get_pixel(65, 385), normal_tile(5, 1));

    let only_puyo_path = asset_dir.join("only.png");
    bit_field.export_only_puyo_image(&only_puyo_path).unwrap();
    let only = read_png(&only_puyo_path);
    assert_eq!(only.dimensions(), (32 * 8, 32 * 14));
    assert_eq!(*only.get_pixel(1, 1), Rgba([0, 0, 0, 0]));
    assert_eq!(*only.get_pixel(33, 385), normal_tile(0, 0));

    let mut transparent_bits = FieldBits::new();
    transparent_bits.set_onebit(0, 1);
    let transparent_path = asset_dir.join("transparent.png");
    bit_field
        .export_image_with_transparent(&transparent_path, &transparent_bits)
        .unwrap();
    let transparent = read_png(&transparent_path);
    assert_eq!(transparent.dimensions(), (32 * 8, 32 * 14));
    assert_eq!(
        *transparent.get_pixel(33, 385),
        composite(normal_tile(5, 1), transparent_tile(0, 0))
    );
}

#[test]
fn simulate_and_hands_exports_write_expected_frames() {
    let asset_dir = create_asset_fixture("simulate-export");
    let _guard = with_puyo2_config(&asset_dir);

    let mut bit_field = BitField::new();
    for &(x, y) in &[(0, 1), (0, 2), (1, 1), (1, 2)] {
        bit_field.set_color(Color::Red, x, y);
    }

    let simulate_dir = asset_dir.join("simulate");
    bit_field.export_simulate_image(&simulate_dir).unwrap();
    for frame in 1..=3 {
        assert!(simulate_dir.join(format!("{frame}.png")).exists());
    }
    let vanish_frame = read_png(&simulate_dir.join("2.png"));
    let (x, y) = field_cell_pixel(0, 1);
    assert_eq!(*vanish_frame.get_pixel(x, y), normal_tile(5, 5));

    let hands_dir = asset_dir.join("hands");
    let hands = vec![Hand {
        puyo_set: PuyoSet {
            axis: Color::Red,
            child: Color::Red,
        },
        position: [1, 0],
    }];
    BitField::new()
        .export_hands_simulate_image(&hands, &hands_dir)
        .unwrap();
    for frame in 1..=3 {
        assert!(hands_dir.join(format!("{frame}.png")).exists());
    }

    let hand_frame = read_png(&hands_dir.join("2.png"));
    assert_eq!(*hand_frame.get_pixel(65, 33), normal_tile(0, 0));

    let placed_frame = read_png(&hands_dir.join("3.png"));
    let (x, y) = hand_field_pixel(1, 1);
    assert_eq!(*placed_frame.get_pixel(x, y), normal_tile(0, 1));
}

#[test]
fn shape_exports_use_shape_sprite_sheet() {
    let asset_dir = create_asset_fixture("shape-export");
    let _guard = with_puyo2_config(&asset_dir);

    let mut square = FieldBits::new();
    for &(x, y) in &[(0, 1), (0, 2), (1, 1), (1, 2)] {
        square.set_onebit(x, y);
    }

    let mut shape_bit_field = ShapeBitField::new();
    shape_bit_field.add_shape(square);

    let export_path = asset_dir.join("shape.png");
    shape_bit_field.export_image(&export_path).unwrap();
    let image = read_png(&export_path);
    assert_eq!(image.dimensions(), (32 * 8, 32 * 14));
    let (x, y) = field_cell_pixel(0, 1);
    assert_eq!(*image.get_pixel(x, y), shape_tile(0, 9));

    let chain_result = shape_bit_field.simulate();
    assert_eq!(chain_result.chains, 1);

    let chain_path = asset_dir.join("chain.png");
    shape_bit_field.export_chain_image(&chain_path).unwrap();
    let chain = read_png(&chain_path);
    assert_eq!(*chain.get_pixel(x, y), shape_tile(0, 9));

    let mut single_shape_bit_field = ShapeBitField::new();
    single_shape_bit_field.add_shape(square);
    let one_shape_path = asset_dir.join("shape-0.png");
    single_shape_bit_field
        .export_shape_image(0, &one_shape_path)
        .unwrap();
    let one_shape = read_png(&one_shape_path);
    assert_eq!(*one_shape.get_pixel(x, y), shape_tile(0, 9));
}
