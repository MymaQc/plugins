use dragonfly_plugin::{PlayerMoveEvent, Plugin, plugin};

#[derive(Default)]
struct MovementGuard;

#[plugin(id = "bedrock-gophers:movement-guard")]
impl Plugin for MovementGuard {
    fn on_move(&self, event: &mut PlayerMoveEvent<'_>) {
        if event.new_position().y < 0.0 {
            event.cancel();
        }
    }
}
