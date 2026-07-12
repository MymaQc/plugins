use dragonfly::{
    PlayerBlockBreakEvent, PlayerBlockPlaceEvent, PlayerDeathEvent, PlayerExperienceGainEvent,
    PlayerFireExtinguishEvent, PlayerFoodLossEvent, PlayerHealEvent, PlayerHeldSlotChangeEvent,
    PlayerHurtEvent, PlayerJoinEvent, PlayerJumpEvent, PlayerPunchAirEvent, PlayerQuitEvent,
    PlayerSleepEvent, PlayerStartBreakEvent, PlayerTeleportEvent, PlayerToggleSneakEvent,
    PlayerToggleSprintEvent, Plugin, plugin,
};

#[derive(Default)]
struct LifecycleLogger;

#[plugin]
impl Plugin for LifecycleLogger {
    fn on_enable(&self) {
        eprintln!("lifecycle-logger enabled");
    }

    fn on_disable(&self) {
        eprintln!("lifecycle-logger disabled");
    }

    fn on_join(&self, event: &mut PlayerJoinEvent<'_>) {
        eprintln!("{} joined", event.name());
    }

    fn on_quit(&self, event: &PlayerQuitEvent<'_>) {
        eprintln!("{} quit", event.name());
    }

    fn on_hurt(&self, event: &mut PlayerHurtEvent<'_>) {
        eprintln!("player hurt for {} by {}", event.damage(), event.source());
    }

    fn on_heal(&self, event: &mut PlayerHealEvent<'_>) {
        eprintln!("player healed for {} by {}", event.health(), event.source());
    }

    fn on_block_break(&self, event: &mut PlayerBlockBreakEvent<'_>) {
        eprintln!("broke {} at {:?}", event.block(), event.position());
    }

    fn on_block_place(&self, event: &mut PlayerBlockPlaceEvent<'_>) {
        eprintln!("placed {} at {:?}", event.block(), event.position());
    }

    fn on_food_loss(&self, event: &mut PlayerFoodLossEvent<'_>) {
        eprintln!("food changed from {} to {}", event.from(), event.to());
    }

    fn on_death(&self, event: &mut PlayerDeathEvent<'_>) {
        eprintln!("player died from {}", event.source());
    }

    fn on_start_break(&self, event: &mut PlayerStartBreakEvent<'_>) {
        eprintln!("started breaking at {:?}", event.position());
    }

    fn on_fire_extinguish(&self, event: &mut PlayerFireExtinguishEvent<'_>) {
        eprintln!("extinguished fire at {:?}", event.position());
    }
    fn on_toggle_sprint(&self, event: &mut PlayerToggleSprintEvent<'_>) {
        eprintln!("sprint={}", event.after());
    }
    fn on_toggle_sneak(&self, event: &mut PlayerToggleSneakEvent<'_>) {
        eprintln!("sneak={}", event.after());
    }
    fn on_jump(&self, event: &PlayerJumpEvent<'_>) {
        if let Some(name) = event.player().name() {
            eprintln!("{name} jumped");
        }
    }
    fn on_teleport(&self, event: &mut PlayerTeleportEvent<'_>) {
        eprintln!("teleporting to {:?}", event.position());
    }
    fn on_experience_gain(&self, event: &mut PlayerExperienceGainEvent<'_>) {
        eprintln!("gaining {} experience", event.amount());
    }
    fn on_punch_air(&self, _event: &mut PlayerPunchAirEvent<'_>) {
        eprintln!("punched air");
    }
    fn on_held_slot_change(&self, event: &mut PlayerHeldSlotChangeEvent<'_>) {
        eprintln!("held slot {} -> {}", event.from(), event.to());
    }
    fn on_sleep(&self, event: &mut PlayerSleepEvent<'_>) {
        eprintln!("sleeping, reminder={}", event.send_reminder());
    }
}
