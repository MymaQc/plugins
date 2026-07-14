#nullable enable

namespace Dragonfly;

internal static class ItemNbtCodec
{
    internal static bool TryEncode(World.Item? item, out byte[] data) => TryEncode(item, 0, out data);

    internal static bool TryEncode(World.Item? item, int depth, out byte[] data)
    {
        if (depth > CrossbowStackNbt.MaxDepth) throw new InvalidDataException("nested item stack is too deep");
        switch (item)
        {
            case Item.Helmet armour:
                data = EncodeArmour(armour.Tier, armour.Trim);
                return true;
            case Item.Chestplate armour:
                data = EncodeArmour(armour.Tier, armour.Trim);
                return true;
            case Item.Leggings armour:
                data = EncodeArmour(armour.Tier, armour.Trim);
                return true;
            case Item.Boots armour:
                data = EncodeArmour(armour.Tier, armour.Trim);
                return true;
            case Item.Firework firework:
                data = EncodeFirework(firework);
                return true;
            case Item.FireworkStar star:
                data = EncodeFireworkStar(star);
                return true;
            case Item.BookAndQuill book:
                data = EncodeBookAndQuill(book);
                return true;
            case Item.WrittenBook book:
                data = EncodeWrittenBook(book);
                return true;
            case Item.Crossbow crossbow:
                data = CrossbowStackNbt.Encode(crossbow.Item, depth + 1);
                return true;
            default:
                if (ItemCodec.TryRawNbt(item, out data)) return true;
                data = [];
                return false;
        }
    }

    internal static World.Item Decode(World.Item item, ReadOnlySpan<byte> data, out bool consumed) =>
        Decode(item, data, 0, out consumed);

    internal static World.Item Decode(World.Item item, ReadOnlySpan<byte> data, int depth, out bool consumed)
    {
        if (depth > CrossbowStackNbt.MaxDepth) throw new InvalidDataException("nested item stack is too deep");
        switch (item)
        {
            case Item.Helmet armour:
                consumed = true;
                var (helmetTier, helmetTrim) = DecodeArmour(armour.Tier, data);
                return new Item.Helmet(helmetTier, helmetTrim);
            case Item.Chestplate armour:
                consumed = true;
                var (chestplateTier, chestplateTrim) = DecodeArmour(armour.Tier, data);
                return new Item.Chestplate(chestplateTier, chestplateTrim);
            case Item.Leggings armour:
                consumed = true;
                var (leggingsTier, leggingsTrim) = DecodeArmour(armour.Tier, data);
                return new Item.Leggings(leggingsTier, leggingsTrim);
            case Item.Boots armour:
                consumed = true;
                var (bootsTier, bootsTrim) = DecodeArmour(armour.Tier, data);
                return new Item.Boots(bootsTier, bootsTrim);
            case Item.Firework firework:
                consumed = true;
                return DecodeFirework(firework, data);
            case Item.FireworkStar star:
                consumed = true;
                return DecodeFireworkStar(star, data);
            case Item.BookAndQuill book:
                consumed = true;
                return DecodeBookAndQuill(book, data);
            case Item.WrittenBook book:
                consumed = true;
                return DecodeWrittenBook(book, data);
            case Item.Crossbow:
                consumed = true;
                return new Item.Crossbow(CrossbowStackNbt.Decode(data, depth + 1));
            default:
                if (ItemCodec.TryWithRawNbt(item, data, out var encoded))
                {
                    consumed = true;
                    return encoded;
                }
                consumed = false;
                return item;
        }
    }

    private static byte[] EncodeArmour(Item.ArmourTier tier, Item.ArmourTrim trim)
    {
        var root = new Nbt.Compound();
        if (tier is Item.ArmourTierLeather leather && leather.Colour != default)
            root["customColor"] = Nbt.Value.Int(ColourARGB(leather.Colour));
        if (!trim.Zero())
        {
            root["Trim"] = Nbt.Value.Compound(new Nbt.Compound
            {
                ["Material"] = Nbt.Value.String(trim.Material!.TrimMaterial()),
                ["Pattern"] = Nbt.Value.String(trim.Template.String()),
            });
        }
        return Nbt.Encode(root);
    }

    private static (Item.ArmourTier Tier, Item.ArmourTrim Trim) DecodeArmour(
        Item.ArmourTier tier,
        ReadOnlySpan<byte> data)
    {
        if (data.IsEmpty) return (tier, default);
        var root = Nbt.Decode(data);
        if (tier is Item.ArmourTierLeather && root.TryGetValue("customColor", out var encodedColour) &&
            encodedColour.Type == Nbt.TagType.Int)
            tier = new Item.ArmourTierLeather(ColourFromARGB(encodedColour.AsInt()));
        return (tier, DecodeArmourTrim(root));
    }

    private static Item.ArmourTrim DecodeArmourTrim(Nbt.Compound root)
    {
        if (!root.TryGetValue("Trim", out var encoded) || encoded.Type != Nbt.TagType.Compound)
            return default;
        var trim = encoded.AsCompound();
        if (!trim.TryGetValue("Material", out var encodedMaterial) || encodedMaterial.Type != Nbt.TagType.String ||
            !trim.TryGetValue("Pattern", out var encodedPattern) || encodedPattern.Type != Nbt.TagType.String ||
            !ArmourCodec.TryTrimMaterial(encodedMaterial.AsString(), out var material) ||
            !ArmourCodec.TryTemplate(encodedPattern.AsString(), out var template))
            return default;
        return new Item.ArmourTrim(template, material);
    }

    private static Color.RGBA ColourFromARGB(int colour)
    {
        var value = unchecked((uint)colour);
        return new Color.RGBA(
            (byte)(value >> 16),
            (byte)(value >> 8),
            (byte)value,
            (byte)(value >> 24));
    }

    private static int ColourARGB(Color.RGBA value)
    {
        if (value.R == 0 && value.G == 0 && value.B == 0) return unchecked((int)0xff000000);
        return unchecked((int)((uint)value.A << 24 | (uint)value.R << 16 | (uint)value.G << 8 | value.B));
    }

    private static byte[] EncodeFirework(Item.Firework firework)
    {
        var explosions = firework.Explosions
            .Select(explosion => Nbt.Value.Compound(EncodeFireworkExplosion(explosion)))
            .ToArray();
        var flight = unchecked((byte)(unchecked(firework.Duration.Ticks - 5_000_000L) / 5_000_000L));
        return Nbt.Encode(new Nbt.Compound
        {
            ["Fireworks"] = Nbt.Value.Compound(new Nbt.Compound
            {
                ["Explosions"] = Nbt.Value.List(Nbt.TagType.Compound, explosions),
                ["Flight"] = Nbt.Value.Byte(flight),
            }),
        });
    }

    private static Item.Firework DecodeFirework(Item.Firework firework, ReadOnlySpan<byte> data)
    {
        if (data.IsEmpty) return firework;
        var root = Nbt.Decode(data);
        if (!root.TryGetValue("Fireworks", out var encoded) || encoded.Type != Nbt.TagType.Compound)
            return firework;
        var fireworks = encoded.AsCompound();
        var explosions = firework.Explosions;
        if (fireworks.TryGetValue("Explosions", out var encodedExplosions) && encodedExplosions.Type == Nbt.TagType.List)
        {
            explosions = encodedExplosions.AsList()
                .Select(value => value.Type == Nbt.TagType.Compound
                    ? DecodeFireworkExplosion(default, value.AsCompound())
                    : throw new InvalidDataException("invalid firework explosion"))
                .ToArray();
        }
        var duration = firework.Duration;
        if (fireworks.TryGetValue("Flight", out var encodedFlight) && encodedFlight.Type == Nbt.TagType.Byte)
            duration = TimeSpan.FromMilliseconds((encodedFlight.AsByte() + 1L) * 500L);
        return new Item.Firework(duration, explosions);
    }

    private static byte[] EncodeFireworkStar(Item.FireworkStar star) => Nbt.Encode(new Nbt.Compound
    {
        ["FireworksItem"] = Nbt.Value.Compound(EncodeFireworkExplosion(star.FireworkExplosion)),
        ["customColor"] = Nbt.Value.Int(ColourARGB(star.FireworkExplosion.Colour)),
    });

    private static Item.FireworkStar DecodeFireworkStar(Item.FireworkStar star, ReadOnlySpan<byte> data)
    {
        if (data.IsEmpty) return star;
        var root = Nbt.Decode(data);
        if (!root.TryGetValue("FireworksItem", out var encoded) || encoded.Type != Nbt.TagType.Compound)
            return star;
        return new Item.FireworkStar(DecodeFireworkExplosion(star.FireworkExplosion, encoded.AsCompound()));
    }

    private static Nbt.Compound EncodeFireworkExplosion(Item.FireworkExplosion explosion)
    {
        var fade = explosion.Fades ? new[] { InvertedColour(explosion.Fade) } : [];
        return new Nbt.Compound
        {
            ["FireworkType"] = Nbt.Value.Byte(explosion.Shape.Uint8()),
            ["FireworkColor"] = Nbt.Value.ByteArray([InvertedColour(explosion.Colour)]),
            ["FireworkFade"] = Nbt.Value.ByteArray(fade),
            ["FireworkFlicker"] = Nbt.Value.Byte(explosion.Twinkle ? (byte)1 : (byte)0),
            ["FireworkTrail"] = Nbt.Value.Byte(explosion.Trail ? (byte)1 : (byte)0),
        };
    }

    private static Item.FireworkExplosion DecodeFireworkExplosion(
        Item.FireworkExplosion explosion,
        Nbt.Compound data)
    {
        var shape = new Item.FireworkShape(GetByte(data, "FireworkType"));
        if (shape.Id >= 5) throw new InvalidDataException("invalid firework shape");
        var colour = explosion.Colour;
        if (data.TryGetValue("FireworkColor", out var encodedColour) && encodedColour.Type == Nbt.TagType.ByteArray)
        {
            var colours = encodedColour.AsByteArray();
            if (colours.Length == 1) colour = ColourFromInverted(colours[0]);
        }
        var fade = explosion.Fade;
        var fades = explosion.Fades;
        if (data.TryGetValue("FireworkFade", out var encodedFade) && encodedFade.Type == Nbt.TagType.ByteArray)
        {
            var values = encodedFade.AsByteArray();
            if (values.Length == 1)
            {
                fade = ColourFromInverted(values[0]);
                fades = true;
            }
        }
        return new Item.FireworkExplosion
        {
            Shape = shape,
            Colour = colour,
            Fade = fade,
            Fades = fades,
            Twinkle = GetByte(data, "FireworkFlicker") == 1,
            Trail = GetByte(data, "FireworkTrail") == 1,
        };
    }

    private static byte GetByte(Nbt.Compound data, string name)
    {
        if (!data.TryGetValue(name, out var value) || value.Type != Nbt.TagType.Byte)
            throw new InvalidDataException($"invalid firework {name}");
        return value.AsByte();
    }

    private static byte InvertedColour(Item.Colour colour) => (byte)(~colour.Uint8() & 0xf);

    private static Item.Colour ColourFromInverted(byte colour) => new(~colour & 0xf);

    private static int ColourARGB(Item.Colour colour)
    {
        var value = colour.RGBA();
        if (value.R == 0 && value.G == 0 && value.B == 0) return unchecked((int)0xff000000);
        return unchecked((int)((uint)value.A << 24 | (uint)value.R << 16 | (uint)value.G << 8 | value.B));
    }

    private static byte[] EncodeBookAndQuill(Item.BookAndQuill book)
    {
        if (book.TotalPages() == 0) return [];
        return Nbt.Encode(new Nbt.Compound
        {
            ["pages"] = EncodePages(book.Pages),
        });
    }

    private static Item.BookAndQuill DecodeBookAndQuill(Item.BookAndQuill book, ReadOnlySpan<byte> data)
    {
        if (data.IsEmpty) return book;
        var pages = new List<string>(book.Pages);
        if (Nbt.Decode(data).TryGetValue("pages", out var encoded) && encoded.Type == Nbt.TagType.List)
        {
            foreach (var page in encoded.AsList())
            {
                if (page.Type != Nbt.TagType.Compound ||
                    !page.AsCompound().TryGetValue("text", out var text) || text.Type != Nbt.TagType.String)
                    continue;
                pages.Add(text.AsString());
            }
        }
        return new Item.BookAndQuill(pages.ToArray());
    }

    private static byte[] EncodeWrittenBook(Item.WrittenBook book) => Nbt.Encode(new Nbt.Compound
    {
        ["pages"] = EncodePages(book.Pages),
        ["author"] = Nbt.Value.String(book.Author),
        ["title"] = Nbt.Value.String(book.Title),
        ["generation"] = Nbt.Value.Byte(book.Generation.Uint8()),
    });

    private static Item.WrittenBook DecodeWrittenBook(Item.WrittenBook book, ReadOnlySpan<byte> data)
    {
        if (data.IsEmpty) return book;
        var root = Nbt.Decode(data);
        var pages = DecodePages(root);
        var title = GetString(root, "title");
        var author = GetString(root, "author");
        var generation = Item.OriginalGeneration();
        if (root.TryGetValue("generation", out var encodedGeneration) && encodedGeneration.Type == Nbt.TagType.Byte)
        {
            generation = encodedGeneration.AsByte() switch
            {
                1 => Item.CopyGeneration(),
                2 => Item.CopyOfCopyGeneration(),
                _ => Item.OriginalGeneration(),
            };
        }
        return new Item.WrittenBook(title, author, generation, pages);
    }

    private static Nbt.Value EncodePages(IEnumerable<string> pages) => Nbt.Value.List(
        Nbt.TagType.Compound,
        pages.Select(page => Nbt.Value.Compound(new Nbt.Compound
        {
            ["text"] = Nbt.Value.String(page),
        })).ToArray());

    private static string[] DecodePages(Nbt.Compound root)
    {
        if (!root.TryGetValue("pages", out var encoded) || encoded.Type != Nbt.TagType.List) return [];
        var pages = new List<string>();
        foreach (var page in encoded.AsList())
        {
            if (page.Type != Nbt.TagType.Compound ||
                !page.AsCompound().TryGetValue("text", out var text) || text.Type != Nbt.TagType.String)
                throw new InvalidDataException("invalid written book page");
            pages.Add(text.AsString());
        }
        return pages.ToArray();
    }

    private static string GetString(Nbt.Compound root, string name) =>
        root.TryGetValue(name, out var value) && value.Type == Nbt.TagType.String
            ? value.AsString()
            : string.Empty;
}
