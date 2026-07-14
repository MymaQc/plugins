#nullable enable
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading;
using Dragonfly;

public sealed class KitchenSink : Plugin
{
    private long _jumps;
    private long _punches;
    private long _quits;
    private long _sneaks;
    private long _sprints;
    private long _transfers;
    private long _commandExecutions;
    private long _diagnostics;

    public override void OnEnable()
    {
        Cmd.Register(Cmd.New(
            "kitchen",
            "Exercises the reflected C# command API.",
            ["ks"],
            new KitchenStatus(this),
            new KitchenEcho(),
            new KitchenMode(),
            new KitchenPing(),
            new KitchenPosition(),
            new KitchenDestination(),
            new KitchenText(),
            new KitchenBlock(),
            new KitchenBiome(),
            new KitchenTick(),
            new KitchenParticle(),
            new KitchenGameMode(),
            new KitchenState(),
            new KitchenItem(),
            new KitchenForm(),
            new KitchenRawFormCommand(),
            new KitchenEffect(),
            new KitchenCrop(),
            new KitchenKinematics()));
        Console.WriteLine("kitchen-sink enabled");
    }

    public override void OnDisable() => Console.WriteLine(
        $"kitchen-sink disabled: jumps={_jumps}, punches={_punches}, " +
        $"sprints={_sprints}, sneaks={_sneaks}, quits={_quits}");

    public override void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot)
    {
        if (!Finite(newPos) || !double.IsFinite(newRot.Yaw) || !double.IsFinite(newRot.Pitch))
            ctx.Cancel();
    }

    public override void HandleJump(Player player) => Increment(ref _jumps);

    public override void HandleTeleport(Player.Context ctx, Vector3 pos)
    {
        if (!Finite(pos)) ctx.Cancel();
    }

    public override void HandleChangeWorld(Player player, World? before, World after) =>
        _ = (player, before, after);

    public override void HandleToggleSprint(Player.Context ctx, bool sprinting)
    {
        if (sprinting) Increment(ref _sprints);
    }

    public override void HandleToggleSneak(Player.Context ctx, bool sneaking)
    {
        if (sneaking) Increment(ref _sneaks);
    }

    public override void HandleChat(Player.Context ctx, ref string message) =>
        message = message.Trim();

    public override void HandleFoodLoss(Player.Context ctx, int from, ref int to) =>
        to = Math.Clamp(to, 0, 20);

    public override void HandleHeal(Player.Context ctx, ref double health, World.HealingSource source)
    {
        health = Math.Max(0, health);
        _ = source;
    }

    public override void HandleHurt(
        Player.Context ctx,
        ref double damage,
        bool immune,
        ref TimeSpan attackImmunity,
        World.DamageSource source)
    {
        damage = Math.Max(0, damage);
        _ = (immune, attackImmunity, source.ReducedByArmour(), source.ReducedByResistance(),
            source.Fire(), source.IgnoreTotem());
    }

    public override void HandleDeath(Player player, World.DamageSource source, ref bool keepInv) =>
        _ = (player, source, keepInv);

    public override void HandleRespawn(Player player, ref Vector3 pos, ref World world) =>
        _ = (player, pos, world);

    public override void HandleSkinChange(Player.Context ctx, ref Skin skin) =>
        _ = (skin.Bounds(), skin.Pix, skin.ModelConfig, skin.Model, skin.Cape, skin.Animations);

    public override void HandleFireExtinguish(Player.Context ctx, Cube.Pos pos) => _ = pos;
    public override void HandleStartBreak(Player.Context ctx, Cube.Pos pos) => _ = pos;

    public override void HandleBlockBreak(
        Player.Context ctx,
        Cube.Pos pos,
        ref Item.Stack[] drops,
        ref int xp)
    {
        drops = [.. drops.Where(drop => !drop.Empty())];
        xp = Math.Max(0, xp);
    }

    public override void HandleBlockPlace(Player.Context ctx, Cube.Pos pos, World.Block block) =>
        _ = (pos, block);

    public override void HandleBlockPick(Player.Context ctx, Cube.Pos pos, World.Block block) =>
        _ = (pos, block);

    public override void HandleItemUse(Player.Context ctx)
    {
        if (!Finite(ctx.Player().Position())) ctx.Cancel();
    }

    public override void HandleItemUseOnBlock(
        Player.Context ctx,
        Cube.Pos pos,
        Cube.Face face,
        Vector3 clickPos) => _ = (pos, face, clickPos);

    public override void HandleItemUseOnEntity(Player.Context ctx, World.Entity entity)
    {
        if (!Finite(entity.Position()) || !Finite(entity.Rotation()))
        {
            entity.Close();
            ctx.Cancel();
        }
        _ = entity.H();
    }

    public override void HandleItemRelease(Player.Context ctx, Item.Stack item, TimeSpan duration) =>
        _ = (item, duration);

    public override void HandleItemConsume(Player.Context ctx, Item.Stack item) => _ = item;

    public override void HandleAttackEntity(
        Player.Context ctx,
        World.Entity entity,
        ref double force,
        ref double height,
        ref bool critical)
    {
        force = Math.Max(0, force);
        height = Math.Max(0, height);
        _ = (entity, critical);
    }
    public override void HandleExperienceGain(Player.Context ctx, ref int amount) => amount = Math.Max(0, amount);

    public override void HandleSignEdit(
        Player.Context ctx,
        Cube.Pos pos,
        bool frontSide,
        string oldText,
        string newText) => _ = (pos, frontSide, oldText, newText);

    public override void HandleSleep(Player.Context ctx, ref bool sendReminder) { }

    public override void HandleLecternPageTurn(
        Player.Context ctx,
        Cube.Pos pos,
        int oldPage,
        ref int newPage) => newPage = Math.Max(0, newPage);

    public override void HandleItemDamage(Player.Context ctx, Item.Stack item, ref int damage) =>
        damage = Math.Max(0, damage);

    public override void HandleItemPickup(Player.Context ctx, ref Item.Stack item) =>
        item = item.WithCustomName(item.CustomName().Trim());
    public override void HandleHeldSlotChange(Player.Context ctx, int from, int to) => _ = (from, to);
    public override void HandleItemDrop(Player.Context ctx, Item.Stack item) => _ = item;

    public override void HandleTransfer(Player.Context ctx, ref Net.UDPAddr address)
    {
        address.Port = Math.Clamp(address.Port, 0, 65535);
        Increment(ref _transfers);
    }

    public override void HandleCommandExecution(Player.Context ctx, Cmd.Command command, string[] args)
    {
        for (var index = 0; index < args.Length; index++) args[index] = args[index].Trim();
        _ = (command.Name(), command.Description(), command.Usage(), command.Aliases());
        Increment(ref _commandExecutions);
    }

    public override void HandlePunchAir(Player.Context ctx) => Increment(ref _punches);
    public override void HandleQuit(Player player) => Increment(ref _quits);

    public override void HandleDiagnostics(Player player, Session.Diagnostics diagnostics)
    {
        _ = (player, diagnostics.AverageFramesPerSecond, diagnostics.AverageServerSimTickTime,
            diagnostics.AverageClientSimTickTime, diagnostics.AverageBeginFrameTime,
            diagnostics.AverageInputTime, diagnostics.AverageRenderTime,
            diagnostics.AverageEndFrameTime, diagnostics.AverageRemainderTimePercent,
            diagnostics.AverageUnaccountedTimePercent);
        Increment(ref _diagnostics);
    }

    private static void Increment(ref long counter) => Interlocked.Increment(ref counter);

    private static bool Finite(Vector3 value) =>
        double.IsFinite(value.X) && double.IsFinite(value.Y) && double.IsFinite(value.Z);

    private static bool Finite(Rotation value) =>
        double.IsFinite(value.Yaw) && double.IsFinite(value.Pitch);

    internal sealed class KitchenStatus(KitchenSink plugin) : Cmd.Runnable
    {
        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx) => output.Printf(
            "jumps={0}, punches={1}, sprints={2}, sneaks={3}, quits={4}",
            plugin._jumps,
            plugin._punches,
            plugin._sprints,
            plugin._sneaks,
            plugin._quits);
    }

    internal sealed class KitchenEcho : Cmd.Runnable
    {
        public Cmd.SubCommand Echo;
        public Cmd.Varargs Message;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx) => output.Print(Message);
    }

    internal enum Mode
    {
        Survival,
        Creative,
        Adventure,
        Spectator,
    }

    internal sealed class KitchenMode : Cmd.Runnable
    {
        public Cmd.SubCommand Mode;
        public Mode Value;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx) => output.Printf("mode={0}", Value);
    }

    internal sealed class KitchenPing : Cmd.Runnable
    {
        public Cmd.SubCommand Ping;
        public Cmd.Optional<Player> Target;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            var (target, _) = Target.Load();
            target ??= source as Player;
            if (target is null)
            {
                output.Error("A player target is required from the console.");
                return;
            }
            output.Printf("{0}'s ping: {1}ms", target.Name(), target.Latency().TotalMilliseconds);
        }
    }

    internal sealed class KitchenPosition : Cmd.Runnable
    {
        public Cmd.SubCommand Position;
        public Vector3 Value;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx) =>
            output.Printf("position={0},{1},{2}", Value.X, Value.Y, Value.Z);
    }

    internal sealed class KitchenKinematics : Cmd.Runnable
    {
        public Cmd.SubCommand Kinematics;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }
            var position = player.Position();
            var velocity = player.Velocity();
            var rotation = player.Rotation();
            player.Teleport(position);
            player.Move(default, 0, 0);
            player.Displace(default);
            player.SetVelocity(velocity);
            output.Printf(
                "position={0},{1},{2}, velocity={3},{4},{5}, rotation={6},{7}",
                position.X, position.Y, position.Z,
                velocity.X, velocity.Y, velocity.Z,
                rotation.Yaw, rotation.Pitch);
        }
    }

    internal sealed class KitchenDestination : Cmd.Runnable
    {
        public Cmd.SubCommand Destination;
        public Destination Value = new();

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx) => output.Printf("destination={0}", Value);
    }

    internal sealed class Destination(string value = "spawn") : Cmd.Enum
    {
        public string Type() => "kitchen_destination";
        public IReadOnlyList<string> Options(Cmd.Source source) => ["spawn", "source"];
        public override string ToString() => value;
    }

    internal enum TextAction
    {
        Message,
        Popup,
        Tip,
        Jukebox,
        NameTag,
        Disconnect,
    }

    internal sealed class KitchenText : Cmd.Runnable
    {
        [Cmd.Tag("text")]
        public Cmd.SubCommand Text;
        [Cmd.Tag("action")]
        public TextAction Action;
        public Cmd.Varargs Content;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }
            switch (Action)
            {
                case TextAction.Message:
                    player.Message(Content, true, 12, 1.5, null);
                    break;
                case TextAction.Popup:
                    player.SendPopup(Content);
                    break;
                case TextAction.Tip:
                    player.SendTip(Content);
                    break;
                case TextAction.Jukebox:
                    player.SendJukeboxPopup(Content);
                    break;
                case TextAction.NameTag:
                    player.SetNameTag(Content);
                    break;
                case TextAction.Disconnect:
                    player.Disconnect(Content);
                    break;
            }
        }
    }

    internal sealed class KitchenBlock : Cmd.Runnable
    {
        public Cmd.SubCommand Block;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (tx is null)
            {
                output.Error("A world transaction is required.");
                return;
            }
            var position = Cube.PosFromVec3(source.Position()).Side(Cube.Face.Down);
            var range = tx.Range();
            var (block, loaded) = tx.BlockLoaded(position);
            var previous = loaded ? block : tx.Block(position);
            var wasSand = previous is Block.Sand;
            var nearbySand = tx.BlocksWithin(position, 2, new Block.Sand());
            var highestLightBlocker = tx.HighestLightBlocker(position.X(), position.Z());
            var highestBlock = tx.HighestBlock(position.X(), position.Z());
            var light = tx.Light(position);
            var skyLight = tx.SkyLight(position);
            var (_, liquidBeforeFound) = tx.Liquid(position);
            tx.SetBlock(position, new Block.Sand(), new World.SetOpts
            {
                DisableBlockUpdates = true,
                DisableLiquidDisplacement = true,
                DisableRedstoneUpdates = true,
            });
            var scheduledWater = new Block.Water(Still: true, Depth: 8, Falling: false);
            tx.SetLiquid(position, scheduledWater);
            var (liquid, liquidFound) = tx.Liquid(position);
            var liquidState = liquid is Block.Water water
                ? $"Water(still={(water.Still ? "true" : "false")},depth={water.Depth}," +
                  $"falling={(water.Falling ? "true" : "false")})"
                : liquid?.GetType().Name ?? "none";
            tx.SetLiquid(position, null);
            tx.SetLiquid(position, scheduledWater);
            var blockUpdateDelay = TimeSpan.FromMilliseconds(250);
            tx.ScheduleBlockUpdate(position, scheduledWater, blockUpdateDelay);
            var firstNearbySand = "none";
            foreach (var nearbyPosition in nearbySand)
            {
                firstNearbySand = nearbyPosition.ToString();
                break;
            }
            output.Printf(
                "block={0}, range={1}..{2}, loaded={3}, was_sand={4}, nearby_sand={5}, " +
                "highest_light_blocker={6}, highest_block={7}, light={8}, sky_light={9}, " +
                "liquid_before={10}, liquid={11}:{12}, scheduled_update=water:{13}ms",
                position,
                range.Min(),
                range.Max(),
                loaded ? "true" : "false",
                wasSand ? "true" : "false",
                firstNearbySand,
                highestLightBlocker,
                highestBlock,
                light,
                skyLight,
                liquidBeforeFound ? "true" : "false",
                liquidFound ? "true" : "false",
                liquidState,
                blockUpdateDelay.TotalMilliseconds);
        }
    }

    internal sealed class KitchenCrop : Cmd.Runnable
    {
        public Cmd.SubCommand Crop;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (tx is null)
            {
                output.Error("A world transaction is required.");
                return;
            }
            var position = Cube.PosFromVec3(source.Position()).Side(Cube.Face.Down);
            var current = tx.Block(position);
            var currentGrowth = current is Block.WheatSeeds wheat ? wheat.Growth : -1;
            tx.SetBlock(position, new Block.WheatSeeds(7));
            output.Printf("crop={0}, planted=7", currentGrowth);
        }
    }

    internal sealed class KitchenBiome : Cmd.Runnable
    {
        public Cmd.SubCommand Biome;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (tx is null)
            {
                output.Error("A world transaction is required.");
                return;
            }
            var position = Cube.PosFromVec3(source.Position());
            var previous = tx.Biome(position);
            World.Biome current = previous;
            var temperature = 0.0;
            var rainingAt = false;
            var snowingAt = false;
            var thunderingAt = false;
            var raining = false;
            var thundering = false;
            tx.SetBiome(position, new Biome.Desert());
            try
            {
                current = tx.Biome(position);
                temperature = tx.Temperature(position);
                rainingAt = tx.RainingAt(position);
                snowingAt = tx.SnowingAt(position);
                thunderingAt = tx.ThunderingAt(position);
                raining = tx.Raining();
                thundering = tx.Thundering();
            }
            finally
            {
                tx.SetBiome(position, previous);
            }
            var restored = tx.Biome(position);
            output.Printf(
                "biome=Desert, applied={0}, temperature={1}, raining_at={2}, snowing_at={3}, " +
                "thundering_at={4}, raining={5}, thundering={6}, restored={7}",
                current is Biome.Desert ? "true" : "false",
                temperature,
                rainingAt ? "true" : "false",
                snowingAt ? "true" : "false",
                thunderingAt ? "true" : "false",
                raining ? "true" : "false",
                thundering ? "true" : "false",
                restored.Equals(previous) ? "true" : "false");
        }
    }

    internal sealed class KitchenTick : Cmd.Runnable
    {
        public Cmd.SubCommand Tick;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (tx is null)
            {
                output.Error("A world transaction is required.");
                return;
            }
            output.Printf("tick={0}", tx.CurrentTick());
        }
    }

    internal sealed class KitchenParticle : Cmd.Runnable
    {
        public Cmd.SubCommand Particle;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (tx is null)
            {
                output.Error("A world transaction is required.");
                return;
            }
            World.Particle[] particles =
            [
                new Particle.Flame(new Color.RGBA(1, 2, 3, 4)),
                new Particle.Dust(new Color.RGBA(5, 6, 7, 8)),
                new Particle.BlockBreak(new Block.Sand()),
                new Particle.PunchBlock(new Block.Sand(), Cube.Face.East),
                new Particle.BlockForceField(),
                new Particle.BoneMeal(true),
                new Particle.Note(Sound.Piano(), 24),
                new Particle.Note(Sound.BassDrum(), 24),
                new Particle.Note(Sound.Snare(), 24),
                new Particle.Note(Sound.ClicksAndSticks(), 24),
                new Particle.Note(Sound.Bass(), 24),
                new Particle.Note(Sound.Flute(), 24),
                new Particle.Note(Sound.Bell(), 24),
                new Particle.Note(Sound.Guitar(), 24),
                new Particle.Note(Sound.Chimes(), 24),
                new Particle.Note(Sound.Xylophone(), 24),
                new Particle.Note(Sound.IronXylophone(), 24),
                new Particle.Note(Sound.CowBell(), 24),
                new Particle.Note(Sound.Didgeridoo(), 24),
                new Particle.Note(Sound.Bit(), 24),
                new Particle.Note(Sound.Banjo(), 24),
                new Particle.Note(Sound.Pling(), 24),
                new Particle.DragonEggTeleport(new Cube.Pos(-3, 4, 5)),
                new Particle.Evaporate(),
                new Particle.WaterDrip(),
                new Particle.LavaDrip(),
                new Particle.Lava(),
                new Particle.DustPlume(),
                new Particle.HugeExplosion(),
                new Particle.EndermanTeleport(),
                new Particle.SnowballPoof(),
                new Particle.EggSmash(),
                new Particle.Splash(new Color.RGBA(9, 10, 11, 12)),
                new Particle.Effect(new Color.RGBA(13, 14, 15, 16)),
                new Particle.EntityFlame(),
            ];
            foreach (var particle in particles)
                tx.AddParticle(source.Position(), particle);
            output.Printf("particles={0}", particles.Length);
        }
    }

    internal sealed class KitchenGameMode : Cmd.Runnable
    {
        [Cmd.Tag("game-mode")]
        public Cmd.SubCommand GameMode;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }
            var current = player.GameMode();
            var (id, registered) = World.GameModeID(current);
            var (roundTrip, found) = World.GameModeByID(id);
            var (roundTripId, roundTripRegistered) = World.GameModeID(roundTrip);
            var custom = new CustomGameMode();
            var (_, customRegistered) = World.GameModeID(custom);
            player.SetGameMode(custom);
            player.SetGameMode(current);
            output.Printf(
                "game_mode_id={0}, registered={1}, round_trip={2}, custom_registered={3}",
                id,
                registered ? "true" : "false",
                registered && found && roundTripRegistered && roundTripId == id ? "true" : "false",
                customRegistered ? "true" : "false");
        }
    }

    internal sealed class KitchenState : Cmd.Runnable
    {
        public Cmd.SubCommand State;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }
            var food = player.Food();
            var health = player.Health();
            var maxHealth = player.MaxHealth();
            var experienceLevel = player.ExperienceLevel();
            var experienceProgress = player.ExperienceProgress();
            var scale = player.Scale();
            var invisible = player.Invisible();
            var immobile = player.Immobile();

            player.SetFood(food);
            player.SetMaxHealth(maxHealth);
            player.SetExperienceLevel(experienceLevel);
            player.SetExperienceProgress(experienceProgress);
            player.SetScale(scale);
            if (invisible) player.SetInvisible();
            else player.SetVisible();
            if (immobile) player.SetImmobile();
            else player.SetMobile();

            output.Printf(
                "food={0}, health={1}/{2}, experience={3}:{4}, scale={5}, invisible={6}, immobile={7}",
                food,
                health,
                maxHealth,
                experienceLevel,
                experienceProgress,
                scale,
                invisible ? "true" : "false",
                immobile ? "true" : "false");
        }
    }

    private sealed record KitchenLiquid(string Type = "kitchen") : World.Liquid
    {
        public string LiquidType() => Type;
    }

    internal sealed class KitchenItem : Cmd.Runnable
    {
        public Cmd.SubCommand Item;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }
            var inventory = player.Inventory();
            var previous = inventory.Item(0);
            var enderChest = player.EnderChestInventory();
            var previousEnderItem = enderChest.Item(0);
            var (mainHand, offHand) = player.HeldItems();
            var sword = Dragonfly.Item.NewStack(
                    new Dragonfly.Item.Sword(Dragonfly.Item.ToolTierDiamond),
                    1)
                .WithCustomName("Kitchen sword")
                .WithLore("Generated from Dragonfly", "Restored after this command");
            try
            {
                inventory.SetItem(0, sword);
                enderChest.SetItem(0, sword);
                player.SetHeldItems(sword, offHand);
                var stored = inventory.Item(0);
                var enderStored = enderChest.Item(0);
                var (held, _) = player.HeldItems();
                var armour = player.Armour();
                var helmet = armour.Helmet();
                armour.SetHelmet(helmet);
                var addedEmpty = inventory.AddItem(default);
                if (stored.Item() is not Dragonfly.Item.Sword typed ||
                    enderStored.Item() is not Dragonfly.Item.Sword || enderChest.Size() != 27)
                {
                    output.Error("Typed item round-trip failed.");
                    return;
                }

                var damagedSword = sword.Damage(10);
                var unbreakableSword = damagedSword.AsUnbreakable();
                var snowballs = Dragonfly.Item.NewStack(new Dragonfly.Item.Snowball(), 8);
                var moreSnowballs = Dragonfly.Item.NewStack(new Dragonfly.Item.Snowball(), 10);
                var (fullSnowballs, remainingSnowballs) = snowballs.AddStack(moreSnowballs);
                var zeroSword = Dragonfly.Item.NewStack(
                    new Dragonfly.Item.Sword(Dragonfly.Item.ToolTierDiamond), 0);
                var persistentElytra = Dragonfly.Item.NewStack(new Dragonfly.Item.Elytra(), 1);
                if (sword.MaxCount() != 1 || sword.MaxDurability() != 1561 || sword.Durability() != 1561 ||
                    sword.AttackDamage() != 8d || damagedSword.Durability() != 1551 ||
                    unbreakableSword.Damage(100).Durability() != 1551 || !unbreakableSword.Unbreakable() ||
                    unbreakableSword.AsBreakable().Unbreakable() || !sword.WithDurability(0).Empty() ||
                    sword.WithAnvilCost(7).AnvilCost() != 7 ||
                    Dragonfly.Item.NewStack(new Dragonfly.Item.Apple(), 1).WithAnvilCost(7).AnvilCost() != 0 ||
                    snowballs.MaxCount() != 16 || fullSnowballs.Count() != 16 || remainingSnowballs.Count() != 2 ||
                    !snowballs.Comparable(moreSnowballs) || snowballs.Equal(moreSnowballs) ||
                    !zeroSword.Grow(1).Item()!.Equals(new Dragonfly.Item.Sword(Dragonfly.Item.ToolTierDiamond)) ||
                    persistentElytra.Damage(433).Empty())
                {
                    output.Error("Stack behavior failed.");
                    return;
                }

                var black = Dragonfly.Item.ColourBlack();
                var lavaChicken = Dragonfly.Sound.DiscLavaChicken();
                if (black.String() != "black" || black.SilverString() != "black" || black.Uint8() != 15 ||
                    black.RGBA() != new Dragonfly.Color.RGBA(29, 29, 33, 255) ||
                    black.SignRGBA() != new Dragonfly.Color.RGBA(0, 0, 0, 255) ||
                    Dragonfly.Sound.Dream().Name() != "Dream" ||
                    Dragonfly.Potion.StrongSlowness().Uint8() != 42 ||
                    lavaChicken.String() != "lava_chicken" || lavaChicken.DisplayName() != "Lava Chicken" ||
                    lavaChicken.Author() != "Hyper Potions")
                {
                    output.Error("Stateful item value methods failed.");
                    return;
                }

                var emptyBucket = new Dragonfly.Item.Bucket();
                var waterContent = Dragonfly.Item.LiquidBucketContent(new Dragonfly.Block.Water(false, 0, false));
                var waterBucket = new Dragonfly.Item.Bucket(waterContent);
                var lavaBucket = new Dragonfly.Item.Bucket(
                    Dragonfly.Item.LiquidBucketContent(new Dragonfly.Block.Lava(false, 0, false)));
                var milkBucket = new Dragonfly.Item.Bucket(Dragonfly.Item.MilkBucketContent());
                var customBucket = new Dragonfly.Item.Bucket(
                    Dragonfly.Item.LiquidBucketContent(new KitchenLiquid()));
                var customLavaBucket = new Dragonfly.Item.Bucket(
                    Dragonfly.Item.LiquidBucketContent(new KitchenLiquid("lava")));
                var (bucketLiquid, bucketLiquidFound) = waterContent.Liquid();
                var lavaFuel = lavaBucket.FuelInfo();
                if (!emptyBucket.Empty() || emptyBucket.MaxCount() != 16 ||
                    emptyBucket.Content.String() != "" || emptyBucket.Content.LiquidType() != "milk" ||
                    !bucketLiquidFound || bucketLiquid is not Dragonfly.Block.Water ||
                    waterContent.String() != "water" || waterContent.LiquidType() != "water" ||
                    waterBucket.Empty() || waterBucket.MaxCount() != 1 || waterBucket.AlwaysConsumable() ||
                    customBucket.Content.String() != "kitchen" ||
                    Dragonfly.Item.NewStack(customBucket, 1).MaxCount() != 1 ||
                    customLavaBucket.FuelInfo().Duration != TimeSpan.FromSeconds(1000) ||
                    milkBucket.Empty() || !milkBucket.AlwaysConsumable() || !milkBucket.CanConsume() ||
                    milkBucket.ConsumeDuration() != TimeSpan.FromMilliseconds(1610) ||
                    lavaFuel.Duration != TimeSpan.FromSeconds(1000) ||
                    lavaFuel.Residue.Count() != 1 || lavaFuel.Residue.Item() is not Dragonfly.Item.Bucket residue ||
                    !residue.Empty())
                {
                    output.Error("Bucket behavior failed.");
                    return;
                }

                Dragonfly.Item.Stack[] variants =
                [
                    Dragonfly.Item.NewStack(new Dragonfly.Item.Arrow(Dragonfly.Potion.NightVision()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.BannerPattern(Dragonfly.Item.CreeperBannerPattern()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.Dye(Dragonfly.Item.ColourBlack()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.GoatHorn(Dragonfly.Sound.Dream()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.Potion(Dragonfly.Potion.StrongSlowness()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.LingeringPotion(Dragonfly.Potion.Healing()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.SplashPotion(Dragonfly.Potion.Harming()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.MusicDisc(Dragonfly.Sound.DiscLavaChicken()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.PotterySherd(Dragonfly.Item.SherdTypeScrape()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.SmithingTemplate(Dragonfly.Item.TemplateBolt()), 1),
                    Dragonfly.Item.NewStack(new Dragonfly.Item.SuspiciousStew(Dragonfly.Item.NauseaStew()), 1),
                    Dragonfly.Item.NewStack(emptyBucket, 1),
                    Dragonfly.Item.NewStack(waterBucket, 1),
                    Dragonfly.Item.NewStack(lavaBucket, 1),
                    Dragonfly.Item.NewStack(milkBucket, 1),
                ];
                foreach (var variant in variants)
                {
                    inventory.SetItem(0, variant);
                    if (!Equals(inventory.Item(0).Item(), variant.Item()))
                    {
                        output.Error("Stateful item round-trip failed.");
                        return;
                    }
                }

                var writable = new Dragonfly.Item.BookAndQuill("alpha")
                    .InsertPage(1, "beta")
                    .SetPage(0, "first")
                    .SwapPages(0, 1);
                var (writablePage, writablePageFound) = writable.Page(1);
                if (!writablePageFound || writablePage != "first" || writable.TotalPages() != 2 ||
                    writable.DeletePage(1).TotalPages() != 1)
                {
                    output.Error("Writable book behavior failed.");
                    return;
                }
                var writableStack = Dragonfly.Item.NewStack(writable, 1);
                var otherWritableStack = Dragonfly.Item.NewStack(new Dragonfly.Item.BookAndQuill("different"), 1);
                var (unchangedWritable, remainingWritable) = writableStack.AddStack(otherWritableStack);
                if (writableStack.Comparable(otherWritableStack) ||
                    unchangedWritable.Count() != 1 || remainingWritable.Count() != 1)
                {
                    output.Error("Writable book comparison failed.");
                    return;
                }
                inventory.SetItem(0, writableStack);
                if (inventory.Item(0).Item() is not Dragonfly.Item.BookAndQuill storedWritable ||
                    storedWritable.Page(0) != ("beta", true) || storedWritable.Page(1) != ("first", true))
                {
                    output.Error("Writable book round-trip failed.");
                    return;
                }

                var written = new Dragonfly.Item.WrittenBook(
                    "Kitchen", "bedrock-gophers", Dragonfly.Item.CopyGeneration(), "page one", "page two");
                inventory.SetItem(0, Dragonfly.Item.NewStack(written, 1));
                if (inventory.Item(0).Item() is not Dragonfly.Item.WrittenBook storedWritten ||
                    storedWritten.Title != "Kitchen" || storedWritten.Author != "bedrock-gophers" ||
                    storedWritten.Generation != Dragonfly.Item.CopyGeneration() ||
                    storedWritten.Page(1) != ("page two", true))
                {
                    output.Error("Written book round-trip failed.");
                    return;
                }

                var explosion = new Dragonfly.Item.FireworkExplosion
                {
                    Shape = Dragonfly.Item.FireworkShapeStar(),
                    Colour = Dragonfly.Item.ColourBlack(),
                    Fade = Dragonfly.Item.ColourRed(),
                    Fades = true,
                    Twinkle = true,
                    Trail = true,
                };
                var firework = new Dragonfly.Item.Firework(TimeSpan.FromMilliseconds(1500), explosion);
                var randomisedDuration = firework.RandomisedDuration();
                var otherFirework = new Dragonfly.Item.Firework(TimeSpan.FromMilliseconds(2000), explosion);
                if (!firework.OffHand() || explosion.Shape.Name() != "Star" || explosion.Shape.String() != "star" ||
                    randomisedDuration < firework.Duration ||
                    randomisedDuration >= firework.Duration + TimeSpan.FromMilliseconds(600) ||
                    Dragonfly.Item.NewStack(firework, 1).Comparable(Dragonfly.Item.NewStack(otherFirework, 1)))
                {
                    output.Error("Firework behavior failed.");
                    return;
                }
                inventory.SetItem(0, Dragonfly.Item.NewStack(firework, 1));
                if (inventory.Item(0).Item() is not Dragonfly.Item.Firework storedFirework ||
                    storedFirework.Duration != firework.Duration || storedFirework.Explosions.Length != 1 ||
                    storedFirework.Explosions[0] != explosion)
                {
                    output.Error("Firework round-trip failed.");
                    return;
                }

                var starExplosion = new Dragonfly.Item.FireworkExplosion
                {
                    Shape = Dragonfly.Item.FireworkShapeBurst(),
                    Colour = Dragonfly.Item.ColourCyan(),
                };
                inventory.SetItem(0, Dragonfly.Item.NewStack(new Dragonfly.Item.FireworkStar(starExplosion), 1));
                if (inventory.Item(0).Item() is not Dragonfly.Item.FireworkStar storedStar ||
                    storedStar.FireworkExplosion != starExplosion)
                {
                    output.Error("Firework star round-trip failed.");
                    return;
                }

                var chargedRocket = Dragonfly.Item.NewStack(firework, 1)
                    .WithCustomName("Charged rocket")
                    .WithLore("Nested stack");
                var crossbow = new Dragonfly.Item.Crossbow(chargedRocket);
                var crossbowDurability = crossbow.DurabilityInfo();
                var crossbowFuel = crossbow.FuelInfo();
                if (crossbow.MaxCount() != 1 || crossbowDurability.MaxDurability != 464 ||
                    crossbowDurability.BrokenItem is null || !crossbowDurability.BrokenItem().Empty() ||
                    crossbowFuel.Duration != TimeSpan.FromSeconds(15) || !crossbowFuel.Residue.Empty() ||
                    crossbow.EnchantmentValue() != 1)
                {
                    output.Error("Crossbow behavior failed.");
                    return;
                }
                inventory.SetItem(0, Dragonfly.Item.NewStack(crossbow, 1));
                if (inventory.Item(0).Item() is not Dragonfly.Item.Crossbow storedCrossbow ||
                    storedCrossbow.Item.CustomName() != "Charged rocket" ||
                    storedCrossbow.Item.Lore() is not ["Nested stack"] ||
                    storedCrossbow.Item.Item() is not Dragonfly.Item.Firework storedRocket ||
                    storedRocket.Duration != firework.Duration || storedRocket.Explosions.Length != 1 ||
                    storedRocket.Explosions[0] != explosion)
                {
                    output.Error("Crossbow round-trip failed.");
                    return;
                }

                var armourTrim = new Dragonfly.Item.ArmourTrim(
                    Dragonfly.Item.TemplateFlow(),
                    new Dragonfly.Item.RedstoneWire());
                var dyedLeather = new Dragonfly.Item.ArmourTierLeather(
                    new Dragonfly.Color.RGBA(1, 2, 3, 255));
                var dyedHelmet = new Dragonfly.Item.Helmet(dyedLeather, armourTrim);
                var helmetDurability = dyedHelmet.DurabilityInfo();
                var copperChestplate = new Dragonfly.Item.Chestplate(new Dragonfly.Item.ArmourTierCopper());
                var copperSmelt = copperChestplate.SmeltInfo();
                var redstoneMaterial = new Dragonfly.Item.RedstoneWire();
                if (Dragonfly.Item.ArmourTiers().Length != 7 ||
                    Dragonfly.Item.ArmourTrimMaterials().Length != 11 ||
                    dyedLeather.BaseDurability() != 55d || dyedLeather.Name() != "leather" ||
                    dyedHelmet.MaxCount() != 1 || dyedHelmet.DefencePoints() != 1d ||
                    dyedHelmet.Toughness() != 0d || dyedHelmet.KnockBackResistance() != 0d ||
                    dyedHelmet.EnchantmentValue() != 15 || helmetDurability.MaxDurability != 55 ||
                    helmetDurability.BrokenItem is null || !helmetDurability.BrokenItem().Empty() ||
                    !((Dragonfly.Item.HelmetType)dyedHelmet).Helmet() ||
                    !dyedHelmet.RepairableBy(Dragonfly.Item.NewStack(new Dragonfly.Item.Leather(), 1)) ||
                    dyedHelmet.RepairableBy(Dragonfly.Item.NewStack(new Dragonfly.Item.Diamond(), 1)) ||
                    copperSmelt.Product.Item() is not Dragonfly.Item.CopperNugget ||
                    copperSmelt.Product.Count() != 1 || copperSmelt.Experience != 0.1d ||
                    copperSmelt.Food || !copperSmelt.Ores ||
                    redstoneMaterial.TrimMaterial() != "redstone" || redstoneMaterial.MaterialColour() != "§m" ||
                    dyedHelmet.WithTrim(default) is not Dragonfly.Item.Helmet untrimmed || !untrimmed.Trim.Zero() ||
                    Dragonfly.Item.NewStack(dyedHelmet, 1).Comparable(Dragonfly.Item.NewStack(untrimmed, 1)))
                {
                    output.Error("Armour behavior failed.");
                    return;
                }

                var armourItems = new List<World.Item>();
                foreach (var tier in Dragonfly.Item.ArmourTiers())
                {
                    armourItems.Add(tier is Dragonfly.Item.ArmourTierLeather ? dyedHelmet : new Dragonfly.Item.Helmet(tier));
                    armourItems.Add(new Dragonfly.Item.Chestplate(tier));
                    armourItems.Add(new Dragonfly.Item.Leggings(tier));
                    armourItems.Add(new Dragonfly.Item.Boots(tier));
                }
                foreach (var armourItem in armourItems)
                {
                    inventory.SetItem(0, Dragonfly.Item.NewStack(armourItem, 1));
                    var storedArmour = inventory.Item(0).Item();
                    if (storedArmour?.GetType() != armourItem.GetType() || !storedArmour.Equals(armourItem))
                    {
                        output.Error("Armour round-trip failed.");
                        return;
                    }
                }
                output.Printf(
                    "item=Sword, tier={0}, count={1}, held={2}, armour_slots={3}, ender_slots={4}, added_empty={5}, variants={6}",
                    typed.Tier.Name,
                    stored.Count(),
                    held.Item() is Dragonfly.Item.Sword ? "true" : "false",
                    armour.Inventory().Size(),
                    enderChest.Size(),
                    addedEmpty,
                    variants.Length + 33);
            }
            finally
            {
                inventory.SetItem(0, previous);
                enderChest.SetItem(0, previousEnderItem);
                player.SetHeldItems(mainHand, offHand);
            }
        }
    }

    internal sealed class KitchenEffect : Cmd.Runnable
    {
        public Cmd.SubCommand Effect;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }

            var timed = Dragonfly.Effect.New(Dragonfly.Effect.Speed, 2, TimeSpan.FromMilliseconds(1500));
            var ticked = timed.TickDuration();
            var ambient = Dragonfly.Effect.NewAmbient(
                Dragonfly.Effect.Regeneration, 1, TimeSpan.FromSeconds(2)).WithoutParticles();
            var infinite = Dragonfly.Effect.NewInfinite(Dragonfly.Effect.FireResistance, 1);
            var instant = Dragonfly.Effect.NewInstant(Dragonfly.Effect.InstantHealth, 1);
            var potent = Dragonfly.Effect.NewInstantWithPotency(Dragonfly.Effect.InstantDamage, 2, 0.5d);
            var (speedID, speedRegistered) = Dragonfly.Effect.ID(Dragonfly.Effect.Speed);
            var (speedType, speedFound) = Dragonfly.Effect.ByID(speedID);
            var (mixed, mixedAmbient) = Dragonfly.Effect.ResultingColour([timed, infinite]);
            var potions = Dragonfly.Potion.All();
            var turtle = Dragonfly.Potion.TurtleMaster().Effects();
            var stews = Dragonfly.Item.StewTypes();
            var saturation = Dragonfly.Item.SaturationDandelionStew().Effects();
            if (ticked.Duration() != TimeSpan.FromMilliseconds(1450) || ticked.Tick() != 1 ||
                ticked.Type() != Dragonfly.Effect.Speed || timed.ParticlesHidden() ||
                !ambient.Ambient() || !ambient.ParticlesHidden() || !infinite.Infinite() ||
                instant.Type() != Dragonfly.Effect.InstantHealth || potent.Level() != 2 ||
                !speedRegistered || !speedFound || speedType != Dragonfly.Effect.Speed ||
                mixed == default || mixedAmbient || potions.Count != 43 ||
                Dragonfly.Potion.From(256) != Dragonfly.Potion.Water() || turtle.Count != 2 ||
                Dragonfly.Potion.From(43).Uint8() != 43 || Dragonfly.Potion.From(43).Effects().Count != 0 ||
                turtle[0].Type() != Dragonfly.Effect.Resistance || turtle[1].Type() != Dragonfly.Effect.Slowness ||
                stews.Count != 13 || saturation.Count != 1 ||
                saturation[0].Duration() != TimeSpan.FromMilliseconds(300))
            {
                output.Error("Effect behavior failed.");
                return;
            }

            var (previous, hadPrevious) = player.Effect(Dragonfly.Effect.Regeneration);
            try
            {
                player.AddEffect(ambient);
                var (active, found) = player.Effect(Dragonfly.Effect.Regeneration);
                var all = player.Effects();
                if (!found || active.Level() != 1 || !active.Ambient() || !active.ParticlesHidden() ||
                    !all.Any(value => value.Type() == Dragonfly.Effect.Regeneration))
                {
                    output.Error("Player effect round-trip failed.");
                    return;
                }
            }
            finally
            {
                player.RemoveEffect(Dragonfly.Effect.Regeneration);
                if (hadPrevious) player.AddEffect(previous);
            }

            output.Printf("effects=28, potions={0}, stews={1}, active=true", potions.Count, stews.Count);
        }
    }

    internal sealed class KitchenForm : Cmd.Runnable
    {
        public Cmd.SubCommand Form;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }
            player.SendForm(KitchenMenu.Create());
        }
    }

    internal sealed class KitchenRawFormCommand : Cmd.Runnable
    {
        [Cmd.Tag("raw-form")]
        public Cmd.SubCommand RawForm;

        public void Run(Cmd.Source source, Cmd.Output output, World.Tx? tx)
        {
            if (source is not Player player)
            {
                output.Error("This command can only be used by a player.");
                return;
            }
            player.SendForm(new KitchenRawForm());
        }
    }

    internal sealed class KitchenRawForm : Form.Value
    {
        public byte[] MarshalJSON()
        {
            var header = Encoding.UTF8.GetString(Form.NewHeader("Custom Form.Value").MarshalJSON());
            var button = Encoding.UTF8.GetString(Form.NewButton("Submit", string.Empty).MarshalJSON());
            return Encoding.UTF8.GetBytes(
                $$"""{"type":"form","title":"Custom Form.Value","content":"Open form interface","elements":[{{header}},{{button}}]}""");
        }

        public void SubmitJSON(byte[]? response, Form.Submitter submitter, World.Tx tx)
        {
            if (submitter is not Player player) return;
            if (response is null)
            {
                player.Message("Custom Form.Value dismissed.");
                return;
            }
            var position = player.Position();
            player.Message(
                $"raw={Encoding.UTF8.GetString(response)}, player={player.Name()}, " +
                $"latency={player.Latency().TotalMilliseconds:0}ms, " +
                $"position={position.X},{position.Y},{position.Z}");
        }
    }

    private sealed class KitchenMenu : Form.MenuSubmittable, Form.Closer
    {
        private static readonly Form.Button CloseButton = Form.NewButton("Close", string.Empty);

        public Form.Button OpenCustom = Form.NewButton(
            "Open every custom element",
            "textures/ui/icon_recipe_nature");
        public Form.Button OpenModal = Form.NewButton(
            "Skip to the modal",
            "https://raw.githubusercontent.com/df-mc/dragonfly/master/.github/assets/logo.png");

        public static Form.Menu Create()
        {
            var menu = Form.NewMenu(new KitchenMenu(), "Kitchen sink forms")
                .WithBody("Dragonfly's reflected menu API from C#.")
                .AddHeader(Form.NewHeader("Generated from Dragonfly"))
                .AddDivider(new Form.Divider())
                .AddLabel(Form.NewLabel("The first two buttons are reflected fields."))
                .AddButton(Form.NewButton("Extra button", string.Empty))
                .WithButtons(CloseButton)
                .WithElements(
                    Form.NewLabel("Menu elements may be appended together."),
                    new Form.Divider());
            return menu.AddLabel(Form.NewLabel(
                $"{menu.Title()}: {menu.Body()} " +
                $"({menu.Buttons().Count} buttons, {menu.Elements().Count} elements)"));
        }

        public void Submit(Form.Submitter submitter, Form.Button pressed, World.Tx tx)
        {
            if (pressed.Equals(OpenModal))
            {
                submitter.SendForm(KitchenModal.Create("Opened directly from the menu."));
                return;
            }
            if (pressed.Equals(CloseButton))
            {
                submitter.CloseForm();
                Message(submitter, "Kitchen form closed.");
                return;
            }
            if (pressed.Equals(OpenCustom))
            {
                submitter.SendForm(KitchenCustom.Create());
                return;
            }
            submitter.SendForm(KitchenCustom.Create());
        }

        public void Close(Form.Submitter submitter, World.Tx tx) =>
            Message(submitter, "Kitchen menu dismissed.");
    }

    private sealed class KitchenCustom : Form.Submittable, Form.Closer
    {
        public Form.Header Header = Form.NewHeader("Every custom element");
        public Form.Divider Divider = new();
        public Form.Label Label = Form.NewLabel("Values are reflected back into these fields.");
        public Form.Input Name = Form.NewInput("Name", "Dragonfly", "Type a name")
            .WithTooltip("A UTF-8 string value.");
        public Form.Toggle Enabled = Form.NewToggle("Enabled", true)
            .WithTooltip("A boolean value.");
        public Form.Slider Power = Form.NewSlider("Power", 0, 10, 0.5, 5)
            .WithTooltip("A bounded numeric value.");
        public Form.Dropdown Colour = Form.NewDropdown(
                "Colour",
                ["Red", "Green", "Blue"],
                1)
            .WithTooltip("An option index.");
        public Form.StepSlider Size = Form.NewStepSlider(
                "Size",
                ["Small", "Medium", "Large"],
                1)
            .WithTooltip("A stepped option index.");

        public static Form.Custom Create()
        {
            var screen = new KitchenCustom();
            var custom = Form.New(screen, "Kitchen custom form");
            screen.Label = Form.NewLabel(
                $"{custom.Title()} contains {custom.Elements().Count} reflected elements.");
            return custom;
        }

        public void Submit(Form.Submitter submitter, World.Tx tx)
        {
            var summary = $"name={Name.Value()}, enabled={Enabled.Value()}, " +
                          $"power={Power.Value():0.0}, colour={Colour.Value()}, size={Size.Value()}";
            submitter.SendForm(KitchenModal.Create(summary));
        }

        public void Close(Form.Submitter submitter, World.Tx tx) =>
            Message(submitter, "Kitchen custom form dismissed.");
    }

    private sealed class KitchenModal : Form.ModalSubmittable, Form.Closer
    {
        public Form.Button Accept = Form.YesButton();
        public Form.Button Reject = Form.NoButton();

        private readonly string _summary;

        private KitchenModal(string summary) => _summary = summary;

        public static Form.Modal Create(string summary)
        {
            var modal = Form.NewModal(new KitchenModal(summary), "Confirm kitchen values")
                .WithBody(summary);
            return modal.WithBody(
                $"{modal.Title()}: {modal.Body()} ({modal.Buttons().Count} choices)");
        }

        public void Submit(Form.Submitter submitter, Form.Button pressed, World.Tx tx) =>
            Message(
                submitter,
                $"{(pressed.Equals(Accept) ? "Accepted" : "Rejected")}: {_summary}");

        public void Close(Form.Submitter submitter, World.Tx tx) =>
            Message(submitter, "Kitchen modal dismissed.");
    }

    private static void Message(Form.Submitter submitter, string message)
    {
        if (submitter is Player player) player.Message(message);
    }

    private sealed class CustomGameMode : World.GameMode
    {
        public bool AllowsEditing() => true;
        public bool AllowsTakingDamage() => true;
        public bool CreativeInventory() => false;
        public bool HasCollision() => true;
        public bool AllowsFlying() => false;
        public bool AllowsInteraction() => true;
        public bool Visible() => true;
        public bool InstantPortalTravel() => false;
    }
}
