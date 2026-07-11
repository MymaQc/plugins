use dragonfly_plugin::{
    CommandEnum, CommandEvent, CommandSource, Dynamic, DynamicCommandEnum, Player, Plugin, Varargs,
    plugin,
};

struct GreetingTargets;

impl DynamicCommandEnum for GreetingTargets {
    fn options(source: CommandSource<'_>) -> Vec<String> {
        let mut players: Vec<_> = source.online_players().map(str::to_owned).collect();
        players.push("everyone".to_owned());
        players
    }
}

#[derive(CommandEnum)]
enum Style {
    Plain,
    Excited,
}

#[derive(Default)]
struct HelloCommand;

#[plugin]
impl Plugin for HelloCommand {
    #[command("hello say")]
    fn say(&self, event: &mut CommandEvent<'_>, style: Style, text: Varargs) {
        let message = match style {
            Style::Plain => format!("Hello, {}: {}", event.source(), text.value()),
            Style::Excited => format!(
                "HELLO, {}! {}",
                event.source().to_uppercase(),
                text.value().to_uppercase()
            ),
        };
        event.reply(&message).expect("command reply fits");
    }

    #[command("hello add")]
    fn add(&self, event: &mut CommandEvent<'_>, left: i64, right: i64) {
        event
            .reply(&format!("{}", left + right))
            .expect("reply fits");
    }

    #[command("hello toggle")]
    fn toggle(&self, event: &mut CommandEvent<'_>, enabled: bool) {
        event
            .reply(&format!("enabled={enabled}"))
            .expect("reply fits");
    }

    #[command("hello echo")]
    fn echo(&self, event: &mut CommandEvent<'_>, text: String) {
        event.reply(&text).expect("reply fits");
    }

    #[command("hello about")]
    fn about(&self, event: &mut CommandEvent<'_>) {
        event
            .reply("Hello from a Rust plugin running in Dragonfly.")
            .expect("reply fits");
    }

    #[command("hello greet")]
    fn greet(&self, event: &mut CommandEvent<'_>, target: Dynamic<GreetingTargets>) {
        event
            .reply(&format!("Greetings, {}!", target.value()))
            .expect("reply fits");
    }

    #[command("hello direct")]
    fn direct(&self, event: &mut CommandEvent<'_>, player: Player) {
        event
            .reply(&format!("player generation={}", player.id().generation()))
            .expect("reply fits");
    }

    #[command("hello maybe")]
    fn maybe(&self, event: &mut CommandEvent<'_>, value: Option<i64>) {
        event
            .reply(&format!("value={value:?}"))
            .expect("reply fits");
    }
}
