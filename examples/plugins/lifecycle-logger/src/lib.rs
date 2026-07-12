use dragonfly::{
    PlayerBlockBreakEvent, PlayerBlockPickEvent, PlayerBlockPlaceEvent, PlayerDeathEvent,
    PlayerExperienceGainEvent, PlayerFireExtinguishEvent, PlayerFoodLossEvent, PlayerHealEvent,
    PlayerHeldSlotChangeEvent, PlayerHurtEvent, PlayerItemConsumeEvent, PlayerItemReleaseEvent,
    PlayerItemUseEvent, PlayerItemUseOnBlockEvent, PlayerJoinEvent, PlayerJumpEvent,
    PlayerLecternPageTurnEvent, PlayerPunchAirEvent, PlayerQuitEvent, PlayerSignEditEvent,
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
    fn on_block_pick(&self, event: &mut PlayerBlockPickEvent<'_>) {
        eprintln!("picked {} at {:?}", event.block(), event.position());
    }
    fn on_lectern_page_turn(&self, event: &mut PlayerLecternPageTurnEvent<'_>) {
        eprintln!("lectern page {} -> {}", event.old_page(), event.new_page());
    }
    fn on_sign_edit(&self, event: &mut PlayerSignEditEvent<'_>) {
        eprintln!(
            "sign changed from {:?} to {:?}",
            event.old_text(),
            event.new_text()
        );
    }
    fn on_item_use(&self, _event: &mut PlayerItemUseEvent<'_>) {
        eprintln!("used an item");
    }
    fn on_item_use_on_block(&self, event: &mut PlayerItemUseOnBlockEvent<'_>) {
        eprintln!(
            "used item on {:?} face {:?}",
            event.position(),
            event.face()
        );
    }
    fn on_item_consume(&self, event: &mut PlayerItemConsumeEvent<'_>) {
        eprintln!(
            "consumed {} x{}",
            event.item().identifier(),
            event.item().count()
        );
    }
    fn on_item_release(&self, event: &mut PlayerItemReleaseEvent<'_>) {
        eprintln!(
            "released {} after {:?}",
            event.item().identifier(),
            event.duration()
        );
    }
}
