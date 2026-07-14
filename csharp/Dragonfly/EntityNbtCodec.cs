namespace Dragonfly;

// Private transport codec for EntityType.DecodeNBT/EncodeNBT. Public entity
// APIs keep Dragonfly's map shape; NBT bytes exist only at ABI boundary.
internal static class EntityNbtCodec
{
    internal static Dictionary<string, object?> Decode(ReadOnlySpan<byte> data)
    {
        if (data.Length == 0) return new(StringComparer.Ordinal);
        var values = StackValueCodec.Decode(data.ToArray());
        return values.ToDictionary(
            entry => entry.Key,
            entry => (object?)entry.Value,
            StringComparer.Ordinal);
    }

    internal static byte[] Encode(IReadOnlyDictionary<string, object?> values)
    {
        ArgumentNullException.ThrowIfNull(values);
        var encoded = new Dictionary<string, object>(values.Count, StringComparer.Ordinal);
        foreach (var (name, value) in values)
        {
            if (value is null)
                throw new ArgumentException("entity NBT compounds cannot contain null values", nameof(values));
            encoded.Add(name, value);
        }
        return StackValueCodec.Encode(encoded);
    }
}
