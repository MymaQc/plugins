use dragonfly::{
    CommandEnum, CommandSource, Context, Dynamic, DynamicCommandEnum, Player, Plugin, Source,
    Varargs, plugin,
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
#[command("hello")]
impl Plugin for HelloCommand {
    #[command]
    fn root(&self, context: &mut Context<'_>) {
        context.source().message("Use a hello subcommand.");
    }

    #[subcommand("say")]
    fn say(&self, context: &mut Context<'_>, style: Style, text: Varargs) {
        let message = match style {
            Style::Plain => format!("Hello, {}: {}", context.source().name(), text.value()),
            Style::Excited => format!(
                "HELLO, {}! {}",
                context.source().name().to_uppercase(),
                text.value().to_uppercase()
            ),
        };
        context.source().message(&message);
    }

    #[subcommand("add")]
    fn add(&self, context: &mut Context<'_>, left: i64, right: i64) {
        context.source().message(&format!("{}", left + right));
    }

    #[subcommand("toggle")]
    fn toggle(&self, context: &mut Context<'_>, enabled: bool) {
        context.source().message(&format!("enabled={enabled}"));
    }

    #[subcommand("echo")]
    fn echo(&self, context: &mut Context<'_>, text: String) {
        context.source().message(&text);
    }

    #[subcommand("about")]
    fn about(&self, context: &mut Context<'_>) {
        context
            .source()
            .message("Hello from a Rust plugin running in Dragonfly.");
    }

    #[subcommand("greet")]
    fn greet(&self, context: &mut Context<'_>, target: Dynamic<GreetingTargets>) {
        context
            .source()
            .message(&format!("Greetings, {}!", target.value()));
    }

    #[subcommand("direct")]
    fn direct(&self, context: &mut Context<'_>, player: Player) {
        context
            .source()
            .message(&format!("player generation={}", player.id().generation()));
    }

    #[subcommand("maybe")]
    fn maybe(&self, context: &mut Context<'_>, value: Option<i64>) {
        context.source().message(&format!("value={value:?}"));
    }
}
