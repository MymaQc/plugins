use dragonfly::{PlayerJoinEvent, PlayerQuitEvent, Plugin, plugin};

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
}
