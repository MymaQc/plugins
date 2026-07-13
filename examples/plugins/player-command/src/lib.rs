use dragonfly::{
    CommandEnum, Context, GameMode, Player, Plugin, Rotation, Varargs, Vec3, block, damage, effect,
    healing, plugin, sound,
};

#[derive(CommandEnum)]
enum Mode {
    Survival,
    Creative,
    Adventure,
    Spectator,
}

impl From<Mode> for GameMode {
    fn from(value: Mode) -> Self {
        match value {
            Mode::Survival => Self::Survival,
            Mode::Creative => Self::Creative,
            Mode::Adventure => Self::Adventure,
            Mode::Spectator => Self::Spectator,
        }
    }
}

#[derive(Default)]
struct PlayerCommand;

#[plugin]
#[command("player")]
impl Plugin for PlayerCommand {
    #[command]
    fn root(&self, context: &mut Context<'_, Player>) {
        context.reply("Use a player subcommand.");
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

    #[subcommand("gamemode")]
    fn game_mode(&self, context: &mut Context<'_, Player>, mode: Mode) {
        context.source().set_game_mode(mode.into());
        context.reply("Game mode set.");
    }

    #[subcommand("heal")]
    fn heal(&self, context: &mut Context<'_, Player>, amount: f64) {
        let healed = context.source().heal(amount, healing::Instant);
        context.reply(&format!(
            "Healed: {healed}, health: {}",
            context.source().health()
        ));
    }

    #[subcommand("hurt")]
    fn hurt(&self, context: &mut Context<'_, Player>, amount: f64) {
        let (damage, vulnerable) = context.source().hurt(amount, damage::Instant);
        context.reply(&format!(
            "Damage: {damage}, vulnerable: {vulnerable}, health: {}",
            context.source().health()
        ));
    }

    #[subcommand("heal-food")]
    fn heal_food(&self, context: &mut Context<'_, Player>, amount: f64, quick: bool) {
        let healed = context.source().heal(amount, healing::Food::new(quick));
        context.reply(&format!("Healed: {healed}"));
    }

    #[subcommand("hurt-attack")]
    fn hurt_attack(&self, context: &mut Context<'_, Player>, amount: f64, attacker: Player) {
        let result = context
            .source()
            .hurt(amount, damage::Attack::new(attacker.entity()));
        context.reply(&format!("Damage: {}", result.0));
    }

    #[subcommand("hurt-projectile")]
    fn hurt_projectile(&self, context: &mut Context<'_, Player>, amount: f64, projectile: Player) {
        let result = context
            .source()
            .hurt(amount, damage::Projectile::new(projectile.entity(), None));
        context.reply(&format!("Damage: {}", result.0));
    }

    #[subcommand("hurt-block")]
    fn hurt_block(&self, context: &mut Context<'_, Player>, amount: f64) {
        let cactus = block::new("minecraft:cactus").with_property("age", 4i32);
        let result = context.source().hurt(amount, damage::Block::new(cactus));
        context.reply(&format!("Damage: {}", result.0));
    }

    #[subcommand("hurt-custom")]
    fn hurt_custom(&self, context: &mut Context<'_, Player>, amount: f64) {
        let source = damage::Custom::new(
            "example:magic",
            damage::Traits::new(true, false, false, true),
            damage::AffectedProtections::FIRE | damage::AffectedProtections::BLAST,
        );
        let result = context.source().hurt(amount, source);
        context.reply(&format!("Damage: {}", result.0));
    }

    #[subcommand("food")]
    fn food(&self, context: &mut Context<'_, Player>, level: i64) {
        context.source().set_food(level as i32);
        context.reply(&format!("Food: {}", context.source().food()));
    }

    #[subcommand("max-health")]
    fn max_health(&self, context: &mut Context<'_, Player>, health: f64) {
        context.source().set_max_health(health);
        context.reply(&format!("Max health: {}", context.source().max_health()));
    }

    #[subcommand("experience-level")]
    fn experience_level(&self, context: &mut Context<'_, Player>, level: i32) {
        context.source().set_experience_level(level);
        context.reply(&format!("Level: {}", context.source().experience_level()));
    }

    #[subcommand("experience-progress")]
    fn experience_progress(&self, context: &mut Context<'_, Player>, progress: f64) {
        context.source().set_experience_progress(progress);
        context.reply(&format!(
            "Progress: {}",
            context.source().experience_progress()
        ));
    }

    #[subcommand("speed")]
    fn speed(&self, context: &mut Context<'_, Player>, level: i32, seconds: i32) {
        let duration = std::time::Duration::from_secs(seconds.max(0) as u64);
        context
            .source()
            .add_effect(effect::new(effect::Speed, level, duration));
        context.reply("Speed effect added.");
    }

    #[subcommand("clear-speed")]
    fn clear_speed(&self, context: &mut Context<'_, Player>) {
        context.source().remove_effect(effect::Speed);
        context.reply("Speed effect removed.");
    }

    #[subcommand("instant-health")]
    fn instant_health(&self, context: &mut Context<'_, Player>, level: i32, potency: f64) {
        context.source().add_effect(effect::instant_with_potency(
            effect::InstantHealth,
            level,
            potency,
        ));
        context.reply("Instant health applied.");
    }

    #[subcommand("name-tag")]
    fn name_tag(&self, context: &mut Context<'_, Player>, name: Varargs) {
        context.source().set_name_tag(name.value());
        context.reply("Name tag set.");
    }

    #[subcommand("scale")]
    fn scale(&self, context: &mut Context<'_, Player>, scale: f64) {
        context.source().set_scale(scale);
        context.reply(&format!("Scale: {}", context.source().scale()));
    }

    #[subcommand("invisible")]
    fn invisible(&self, context: &mut Context<'_, Player>, invisible: bool) {
        context.source().set_invisible(invisible);
        context.reply(&format!("Invisible: {}", context.source().invisible()));
    }

    #[subcommand("immobile")]
    fn immobile(&self, context: &mut Context<'_, Player>, immobile: bool) {
        context.source().set_immobile(immobile);
        context.reply(&format!("Immobile: {}", context.source().immobile()));
    }

    #[subcommand("sound")]
    fn sound(&self, context: &mut Context<'_, Player>) {
        context.source().play_sound(sound::LevelUp);
        context.reply("Played level-up sound.");
    }

    #[subcommand("disconnect")]
    fn disconnect(&self, context: &mut Context<'_, Player>) {
        context.source().disconnect("Disconnected by Rust plugin.");
    }

    #[subcommand("kick")]
    fn kick(&self, context: &mut Context<'_, Player>) {
        context.source().kick("Kicked by Rust plugin.");
    }

    #[subcommand("hide")]
    fn hide(&self, context: &mut Context<'_, Player>, target: Player) {
        context.source().hide_entity(target.entity());
        context.reply("Entity hidden.");
    }

    #[subcommand("show")]
    fn show(&self, context: &mut Context<'_, Player>, target: Player) {
        context.source().show_entity(target.entity());
        context.reply("Entity shown.");
    }

    #[subcommand("skin-copy")]
    fn skin_copy(&self, context: &mut Context<'_, Player>) {
        let skin = context.source().skin();
        context.source().set_skin(&skin);
        context.reply("Skin copied.");
    }
}
