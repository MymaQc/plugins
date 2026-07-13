#![no_std]

include!("generated.rs");

#[cfg(test)]
mod tests {
    use super::*;
    use core::mem::{align_of, offset_of, size_of};

    #[test]
    fn movement_layout_is_stable() {
        assert_eq!(size_of::<DfPlayerId>(), 24);
        assert_eq!(align_of::<DfPlayerId>(), 8);
        assert_eq!(size_of::<DfRotation>(), 16);
        assert_eq!(size_of::<DfPlayerMoveInput>(), 88);
        assert_eq!(align_of::<DfPlayerMoveInput>(), 8);
        assert_eq!(size_of::<DfPlayerMoveState>(), 1);
    }

    #[test]
    #[cfg(target_pointer_width = "64")]
    fn skin_layout_is_stable() {
        assert_eq!(size_of::<DfSkinAnimationInfo>(), 40);
        assert_eq!(align_of::<DfSkinAnimationInfo>(), 8);
        assert_eq!(offset_of!(DfSkinAnimationInfo, frame_count), 16);
        assert_eq!(offset_of!(DfSkinAnimationInfo, pixels_len), 32);

        assert_eq!(size_of::<DfSkinInfo>(), 88);
        assert_eq!(align_of::<DfSkinInfo>(), 8);
        assert_eq!(offset_of!(DfSkinInfo, play_fab_id_len), 16);
        assert_eq!(offset_of!(DfSkinInfo, cape_width), 64);
        assert_eq!(offset_of!(DfSkinInfo, cape_pixels_len), 72);

        assert_eq!(size_of::<DfSkinData>(), 184);
        assert_eq!(align_of::<DfSkinData>(), 8);
        assert_eq!(offset_of!(DfSkinData, animation_pixels), 168);

        assert_eq!(size_of::<DfSkinAnimationView>(), 48);
        assert_eq!(align_of::<DfSkinAnimationView>(), 8);
        assert_eq!(offset_of!(DfSkinAnimationView, frame_count), 16);
        assert_eq!(offset_of!(DfSkinAnimationView, pixels), 32);

        assert_eq!(size_of::<DfSkinView>(), 152);
        assert_eq!(align_of::<DfSkinView>(), 8);
        assert_eq!(offset_of!(DfSkinView, play_fab_id), 16);
        assert_eq!(offset_of!(DfSkinView, cape_width), 112);
        assert_eq!(offset_of!(DfSkinView, animations), 136);
    }

    #[test]
    #[cfg(target_pointer_width = "64")]
    fn host_v2_layout_is_stable() {
        assert_eq!(size_of::<DfHostApiV2>(), 120);
        assert_eq!(align_of::<DfHostApiV2>(), 8);
        assert_eq!(offset_of!(DfHostApiV2, context), 8);
        assert_eq!(offset_of!(DfHostApiV2, player_text), 16);
        assert_eq!(offset_of!(DfHostApiV2, player_skin_open), 80);
        assert_eq!(offset_of!(DfHostApiV2, player_skin_set), 112);
    }
}
