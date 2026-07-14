using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class World : IEquatable<World>
{
    private readonly ulong _invocation;
    internal WorldId Id { get; }
    internal ulong Invocation => _invocation;

    internal World(ulong invocation, WorldId id)
    {
        _invocation = invocation;
        Id = id;
    }

    public bool Equals(World? other) => other is not null && Id.Value == other.Id.Value;
    public override bool Equals(object? obj) => obj is World other && Equals(other);
    public override int GetHashCode() => Id.Value.GetHashCode();

    public sealed class Entity : IEquatable<Entity>
    {
        internal EntityId Id { get; }
        internal ulong Invocation { get; }

        internal Entity(ulong invocation, EntityId id)
        {
            Invocation = invocation;
            Id = id;
        }

        public bool Equals(Entity? other) => other is not null && SameId(Id, other.Id);
        public override bool Equals(object? obj) => obj is Entity other && Equals(other);

        public override unsafe int GetHashCode()
        {
            var hash = new HashCode();
            var id = Id;
            hash.Add(id.Generation);
            for (var index = 0; index < 16; index++) hash.Add(id.Bytes[index]);
            return hash.ToHashCode();
        }

        private static unsafe bool SameId(EntityId left, EntityId right)
        {
            if (left.Generation != right.Generation) return false;
            for (var index = 0; index < 16; index++)
                if (left.Bytes[index] != right.Bytes[index]) return false;
            return true;
        }
    }

    internal readonly record struct DamageProperties(
        bool ReducedByArmour,
        bool ReducedByResistance,
        bool Fire,
        bool IgnoresTotem);

    public interface DamageSource
    {
        bool ReducedByArmour();
        bool ReducedByResistance();
        bool Fire();
        bool IgnoreTotem();
    }

    public sealed class CustomDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal CustomDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class AttackDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        private readonly Entity? _attacker;
        internal AttackDamageSource(DamageProperties properties, Entity? attacker) =>
            (_properties, _attacker) = (properties, attacker);
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
        public Entity? Attacker() => _attacker;
    }

    public sealed class BlockDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        private readonly Block _block;
        internal BlockDamageSource(DamageProperties properties, Block block) =>
            (_properties, _block) = (properties, block);
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
        public Block Block() => _block;
    }

    public sealed class DrowningDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal DrowningDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class ExplosionDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal ExplosionDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class FallDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal FallDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class FireDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal FireDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class GlideDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal GlideDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class InstantDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal InstantDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class LavaDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal LavaDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class LightningDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal LightningDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class MagmaDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal MagmaDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class PoisonDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        private readonly bool _fatal;
        internal PoisonDamageSource(DamageProperties properties, bool fatal) =>
            (_properties, _fatal) = (properties, fatal);
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
        public bool Fatal() => _fatal;
    }

    public sealed class ProjectileDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        private readonly Entity? _projectile;
        private readonly Entity? _owner;
        internal ProjectileDamageSource(
            DamageProperties properties,
            Entity? projectile,
            Entity? owner)
        {
            _properties = properties;
            _projectile = projectile;
            _owner = owner;
        }
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
        public Entity? Projectile() => _projectile;
        public Entity? Owner() => _owner;
    }

    public sealed class StarvationDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal StarvationDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class SuffocationDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal SuffocationDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class ThornsDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        private readonly Entity? _owner;
        internal ThornsDamageSource(DamageProperties properties, Entity? owner) =>
            (_properties, _owner) = (properties, owner);
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
        public Entity? Owner() => _owner;
    }

    public sealed class VoidDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal VoidDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public sealed class WitherDamageSource : DamageSource
    {
        private readonly DamageProperties _properties;
        internal WitherDamageSource(DamageProperties properties) => _properties = properties;
        public bool ReducedByArmour() => _properties.ReducedByArmour;
        public bool ReducedByResistance() => _properties.ReducedByResistance;
        public bool Fire() => _properties.Fire;
        public bool IgnoreTotem() => _properties.IgnoresTotem;
    }

    public interface HealingSource { }

    public sealed class CustomHealingSource : HealingSource { internal CustomHealingSource() { } }

    public sealed class FoodHealingSource : HealingSource
    {
        private readonly bool _quickRegeneration;
        internal FoodHealingSource(bool quickRegeneration) =>
            _quickRegeneration = quickRegeneration;
        public bool QuickRegeneration() => _quickRegeneration;
    }

    public sealed class InstantHealingSource : HealingSource { internal InstantHealingSource() { } }
    public sealed class RegenerationHealingSource : HealingSource { internal RegenerationHealingSource() { } }
}
