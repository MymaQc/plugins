#![no_std]

include!("generated.rs");

#[cfg(test)]
mod tests {
    use super::*;
    use core::mem::{align_of, size_of};

    #[test]
    fn movement_layout_is_stable() {
        assert_eq!(size_of::<DfPlayerId>(), 24);
        assert_eq!(align_of::<DfPlayerId>(), 8);
        assert_eq!(size_of::<DfRotation>(), 16);
        assert_eq!(size_of::<DfPlayerMoveInput>(), 88);
        assert_eq!(align_of::<DfPlayerMoveInput>(), 8);
        assert_eq!(size_of::<DfPlayerMoveState>(), 1);
    }
}
