pub fn long_bonus(mut num_erased: usize) -> usize {
    const LONG_BONUS_TABLE: [usize; 12] = [0, 0, 0, 0, 0, 2, 3, 4, 5, 6, 7, 10];
    if num_erased > 11 {
        num_erased = 11;
    }
    LONG_BONUS_TABLE[num_erased]
}

pub fn color_bonus(num_colors: usize) -> usize {
    const COLOR_BONUS_TABLE: [usize; 6] = [0, 0, 3, 6, 12, 24];
    COLOR_BONUS_TABLE[num_colors]
}

pub fn rensa_bonus(nth_rensa: usize) -> usize {
    const RENSA_BONUS_TABLE: [usize; 20] = [
        0, 0, 8, 16, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 480, 512,
    ];
    RENSA_BONUS_TABLE[nth_rensa]
}

pub fn calc_rensa_bonus_coef(
    rensa_bonus_coef: usize,
    long_bonus_coef: usize,
    color_bonus_coef: usize,
) -> usize {
    let coef = rensa_bonus_coef + long_bonus_coef + color_bonus_coef;
    if coef == 0 {
        return 1;
    }
    if coef > 999 {
        return 999;
    }
    coef
}
