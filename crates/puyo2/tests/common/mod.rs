#![allow(dead_code)]

use std::{
    env, fs,
    path::{Path, PathBuf},
    process::{Child, Command, Output, Stdio},
    sync::{Mutex, MutexGuard, OnceLock},
    time::{SystemTime, UNIX_EPOCH},
};

use image::{Rgba, RgbaImage};

pub fn run_bin(name: &str, args: &[&str], stdin: Option<&str>) -> Output {
    run_bin_with_env(name, args, stdin, &[])
}

pub fn bin_path(name: &str) -> PathBuf {
    PathBuf::from(
        env::var(format!("CARGO_BIN_EXE_{name}"))
            .unwrap_or_else(|_| panic!("missing binary path for {name}")),
    )
}

pub fn run_bin_with_env(
    name: &str,
    args: &[&str],
    stdin: Option<&str>,
    envs: &[(&str, &str)],
) -> Output {
    let exe = env::var(format!("CARGO_BIN_EXE_{name}"))
        .unwrap_or_else(|_| panic!("missing binary path for {name}"));
    let mut command = Command::new(exe);
    command.args(args);
    for (key, value) in envs {
        command.env(key, value);
    }
    if let Some(stdin) = stdin {
        command.stdin(Stdio::piped());
        command.stdout(Stdio::piped());
        command.stderr(Stdio::piped());
        let mut child = command.spawn().expect("failed to spawn binary");
        std::io::Write::write_all(child.stdin.as_mut().unwrap(), stdin.as_bytes())
            .expect("failed to write stdin");
        child.wait_with_output().expect("failed to wait on binary")
    } else {
        command.output().expect("failed to run binary")
    }
}

pub fn spawn_bin_with_env(name: &str, args: &[&str], envs: &[(&str, &str)]) -> Child {
    let mut command = Command::new(bin_path(name));
    command
        .args(args)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    for (key, value) in envs {
        command.env(key, value);
    }
    command.spawn().expect("failed to spawn binary")
}

pub fn output_strings(output: &Output) -> (String, String) {
    (
        String::from_utf8(output.stdout.clone()).expect("stdout must be utf-8"),
        String::from_utf8(output.stderr.clone()).expect("stderr must be utf-8"),
    )
}

pub fn write_temp_file(prefix: &str, content: &str) -> PathBuf {
    let nonce = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_nanos();
    let path = env::temp_dir().join(format!("puyo2-{prefix}-{nonce}.json"));
    fs::write(&path, content).expect("failed to write temp file");
    path
}

pub fn create_asset_fixture(prefix: &str) -> PathBuf {
    let nonce = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_nanos();
    let dir = env::temp_dir().join(format!("puyo2-assets-{prefix}-{nonce}"));
    fs::create_dir_all(&dir).expect("failed to create asset dir");
    write_sprite_sheet(&dir.join("puyos.png"), 20, 16, false);
    write_sprite_sheet(&dir.join("puyos_transparent.png"), 20, 16, true);
    write_shape_sprite_sheet(&dir.join("puyos_shape.png"), 20, 16);
    dir
}

pub fn read_png(path: &Path) -> RgbaImage {
    image::open(path)
        .unwrap_or_else(|err| panic!("failed to read png {path:?}: {err}"))
        .into_rgba8()
}

pub struct Puyo2ConfigGuard {
    _guard: MutexGuard<'static, ()>,
    previous: Option<std::ffi::OsString>,
}

impl Drop for Puyo2ConfigGuard {
    fn drop(&mut self) {
        unsafe {
            if let Some(previous) = &self.previous {
                env::set_var("PUYO2_CONFIG", previous);
            } else {
                env::remove_var("PUYO2_CONFIG");
            }
        }
    }
}

pub fn with_puyo2_config(path: &Path) -> Puyo2ConfigGuard {
    let guard = env_lock()
        .lock()
        .unwrap_or_else(|poisoned| poisoned.into_inner());
    let previous = env::var_os("PUYO2_CONFIG");
    unsafe {
        env::set_var("PUYO2_CONFIG", path);
    }
    Puyo2ConfigGuard {
        _guard: guard,
        previous,
    }
}

fn env_lock() -> &'static Mutex<()> {
    static LOCK: OnceLock<Mutex<()>> = OnceLock::new();
    LOCK.get_or_init(|| Mutex::new(()))
}

fn write_sprite_sheet(path: &Path, width_tiles: u32, height_tiles: u32, transparent: bool) {
    let mut image = RgbaImage::new(width_tiles * 32, height_tiles * 32);
    for tile_y in 0..height_tiles {
        for tile_x in 0..width_tiles {
            let color = if transparent {
                Rgba([
                    (tile_x as u8).wrapping_mul(11).wrapping_add(3),
                    (tile_y as u8).wrapping_mul(17).wrapping_add(5),
                    120,
                    96,
                ])
            } else {
                Rgba([
                    (tile_x as u8).wrapping_mul(11).wrapping_add(3),
                    (tile_y as u8).wrapping_mul(17).wrapping_add(5),
                    220,
                    255,
                ])
            };
            fill_tile(&mut image, tile_x, tile_y, color);
        }
    }
    image.save(path).expect("failed to write sprite sheet");
}

fn write_shape_sprite_sheet(path: &Path, width_tiles: u32, height_tiles: u32) {
    let mut image = RgbaImage::new(width_tiles * 32, height_tiles * 32);
    for tile_y in 0..height_tiles {
        for tile_x in 0..width_tiles {
            let color = Rgba([
                (tile_x as u8).wrapping_mul(13).wrapping_add(7),
                (tile_y as u8).wrapping_mul(19).wrapping_add(9),
                40,
                255,
            ]);
            fill_tile(&mut image, tile_x, tile_y, color);
        }
    }
    image
        .save(path)
        .expect("failed to write shape sprite sheet");
}

fn fill_tile(image: &mut RgbaImage, tile_x: u32, tile_y: u32, color: Rgba<u8>) {
    for y in tile_y * 32..(tile_y + 1) * 32 {
        for x in tile_x * 32..(tile_x + 1) * 32 {
            image.put_pixel(x, y, color);
        }
    }
}
