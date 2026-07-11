use dragonfly::{
    PlayerHealEvent, PlayerHurtEvent, PlayerJoinEvent, PlayerQuitEvent, Plugin, plugin,
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
}
