use dragonfly::{
    PlayerBlockBreakEvent, PlayerBlockPlaceEvent, PlayerDeathEvent, PlayerFireExtinguishEvent,
    PlayerFoodLossEvent, PlayerHealEvent, PlayerHurtEvent, PlayerJoinEvent, PlayerQuitEvent,
    PlayerStartBreakEvent, Plugin, plugin,
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
}
