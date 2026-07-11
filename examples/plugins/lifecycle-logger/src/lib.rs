use dragonfly_plugin::{Plugin, plugin};

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
}
