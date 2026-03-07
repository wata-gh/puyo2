use std::{
    env, io,
    path::{Path, PathBuf},
};

use image::{GenericImage, RgbaImage, imageops};

pub(crate) const TILE_SIZE: u32 = 32;

pub(crate) fn load_sprite(name: &str) -> io::Result<RgbaImage> {
    let path = resolve_asset_path(name)?;
    let image = image::open(&path).map_err(io::Error::other)?;
    Ok(image.into_rgba8())
}

pub(crate) fn new_canvas(width_tiles: u32, height_tiles: u32) -> RgbaImage {
    RgbaImage::new(width_tiles * TILE_SIZE, height_tiles * TILE_SIZE)
}

pub(crate) fn overlay_tile(
    source: &RgbaImage,
    tile_x: u32,
    tile_y: u32,
    target: &mut RgbaImage,
    dst_x: u32,
    dst_y: u32,
) {
    let tile = imageops::crop_imm(
        source,
        tile_x * TILE_SIZE,
        tile_y * TILE_SIZE,
        TILE_SIZE,
        TILE_SIZE,
    )
    .to_image();
    imageops::overlay(target, &tile, i64::from(dst_x), i64::from(dst_y));
}

#[allow(dead_code)]
pub(crate) fn copy_tile(
    source: &RgbaImage,
    tile_x: u32,
    tile_y: u32,
    target: &mut RgbaImage,
    dst_x: u32,
    dst_y: u32,
) {
    let tile = imageops::crop_imm(
        source,
        tile_x * TILE_SIZE,
        tile_y * TILE_SIZE,
        TILE_SIZE,
        TILE_SIZE,
    )
    .to_image();
    let _ = target.copy_from(&tile, dst_x, dst_y);
}

pub(crate) fn write_png(path: &Path, image: &RgbaImage) -> io::Result<()> {
    image.save(path).map_err(io::Error::other)
}

fn resolve_asset_path(name: &str) -> io::Result<PathBuf> {
    if let Some(config) = env::var_os("PUYO2_CONFIG") {
        return Ok(PathBuf::from(config).join(name));
    }
    Ok(env::current_dir()?.join("images").join(name))
}
