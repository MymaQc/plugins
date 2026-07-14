#nullable enable
using System.Text;
using Dragonfly.Native;

namespace Dragonfly;

// Crossbows embed a complete item stack. This private transport keeps that stack
// structured until the Go host restores Dragonfly's native item.Crossbow value.
internal static class CrossbowStackNbt
{
    internal const int MaxDepth = 16;
    private const string VersionName = "bedrock_gophers_version";
    private const int Version = 1;
    private const int MaxEntries = 256;
    private const int MaxIdentifierBytes = 256;
    private const int MaxTextBytes = 4096;

    internal static byte[] Encode(Item.Stack stack, int depth)
    {
        if (stack.Empty()) return [];
        if (depth > MaxDepth) throw new InvalidDataException("nested item stack is too deep");
        if (!stack.TryEncode(out var identifier, out var metadata))
            throw new InvalidDataException("nested item stack cannot be encoded");

        var loreLines = stack.Lore();
        var stackEnchantments = stack.EncodedEnchantments;
        if (Encoding.UTF8.GetByteCount(identifier) > MaxIdentifierBytes ||
            Encoding.UTF8.GetByteCount(stack.CustomName()) > MaxTextBytes ||
            loreLines.Length > MaxEntries || stackEnchantments.Length > MaxEntries ||
            loreLines.Any(line => Encoding.UTF8.GetByteCount(line) > MaxTextBytes) ||
            stackEnchantments.Any(enchantment => enchantment.Level == 0))
            throw new ArgumentException("nested item stack exceeds server limits", nameof(stack));

        var itemNbt = ItemNbtCodec.TryEncode(stack.RawItem, depth, out var encodedItemNbt)
            ? encodedItemNbt
            : stack.RawItemNbt;
        var enchantments = stackEnchantments.Select(enchantment => Nbt.Value.Compound(new Nbt.Compound
        {
            ["id"] = Nbt.Value.Int(checked((int)enchantment.Id)),
            ["level"] = Nbt.Value.Int(checked((int)enchantment.Level)),
        })).ToArray();
        var lore = loreLines.Select(line => Nbt.Value.String(line)).ToArray();
        var charged = new Nbt.Compound
        {
            [VersionName] = Nbt.Value.Int(Version),
            ["identifier"] = Nbt.Value.String(identifier),
            ["metadata"] = Nbt.Value.Int(metadata),
            ["count"] = Nbt.Value.Int(stack.Count()),
            ["damage"] = Nbt.Value.Int(checked((int)stack.DamageValue)),
            ["unbreakable"] = Nbt.Value.Byte(stack.IsUnbreakable ? (byte)1 : (byte)0),
            ["anvilCost"] = Nbt.Value.Int(stack.AnvilCostValue),
            ["customName"] = Nbt.Value.String(stack.CustomName()),
            ["lore"] = Nbt.Value.List(Nbt.TagType.String, lore),
            ["itemNbt"] = Nbt.Value.ByteArray(itemNbt),
            ["valuesNbt"] = Nbt.Value.ByteArray(stack.ValuesNbt),
            ["enchantments"] = Nbt.Value.List(Nbt.TagType.Compound, enchantments),
        };
        return Nbt.Encode(new Nbt.Compound { ["chargedItem"] = Nbt.Value.Compound(charged) });
    }

    internal static Item.Stack Decode(ReadOnlySpan<byte> data, int depth)
    {
        if (data.IsEmpty || depth > MaxDepth || !Nbt.TryDecode(data, out var root) || root is null ||
            !root.TryGetValue("chargedItem", out var encoded) || encoded.Type != Nbt.TagType.Compound)
            return default;
        return TryDecodeStack(encoded.AsCompound(), depth, out var stack) ? stack : default;
    }

    private static bool TryDecodeStack(Nbt.Compound data, int depth, out Item.Stack stack)
    {
        stack = default;
        try
        {
            if (depth > MaxDepth || GetInt(data, VersionName) != Version)
                return false;
            var identifier = GetString(data, "identifier");
            var metadata = GetInt(data, "metadata");
            var count = GetInt(data, "count");
            var damage = GetInt(data, "damage");
            var unbreakable = GetByte(data, "unbreakable");
            var anvilCost = GetInt(data, "anvilCost");
            var customName = GetString(data, "customName");
            if (identifier.Length == 0 || count <= 0 || damage < 0 || unbreakable > 1)
                return false;

            var lore = GetList(data, "lore", Nbt.TagType.String)
                .Select(value => value.AsString()).ToArray();
            var itemNbt = GetBytes(data, "itemNbt");
            var valuesNbt = GetBytes(data, "valuesNbt");
            var enchantments = GetList(data, "enchantments", Nbt.TagType.Compound)
                .Select(value =>
                {
                    var enchantment = value.AsCompound();
                    return new ItemEnchantment
                    {
                        Id = checked((uint)GetInt(enchantment, "id")),
                        Level = checked((uint)GetInt(enchantment, "level")),
                    };
                }).ToArray();

            var item = ItemCodec.Decode(identifier, metadata);
            item = ItemNbtCodec.Decode(item, itemNbt, depth, out var consumed);
            stack = new Item.Stack(
                item,
                count,
                checked((uint)damage),
                unbreakable == 1,
                anvilCost,
                customName,
                lore,
                consumed ? null : itemNbt,
                valuesNbt,
                enchantments);
            return true;
        }
        catch (Exception exception) when (exception is InvalidDataException or InvalidOperationException or OverflowException)
        {
            return false;
        }
    }

    private static int GetInt(Nbt.Compound data, string name) =>
        data.TryGetValue(name, out var value) && value.Type == Nbt.TagType.Int
            ? value.AsInt()
            : throw new InvalidDataException($"invalid nested item {name}");

    private static byte GetByte(Nbt.Compound data, string name) =>
        data.TryGetValue(name, out var value) && value.Type == Nbt.TagType.Byte
            ? value.AsByte()
            : throw new InvalidDataException($"invalid nested item {name}");

    private static string GetString(Nbt.Compound data, string name) =>
        data.TryGetValue(name, out var value) && value.Type == Nbt.TagType.String
            ? value.AsString()
            : throw new InvalidDataException($"invalid nested item {name}");

    private static byte[] GetBytes(Nbt.Compound data, string name)
    {
        if (!data.TryGetValue(name, out var value))
            throw new InvalidDataException($"invalid nested item {name}");
        if (value.Type == Nbt.TagType.ByteArray) return value.AsByteArray();
        if (value.Type != Nbt.TagType.List)
            throw new InvalidDataException($"invalid nested item {name}");
        var list = value.AsList();
        if (list.Count == 0 && list.ElementType == Nbt.TagType.End) return [];
        if (list.ElementType != Nbt.TagType.Byte)
            throw new InvalidDataException($"invalid nested item {name}");
        return list.Select(entry => entry.AsByte()).ToArray();
    }

    private static Nbt.ListValue GetList(Nbt.Compound data, string name, Nbt.TagType elementType)
    {
        if (!data.TryGetValue(name, out var value) || value.Type != Nbt.TagType.List)
            throw new InvalidDataException($"invalid nested item {name}");
        var list = value.AsList();
        if (list.Count == 0 && list.ElementType == Nbt.TagType.End) return list;
        if (list.ElementType != elementType)
            throw new InvalidDataException($"invalid nested item {name}");
        return list;
    }
}
