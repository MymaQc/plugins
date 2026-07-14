// Code generated from Dragonfly server/player/handler.go. DO NOT EDIT.
#nullable enable
namespace Dragonfly;

public sealed partial class Player
{
    public interface Handler
    {
        void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot);
        void HandleJump(Player p);
        void HandleTeleport(Player.Context ctx, Vector3 pos);
        void HandleChangeWorld(Player p, World? before, World after);
        void HandleToggleSprint(Player.Context ctx, bool after);
        void HandleToggleSneak(Player.Context ctx, bool after);
        void HandleChat(Player.Context ctx, ref string message);
        void HandleFoodLoss(Player.Context ctx, int from, ref int to);
        void HandleHeal(Player.Context ctx, ref double health, World.HealingSource src);
        void HandleHurt(Player.Context ctx, ref double damage, bool immune, ref TimeSpan attackImmunity, World.DamageSource src);
        void HandleDeath(Player p, World.DamageSource src, ref bool keepInv);
        void HandleRespawn(Player p, ref Vector3 pos, ref World w);
        void HandleSkinChange(Player.Context ctx, ref Skin skin);
        void HandleFireExtinguish(Player.Context ctx, Cube.Pos pos);
        void HandleStartBreak(Player.Context ctx, Cube.Pos pos);
        void HandleBlockBreak(Player.Context ctx, Cube.Pos pos, ref Item.Stack[] drops, ref int xp);
        void HandleBlockPlace(Player.Context ctx, Cube.Pos pos, World.Block b);
        void HandleBlockPick(Player.Context ctx, Cube.Pos pos, World.Block b);
        void HandleItemUse(Player.Context ctx);
        void HandleItemUseOnBlock(Player.Context ctx, Cube.Pos pos, Cube.Face face, Vector3 clickPos);
        void HandleItemUseOnEntity(Player.Context ctx, World.Entity e);
        void HandleItemRelease(Player.Context ctx, Item.Stack item, TimeSpan dur);
        void HandleItemConsume(Player.Context ctx, Item.Stack item);
        void HandleAttackEntity(Player.Context ctx, World.Entity e, ref double force, ref double height, ref bool critical);
        void HandleExperienceGain(Player.Context ctx, ref int amount);
        void HandlePunchAir(Player.Context ctx);
        void HandleSignEdit(Player.Context ctx, Cube.Pos pos, bool frontSide, string oldText, string newText);
        void HandleSleep(Player.Context ctx, ref bool sendReminder);
        void HandleLecternPageTurn(Player.Context ctx, Cube.Pos pos, int oldPage, ref int newPage);
        void HandleItemDamage(Player.Context ctx, Item.Stack i, ref int damage);
        void HandleItemPickup(Player.Context ctx, ref Item.Stack i);
        void HandleHeldSlotChange(Player.Context ctx, int from, int to);
        void HandleItemDrop(Player.Context ctx, Item.Stack s);
        void HandleTransfer(Player.Context ctx, ref Net.UDPAddr addr);
        void HandleCommandExecution(Player.Context ctx, Cmd.Command command, string[] args);
        void HandleQuit(Player p);
        void HandleDiagnostics(Player p, Session.Diagnostics d);
    }
}

public abstract partial class Plugin
{
    [HandlerSubscription(1UL)]
    public virtual void HandleMove(Player.Context ctx, Vector3 newPos, Rotation newRot) { }
    [HandlerSubscription(16384UL)]
    public virtual void HandleJump(Player p) { }
    [HandlerSubscription(32768UL)]
    public virtual void HandleTeleport(Player.Context ctx, Vector3 pos) { }
    [HandlerSubscription(2147483648UL)]
    public virtual void HandleChangeWorld(Player p, World? before, World after) { }
    [HandlerSubscription(4096UL)]
    public virtual void HandleToggleSprint(Player.Context ctx, bool after) { }
    [HandlerSubscription(8192UL)]
    public virtual void HandleToggleSneak(Player.Context ctx, bool after) { }
    [HandlerSubscription(2UL)]
    public virtual void HandleChat(Player.Context ctx, ref string message) { }
    [HandlerSubscription(256UL)]
    public virtual void HandleFoodLoss(Player.Context ctx, int from, ref int to) { }
    [HandlerSubscription(32UL)]
    public virtual void HandleHeal(Player.Context ctx, ref double health, World.HealingSource src) { }
    [HandlerSubscription(16UL)]
    public virtual void HandleHurt(Player.Context ctx, ref double damage, bool immune, ref TimeSpan attackImmunity, World.DamageSource src) { }
    [HandlerSubscription(512UL)]
    public virtual void HandleDeath(Player p, World.DamageSource src, ref bool keepInv) { }
    [HandlerSubscription(4294967296UL)]
    public virtual void HandleRespawn(Player p, ref Vector3 pos, ref World w) { }
    [HandlerSubscription(8589934592UL)]
    public virtual void HandleSkinChange(Player.Context ctx, ref Skin skin) { }
    [HandlerSubscription(2048UL)]
    public virtual void HandleFireExtinguish(Player.Context ctx, Cube.Pos pos) { }
    [HandlerSubscription(1024UL)]
    public virtual void HandleStartBreak(Player.Context ctx, Cube.Pos pos) { }
    [HandlerSubscription(64UL)]
    public virtual void HandleBlockBreak(Player.Context ctx, Cube.Pos pos, ref Item.Stack[] drops, ref int xp) { }
    [HandlerSubscription(128UL)]
    public virtual void HandleBlockPlace(Player.Context ctx, Cube.Pos pos, World.Block b) { }
    [HandlerSubscription(1048576UL)]
    public virtual void HandleBlockPick(Player.Context ctx, Cube.Pos pos, World.Block b) { }
    [HandlerSubscription(8388608UL)]
    public virtual void HandleItemUse(Player.Context ctx) { }
    [HandlerSubscription(16777216UL)]
    public virtual void HandleItemUseOnBlock(Player.Context ctx, Cube.Pos pos, Cube.Face face, Vector3 clickPos) { }
    [HandlerSubscription(1073741824UL)]
    public virtual void HandleItemUseOnEntity(Player.Context ctx, World.Entity e) { }
    [HandlerSubscription(67108864UL)]
    public virtual void HandleItemRelease(Player.Context ctx, Item.Stack item, TimeSpan dur) { }
    [HandlerSubscription(33554432UL)]
    public virtual void HandleItemConsume(Player.Context ctx, Item.Stack item) { }
    [HandlerSubscription(536870912UL)]
    public virtual void HandleAttackEntity(Player.Context ctx, World.Entity e, ref double force, ref double height, ref bool critical) { }
    [HandlerSubscription(65536UL)]
    public virtual void HandleExperienceGain(Player.Context ctx, ref int amount) { }
    [HandlerSubscription(131072UL)]
    public virtual void HandlePunchAir(Player.Context ctx) { }
    [HandlerSubscription(4194304UL)]
    public virtual void HandleSignEdit(Player.Context ctx, Cube.Pos pos, bool frontSide, string oldText, string newText) { }
    [HandlerSubscription(524288UL)]
    public virtual void HandleSleep(Player.Context ctx, ref bool sendReminder) { }
    [HandlerSubscription(2097152UL)]
    public virtual void HandleLecternPageTurn(Player.Context ctx, Cube.Pos pos, int oldPage, ref int newPage) { }
    [HandlerSubscription(134217728UL)]
    public virtual void HandleItemDamage(Player.Context ctx, Item.Stack i, ref int damage) { }
    [HandlerSubscription(137438953472UL)]
    public virtual void HandleItemPickup(Player.Context ctx, ref Item.Stack i) { }
    [HandlerSubscription(262144UL)]
    public virtual void HandleHeldSlotChange(Player.Context ctx, int from, int to) { }
    [HandlerSubscription(268435456UL)]
    public virtual void HandleItemDrop(Player.Context ctx, Item.Stack s) { }
    [HandlerSubscription(274877906944UL)]
    public virtual void HandleTransfer(Player.Context ctx, ref Net.UDPAddr addr) { }
    [HandlerSubscription(549755813888UL)]
    public virtual void HandleCommandExecution(Player.Context ctx, Cmd.Command command, string[] args) { }
    [HandlerSubscription(8UL)]
    public virtual void HandleQuit(Player p) { }
    [HandlerSubscription(1099511627776UL)]
    public virtual void HandleDiagnostics(Player p, Session.Diagnostics d) { }
}
