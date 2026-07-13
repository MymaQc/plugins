use std::time::Duration;

use dragonfly::{Context, Player, Plugin, Vec3, World, block, entity, item, plugin};

#[dragonfly::entity(network = "minecraft:armor_stand", width = 0.5, height = 1.975)]
struct Marker;

#[derive(Default)]
struct EntityCommand;

#[plugin]
#[command("entity")]
impl Plugin for EntityCommand {
    #[command]
    fn root(&self, context: &mut Context<'_, Player>) {
        context.source().message(
            "Use /entity text, /entity marker, /entity lightning, /entity tnt, /entity snowball, /entity sword, /entity sand, or /entity list.",
        );
    }

    #[subcommand("text")]
    fn text(&self, context: &mut Context<'_, Player>) {
        self.spawn(context, entity::Text::new("Native Rust entity"));
    }

    #[subcommand("marker")]
    fn marker(&self, context: &mut Context<'_, Player>) {
        self.spawn(context, Marker);
    }

    #[subcommand("lightning")]
    fn lightning(&self, context: &mut Context<'_, Player>) {
        self.spawn(context, entity::Lightning::new());
    }

    #[subcommand("tnt")]
    fn tnt(&self, context: &mut Context<'_, Player>) {
        self.spawn(context, entity::TNT::new(Duration::from_secs(4)));
    }

    #[subcommand("snowball")]
    fn snowball(&self, context: &mut Context<'_, Player>) {
        let player = context.source();
        let Some((world, position)) = world_and_position(player) else {
            player.message("World or position unavailable.");
            return;
        };
        let options = entity::SpawnOptions::new(above(position)).velocity(Vec3 {
            x: 0.0,
            y: 0.2,
            z: 1.5,
        });
        if world
            .spawn_entity(entity::Snowball::new(player.entity()), options)
            .is_some()
        {
            player.message("Snowball spawned.");
        } else {
            player.message("Could not spawn snowball.");
        }
    }

    #[subcommand("sword")]
    fn sword(&self, context: &mut Context<'_, Player>) {
        let sword = item::new(item::Sword::new(item::ToolTier::Diamond), 1);
        self.spawn(context, entity::DroppedItem::new(sword));
    }

    #[subcommand("sand")]
    fn sand(&self, context: &mut Context<'_, Player>) {
        self.spawn(
            context,
            entity::FallingBlock::new(block::new("minecraft:sand")),
        );
    }

    #[subcommand("list")]
    fn list(&self, context: &mut Context<'_, Player>) {
        let player = context.source();
        let Some(world) = player.entity().world() else {
            player.message("World unavailable.");
            return;
        };
        let Some(entities) = world.entities() else {
            context.source().message("Could not list entities.");
            return;
        };
        let players = world.players().map_or(0, |players| players.len());
        context.source().message(&format!(
            "{} entities, {} players.",
            entities.len(),
            players
        ));
    }
}

impl EntityCommand {
    fn spawn(&self, context: &mut Context<'_, Player>, descriptor: impl entity::Spawnable) {
        let player = context.source();
        let Some((world, position)) = world_and_position(player) else {
            player.message("World or position unavailable.");
            return;
        };
        if world
            .spawn_entity(descriptor, entity::SpawnOptions::new(above(position)))
            .is_some()
        {
            player.message("Entity spawned.");
        } else {
            player.message("Could not spawn entity.");
        }
    }
}

fn world_and_position(player: Player) -> Option<(World, Vec3)> {
    let entity = player.entity();
    Some((entity.world()?, entity.position()?))
}

fn above(mut position: Vec3) -> Vec3 {
    position.y += 1.5;
    position
}
