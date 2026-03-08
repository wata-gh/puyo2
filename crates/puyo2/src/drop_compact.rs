use std::sync::LazyLock;

static BYTE_COMPACT_TABLE: LazyLock<Box<[u8; 1 << 16]>> = LazyLock::new(|| {
    let mut table = Box::new([0u8; 1 << 16]);
    let mut keep_mask = 0usize;
    while keep_mask < 256 {
        let mut byte = 0usize;
        while byte < 256 {
            table[(keep_mask << 8) | byte] = compact_byte_reference(byte as u8, keep_mask as u8);
            byte += 1;
        }
        keep_mask += 1;
    }
    table
});

#[inline]
pub(crate) fn compact_lane_u16(lane: u16, vanished: u16) -> u16 {
    let keep = !vanished;
    let low_keep = keep as u8;
    let high_keep = (keep >> 8) as u8;
    let low = BYTE_COMPACT_TABLE[(usize::from(low_keep) << 8) | usize::from(lane as u8)] as u16;
    let high =
        BYTE_COMPACT_TABLE[(usize::from(high_keep) << 8) | usize::from((lane >> 8) as u8)] as u16;
    low | (high << low_keep.count_ones())
}

#[inline]
fn compact_byte_reference(byte: u8, keep_mask: u8) -> u8 {
    let mut compacted = 0u8;
    let mut dst = 0u8;
    for src in 0..8 {
        let keep = (keep_mask >> src) & 1;
        let bit = ((byte >> src) & 1) & keep;
        compacted |= bit << dst;
        dst += keep;
    }
    compacted
}

#[cfg(test)]
mod tests {
    use super::compact_lane_u16;

    #[test]
    fn compact_byte_table_matches_reference_for_all_inputs() {
        for keep_mask in 0u8..=u8::MAX {
            for byte in 0u8..=u8::MAX {
                assert_eq!(
                    super::BYTE_COMPACT_TABLE[(usize::from(keep_mask) << 8) | usize::from(byte)],
                    super::compact_byte_reference(byte, keep_mask)
                );
            }
        }
    }

    #[test]
    fn compact_lane_u16_matches_reference_on_random_inputs() {
        let mut seed = 0xd6e8_feb8_6659_fd93u64;
        for _ in 0..200_000 {
            let lane = next_u64(&mut seed) as u16;
            let vanished = next_u64(&mut seed) as u16;
            assert_eq!(
                compact_lane_u16(lane, vanished),
                compact_lane_reference(lane, vanished)
            );
        }
    }

    fn compact_lane_reference(lane: u16, vanished: u16) -> u16 {
        let mut compacted = 0u16;
        let mut dst = 0u16;
        for src in 0..16 {
            let keep = (((vanished >> src) & 1) ^ 1) as u16;
            let bit = ((lane >> src) & 1) & keep;
            compacted |= bit << dst;
            dst += keep;
        }
        compacted
    }

    fn next_u64(seed: &mut u64) -> u64 {
        *seed = seed.wrapping_mul(6364136223846793005).wrapping_add(1);
        *seed
    }
}
