use dragonfly::{Enchantment, Event, Plugin, Title, Value, plugin};

#[derive(Default)]
struct LifecycleLogger;

#[plugin]
impl Plugin for LifecycleLogger {
    fn on_enable(&self) {
        eprintln!("lifecycle-logger enabled");
    }

    fn on_disable(&self) {
        eprintln!("lifecycle-logger disabled");
    }

    fn on_join(&self, event: &mut Event::PlayerJoin<'_>) {
        eprintln!("{} joined", event.name());
        let player = event.player();
        player.message("Welcome from a Rust plugin.");
        player.send_tip("Rust tip");
        player.send_popup("Rust popup");
        player.send_jukebox_popup("Rust jukebox popup");
        player.send_title(&Title::new("Rust plugin").subtitle("Native Dragonfly"));
    }

    fn on_quit(&self, event: &Event::PlayerQuit<'_>) {
        eprintln!("{} quit", event.name());
    }

    fn on_hurt(&self, event: &mut Event::PlayerHurt<'_>) {
        eprintln!(
            "player hurt for {} by {}",
            event.damage(),
            event.damage_source().name()
        );
    }

    fn on_heal(&self, event: &mut Event::PlayerHeal<'_>) {
        eprintln!(
            "player healed for {} by {}",
            event.health(),
            event.healing_source().name()
        );
    }

    fn on_block_break(&self, event: &mut Event::PlayerBlockBreak<'_>) {
        eprintln!("broke {} at {:?}", event.block(), event.position());
    }

    fn on_block_place(&self, event: &mut Event::PlayerBlockPlace<'_>) {
        eprintln!("placed {} at {:?}", event.block(), event.position());
    }

    fn on_food_loss(&self, event: &mut Event::PlayerFoodLoss<'_>) {
        eprintln!("food changed from {} to {}", event.from(), event.to());
    }

    fn on_death(&self, event: &mut Event::PlayerDeath<'_>) {
        eprintln!("player died from {}", event.damage_source().name());
    }

    fn on_start_break(&self, event: &mut Event::PlayerStartBreak<'_>) {
        eprintln!("started breaking at {:?}", event.position());
    }

    fn on_fire_extinguish(&self, event: &mut Event::PlayerFireExtinguish<'_>) {
        eprintln!("extinguished fire at {:?}", event.position());
    }
    fn on_toggle_sprint(&self, event: &mut Event::PlayerToggleSprint<'_>) {
        eprintln!("sprint={}", event.after());
    }
    fn on_toggle_sneak(&self, event: &mut Event::PlayerToggleSneak<'_>) {
        eprintln!("sneak={}", event.after());
    }
    fn on_jump(&self, event: &Event::PlayerJump<'_>) {
        if let Some(name) = event.player().name() {
            eprintln!("{name} jumped");
        }
    }
    fn on_teleport(&self, event: &mut Event::PlayerTeleport<'_>) {
        eprintln!("teleporting to {:?}", event.position());
    }
    fn on_experience_gain(&self, event: &mut Event::PlayerExperienceGain<'_>) {
        eprintln!("gaining {} experience", event.amount());
    }
    fn on_punch_air(&self, _event: &mut Event::PlayerPunchAir<'_>) {
        eprintln!("punched air");
    }
    fn on_held_slot_change(&self, event: &mut Event::PlayerHeldSlotChange<'_>) {
        eprintln!("held slot {} -> {}", event.from(), event.to());
    }
    fn on_sleep(&self, event: &mut Event::PlayerSleep<'_>) {
        eprintln!("sleeping, reminder={}", event.send_reminder());
    }
    fn on_block_pick(&self, event: &mut Event::PlayerBlockPick<'_>) {
        eprintln!("picked {} at {:?}", event.block(), event.position());
    }
    fn on_lectern_page_turn(&self, event: &mut Event::PlayerLecternPageTurn<'_>) {
        eprintln!("lectern page {} -> {}", event.old_page(), event.new_page());
    }
    fn on_sign_edit(&self, event: &mut Event::PlayerSignEdit<'_>) {
        eprintln!(
            "sign changed from {:?} to {:?}",
            event.old_text(),
            event.new_text()
        );
    }
    fn on_item_use(&self, _event: &mut Event::PlayerItemUse<'_>) {
        eprintln!("used an item");
    }
    fn on_item_use_on_block(&self, event: &mut Event::PlayerItemUseOnBlock<'_>) {
        eprintln!(
            "used item on {:?} face {:?}",
            event.position(),
            event.face()
        );
    }
    fn on_item_consume(&self, event: &mut Event::PlayerItemConsume<'_>) {
        let item = event.item();
        eprintln!("consumed {} x{}", item.identifier(), item.count());
        if item.custom_name() == "__snapshot_test__"
            && item.damage() == 7
            && item.unbreakable()
            && item.anvil_cost() == 4
            && item.lore() == ["one", "two"]
            && item.nbt().get("kind") == Some(&Value::String("event".to_owned()))
            && item.value("owner") == Some(&Value::String("rust".to_owned()))
            && item
                .enchantment(Enchantment::Sharpness)
                .is_some_and(|enchantment| enchantment.level() == 5)
        {
            event.cancel();
        }
    }
    fn on_item_release(&self, event: &mut Event::PlayerItemRelease<'_>) {
        eprintln!(
            "released {} after {:?}",
            event.item().identifier(),
            event.duration()
        );
    }
    fn on_item_damage(&self, event: &mut Event::PlayerItemDamage<'_>) {
        eprintln!(
            "damaging {} by {}",
            event.item().identifier(),
            event.damage()
        );
    }
    fn on_item_drop(&self, event: &mut Event::PlayerItemDrop<'_>) {
        eprintln!(
            "dropping {} x{}",
            event.item().identifier(),
            event.item().count()
        );
    }
    fn on_attack_entity(&self, event: &mut Event::PlayerAttackEntity<'_>) {
        eprintln!(
            "attacking {:?}, critical={}, knockback=({}, {})",
            event.target().id(),
            event.critical(),
            event.knockback_force(),
            event.knockback_height()
        );
    }
    fn on_item_use_on_entity(&self, event: &mut Event::PlayerItemUseOnEntity<'_>) {
        eprintln!("used item on {:?}", event.target().id());
    }

    fn on_change_world(&self, event: &Event::PlayerChangeWorld<'_>) {
        eprintln!(
            "player {:?} changed world from {:?} to {:?}",
            event.player().id(),
            event.before(),
            event.after()
        );
    }
}
