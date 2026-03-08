use std::{fmt, io, path::Path};

use serde::{Deserialize, Serialize};

use crate::{
    BitField, Color, FieldBits, ShapeRensaResult,
    drop_compact::compact_lane_u16,
    render::{TILE_SIZE, copy_tile, load_sprite, new_canvas, overlay_tile, write_png},
};

const SHAPE_LABELS: [char; 20] = [
    '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
    'k',
];

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct ShapeBitField {
    pub shapes: Vec<FieldBits>,
    pub original_shapes: Vec<FieldBits>,
    pub chain_ordered_shapes: Vec<Vec<FieldBits>>,
    pub key_string: Option<String>,
}

impl ShapeBitField {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn from_field_string(param: &str) -> Self {
        let mut shape_bit_field = Self::new();
        let max = param
            .chars()
            .filter_map(letter_to_num)
            .max()
            .unwrap_or_default();
        let mut shapes = vec![FieldBits::new(); max];

        for (index, ch) in param.chars().enumerate() {
            if ch == '.' {
                continue;
            }
            let n = letter_to_num(ch)
                .unwrap_or_else(|| panic!("unsupported shape label {ch:?} in field string"));
            let x = index % 6;
            let y = 13usize.saturating_sub(index / 6);
            shapes[n - 1].set_onebit(x, y);
        }

        for shape in shapes {
            shape_bit_field.add_shape(shape);
        }
        shape_bit_field
    }

    pub fn add_shape(&mut self, shape: FieldBits) {
        self.shapes.push(shape);
        self.original_shapes.push(shape);
        self.key_string = None;
    }

    pub fn drop(&mut self, field_bits: FieldBits) {
        let mut active = [(0usize, 0u32, 0u64, 0u16); 6];
        let mut active_len = 0usize;
        for x in 0..6 {
            let idx = x >> 2;
            let shift = ((x & 3) * 16) as u32;
            let vanished_lane = ((field_bits.m[idx] >> shift) & 0xffff) as u16;
            if vanished_lane == 0 {
                continue;
            }
            active[active_len] = (idx, shift, 0xffffu64 << shift, vanished_lane);
            active_len += 1;
        }
        for shape in &mut self.shapes {
            for &(idx, shift, lane_mask, vanished_lane) in &active[..active_len] {
                let lane = ((shape.m[idx] >> shift) & 0xffff) as u16;
                let compacted = compact_lane_u16(lane, vanished_lane) as u64;
                shape.m[idx] = (shape.m[idx] & !lane_mask) | (compacted << shift);
            }
        }
        self.key_string = None;
    }

    pub fn fill_chainable_color(&self) -> Option<BitField> {
        if self.shape_count() == 0 {
            return Some(BitField::new());
        }
        let mut targets = vec![Color::Empty; self.shape_count()];
        targets[0] = Color::Red;
        self.find_color(targets, 0)
    }

    pub fn find_vanishing_field_bits_num(&self) -> Vec<usize> {
        let mut vanishing = Vec::new();
        for (index, shape) in self.shapes.iter().enumerate() {
            let vanished = shape.mask_field12().find_vanishing_bits();
            if !vanished.is_empty() {
                vanishing.push(index);
            }
        }
        vanishing
    }

    pub fn is_empty(&self) -> bool {
        self.shapes.iter().all(FieldBits::is_empty)
    }

    pub fn original_overall_shape(&self) -> FieldBits {
        self.original_shapes
            .iter()
            .copied()
            .fold(FieldBits::new(), |acc, shape| acc.or(shape))
    }

    pub fn overall_shape(&self) -> FieldBits {
        self.shapes
            .iter()
            .copied()
            .fold(FieldBits::new(), |acc, shape| acc.or(shape))
    }

    pub fn shape_count(&self) -> usize {
        self.shapes.len()
    }

    pub fn shape_num(&self, x: usize, y: usize) -> isize {
        for (index, shape) in self.shapes.iter().enumerate() {
            if shape.onebit(x, y) > 0 {
                return index as isize;
            }
        }
        -1
    }

    pub fn simulate1(&mut self) -> Vec<usize> {
        let vanishing_nums = self.find_vanishing_field_bits_num();
        let mut vanished = FieldBits::new();
        for index in &vanishing_nums {
            vanished = vanished.or(self.shapes[*index].mask_field12().find_vanishing_bits());
        }
        if vanished.is_empty() {
            return Vec::new();
        }
        self.drop(vanished);

        let chain_shapes = vanishing_nums
            .iter()
            .map(|index| self.original_shapes[*index])
            .collect();
        self.chain_ordered_shapes.push(chain_shapes);
        vanishing_nums
    }

    pub fn simulate(&mut self) -> ShapeRensaResult {
        let mut result = ShapeRensaResult::new();
        loop {
            let vanished = self.simulate1();
            if vanished.is_empty() {
                break;
            }
            result.add_chain();
        }
        result.set_shape_bit_field(self.clone());
        result
    }

    pub fn insert_shape(&mut self, field_bits: FieldBits) {
        let overall = self.overall_shape();
        let and = overall.and(field_bits);
        for x in 0..6 {
            let col = and.col_bits(x);
            if col == 0 {
                continue;
            }
            let mut shift_target = field_bits;
            let shift_base = if x > 3 { (x - 4) * 16 } else { x * 16 };
            let start = 64usize.saturating_sub((col >> shift_base).leading_zeros() as usize);
            for y in start..16 {
                shift_target.set_onebit(x, y);
            }
            self.shift_col(field_bits, shift_target, x);
        }
        self.add_shape(field_bits);
    }

    pub fn expand_3_puyo_shapes(&self) -> Self {
        let mut cloned = self.clone();
        for index in 0..cloned.shapes.len() {
            if cloned.shapes[index].popcount() == 3 {
                let mut overall = cloned.overall_shape();
                overall.m[0] = !overall.m[0];
                overall.m[1] = !overall.m[1];
                overall = overall.mask_field13();
                let expanded = cloned.shapes[index].expand1(overall);
                cloned.shapes[index] = cloned.shapes[index].or(expanded);
            }
        }
        cloned.key_string = None;
        cloned
    }

    pub fn key_string(&mut self) -> &str {
        if self.key_string.is_none() {
            let mut output = String::new();
            for shape in &self.shapes {
                output.push('_');
                output.push_str(&format!("{:x}:{:x}", shape.m[0], shape.m[1]));
            }
            self.key_string = Some(output);
        }
        self.key_string.as_deref().unwrap()
    }

    pub fn field_string(&self) -> String {
        let mut buf = vec!['.'; 78];
        let overall = self.overall_shape();
        for y in (1..=13).rev() {
            for x in 0..6 {
                if overall.onebit(x, y) == 0 {
                    continue;
                }
                let idx = (13 - y) * 6 + x;
                for (shape_index, shape) in self.shapes.iter().enumerate() {
                    if shape.onebit(x, y) > 0 {
                        buf[idx] = num_to_letter(shape_index + 1);
                        break;
                    }
                }
            }
        }
        buf.into_iter().collect()
    }

    pub fn chain_ordered_field_string(&self) -> String {
        let mut ordered_shapes = Vec::new();
        for chain_shapes in &self.chain_ordered_shapes {
            if let Some(shape) = chain_shapes.first() {
                ordered_shapes.push(*shape);
            }
        }
        for (index, shape) in self.shapes.iter().enumerate() {
            if !shape.is_empty() {
                ordered_shapes.push(self.original_shapes[index]);
            }
        }

        let mut output = String::new();
        for y in (1..=13).rev() {
            for x in 0..6 {
                let mut empty = true;
                for (index, shape) in ordered_shapes.iter().enumerate() {
                    if shape.onebit(x, y) > 0 {
                        output.push(num_to_letter(index + 1));
                        empty = false;
                        break;
                    }
                }
                if empty {
                    output.push('.');
                }
            }
        }
        output
    }

    pub fn to_chain_shapes(&self) -> Vec<FieldBits> {
        let mut cloned = self.clone();
        let mut shapes = Vec::new();
        let vanishing_nums = cloned.find_vanishing_field_bits_num();
        if vanishing_nums.is_empty() {
            return shapes;
        }
        shapes.push(self.shapes[vanishing_nums[0]]);
        loop {
            cloned.simulate1();
            let vanishing_nums = cloned.find_vanishing_field_bits_num();
            if vanishing_nums.is_empty() {
                break;
            }
            shapes.push(cloned.shapes[vanishing_nums[0]]);
        }
        shapes
    }

    pub fn to_chain_shapes_u64_array(&self) -> Vec<[u64; 2]> {
        self.to_chain_shapes()
            .into_iter()
            .map(|shape| shape.to_int_array())
            .collect()
    }

    pub fn export_image<P: AsRef<Path>>(&self, name: P) -> io::Result<()> {
        let field = load_sprite("puyos.png")?;
        let puyo = load_sprite("puyos_shape.png")?;
        let mut out = new_canvas(8, 14);
        draw_field(&field, &mut out);

        for y in (1..=13).rev() {
            for x in 0..6 {
                for (index, shape) in self.shapes.iter().enumerate() {
                    if shape.onebit(x, y) > 0 {
                        draw_shape_puyo(shape, index as u32, x, y, &puyo, &mut out);
                    }
                }
            }
        }

        write_png(name.as_ref(), &out)
    }

    pub fn export_chain_image<P: AsRef<Path>>(&self, name: P) -> io::Result<()> {
        let field = load_sprite("puyos.png")?;
        let puyo = load_sprite("puyos_shape.png")?;
        let mut out = new_canvas(8, 14);
        draw_field(&field, &mut out);

        for y in (1..=13).rev() {
            for x in 0..6 {
                for (index, chain_shapes) in self.chain_ordered_shapes.iter().enumerate() {
                    for shape in chain_shapes {
                        if shape.onebit(x, y) > 0 {
                            draw_shape_puyo(shape, index as u32, x, y, &puyo, &mut out);
                        }
                    }
                }
            }
        }

        write_png(name.as_ref(), &out)
    }

    pub fn export_shape_image<P: AsRef<Path>>(&self, n: usize, name: P) -> io::Result<()> {
        let field = load_sprite("puyos.png")?;
        let puyo = load_sprite("puyos_shape.png")?;
        let mut out = new_canvas(8, 14);
        draw_field(&field, &mut out);

        let shape = self.shapes[n];
        for y in (1..=13).rev() {
            for x in 0..6 {
                if shape.onebit(x, y) > 0 {
                    draw_shape_puyo(&shape, n as u32, x, y, &puyo, &mut out);
                }
            }
        }

        write_png(name.as_ref(), &out)
    }

    fn find_possibilities(&self, colors: &[Color], num: usize) -> Vec<Color> {
        if colors[num] != Color::Empty {
            return vec![colors[num]];
        }
        let mut palette = vec![Color::Red, Color::Blue, Color::Yellow, Color::Green];
        for (index, shape) in self.shapes.iter().enumerate() {
            if index == num {
                continue;
            }
            let expanded = self.shapes[num].expand(*shape);
            if !expanded.is_empty() {
                let vanished = expanded.find_vanishing_bits();
                if !vanished.is_empty() {
                    palette.retain(|color| *color != colors[index]);
                }
            }
        }
        palette
    }

    fn find_color(&self, targets: Vec<Color>, idx: usize) -> Option<BitField> {
        if idx == self.shape_count() {
            let chain_shapes = self.to_chain_shapes_u64_array();
            let overall = self.overall_shape();
            let mut bit_field = BitField::new();
            for x in 0..6 {
                for y in 0..14 {
                    if overall.onebit(x, y) > 0 {
                        for (shape_index, shape) in self.shapes.iter().enumerate() {
                            if shape.onebit(x, y) > 0 {
                                bit_field.set_color(targets[shape_index], x, y);
                                break;
                            }
                        }
                    }
                }
            }

            let mut simulated = bit_field.clone();
            let result = simulated.simulate();
            if result.chains == self.shape_count() {
                let simulated_shapes = bit_field.to_chain_shapes_u64_array();
                let same_chain = simulated_shapes.len() == chain_shapes.len()
                    && simulated_shapes
                        .iter()
                        .zip(chain_shapes.iter())
                        .all(|(left, right)| left == right);
                if same_chain {
                    return Some(bit_field);
                }
            }
            return None;
        }

        for color in self.find_possibilities(&targets, idx) {
            let mut new_targets = targets.clone();
            new_targets[idx] = color;
            if let Some(bit_field) = self.find_color(new_targets, idx + 1) {
                return Some(bit_field);
            }
        }
        None
    }

    fn shift_col(&mut self, insert_shape: FieldBits, shift_target: FieldBits, col_num: usize) {
        for index in 0..self.shapes.len() {
            let mut shape = self.shapes[index];
            let and = shape.and(shift_target);
            let col_bits = and.col_bits(col_num);
            if col_bits == 0 {
                continue;
            }

            let shift_bits = insert_shape.col_bits(col_num).count_ones();
            let shift_col = col_bits << shift_bits;
            if col_num < 4 {
                let shift = FieldBits::with_matrix([shift_col, 0]);
                let mut untarget =
                    FieldBits::with_matrix([!(shift_target.m[0] | (1u64 << (col_num * 16))), 0]);
                untarget.m[0] = untarget.col_bits(col_num);
                untarget = untarget.and(shape);
                match col_num {
                    0 => shape.m[0] &= !0xffff,
                    1 => shape.m[0] &= !0xffff0000,
                    2 => shape.m[0] &= !0xffff00000000,
                    3 => shape.m[0] &= !0xffff000000000000,
                    _ => {}
                }
                shape = shape.or(shift);
                shape = shape.or(untarget);
            } else {
                let shift = FieldBits::with_matrix([0, shift_col]);
                let mut untarget = FieldBits::with_matrix([
                    0,
                    !(shift_target.m[1] | (1u64 << ((col_num - 4) * 16))),
                ]);
                untarget.m[1] = untarget.col_bits(col_num);
                untarget = untarget.and(shape);
                match col_num {
                    4 => shape.m[1] &= !0xffff,
                    5 => shape.m[1] &= !0xffff0000,
                    _ => {}
                }
                shape = shape.or(shift);
                shape = shape.or(untarget);
            }

            self.shapes[index] = shape;
            self.original_shapes[index] = shape;
        }
        self.key_string = None;
    }
}

impl fmt::Display for ShapeBitField {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for y in (1..=14).rev() {
            write!(f, "{y:02}: ")?;
            for x in 0..6 {
                let mut empty = true;
                for (index, shape) in self.shapes.iter().enumerate() {
                    if shape.onebit(x, y) > 0 {
                        write!(f, "{}", num_to_letter(index + 1))?;
                        empty = false;
                        break;
                    }
                }
                if empty {
                    write!(f, ".")?;
                }
            }
            writeln!(f)?;
        }
        Ok(())
    }
}

fn draw_field(puyo: &image::RgbaImage, out: &mut image::RgbaImage) {
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

fn draw_shape_puyo(
    field_bits: &FieldBits,
    shape_num: u32,
    x: usize,
    y: usize,
    puyo: &image::RgbaImage,
    out: &mut image::RgbaImage,
) {
    let mut tile_y = 0u32;
    let up = field_bits.onebit(x, y + 1) > 0;
    let down = y > 0 && field_bits.onebit(x, y - 1) > 0;
    let left = x > 0 && field_bits.onebit(x - 1, y) > 0;
    let right = x < 5 && field_bits.onebit(x + 1, y) > 0;

    if up && !down && !left && !right {
        tile_y = 1;
    } else if !up && down && !left && !right {
        tile_y = 2;
    } else if up && down && !left && !right {
        tile_y = 3;
    } else if !up && !down && left && !right {
        tile_y = 4;
    } else if up && !down && left && !right {
        tile_y = 5;
    } else if !up && down && left && !right {
        tile_y = 6;
    } else if up && down && left && !right {
        tile_y = 7;
    } else if !up && !down && !left && right {
        tile_y = 8;
    } else if up && !down && !left && right {
        tile_y = 9;
    } else if !up && down && !left && right {
        tile_y = 10;
    } else if up && down && !left && right {
        tile_y = 11;
    } else if !up && !down && left && right {
        tile_y = 12;
    } else if up && !down && left && right {
        tile_y = 13;
    } else if !up && down && left && right {
        tile_y = 14;
    } else if up && down && left && right {
        tile_y = 15;
    }

    overlay_tile(
        puyo,
        shape_num,
        tile_y,
        out,
        ((x + 1) as u32) * TILE_SIZE,
        ((13 - y) as u32) * TILE_SIZE,
    );
}

fn letter_to_num(ch: char) -> Option<usize> {
    SHAPE_LABELS
        .iter()
        .position(|label| *label == ch)
        .map(|idx| idx + 1)
}

fn num_to_letter(num: usize) -> char {
    SHAPE_LABELS
        .get(num - 1)
        .copied()
        .unwrap_or_else(|| panic!("shape label index out of range: {num}"))
}
