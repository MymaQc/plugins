#nullable enable
using System;
using System.Collections.Generic;
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
            new KitchenItem(),
            new KitchenForm(),
            new KitchenRawFormCommand()));
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

    public override void HandlePunchAir(Player.Context ctx) => Increment(ref _punches);
    public override void HandleQuit(Player player) => Increment(ref _quits);

    private static void Increment(ref long counter) => Interlocked.Increment(ref counter);

    private static bool Finite(Vector3 value) =>
        double.IsFinite(value.X) && double.IsFinite(value.Y) && double.IsFinite(value.Z);

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
            var (mainHand, offHand) = player.HeldItems();
            var sword = Dragonfly.Item.NewStack(
                    new Dragonfly.Item.Sword(Dragonfly.Item.ToolTierDiamond),
                    1)
                .WithCustomName("Kitchen sword")
                .WithLore("Generated from Dragonfly", "Restored after this command");
            try
            {
                inventory.SetItem(0, sword);
                player.SetHeldItems(sword, offHand);
                var stored = inventory.Item(0);
                var (held, _) = player.HeldItems();
                var armour = player.Armour();
                var helmet = armour.Helmet();
                armour.SetHelmet(helmet);
                var addedEmpty = inventory.AddItem(default);
                if (stored.Item() is not Dragonfly.Item.Sword typed)
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
                output.Printf(
                    "item=Sword, tier={0}, count={1}, held={2}, armour_slots={3}, added_empty={4}, variants={5}",
                    typed.Tier.Name,
                    stored.Count(),
                    held.Item() is Dragonfly.Item.Sword ? "true" : "false",
                    armour.Inventory().Size(),
                    addedEmpty,
                    variants.Length);
            }
            finally
            {
                inventory.SetItem(0, previous);
                player.SetHeldItems(mainHand, offHand);
            }
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
