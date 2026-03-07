use std::{
    io,
    path::{Path, PathBuf},
};

use image::RgbaImage;

use crate::{
    BitField, Color, FieldBits, Hand,
    render::{TILE_SIZE, copy_tile, load_sprite, new_canvas, overlay_tile, write_png},
};

impl BitField {
    pub fn export_image<P: AsRef<Path>>(&self, name: P) -> io::Result<()> {
        let puyo = load_sprite("puyos.png")?;
        let out = self.draw_field_and_puyo(&puyo);
        write_png(name.as_ref(), &out)
    }

    pub fn export_simulate_image<P: AsRef<Path>>(&self, path: P) -> io::Result<()> {
        let path = path.as_ref();
        std::fs::create_dir_all(path)?;
        let mut bit_field = self.clone();
        let mut index = 1usize;
        bit_field.export_image(path.join(format!("{index}.png")))?;
        loop {
            let vanished = bit_field.find_vanishing_bits();
            if vanished.is_empty() {
                return Ok(());
            }
            index += 1;
            bit_field.export_image_with_vanish(path.join(format!("{index}.png")), &vanished)?;
            bit_field.drop_vanished(vanished);
            index += 1;
            bit_field.export_image(path.join(format!("{index}.png")))?;
        }
    }

    pub fn export_hands_simulate_image<P: AsRef<Path>>(
        &self,
        hands: &[Hand],
        path: P,
    ) -> io::Result<()> {
        let path = path.as_ref();
        std::fs::create_dir_all(path)?;
        let mut bit_field = self.clone();
        let mut index = 1usize;
        bit_field.export_hand_image(path.join(format!("{index}.png")), None)?;
        let before_drop = bit_field.mattulwan_editor_param();
        bit_field.drop_vanished(bit_field.bits(Color::Empty).mask_field12());
        if before_drop != bit_field.mattulwan_editor_param() {
            index += 1;
            bit_field.export_hand_image(path.join(format!("{index}.png")), None)?;
        }
        for hand in hands {
            index += 1;
            bit_field.export_hand_image(path.join(format!("{index}.png")), Some(hand))?;
            let (placed, _) = bit_field.place_puyo(hand.puyo_set, hand.position);
            if !placed {
                return Err(io::Error::other("can not place puyo."));
            }
            index += 1;
            bit_field.export_hand_image(path.join(format!("{index}.png")), None)?;
            loop {
                let vanished = bit_field.find_vanishing_bits();
                if vanished.is_empty() {
                    break;
                }
                index += 1;
                bit_field.export_hand_image_with_vanish(
                    path.join(format!("{index}.png")),
                    &vanished,
                    None,
                )?;
                bit_field.drop_vanished(vanished);
                index += 1;
                bit_field.export_hand_image(path.join(format!("{index}.png")), None)?;
            }
        }
        Ok(())
    }

    pub fn export_only_puyo_image<P: AsRef<Path>>(&self, name: P) -> io::Result<()> {
        let puyo = load_sprite("puyos.png")?;
        let mut out = new_canvas(8, 14);
        for y in (1..=13).rev() {
            for x in 0..6 {
                let color = self.color(x, y);
                if color == Color::Empty {
                    continue;
                }
                if color == Color::Ojama {
                    overlay_tile(
                        &puyo,
                        5,
                        2,
                        &mut out,
                        ((x + 1) as u32) * TILE_SIZE,
                        ((13 - y) as u32) * TILE_SIZE,
                    );
                } else {
                    self.draw_puyo(color, x, y, &puyo, &mut out);
                }
            }
        }
        write_png(name.as_ref(), &out)
    }

    pub fn export_image_with_transparent<P: AsRef<Path>>(
        &self,
        name: P,
        trans: &FieldBits,
    ) -> io::Result<()> {
        let puyo = load_sprite("puyos.png")?;
        let transparent = load_sprite("puyos_transparent.png")?;
        let mut out = new_canvas(8, 14);
        self.draw_field(&puyo, &mut out);

        for y in (1..=13).rev() {
            for x in 0..6 {
                let color = self.color(x, y);
                if color == Color::Empty {
                    continue;
                }
                let source = if trans.onebit(x, y) == 0 {
                    &puyo
                } else {
                    &transparent
                };
                if color == Color::Ojama {
                    overlay_tile(
                        source,
                        5,
                        2,
                        &mut out,
                        ((x + 1) as u32) * TILE_SIZE,
                        ((13 - y) as u32) * TILE_SIZE,
                    );
                } else {
                    self.draw_puyo(color, x, y, source, &mut out);
                }
            }
        }

        write_png(name.as_ref(), &out)
    }

    fn export_image_with_vanish<P: AsRef<Path>>(
        &self,
        name: P,
        vanish: &FieldBits,
    ) -> io::Result<()> {
        let puyo = load_sprite("puyos.png")?;
        let out = self.draw_field_and_puyo_with_vanish(&puyo, vanish);
        write_png(name.as_ref(), &out)
    }

    fn export_hand_image<P: AsRef<Path>>(&self, name: P, hand: Option<&Hand>) -> io::Result<()> {
        let puyo = load_sprite("puyos.png")?;
        let mut out = new_canvas(8, 17);
        let field_and_puyo = self.draw_field_and_puyo(&puyo);
        if let Some(hand) = hand {
            self.draw_hand(hand, &puyo, &mut out);
        }
        image::imageops::overlay(&mut out, &field_and_puyo, 0, i64::from(3 * TILE_SIZE));
        write_png(name.as_ref(), &out)
    }

    fn export_hand_image_with_vanish<P: AsRef<Path>>(
        &self,
        name: P,
        vanish: &FieldBits,
        hand: Option<&Hand>,
    ) -> io::Result<()> {
        let puyo = load_sprite("puyos.png")?;
        let field_and_puyo = self.draw_field_and_puyo_with_vanish(&puyo, vanish);
        let mut out = new_canvas(8, 17);
        if let Some(hand) = hand {
            self.draw_hand(hand, &puyo, &mut out);
        }
        image::imageops::overlay(&mut out, &field_and_puyo, 0, i64::from(3 * TILE_SIZE));
        write_png(name.as_ref(), &out)
    }

    fn draw_hand(&self, hand: &Hand, puyo: &RgbaImage, out: &mut RgbaImage) {
        let axis_x = self.puyo_image_pos_x(hand.puyo_set.axis);
        overlay_tile(
            puyo,
            axis_x,
            0,
            out,
            ((hand.position[0] + 1) as u32) * TILE_SIZE,
            TILE_SIZE,
        );

        let (xoffset, yoffset) = match hand.position[1] {
            0 => (0i32, -32),
            1 => (32, 0),
            2 => (0, 32),
            3 => (-32, 0),
            _ => (0, -32),
        };
        let child_x = ((hand.position[0] + 1) as i32 * TILE_SIZE as i32 + xoffset) as u32;
        let child_y = (TILE_SIZE as i32 + yoffset) as u32;
        let child_tile_x = self.puyo_image_pos_x(hand.puyo_set.child);
        overlay_tile(puyo, child_tile_x, 0, out, child_x, child_y);
    }

    fn draw_field(&self, puyo: &RgbaImage, out: &mut RgbaImage) {
        for y in (0..=13).rev() {
            copy_tile(puyo, 5, 0, out, 0, (13 - y) * TILE_SIZE);
            copy_tile(puyo, 5, 0, out, 7 * TILE_SIZE, (13 - y) * TILE_SIZE);
            for x in 0..8 {
                if (x == 0 || x == 7 || y == 0) && y != 13 {
                    copy_tile(puyo, 5, 0, out, x * TILE_SIZE, 13 * TILE_SIZE);
                } else {
                    copy_tile(puyo, 5, 1, out, x * TILE_SIZE, (13 - y) * TILE_SIZE);
                }
            }
        }
    }

    fn draw_field_and_puyo(&self, puyo: &RgbaImage) -> RgbaImage {
        let mut out = new_canvas(8, 14);
        self.draw_field(puyo, &mut out);

        for y in (1..=13).rev() {
            for x in 0..6 {
                let color = self.color(x, y);
                if color == Color::Empty {
                    continue;
                }
                if color == Color::Ojama {
                    overlay_tile(
                        puyo,
                        5,
                        2,
                        &mut out,
                        ((x + 1) as u32) * TILE_SIZE,
                        ((13 - y) as u32) * TILE_SIZE,
                    );
                } else {
                    self.draw_puyo(color, x, y, puyo, &mut out);
                }
            }
        }
        out
    }

    fn draw_field_and_puyo_with_vanish(&self, puyo: &RgbaImage, vanish: &FieldBits) -> RgbaImage {
        let mut out = new_canvas(8, 14);
        self.draw_field(puyo, &mut out);

        for y in (1..=13).rev() {
            for x in 0..6 {
                let color = self.color(x, y);
                if color == Color::Empty {
                    continue;
                }
                if vanish.onebit(x, y) == 0 {
                    if color == Color::Ojama {
                        overlay_tile(
                            puyo,
                            5,
                            2,
                            &mut out,
                            ((x + 1) as u32) * TILE_SIZE,
                            ((13 - y) as u32) * TILE_SIZE,
                        );
                    } else {
                        self.draw_puyo(color, x, y, puyo, &mut out);
                    }
                } else {
                    overlay_tile(
                        puyo,
                        5,
                        vanish_tile_y(color),
                        &mut out,
                        ((x + 1) as u32) * TILE_SIZE,
                        ((13 - y) as u32) * TILE_SIZE,
                    );
                }
            }
        }

        out
    }

    fn puyo_image_pos_x(&self, color: Color) -> u32 {
        match color {
            Color::Red => 0,
            Color::Green => 1,
            Color::Blue => 2,
            Color::Yellow => 3,
            Color::Purple => 4,
            _ => 5,
        }
    }

    fn draw_puyo(&self, color: Color, x: usize, y: usize, puyo: &RgbaImage, out: &mut RgbaImage) {
        let up = self.color(x, y + 1) == color;
        let down = y > 0 && self.color(x, y - 1) == color;
        let left = x > 0 && self.color(x - 1, y) == color;
        let right = x < 5 && self.color(x + 1, y) == color;

        overlay_tile(
            puyo,
            self.puyo_image_pos_x(color),
            puyo_tile_y(up, down, left, right),
            out,
            ((x + 1) as u32) * TILE_SIZE,
            ((13 - y) as u32) * TILE_SIZE,
        );
    }
}

fn puyo_tile_y(up: bool, down: bool, left: bool, right: bool) -> u32 {
    match (up, down, left, right) {
        (false, false, false, false) => 0,
        (true, false, false, false) => 1,
        (false, true, false, false) => 2,
        (true, true, false, false) => 3,
        (false, false, true, false) => 4,
        (true, false, true, false) => 5,
        (false, true, true, false) => 6,
        (true, true, true, false) => 7,
        (false, false, false, true) => 8,
        (true, false, false, true) => 9,
        (false, true, false, true) => 10,
        (true, true, false, true) => 11,
        (false, false, true, true) => 12,
        (true, false, true, true) => 13,
        (false, true, true, true) => 14,
        (true, true, true, true) => 15,
    }
}

fn vanish_tile_y(color: Color) -> u32 {
    match color {
        Color::Red => 5,
        Color::Green => 6,
        Color::Blue => 7,
        Color::Yellow => 8,
        Color::Purple => 9,
        _ => 10,
    }
}

#[allow(dead_code)]
fn frame_path(dir: &Path, index: usize) -> PathBuf {
    dir.join(format!("{index}.png"))
}
