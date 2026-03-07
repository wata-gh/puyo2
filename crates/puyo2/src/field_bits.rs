use std::fmt;

use serde::{Deserialize, Serialize};

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct FieldBits {
    pub m: [u64; 2],
}

impl FieldBits {
    pub const fn new() -> Self {
        Self { m: [0, 0] }
    }

    pub const fn with_matrix(m: [u64; 2]) -> Self {
        Self { m }
    }

    fn bit_char(bit: bool) -> char {
        if bit { 'o' } else { '.' }
    }

    pub const fn is_empty(&self) -> bool {
        self.m[0] == 0 && self.m[1] == 0
    }

    pub const fn equals(&self, other: &Self) -> bool {
        self.m[0] == other.m[0] && self.m[1] == other.m[1]
    }

    pub fn onebit(&self, x: usize, y: usize) -> u64 {
        let idx = x >> 2;
        let pos = (x & 3) * 16 + y;
        self.m[idx] & (1u64 << pos)
    }

    pub fn set_onebit(&mut self, x: usize, y: usize) {
        let idx = x >> 2;
        let pos = (x & 3) * 16 + y;
        self.m[idx] |= 1u64 << pos;
    }

    #[inline]
    pub fn col_bits(&self, column: usize) -> u64 {
        match column {
            0 => self.m[0] & 0xffff,
            1 => self.m[0] & 0xffff0000,
            2 => self.m[0] & 0xffff00000000,
            3 => self.m[0] & 0xffff000000000000,
            4 => self.m[1] & 0xffff,
            5 => self.m[1] & 0xffff0000,
            _ => panic!("column number must be 0-5. passed {column}"),
        }
    }

    #[inline]
    pub fn shifted_col_bits(&self, column: usize) -> u64 {
        let mut col = self.col_bits(column);
        if column < 4 {
            col >>= 16 * column;
        } else {
            col >>= 16 * (column - 4);
        }
        col
    }

    pub fn mask_field13(&self) -> Self {
        Self::with_matrix([self.m[0] & 0x3FFE3FFE3FFE3FFE, self.m[1] & 0x3FFE3FFE])
    }

    #[inline]
    pub fn mask_field12(&self) -> Self {
        Self::with_matrix([self.m[0] & 0x1FFE1FFE1FFE1FFE, self.m[1] & 0x1FFE1FFE])
    }

    #[inline]
    pub fn and(self, other: Self) -> Self {
        Self::with_matrix([self.m[0] & other.m[0], self.m[1] & other.m[1]])
    }

    #[inline]
    pub fn or(self, other: Self) -> Self {
        Self::with_matrix([self.m[0] | other.m[0], self.m[1] | other.m[1]])
    }

    pub fn not(self) -> Self {
        Self::with_matrix([!self.m[0], !self.m[1]])
    }

    pub fn and_not_mut(&mut self, other: Self) {
        self.m[0] &= !other.m[0];
        self.m[1] &= !other.m[1];
    }

    #[inline]
    pub fn popcount(&self) -> usize {
        self.m[0].count_ones() as usize + self.m[1].count_ones() as usize
    }

    pub fn expand1(&self, mask: Self) -> Self {
        let r4 = self.m[0] & 0xffff000000000000;
        let r5 = self.m[1] & 0xffff;
        let m0 = self.m[0]
            | (self.m[0] << 1)
            | (self.m[0] >> 1)
            | (self.m[0] << 16)
            | (self.m[0] >> 16)
            | (r5 << 48);
        let m1 = self.m[1]
            | (self.m[1] << 1)
            | (self.m[1] >> 1)
            | (self.m[1] << 16)
            | (self.m[1] >> 16)
            | (r4 >> 48);
        Self::with_matrix([m0 & mask.m[0], m1 & mask.m[1]])
    }

    pub fn expand(&self, mask: Self) -> Self {
        let mut current = *self;
        loop {
            let next = current.expand1(mask);
            if next == current {
                return next;
            }
            current = next;
        }
    }

    pub fn fast_lift(&self, y: usize) -> Self {
        let mut m0 = 0u64;
        match y / 16 {
            1 => m0 = self.m[0] >> 48,
            2 => m0 = self.m[0] >> 32,
            3 => m0 = self.m[0] >> 16,
            4 => m0 = self.m[0],
            5 => m0 = self.m[0] << 16,
            _ => {}
        }
        Self::with_matrix([self.m[0] << y, (self.m[1] << y) | m0])
    }

    pub fn find_vanishing_bits(&self) -> Self {
        let r4 = self.m[0] & 0xffff000000000000;
        let r5 = self.m[1] & 0xffff;
        let u = [self.m[0] & (self.m[0] << 1), self.m[1] & (self.m[1] << 1)];
        let d = [self.m[0] & (self.m[0] >> 1), self.m[1] & (self.m[1] >> 1)];
        let l = [
            self.m[0] & ((self.m[0] >> 16) | (r5 << 48)),
            self.m[1] & (self.m[1] >> 16),
        ];
        let r = [
            self.m[0] & (self.m[0] << 16),
            self.m[1] & ((self.m[1] << 16) | (r4 >> 48)),
        ];

        let ud_and = [u[0] & d[0], u[1] & d[1]];
        let lr_and = [l[0] & r[0], l[1] & r[1]];
        let ud_or = [u[0] | d[0], u[1] | d[1]];
        let lr_or = [l[0] | r[0], l[1] | r[1]];
        let threes_ud_and_and_lr_or = [ud_and[0] & lr_or[0], ud_and[1] & lr_or[1]];
        let threes_lr_and_and_ud_or = [lr_and[0] & ud_or[0], lr_and[1] & ud_or[1]];
        let threes = [
            threes_ud_and_and_lr_or[0] | threes_lr_and_and_ud_or[0],
            threes_ud_and_and_lr_or[1] | threes_lr_and_and_ud_or[1],
        ];

        let twos_ud_and_or_lr_and = [ud_and[0] | lr_and[0], ud_and[1] | lr_and[1]];
        let twos_ud_or_and_lr_or = [ud_or[0] & lr_or[0], ud_or[1] & lr_or[1]];
        let twos = [
            twos_ud_and_or_lr_and[0] | twos_ud_or_and_lr_or[0],
            twos_ud_and_or_lr_and[1] | twos_ud_or_and_lr_or[1],
        ];

        let two_d = [(twos[0] >> 1) & twos[0], (twos[1] >> 1) & twos[1]];
        let two_l = [
            ((twos[0] >> 16) | ((twos[1] & 0xffff) << 48)) & twos[0],
            (twos[1] >> 16) & twos[1],
        ];

        let mut vanishing = Self::new();
        vanishing.m[0] |= threes[0] | two_d[0] | two_l[0];
        vanishing.m[1] |= threes[1] | two_d[1] | two_l[1];
        if vanishing.is_empty() {
            return vanishing;
        }

        let two_u = [(twos[0] << 1) & twos[0], (twos[1] << 1) & twos[1]];
        let two_r = [
            ((twos[0] << 16) & twos[0]),
            ((twos[1] << 16) | (twos[0] >> 48)) & twos[1],
        ];

        vanishing.m[0] |= two_u[0] | two_r[0];
        vanishing.m[1] |= two_u[1] | two_r[1];
        vanishing.expand1(*self)
    }

    pub fn has_vanishing_bits(&self) -> bool {
        let r4 = self.m[0] & 0xffff000000000000;
        let r5 = self.m[1] & 0xffff;
        let u = [self.m[0] & (self.m[0] << 1), self.m[1] & (self.m[1] << 1)];
        let d = [self.m[0] & (self.m[0] >> 1), self.m[1] & (self.m[1] >> 1)];
        let l = [
            self.m[0] & ((self.m[0] >> 16) | (r5 << 48)),
            self.m[1] & (self.m[1] >> 16),
        ];
        let r = [
            self.m[0] & (self.m[0] << 16),
            self.m[1] & ((self.m[1] << 16) | (r4 >> 48)),
        ];

        let ud_and = [u[0] & d[0], u[1] & d[1]];
        let lr_and = [l[0] & r[0], l[1] & r[1]];
        let ud_or = [u[0] | d[0], u[1] | d[1]];
        let lr_or = [l[0] | r[0], l[1] | r[1]];
        let threes = [
            (ud_and[0] & lr_or[0]) | (lr_and[0] & ud_or[0]),
            (ud_and[1] & lr_or[1]) | (lr_and[1] & ud_or[1]),
        ];
        let twos = [
            ud_and[0] | lr_and[0] | (ud_or[0] & lr_or[0]),
            ud_and[1] | lr_and[1] | (ud_or[1] & lr_or[1]),
        ];
        let two_d = [(twos[0] >> 1) & twos[0], (twos[1] >> 1) & twos[1]];
        let two_l = [
            ((twos[0] >> 16) | ((twos[1] & 0xffff) << 48)) & twos[0],
            (twos[1] >> 16) & twos[1],
        ];

        (threes[0] | two_d[0] | two_l[0]) != 0 || (threes[1] | two_d[1] | two_l[1]) != 0
    }

    pub fn iterate_bit_with_masking<F>(&self, mut callback: F)
    where
        F: FnMut(Self) -> Self,
    {
        let mut current = *self;
        while !current.is_empty() {
            let len0 = 64usize.saturating_sub(current.m[0].leading_zeros() as usize);
            if len0 > 0 {
                let mask = callback(Self::with_matrix([1u64 << (len0 - 1), 0]));
                current.and_not_mut(mask);
            }
            let len1 = 64usize.saturating_sub(current.m[1].leading_zeros() as usize);
            if len1 > 0 {
                let mask = callback(Self::with_matrix([0, 1u64 << (len1 - 1)]));
                current.and_not_mut(mask);
            }
        }
    }

    pub const fn to_int_array(&self) -> [u64; 2] {
        self.m
    }
}

impl fmt::Display for FieldBits {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for y in (0..=14).rev() {
            write!(f, "{y:02}: ")?;
            for x in 0..6 {
                let idx = x >> 2;
                let pos = (x & 3) * 16 + y;
                let bit = ((self.m[idx] >> pos) & 1) == 1;
                write!(f, "{}", Self::bit_char(bit))?;
            }
            writeln!(f)?;
        }
        Ok(())
    }
}
