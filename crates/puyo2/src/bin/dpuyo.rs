#[path = "../cli_common.rs"]
mod cli_common;
#[path = "../render.rs"]
mod render;

use std::{io::Read, process};

use image::RgbaImage;

use crate::{
    cli_common::{collect_args, parse_flag_value},
    render::{TILE_SIZE, load_sprite, new_canvas, overlay_tile, write_png},
};

fn tile_x_for_puyo(ch: char) -> Option<u32> {
    match ch {
        'r' => Some(0),
        'g' => Some(1),
        'b' => Some(2),
        'y' => Some(3),
        'p' => Some(4),
        '.' => None,
        _ => None,
    }
}

fn place_puyo(canvas: &mut RgbaImage, puyo_image: &RgbaImage, puyo: char, x: usize, y: usize) {
    if let Some(tile_x) = tile_x_for_puyo(puyo) {
        overlay_tile(
            puyo_image,
            tile_x,
            0,
            canvas,
            x as u32 * TILE_SIZE,
            y as u32 * TILE_SIZE,
        );
    }
}

fn main() {
    let mut width = 10u32;
    let mut height = 10u32;
    let mut output_path: Option<String> = None;

    let mut args = collect_args().into_iter();
    while let Some(arg) = args.next() {
        if let Some(value) =
            parse_flag_value(&arg, &mut args, "-width", "--width").unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            width = value.parse::<u32>().unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            });
            continue;
        }
        if let Some(value) = parse_flag_value(&arg, &mut args, "-height", "--height")
            .unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            })
        {
            height = value.parse::<u32>().unwrap_or_else(|err| {
                eprintln!("{err}");
                process::exit(1);
            });
            continue;
        }
        output_path = Some(arg);
    }

    let output_path = output_path.unwrap_or_else(|| {
        eprintln!("missing output path");
        process::exit(1);
    });

    let puyo_image = load_sprite("puyos.png").unwrap_or_else(|err| {
        eprintln!("{err}");
        process::exit(1);
    });
    let mut canvas = new_canvas(width, height);
    let mut input = String::new();
    std::io::stdin()
        .read_to_string(&mut input)
        .unwrap_or_else(|err| {
            eprintln!("{err}");
            process::exit(1);
        });
    for (row, line) in input.lines().enumerate() {
        for (column, puyo) in line.chars().enumerate() {
            place_puyo(&mut canvas, &puyo_image, puyo, column, row);
        }
    }
    write_png(std::path::Path::new(&output_path), &canvas).unwrap_or_else(|err| {
        eprintln!("{err}");
        process::exit(1);
    });
}
