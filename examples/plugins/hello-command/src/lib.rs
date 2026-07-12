use dragonfly::{
    CommandEnum, CommandSource, Context, Dynamic, DynamicCommandEnum, Player, Plugin, Rotation,
    Varargs, Vec3, plugin,
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
        context.reply("Use a hello subcommand.");
    }

    #[subcommand("say")]
    fn say(&self, context: &mut Context<'_>, style: Style, text: Varargs) {
        let message = match style {
            Style::Plain => format!("Hello, {}: {}", context.source_name(), text.value()),
            Style::Excited => format!(
                "HELLO, {}! {}",
                context.source_name().to_uppercase(),
                text.value().to_uppercase()
            ),
        };
        context.reply(&message);
    }

    #[subcommand("add")]
    fn add(&self, context: &mut Context<'_>, left: i64, right: i64) {
        context.reply(&format!("{}", left + right));
    }

    #[subcommand("toggle")]
    fn toggle(&self, context: &mut Context<'_>, enabled: bool) {
        context.reply(&format!("enabled={enabled}"));
    }

    #[subcommand("echo")]
    fn echo(&self, context: &mut Context<'_>, text: String) {
        context.reply(&text);
    }

    #[subcommand("about")]
    fn about(&self, context: &mut Context<'_>) {
        context.reply("Hello from a Rust plugin running in Dragonfly.");
    }

    #[subcommand("greet")]
    fn greet(&self, context: &mut Context<'_>, target: Dynamic<GreetingTargets>) {
        context.reply(&format!("Greetings, {}!", target.value()));
    }

    #[subcommand("direct")]
    fn direct(&self, context: &mut Context<'_>, player: Player) {
        context.reply(&format!("player generation={}", player.id().generation()));
    }

    #[subcommand("maybe")]
    fn maybe(&self, context: &mut Context<'_>, value: Option<i64>) {
        context.reply(&format!("value={value:?}"));
    }

    #[subcommand("teleport")]
    fn teleport(&self, context: &mut Context<'_, Player>, x: f64, y: f64, z: f64) {
        context.source().teleport(Vec3 { x, y, z });
        context.reply("Teleported.");
    }

    #[subcommand("move")]
    fn move_by(&self, context: &mut Context<'_, Player>, x: f64, y: f64, z: f64) {
        context.source().move_by(Vec3 { x, y, z }, 0.0, 0.0);
        context.reply("Moved.");
    }

    #[subcommand("velocity")]
    fn velocity(&self, context: &mut Context<'_, Player>, x: f64, y: f64, z: f64) {
        context.source().set_velocity(Vec3 { x, y, z });
        context.reply("Velocity set.");
    }

    #[subcommand("rotation")]
    fn rotation(&self, context: &mut Context<'_, Player>) {
        let rotation = context.source().rotation();
        context.reply(&format!("yaw={} pitch={}", rotation.yaw, rotation.pitch));
    }

    #[subcommand("face")]
    fn face(&self, context: &mut Context<'_, Player>, yaw: f64, pitch: f64) {
        context.source().face(Rotation { yaw, pitch });
        context.reply("Rotation set.");
    }
}
