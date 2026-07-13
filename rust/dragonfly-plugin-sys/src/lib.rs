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
        assert_eq!(size_of::<DfPlayerMoveInput>(), 96);
        assert_eq!(align_of::<DfPlayerMoveInput>(), 8);
        assert_eq!(offset_of!(DfPlayerMoveInput, invocation), 0);
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
    fn host_v18_layout_is_stable() {
        assert_eq!(size_of::<DfInventoryId>(), 32);
        assert_eq!(size_of::<DfItemStackInfo>(), 80);
        assert_eq!(size_of::<DfItemStackSnapshot>(), 88);
        assert_eq!(size_of::<DfItemStackSnapshot>(), 88);
        assert_eq!(size_of::<DfItemStackData>(), 152);
        assert_eq!(size_of::<DfItemStackViewV3>(), 120);
        assert_eq!(size_of::<DfWorldId>(), 8);
        assert_eq!(size_of::<DfBlockData>(), 48);
        assert_eq!(size_of::<DfBlockView>(), 32);
        assert_eq!(size_of::<DfEntitySpawnOptions>(), 80);
        assert_eq!(size_of::<DfEntityState>(), 128);
        assert_eq!(offset_of!(DfEntityState, capabilities), 64);
        assert_eq!(offset_of!(DfEntityState, world), 72);
        assert_eq!(offset_of!(DfEntityState, entity_type), 80);
        assert_eq!(size_of::<DfPlayerAttackEntityInput>(), 56);
        assert_eq!(offset_of!(DfPlayerAttackEntityInput, player), 8);
        assert_eq!(offset_of!(DfPlayerAttackEntityInput, target), 32);
        assert_eq!(size_of::<DfPlayerAttackEntityState>(), 32);
        assert_eq!(offset_of!(DfPlayerAttackEntityState, knockback_force), 8);
        assert_eq!(offset_of!(DfPlayerAttackEntityState, knockback_height), 16);
        assert_eq!(offset_of!(DfPlayerAttackEntityState, critical), 24);
        assert_eq!(size_of::<DfPlayerItemUseOnEntityInput>(), 56);
        assert_eq!(offset_of!(DfPlayerItemUseOnEntityInput, player), 8);
        assert_eq!(offset_of!(DfPlayerItemUseOnEntityInput, target), 32);
        assert_eq!(size_of::<DfPlayerItemUseOnEntityState>(), 1);
        assert_eq!(size_of::<DfPlayerChangeWorldInput>(), 48);
        assert_eq!(offset_of!(DfPlayerChangeWorldInput, player), 8);
        assert_eq!(offset_of!(DfPlayerChangeWorldInput, before), 32);
        assert_eq!(offset_of!(DfPlayerChangeWorldInput, after), 40);
        assert_eq!(size_of::<DfPlayerChangeWorldState>(), 1);
        assert_eq!(size_of::<DfPlayerRespawnInput>(), 32);
        assert_eq!(offset_of!(DfPlayerRespawnInput, player), 8);
        assert_eq!(size_of::<DfPlayerRespawnState>(), 32);
        assert_eq!(offset_of!(DfPlayerRespawnState, position), 0);
        assert_eq!(offset_of!(DfPlayerRespawnState, world), 24);
        assert_eq!(size_of::<DfPlayerSkinChangeInput>(), 40);
        assert_eq!(offset_of!(DfPlayerSkinChangeInput, player), 8);
        assert_eq!(offset_of!(DfPlayerSkinChangeInput, snapshot), 32);
        assert_eq!(size_of::<DfPlayerSkinChangeState>(), 1);
        assert_eq!(size_of::<DfParticleViewV1>(), 40);
        assert_eq!(offset_of!(DfParticleViewV1, block), 32);
        assert_eq!(size_of::<DfSoundViewV1>(), 40);
        assert_eq!(offset_of!(DfSoundViewV1, scalar), 16);
        assert_eq!(offset_of!(DfSoundViewV1, item), 32);
        assert_eq!(size_of::<DfDamageSourceView>(), 88);
        assert_eq!(offset_of!(DfDamageSourceView, entity), 24);
        assert_eq!(offset_of!(DfDamageSourceView, block), 72);
        assert_eq!(size_of::<DfHealingSourceView>(), 24);
        assert_eq!(size_of::<DfPlayerHealResult>(), 8);
        assert_eq!(size_of::<DfPlayerHurtResult>(), 16);
        assert_eq!(size_of::<DfEffectView>(), 32);
        assert_eq!(offset_of!(DfEffectView, potency), 16);
        assert_eq!(offset_of!(DfEffectView, mode), 24);
        assert_eq!(size_of::<DfEffectBuffer>(), 24);
        assert_eq!(offset_of!(DfEffectBuffer, data), 0);
        assert_eq!(offset_of!(DfEffectBuffer, len), 8);
        assert_eq!(offset_of!(DfEffectBuffer, capacity), 16);
        assert_eq!(size_of::<DfWorldOpenSpecV1>(), 80);
        assert_eq!(offset_of!(DfWorldOpenSpecV1, provider_path), 8);
        assert_eq!(offset_of!(DfWorldOpenSpecV1, fixed_time), 40);
        assert_eq!(offset_of!(DfWorldOpenSpecV1, open_mode), 48);
        assert_eq!(offset_of!(DfWorldOpenSpecV1, read_only), 76);
        assert_eq!(size_of::<DfHostApiV18>(), 480);
        assert_eq!(align_of::<DfHostApiV18>(), 8);
        assert_eq!(offset_of!(DfHostApiV18, context), 8);
        assert_eq!(offset_of!(DfHostApiV18, player_text), 16);
        assert_eq!(offset_of!(DfHostApiV18, player_skin_open), 80);
        assert_eq!(offset_of!(DfHostApiV18, player_skin_set), 112);
        assert_eq!(offset_of!(DfHostApiV18, inventory_size), 120);
        assert_eq!(offset_of!(DfHostApiV18, player_held_slot_set), 200);
        assert_eq!(offset_of!(DfHostApiV18, player_scoreboard), 208);
        assert_eq!(offset_of!(DfHostApiV18, player_form_close), 232);
        assert_eq!(offset_of!(DfHostApiV18, world_lookup), 240);
        assert_eq!(offset_of!(DfHostApiV18, world_spawn_set), 320);
        assert_eq!(offset_of!(DfHostApiV18, world_entity_spawn), 328);
        assert_eq!(offset_of!(DfHostApiV18, entity_despawn), 384);
        assert_eq!(offset_of!(DfHostApiV18, world_particle_add), 392);
        assert_eq!(offset_of!(DfHostApiV18, world_sound_play), 400);
        assert_eq!(offset_of!(DfHostApiV18, player_sound_play), 408);
        assert_eq!(offset_of!(DfHostApiV18, player_heal), 416);
        assert_eq!(offset_of!(DfHostApiV18, player_hurt), 424);
        assert_eq!(offset_of!(DfHostApiV18, skin_snapshot_info), 432);
        assert_eq!(offset_of!(DfHostApiV18, skin_snapshot_set), 440);
        assert_eq!(offset_of!(DfHostApiV18, world_open_spec), 448);
        assert_eq!(offset_of!(DfHostApiV18, player_transfer), 456);
        assert_eq!(offset_of!(DfHostApiV18, player_effects), 464);
        assert_eq!(offset_of!(DfHostApiV18, player_effects_clear), 472);
    }

    #[test]
    #[cfg(target_pointer_width = "64")]
    fn entity_v3_layout_is_stable() {
        assert_eq!(DF_ABI_VERSION, 3);
        assert_eq!(DF_HOST_ABI_VERSION, 18);
        assert_eq!(size_of::<DfEntityTypeDescriptorV2>(), 144);
        assert_eq!(offset_of!(DfEntityTypeDescriptorV2, type_key), 80);
        assert_eq!(offset_of!(DfEntityTypeDescriptorV2, family), 88);
        assert_eq!(offset_of!(DfEntityTypeDescriptorV2, initial_health), 96);
        assert_eq!(offset_of!(DfEntityTypeDescriptorV2, state_version), 120);
        assert_eq!(offset_of!(DfEntityTypeDescriptorV2, gravity), 128);
        assert_eq!(size_of::<DfEntitySpawnViewV3>(), 200);
        assert_eq!(offset_of!(DfEntitySpawnViewV3, custom_instance), 176);
        assert_eq!(size_of::<DfEntityTickInput>(), 48);
        assert_eq!(offset_of!(DfEntityTickInput, current), 32);
        assert_eq!(size_of::<DfEntityHurtInput>(), 136);
        assert_eq!(size_of::<DfEntityHurtState>(), 16);
        assert_eq!(size_of::<DfEntityHealInput>(), 72);
        assert_eq!(size_of::<DfEntityHealState>(), 16);
        assert_eq!(size_of::<DfEntityDeathInput>(), 136);
        assert_eq!(size_of::<DfEntityDeathState>(), 1);
        assert_eq!(size_of::<DfPluginApiV3>(), 128);
        assert_eq!(offset_of!(DfPluginApiV3, entity_type_count), 64);
        assert_eq!(offset_of!(DfPluginApiV3, entity_type_at), 72);
        assert_eq!(offset_of!(DfPluginApiV3, handle_entity), 80);
        assert_eq!(offset_of!(DfPluginApiV3, handle_event), 120);
    }
}
