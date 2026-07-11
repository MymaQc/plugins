use dragonfly_plugin::{Command, CommandEnum, CommandEvent, Plugin, plugin};

#[derive(CommandEnum)]
enum Style {
    Plain,
    Excited,
}

#[derive(Command)]
#[command(name = "hello", description = "Greets the command source")]
enum Hello {
    Say { style: Style },
    About,
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
            } => format!("Hello, {}.", event.source()),
            Hello::Say {
                style: Style::Excited,
            } => format!("HELLO, {}!", event.source().to_uppercase()),
            Hello::About => "Hello from a Rust plugin running in Dragonfly.".to_owned(),
        };
        event
            .reply(&message)
            .expect("command reply fits the runtime output buffer");
    }
}
