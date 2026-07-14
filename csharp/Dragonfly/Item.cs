#nullable enable
using System.Globalization;
using Dragonfly.Native;

namespace Dragonfly;

public static partial class Item
{
    public static Stack NewStack(World.Item item, int count)
    {
        ArgumentNullException.ThrowIfNull(item);
        if (count < 0) throw new ArgumentOutOfRangeException(nameof(count));
        return new Stack(item, count);
    }

    public readonly struct Stack
    {
        private readonly World.Item? _item;
        private readonly int _count;
        private readonly uint _damage;
        private readonly bool _unbreakable;
        private readonly int _anvilCost;
        private readonly string? _customName;
        private readonly string[]? _lore;
        private readonly byte[]? _itemNbt;
        private readonly byte[]? _valuesNbt;
        private readonly ItemEnchantment[]? _enchantments;

        internal Stack(
            World.Item? item,
            int count,
            uint damage = 0,
            bool unbreakable = false,
            int anvilCost = 0,
            string? customName = null,
            string[]? lore = null,
            byte[]? itemNbt = null,
            byte[]? valuesNbt = null,
            ItemEnchantment[]? enchantments = null)
        {
            _item = item;
            _count = Math.Max(count, 0);
            _damage = damage;
            _unbreakable = unbreakable;
            _anvilCost = anvilCost;
            _customName = customName;
            _lore = lore;
            _itemNbt = itemNbt;
            _valuesNbt = valuesNbt;
            _enchantments = enchantments;
        }

        public int Count() => _count;
        public int MaxCount() => ItemCapabilities.MaxCount(_item);
        public bool Empty() => _count == 0 || _item is null || ItemCodec.IsAir(_item);
        public World.Item? Item() => Empty() ? null : _item;

        public Stack Grow(int count) => Copy(count: Math.Max(0, _count + count));

        public int Durability() => ItemCapabilities.TryDurability(Item(), out var info)
            ? unchecked((int)((long)info.MaxDurability - _damage))
            : -1;

        public int MaxDurability() => ItemCapabilities.TryDurability(Item(), out var info)
            ? info.MaxDurability
            : -1;

        public Stack Damage(int damage)
        {
            if (!ItemCapabilities.TryDurability(Item(), out var info) || _unbreakable) return this;

            var durability = (long)info.MaxDurability - _damage;
            var resultingDurability = durability - damage;
            if (resultingDurability <= 0) return info.Persistent ? this : info.BrokenStack;
            if (resultingDurability > info.MaxDurability) return Copy(damage: 0);
            return Copy(damage: checked((uint)((long)_damage + damage)));
        }

        public Stack WithDurability(int durability)
        {
            if (!ItemCapabilities.TryDurability(Item(), out var info)) return this;
            if (durability > info.MaxDurability) return Copy(damage: 0);
            if (durability == 0) return info.BrokenStack;
            return Copy(damage: checked((uint)((long)info.MaxDurability - durability)));
        }

        public bool Unbreakable() => _unbreakable;

        public Stack AsUnbreakable() => ItemCapabilities.TryDurability(Item(), out _)
            ? Copy(unbreakable: true)
            : this;

        public Stack AsBreakable() => ItemCapabilities.TryDurability(Item(), out _)
            ? Copy(unbreakable: false)
            : this;

        public double AttackDamage() => ItemCapabilities.AttackDamage(Item());

        public string CustomName() => _customName ?? string.Empty;

        public Stack WithCustomName(params object?[] values) => Copy(
            customName: string.Join(" ", values.Select(FormatValue)));

        public string[] Lore() => Empty() ? [] : (string[])(_lore?.Clone() ?? Array.Empty<string>());

        public Stack WithLore(params string[] lines)
        {
            ArgumentNullException.ThrowIfNull(lines);
            return Copy(lore: (string[])lines.Clone());
        }

        public int AnvilCost() => _anvilCost;

        public Stack WithAnvilCost(int anvilCost) => ItemCapabilities.AllowsAnvilCost(Item())
            ? Copy(anvilCost: anvilCost)
            : this;

        public (Stack A, Stack B) AddStack(Stack other)
        {
            if (_count >= MaxCount() || !Comparable(other)) return (this, other);
            var added = Math.Min(MaxCount() - _count, other._count);
            return (Copy(count: _count + added), other.Copy(count: other._count - added));
        }

        public bool Equal(Stack other) => Comparable(other) &&
            _count == other._count && _damage == other._damage;

        public bool Comparable(Stack other)
        {
            if (Empty() || other.Empty()) return true;
            if (!TryEncode(out var identifier, out var metadata) ||
                !other.TryEncode(out var otherIdentifier, out var otherMetadata) ||
                identifier != otherIdentifier || metadata != otherMetadata ||
                _anvilCost != other._anvilCost || CustomName() != other.CustomName() ||
                !SequenceEqual(_lore, other._lore) ||
                !EnchantmentsEqual(_enchantments, other._enchantments) ||
                !SequenceEqual(_valuesNbt, other._valuesNbt) ||
                !SequenceEqual(ItemNbt, other.ItemNbt)) return false;
            return true;
        }

        internal uint DamageValue => _damage;
        internal bool IsUnbreakable => _unbreakable;
        internal int AnvilCostValue => _anvilCost;
        internal World.Item? RawItem => _item;
        internal byte[] RawItemNbt => _itemNbt ?? [];
        internal byte[] ItemNbt => ItemNbtCodec.TryEncode(_item, out var encoded) ? encoded : _itemNbt ?? [];
        internal byte[] ValuesNbt => _valuesNbt ?? [];
        internal ItemEnchantment[] Enchantments => _enchantments ?? [];

        internal bool TryEncode(out string identifier, out int metadata)
        {
            if (_item is not null) return ItemCodec.TryEncode(_item, out identifier, out metadata);
            identifier = string.Empty;
            metadata = 0;
            return false;
        }

        private Stack Copy(
            int? count = null,
            uint? damage = null,
            bool? unbreakable = null,
            int? anvilCost = null,
            string? customName = null,
            string[]? lore = null) => new(
                _item,
                count ?? _count,
                damage ?? _damage,
                unbreakable ?? _unbreakable,
                anvilCost ?? _anvilCost,
                customName ?? _customName,
                lore ?? _lore,
                _itemNbt,
                _valuesNbt,
                _enchantments);

        private static bool SequenceEqual<T>(T[]? left, T[]? right) where T : IEquatable<T> =>
            (left ?? []).AsSpan().SequenceEqual(right ?? []);

        private static bool EnchantmentsEqual(ItemEnchantment[]? left, ItemEnchantment[]? right)
        {
            var leftSpan = (left ?? []).AsSpan();
            var rightSpan = (right ?? []).AsSpan();
            if (leftSpan.Length != rightSpan.Length) return false;
            for (var index = 0; index < leftSpan.Length; index++)
            {
                if (leftSpan[index].Id != rightSpan[index].Id || leftSpan[index].Level != rightSpan[index].Level)
                    return false;
            }
            return true;
        }

        private static string FormatValue(object? value) => value switch
        {
            null => "<nil>",
            bool boolean => boolean ? "true" : "false",
            IFormattable formattable => formattable.ToString(null, CultureInfo.InvariantCulture) ?? string.Empty,
            _ => value.ToString() ?? string.Empty,
        };
    }
}
