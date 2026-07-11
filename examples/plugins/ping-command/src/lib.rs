use dragonfly::{Context, Player, Plugin, plugin};

#[plugin]
#[command("ping")]
impl Plugin for PingCommand {
    #[command]
    fn ping(&self, context: &mut Context<'_, Player>, target: Option<Player>) {
        let player = target.unwrap_or_else(|| context.source());
        let Some(latency) = player.latency() else {
            context.fail("Player is no longer online.");
            return;
        };
        context.reply(&format!(
            "{}'s ping: {}ms",
            player.name().unwrap_or("unknown"),
            latency.as_millis()
        ));
    }
}

#[derive(Default)]
struct PingCommand;
