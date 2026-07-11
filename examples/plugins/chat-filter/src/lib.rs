use dragonfly::{PlayerChatEvent, Plugin, plugin};

#[derive(Default)]
struct ChatFilter;

#[plugin]
impl Plugin for ChatFilter {
    fn on_chat(&self, event: &mut PlayerChatEvent<'_>) {
        if event.message().eq_ignore_ascii_case("blocked") {
            event.cancel();
            return;
        }

        let filtered = event.message().replace("foo", "bar");
        event
            .replace_message(&filtered)
            .expect("filtered message exceeds host capacity");
    }
}
