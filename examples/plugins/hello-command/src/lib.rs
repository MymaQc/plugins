use dragonfly_plugin::{
    Command, CommandEnum, CommandEvent, CommandSource, Dynamic, DynamicCommandEnum, Plugin, plugin,
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

#[derive(Command)]
#[command(name = "hello", description = "Greets the command source")]
enum Hello {
    Say { style: Style, text: String },
    Add { left: i64, right: i64 },
    Toggle { enabled: bool },
    Echo { text: String },
    About,
    Greet { target: Dynamic<GreetingTargets> },
}

#[derive(Default)]
struct HelloCommand;

#[plugin]
impl Plugin for HelloCommand {
    #[command]
    fn hello(&self, event: &mut CommandEvent<'_>, command: Hello) {
        let message = match command {
            Hello::Say {
                style: Style::Plain,
                text,
            } => format!("Hello, {}: {text}", event.source()),
            Hello::Say {
                style: Style::Excited,
                text,
            } => format!(
                "HELLO, {}! {}",
                event.source().to_uppercase(),
                text.to_uppercase()
            ),
            Hello::Add { left, right } => format!("{}", left + right),
            Hello::Toggle { enabled } => format!("enabled={enabled}"),
            Hello::Echo { text } => text,
            Hello::About => "Hello from a Rust plugin running in Dragonfly.".to_owned(),
            Hello::Greet { target } => format!("Greetings, {}!", target.value()),
        };
        event
            .reply(&message)
            .expect("command reply fits the runtime output buffer");
    }
}
