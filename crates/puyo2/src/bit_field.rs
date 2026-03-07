use std::fmt;

use serde::{Deserialize, Serialize};
use thiserror::Error;

use crate::{
    CHIGIRI_FRAMES_TABLE, Color, FieldBits, HandParseError, NthResult, PuyoSet, PuyoSetPlacement,
    RensaResult, SET_FRAMES_TABLE, ShapeBitField, SingleResult, calc_rensa_bonus_coef, color_bonus,
    expand_mattulwan_param, haipuyo_to_puyo_sets, long_bonus, rensa_bonus,
};

const BASIC_COLORS: [Color; 4] = [Color::Red, Color::Blue, Color::Yellow, Color::Green];

#[derive(Clone, Debug, Error, PartialEq, Eq)]
pub enum BasicBitFieldError {
    #[error(transparent)]
    InvalidHaipuyo(#[from] HandParseError),
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct BitField {
    pub m: [[u64; 2]; 3],
    pub table: [Color; Color::COUNT],
    pub colors: Vec<Color>,
}

impl Default for BitField {
    fn default() -> Self {
        Self::new()
    }
}

impl BitField {
    pub fn new() -> Self {
        let mut table = [Color::Empty; Color::COUNT];
        for color in BASIC_COLORS {
            table[color.idx()] = color;
        }
        Self {
            m: [[0, 0], [0, 0], [0, 0]],
            table,
            colors: BASIC_COLORS.to_vec(),
        }
    }

    pub fn with_colors(colors: Vec<Color>) -> Self {
        let has_purple = colors.contains(&Color::Purple);
        let mut table = [Color::Empty; Color::COUNT];
        if has_purple {
            for color in BASIC_COLORS {
                if colors.contains(&color) {
                    table[color.idx()] = color;
                }
            }
            for color in BASIC_COLORS {
                if table[color.idx()] == Color::Empty {
                    table[Color::Purple.idx()] = color;
                    table[color.idx()] = Color::Purple;
                    break;
                }
            }
        } else {
            for color in BASIC_COLORS {
                table[color.idx()] = color;
            }
        }
        let mut field = Self {
            m: [[0, 0], [0, 0], [0, 0]],
            table,
            colors: Vec::new(),
        };
        field.reset_colors();
        field
    }

    pub fn with_table(table: [Color; Color::COUNT]) -> Self {
        let mut field = Self {
            m: [[0, 0], [0, 0], [0, 0]],
            table,
            colors: Vec::new(),
        };
        field.reset_colors();
        field
    }

    pub fn with_table_and_colors(table: [Color; Color::COUNT], colors: Vec<Color>) -> Self {
        Self {
            m: [[0, 0], [0, 0], [0, 0]],
            table,
            colors,
        }
    }

    pub fn with_matrix(m: [[u64; 2]; 3]) -> Self {
        let mut field = Self::new();
        field.m = m;
        field
    }

    pub fn from_mattulwan(field: &str) -> Self {
        let mut bit_field = Self {
            m: [[0, 0], [0, 0], [0, 0]],
            table: [Color::Empty; Color::COUNT],
            colors: Vec::new(),
        };
        bit_field.build_colors_by_field_string(field);
        bit_field.set_mattulwan_param(field);
        bit_field
    }

    pub fn from_mattulwan_and_haipuyo(
        field: &str,
        haipuyo: &str,
    ) -> Result<Self, BasicBitFieldError> {
        let mut bit_field = Self::from_mattulwan(field);
        let table_colors = bit_field.table_colors(bit_field.table);
        let puyo_sets = haipuyo_to_puyo_sets(haipuyo)?;
        let (_, colors) = bit_field.create_table_and_colors(&puyo_sets);
        bit_field.merge_table_colors(table_colors, colors);
        Ok(bit_field)
    }

    fn color_bits(&self, color: Color) -> [u64; 3] {
        match color {
            Color::Empty => [0, 0, 0],
            Color::Ojama => [1, 0, 0],
            Color::Wall => [0, 1, 0],
            Color::Iron => [1, 1, 0],
            Color::Red => [0, 0, 1],
            Color::Blue => [1, 0, 1],
            Color::Yellow => [0, 1, 1],
            Color::Green => [1, 1, 1],
            Color::Purple => panic!("purple must be converted before bit encoding"),
        }
    }

    fn color_char(&self, color: Color) -> char {
        match color {
            Color::Empty => '.',
            Color::Ojama => 'O',
            Color::Wall => 'W',
            Color::Iron => 'I',
            Color::Red => 'R',
            Color::Blue => 'B',
            Color::Yellow => 'Y',
            Color::Green => 'G',
            Color::Purple => 'P',
        }
    }

    fn convert_color(&self, puyo: Color) -> Color {
        if puyo.is_special() {
            return puyo;
        }
        let has_table = BASIC_COLORS
            .into_iter()
            .chain(std::iter::once(Color::Purple))
            .any(|color| self.table[color.idx()] != Color::Empty);
        if !has_table {
            return puyo;
        }
        let converted = self.table[puyo.idx()];
        if converted == Color::Empty {
            panic!("convert table can not convert color {puyo:?}");
        }
        converted
    }

    fn build_colors_by_field_string(&mut self, field: &str) {
        self.table = [Color::Empty; Color::COUNT];
        let mut color_count = 0;
        for ch in expand_mattulwan_param(field).chars() {
            match ch {
                'a' => {}
                'b' => {
                    if self.table[Color::Red.idx()] == Color::Empty {
                        self.table[Color::Red.idx()] = Color::Red;
                        color_count += 1;
                    }
                }
                'c' => {
                    if self.table[Color::Blue.idx()] == Color::Empty {
                        self.table[Color::Blue.idx()] = Color::Blue;
                        color_count += 1;
                    }
                }
                'd' => {
                    if self.table[Color::Yellow.idx()] == Color::Empty {
                        self.table[Color::Yellow.idx()] = Color::Yellow;
                        color_count += 1;
                    }
                }
                'e' => {
                    if self.table[Color::Green.idx()] == Color::Empty {
                        self.table[Color::Green.idx()] = Color::Green;
                        color_count += 1;
                    }
                }
                'f' => {
                    if self.table[Color::Purple.idx()] == Color::Empty {
                        self.table[Color::Purple.idx()] = Color::Purple;
                        color_count += 1;
                    }
                }
                'g' => {}
                _ => panic!("only supports puyo color a,b,c,d,e,f,g. passed {ch:?}"),
            }
            if color_count == 4 {
                break;
            }
        }
        if self.table[Color::Purple.idx()] == Color::Purple {
            for color in BASIC_COLORS {
                if self.table[color.idx()] == Color::Empty {
                    self.table[color.idx()] = Color::Purple;
                    self.table[Color::Purple.idx()] = color;
                    break;
                }
            }
        }
        self.reset_colors();
    }

    fn create_table_and_colors(
        &self,
        puyo_sets: &[PuyoSet],
    ) -> ([Color; Color::COUNT], Vec<Color>) {
        let mut colors = Vec::new();
        let mut table = [Color::Empty; Color::COUNT];
        for puyo_set in puyo_sets {
            table[puyo_set.axis.idx()] = puyo_set.axis;
            table[puyo_set.child.idx()] = puyo_set.child;
        }
        if table[Color::Purple.idx()] == Color::Purple {
            let mut assigned = false;
            for color in BASIC_COLORS {
                if table[color.idx()] == Color::Empty {
                    table[color.idx()] = Color::Purple;
                    table[Color::Purple.idx()] = color;
                    assigned = true;
                    break;
                }
                colors.push(color);
            }
            if assigned {
                colors.push(Color::Purple);
            }
        } else {
            for color in BASIC_COLORS {
                if table[color.idx()] != Color::Empty {
                    colors.push(color);
                }
            }
        }
        (table, colors)
    }

    fn merge_table_colors(&mut self, table_colors: Vec<Color>, mut colors: Vec<Color>) {
        for color in table_colors {
            if !colors.contains(&color) {
                colors.push(color);
            }
        }
        let mut new_table = [Color::Empty; Color::COUNT];
        for color in &colors {
            new_table[color.idx()] = *color;
        }
        if new_table[Color::Purple.idx()] == Color::Purple {
            for color in BASIC_COLORS {
                if new_table[color.idx()] == Color::Empty {
                    new_table[color.idx()] = Color::Purple;
                    new_table[Color::Purple.idx()] = color;
                    break;
                }
            }
        }
        self.table = new_table;
        self.reset_colors();
    }

    fn reset_colors(&mut self) {
        self.colors = self.table_colors(self.table);
    }

    fn table_colors(&self, table: [Color; Color::COUNT]) -> Vec<Color> {
        let mut colors = Vec::new();
        for color in BASIC_COLORS {
            let value = table[color.idx()];
            if value == Color::Purple {
                colors.push(Color::Purple);
            } else if color == value {
                colors.push(color);
            }
        }
        colors
    }

    pub fn bits(&self, color: Color) -> FieldBits {
        match self.convert_color(color) {
            Color::Empty => FieldBits::with_matrix([
                !(self.m[0][0] | self.m[1][0] | self.m[2][0]),
                !(self.m[0][1] | self.m[1][1] | self.m[2][1]),
            ]),
            Color::Ojama => FieldBits::with_matrix([
                self.m[0][0] & !(self.m[1][0] | self.m[2][0]),
                self.m[0][1] & !(self.m[1][1] | self.m[2][1]),
            ]),
            Color::Wall => FieldBits::with_matrix([
                self.m[1][0] & !(self.m[0][0] | self.m[2][0]),
                self.m[1][1] & !(self.m[0][1] | self.m[2][1]),
            ]),
            Color::Iron => FieldBits::with_matrix([
                self.m[0][0] & self.m[1][0] & !self.m[2][0],
                self.m[0][1] & self.m[1][1] & !self.m[2][1],
            ]),
            Color::Red => FieldBits::with_matrix([
                self.m[2][0] & !(self.m[0][0] | self.m[1][0]),
                self.m[2][1] & !(self.m[0][1] | self.m[1][1]),
            ]),
            Color::Blue => FieldBits::with_matrix([
                self.m[2][0] & self.m[0][0] & !self.m[1][0],
                self.m[2][1] & self.m[0][1] & !self.m[1][1],
            ]),
            Color::Yellow => FieldBits::with_matrix([
                self.m[2][0] & self.m[1][0] & !self.m[0][0],
                self.m[2][1] & self.m[1][1] & !self.m[0][1],
            ]),
            Color::Green => FieldBits::with_matrix([
                self.m[2][0] & self.m[1][0] & self.m[0][0],
                self.m[2][1] & self.m[1][1] & self.m[0][1],
            ]),
            Color::Purple => unreachable!(),
        }
    }

    pub fn clone_for_simulation(&self) -> Self {
        Self {
            m: self.m,
            table: self.table,
            colors: self.colors.clone(),
        }
    }

    pub fn color(&self, x: usize, y: usize) -> Color {
        let idx = x >> 2;
        let pos = (x & 3) * 16 + y;
        let color_bits = ((self.m[0][idx] >> pos) & 1)
            | (((self.m[1][idx] >> pos) & 1) << 1)
            | (((self.m[2][idx] >> pos) & 1) << 2);
        self.convert_color(Color::from_bits(color_bits as u8))
    }

    pub fn drop_vanished(&mut self, vanished: FieldBits) {
        let mut dropmask = [0u64; 2];
        for x in 0..6 {
            let idx = x >> 2;
            let vc = vanished.col_bits(x).count_ones();
            let rotated = (((1u64 << vc) - 1).rotate_left(14u32 - vc)) << ((x & 3) * 16);
            dropmask[idx] |= rotated;
        }
        for plane in 0..self.m.len() {
            let r0 = extract(self.m[plane][0], !vanished.m[0]);
            let r1 = extract(self.m[plane][1], !vanished.m[1]);
            self.m[plane][0] = deposit(r0, !dropmask[0]);
            self.m[plane][1] = deposit(r1, !dropmask[1]);
        }
    }

    pub fn equals(&self, other: &Self) -> bool {
        self.m == other.m
    }

    pub fn equal_chain(&self, other: &Self) -> bool {
        let shapes = self.to_chain_shapes_u64_array();
        let other_shapes = other.to_chain_shapes_u64_array();
        if shapes.len() != other_shapes.len() {
            return false;
        }
        for (left, right) in shapes.iter().zip(other_shapes.iter()) {
            if left != right {
                return false;
            }
        }
        true
    }

    pub fn find_vanishing_bits(&self) -> FieldBits {
        let mut vanished = FieldBits::new();
        for color in &self.colors {
            vanished = vanished.or(self.bits(*color).mask_field12().find_vanishing_bits());
        }
        let ojama = vanished.expand1(self.bits(Color::Ojama));
        vanished.or(ojama)
    }

    pub fn flip_horizontal(&mut self) -> &mut Self {
        let mut m = [[0u64; 2]; 3];
        for (i, plane) in self.m.iter().enumerate() {
            m[i][1] = (plane[0] & 0xffff) << 16;
            m[i][1] |= (plane[0] & 0xffff0000) >> 16;
            m[i][0] = (plane[0] & 0xffff00000000) << 16;
            m[i][0] |= (plane[0] & 0xffff000000000000) >> 16;
            m[i][0] |= (plane[1] & 0xffff) << 16;
            m[i][0] |= (plane[1] & 0xffff0000) >> 16;
        }
        self.m = m;
        self
    }

    pub fn is_empty(&self) -> bool {
        self.m[0][0] == 0
            && self.m[1][0] == 0
            && self.m[2][0] == 0
            && self.m[0][1] == 0
            && self.m[1][1] == 0
            && self.m[2][1] == 0
    }

    pub fn mask_field(&self, mask: &FieldBits) -> Self {
        let mut masked = self.clone();
        for plane in 0..3 {
            masked.m[plane][0] &= mask.m[0];
            masked.m[plane][1] &= mask.m[1];
        }
        masked
    }

    pub fn mattulwan_editor_param(&self) -> String {
        let mut output = String::with_capacity(78);
        for y in (1..=13).rev() {
            for x in 0..6 {
                let ch = match self.color_char(self.color(x, y)) {
                    'R' => 'b',
                    'B' => 'c',
                    'Y' => 'd',
                    'G' => 'e',
                    'P' => 'f',
                    '.' => 'a',
                    'O' => 'g',
                    other => panic!("unsupported color char {other}"),
                };
                output.push(ch);
            }
        }
        output
    }

    pub fn mattulwan_editor_url(&self) -> String {
        format!(
            "https://pndsng.com/puyo/index.html?{}",
            self.mattulwan_editor_param()
        )
    }

    pub fn normalize(&self) -> Self {
        let mut normalized = String::new();
        let mut table = std::collections::BTreeMap::new();
        let mut colors = vec!['b', 'c', 'd', 'e', 'f'];
        for ch in self.mattulwan_editor_param().chars() {
            if ch == 'a' {
                normalized.push('a');
            } else if let Some(mapped) = table.get(&ch) {
                normalized.push(*mapped);
            } else {
                let mapped = colors.remove(0);
                table.insert(ch, mapped);
                normalized.push(mapped);
            }
        }
        Self::from_mattulwan(&normalized)
    }

    pub fn overall_shape(&self) -> FieldBits {
        let mut shape = FieldBits::new();
        for color in &self.colors {
            shape = shape.or(self.bits(*color));
        }
        shape.or(self.bits(Color::Ojama)).or(self.bits(Color::Iron))
    }

    pub fn place_puyo_with_placement(&mut self, placement: &PuyoSetPlacement) -> bool {
        let Some(puyo_set) = placement.puyo_set else {
            return false;
        };
        if placement.axis_x == placement.child_x && placement.axis_y == placement.child_y {
            return false;
        }
        if placement.axis_y > 13 {
            return false;
        }
        if !self.can_set_color(placement.axis_x, placement.axis_y) {
            return false;
        }
        if !self.can_set_color(placement.child_x, placement.child_y) {
            return false;
        }

        self.set_color(
            puyo_set.axis,
            placement.axis_x as usize,
            placement.axis_y as usize,
        );
        if placement.child_y <= 13 {
            self.set_color(
                puyo_set.child,
                placement.child_x as usize,
                placement.child_y as usize,
            );
        }
        true
    }

    fn can_set_color(&self, x: isize, y: isize) -> bool {
        if !(0..=5).contains(&x) || !(0..=14).contains(&y) {
            return false;
        }
        self.color(x as usize, y as usize) == Color::Empty
    }

    pub fn set_color(&mut self, color: Color, x: usize, y: usize) {
        if self.colors.len() < 4 && !color.is_special() {
            let found = self.colors.contains(&color);
            if !found {
                self.colors.push(color);
                if color == Color::Purple {
                    for basic in BASIC_COLORS {
                        if self.table[basic.idx()] == Color::Empty {
                            self.table[basic.idx()] = Color::Purple;
                            self.table[Color::Purple.idx()] = basic;
                            break;
                        }
                    }
                } else {
                    self.table[color.idx()] = color;
                }
            }
        }
        let encoded = self.color_bits(self.convert_color(color));
        let pos = (x & 3) * 16 + y;
        let idx = x >> 2;
        for (plane, bit) in encoded.iter().enumerate() {
            if *bit == 1 {
                self.m[plane][idx] |= *bit << pos;
            } else {
                let posbit = 1u64 << pos;
                if self.m[plane][idx] & posbit > 0 {
                    self.m[plane][idx] -= posbit;
                }
            }
        }
    }

    pub fn set_color_with_field_bits(&mut self, color: Color, bits: FieldBits) {
        let encoded = self.color_bits(self.convert_color(color));
        let mask = [!bits.m[0], !bits.m[1]];
        for (plane, bit) in encoded.iter().enumerate() {
            self.m[plane][0] &= mask[0];
            self.m[plane][1] &= mask[1];
            if *bit == 1 {
                self.m[plane][0] |= bits.m[0];
                self.m[plane][1] |= bits.m[1];
            }
        }
    }

    pub fn set_mattulwan_param(&mut self, field: &str) {
        for (i, ch) in expand_mattulwan_param(field).chars().enumerate() {
            let x = i % 6;
            let y = 13 - i / 6;
            let puyo = match ch {
                'a' => Color::Empty,
                'b' => Color::Red,
                'c' => Color::Blue,
                'd' => Color::Yellow,
                'e' => Color::Green,
                'f' => Color::Purple,
                'g' => Color::Ojama,
                _ => panic!("only supports puyo color a,b,c,d,e,f,g. passed {ch:?}"),
            };
            if puyo != Color::Empty {
                self.set_color(puyo, x, y);
            }
        }
    }

    pub fn simulate(&mut self) -> RensaResult {
        let mut result = RensaResult::new();
        while self.simulate1() {
            result.add_chain();
        }
        result.set_bit_field(self.clone());
        result
    }

    pub fn simulate_detail(&mut self) -> RensaResult {
        let mut result = RensaResult::new();
        loop {
            let mut num_colors = 0usize;
            let mut num_erased = 0usize;
            let mut long_bonus_coef = 0usize;
            let mut vanished = FieldBits::new();
            let mut nth = NthResult {
                nth: result.chains + 1,
                erased_puyos: Vec::new(),
            };

            for color in self.colors.clone() {
                let vb = self.bits(color).mask_field12().find_vanishing_bits();
                if !vb.is_empty() {
                    num_colors += 1;
                    let pop_count = vb.popcount();
                    num_erased += pop_count;
                    vanished = vanished.or(vb);

                    if pop_count <= 7 {
                        nth.erased_puyos.push(SingleResult {
                            color,
                            connected: pop_count,
                        });
                        long_bonus_coef += long_bonus(pop_count);
                        continue;
                    }

                    vb.iterate_bit_with_masking(|candidate| {
                        let expanded = candidate.expand(vb);
                        let pop_count = expanded.popcount();
                        nth.erased_puyos.push(SingleResult {
                            color,
                            connected: pop_count,
                        });
                        long_bonus_coef += long_bonus(pop_count);
                        expanded
                    });
                }
            }

            if num_colors == 0 {
                break;
            }

            result.nth_results.push(nth);
            vanished = vanished.or(vanished.expand1(self.bits(Color::Ojama)));
            result.add_chain();
            let color_bonus_coef = color_bonus(num_colors);
            let coef = calc_rensa_bonus_coef(
                rensa_bonus(result.chains),
                long_bonus_coef,
                color_bonus_coef,
            );
            result.add_erased(num_erased);
            result.add_score(10 * num_erased * coef);
            let heights = self.create_heights();
            result.quick = true;
            for (x, height) in heights.iter().enumerate() {
                let col = vanished.shifted_col_bits(x);
                if col != 0 {
                    let vh = 64usize.saturating_sub(col.leading_zeros() as usize) - 1;
                    if *height > vh {
                        result.quick = false;
                        break;
                    }
                }
            }
            self.drop_vanished(vanished);
        }
        result.set_bit_field(self.clone());
        result
    }

    pub fn simulate1(&mut self) -> bool {
        let vanished = self.find_vanishing_bits();
        if vanished.is_empty() {
            return false;
        }
        self.drop_vanished(vanished);
        true
    }

    pub fn to_chain_shapes(&self) -> Vec<FieldBits> {
        let mut cloned = self.clone();
        let mut shapes = Vec::new();
        loop {
            let vanished = cloned.find_vanishing_bits();
            if vanished.is_empty() {
                break;
            }
            shapes.push(vanished);
            cloned.drop_vanished(vanished);
        }
        shapes
    }

    pub fn to_chain_shapes_u64_array(&self) -> Vec<[u64; 2]> {
        self.to_chain_shapes()
            .into_iter()
            .map(|shape| shape.to_int_array())
            .collect()
    }

    pub fn to_shape_bit_field(&self) -> ShapeBitField {
        let mut shape_bit_field = ShapeBitField::new();
        for shape in self.to_chain_shapes() {
            shape_bit_field.add_shape(shape);
        }
        shape_bit_field
    }

    pub fn trim_left(&mut self) -> &mut Self {
        let mut mv = 0;
        if self.m[2][0] & 0xffff == 0 {
            mv += 1;
            if self.m[2][0] & 0xffff0000 == 0 {
                mv += 1;
                if self.m[2][0] & 0xffff00000000 == 0 {
                    mv += 1;
                    if self.m[2][0] & 0xffff000000000000 == 0 {
                        mv += 1;
                        if self.m[2][1] & 0xffff == 0 {
                            mv += 1;
                            if self.m[2][1] & 0xffff0000 == 0 {
                                return self;
                            }
                        }
                    }
                }
            }
        } else {
            return self;
        }

        match mv {
            1 => {
                for plane in 0..3 {
                    self.m[plane][0] >>= 16;
                    self.m[plane][0] |= (self.m[plane][1] << 48) & 0xffff000000000000;
                    self.m[plane][1] >>= 16;
                }
            }
            2 => {
                for plane in 0..3 {
                    self.m[plane][0] >>= 32;
                    self.m[plane][0] |= (self.m[plane][1] << 32) & 0xffffffff00000000;
                    self.m[plane][1] = 0;
                }
            }
            3 => {
                for plane in 0..3 {
                    self.m[plane][0] >>= 48;
                    self.m[plane][0] |= (self.m[plane][1] << 16) & 0x0000ffffffff0000;
                    self.m[plane][1] = 0;
                }
            }
            4 => {
                for plane in 0..3 {
                    self.m[plane][0] = self.m[plane][1];
                    self.m[plane][1] = 0;
                }
            }
            5 => {
                for plane in 0..3 {
                    self.m[plane][0] = self.m[plane][1] >> 16;
                    self.m[plane][1] = 0;
                }
            }
            _ => {}
        }
        self
    }

    pub fn create_heights(&self) -> [usize; 6] {
        let empty = self.bits(Color::Empty);
        let mut heights = [0usize; 6];
        for (column, slot) in heights.iter_mut().enumerate() {
            let mut empty_bits = empty.col_bits(column);
            if column < 4 {
                empty_bits >>= 16 * column;
            } else {
                empty_bits >>= 16 * (column - 4);
            }
            empty_bits |= 0xC000;
            *slot = 16 - empty_bits.count_ones() as usize;
        }
        heights
    }

    pub fn search_placement_for_pos(
        &self,
        puyo_set: &PuyoSet,
        pos: [usize; 2],
    ) -> Option<PuyoSetPlacement> {
        let heights = self.create_heights();
        self.search_placement_for_pos_with_heights(puyo_set, pos, heights)
    }

    pub(crate) fn search_placement_for_pos_with_heights(
        &self,
        puyo_set: &PuyoSet,
        pos: [usize; 2],
        heights: [usize; 6],
    ) -> Option<PuyoSetPlacement> {
        let ax = pos[0] as isize;
        let mut cx = pos[0] as isize;

        let y = heights[ax as usize] as isize + 1;
        let mut ay = y;
        let mut cy = y + 1;
        match pos[1] {
            0 => {}
            1 => {
                cx += 1;
                cy = heights[cx as usize] as isize + 1;
            }
            2 => {
                ay = cy;
                cy = y;
            }
            3 => {
                cx -= 1;
                cy = heights[cx as usize] as isize + 1;
            }
            _ => return None,
        }

        let mut x = 0isize;
        if ax != 2 || cx != 2 {
            if ax > 2 || cx > 2 {
                x = ax.max(cx);
            } else if ax < 2 || cx < 2 {
                x = ax.min(cx);
            }
        }

        if x != 2 {
            if x > 2 {
                for i in 3..x {
                    let h = heights[i as usize];
                    if h >= 13 {
                        return None;
                    }
                    if h == 12 {
                        let mut has_step = heights[1] >= 12 && heights[3] >= 12;
                        if !has_step {
                            for j in (0..i).rev() {
                                if heights[j as usize] >= 13 {
                                    break;
                                }
                                if heights[j as usize] == 11 {
                                    has_step = true;
                                }
                            }
                        }
                        if !has_step {
                            return None;
                        }
                    }
                }
            } else {
                for i in (x..=1).rev() {
                    let h = heights[i as usize];
                    if h >= 13 {
                        return None;
                    }
                    if h == 12 {
                        let mut has_step = heights[1] >= 12 && heights[3] >= 12;
                        if !has_step {
                            for j in (i + 1) as usize..heights.len() {
                                if heights[j] >= 13 {
                                    break;
                                }
                                if heights[j] == 11 {
                                    has_step = true;
                                    break;
                                }
                            }
                        }
                        if !has_step {
                            return None;
                        } else {
                            break;
                        }
                    }
                }
            }
        }

        if ay > 13 {
            return None;
        }
        if cy == 14 && self.color(cx as usize, cy as usize) != Color::Empty {
            return None;
        }

        let mut placement = PuyoSetPlacement {
            puyo_set: Some(*puyo_set),
            pos,
            axis_x: ax,
            axis_y: ay,
            child_x: cx,
            child_y: cy,
            chigiri: ax != cx && ay != cy,
            frames: 0,
        };

        if placement.axis_x < 0 || placement.axis_x as usize >= SET_FRAMES_TABLE[0].len() {
            return None;
        }
        if placement.axis_x == placement.child_x || placement.axis_y == placement.child_y {
            if placement.axis_y < 0 || placement.axis_y as usize >= SET_FRAMES_TABLE.len() {
                return None;
            }
            placement.frames =
                SET_FRAMES_TABLE[placement.axis_y as usize][placement.axis_x as usize];
        } else {
            let mut higher = placement.axis_y;
            let mut steps = placement.axis_y - placement.child_y;
            if higher < placement.child_y {
                higher = placement.child_y;
                steps = placement.child_y - placement.axis_y;
            }
            if higher < 0 || higher as usize >= SET_FRAMES_TABLE.len() {
                return None;
            }
            if steps < 0 || steps as usize >= CHIGIRI_FRAMES_TABLE.len() {
                return None;
            }
            placement.frames = SET_FRAMES_TABLE[higher as usize][placement.axis_x as usize]
                + CHIGIRI_FRAMES_TABLE[steps as usize];
        }
        Some(placement)
    }

    pub fn place_puyo(&mut self, puyo_set: PuyoSet, pos: [usize; 2]) -> (bool, bool) {
        if let Some(placement) = self.search_placement_for_pos(&puyo_set, pos) {
            self.place_puyo_with_placement(&placement);
            return (true, placement.chigiri);
        }
        (false, false)
    }
}

impl fmt::Display for BitField {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for y in (1..=14).rev() {
            write!(f, "{y:02}: ")?;
            for x in 0..6 {
                write!(f, "{}", self.color_char(self.color(x, y)))?;
            }
            writeln!(f)?;
        }
        Ok(())
    }
}

fn extract(x: u64, mut mask: u64) -> u64 {
    let mut result = 0u64;
    let mut next = 1u64;
    loop {
        let lsb = mask & mask.wrapping_neg();
        if lsb == 0 {
            return result;
        }
        mask ^= lsb;
        if x & lsb != 0 {
            result |= next;
        }
        next <<= 1;
    }
}

fn deposit(x: u64, mut mask: u64) -> u64 {
    let mut result = 0u64;
    let mut next = 1u64;
    loop {
        let lsb = mask & mask.wrapping_neg();
        if lsb == 0 {
            return result;
        }
        mask ^= lsb;
        if x & next != 0 {
            result |= lsb;
        }
        next <<= 1;
    }
}
