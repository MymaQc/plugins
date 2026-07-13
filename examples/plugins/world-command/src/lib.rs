use dragonfly::{
    BlockPos, Context, Dimension, OpenMode, Player, Plugin, RandomTicks, SavePolicy, TimePolicy,
    WeatherPolicy, World, WorldSpec, block, plugin,
};

#[derive(Default)]
struct WorldCommand;

#[plugin]
#[command("world")]
impl Plugin for WorldCommand {
    #[command]
    fn root(&self, context: &mut Context<'_, Player>) {
        context
            .source()
            .message("Use /world inspect, /world set-stone, /world open, or /world open-spec.");
    }

    #[subcommand("inspect")]
    fn inspect(&self, context: &mut Context<'_, Player>, x: i64, y: i64, z: i64) {
        let Some(position) = position(x, y, z) else {
            context.source().message("Coordinates must fit i32.");
            return;
        };
        let Some(world) = World::overworld() else {
            context.source().message("Overworld is unavailable.");
            return;
        };
        let Some(block) = world.block(position) else {
            context.source().message("Could not read block.");
            return;
        };
        context
            .source()
            .message(&format!("{} {:?}", block.identifier(), block.properties()));
    }

    #[subcommand("set-stone")]
    fn set(&self, context: &mut Context<'_, Player>, x: i64, y: i64, z: i64) {
        let Some(position) = position(x, y, z) else {
            context.source().message("Coordinates must fit i32.");
            return;
        };
        let Some(world) = World::overworld() else {
            context.source().message("Overworld is unavailable.");
            return;
        };
        world.set_block(position, block::Stone);
        context.source().message("Block queued.");
    }

    #[subcommand("open")]
    fn open(&self, context: &mut Context<'_, Player>, name: String) {
        let Some(world) = World::open(&name, Dimension::Overworld) else {
            context.source().message("Could not open world.");
            return;
        };
        world.set_time(6000);
        world.set_spawn(BlockPos { x: 0, y: 64, z: 0 });
        context.source().message(&format!("Opened {name}."));
    }

    #[subcommand("open-spec")]
    fn open_spec(&self, context: &mut Context<'_, Player>, name: String) {
        let spec = WorldSpec::persistent("examples/managed")
            .open_mode(OpenMode::OpenOrCreate)
            .save(SavePolicy::Manual)
            .random_ticks(RandomTicks::Disabled)
            .time(TimePolicy::Fixed(6000))
            .weather(WeatherPolicy::Clear);
        let Some(_world) = World::open_with(&name, &spec) else {
            context.source().message("Could not open specified world.");
            return;
        };
        context
            .source()
            .message(&format!("Opened {name} from a typed specification."));
    }
}

fn position(x: i64, y: i64, z: i64) -> Option<BlockPos> {
    Some(BlockPos {
        x: i32::try_from(x).ok()?,
        y: i32::try_from(y).ok()?,
        z: i32::try_from(z).ok()?,
    })
}
