#nullable enable
using System.Collections;

namespace Dragonfly;

internal static class StackValueCodec
{
    internal static Dictionary<string, object> Decode(byte[]? data)
    {
        if (data is null || data.Length == 0) return new(StringComparer.Ordinal);
        var root = Nbt.Decode(data);
        return root.ToDictionary(entry => entry.Key, entry => Decode(entry.Value), StringComparer.Ordinal);
    }

    internal static byte[] Encode(IReadOnlyDictionary<string, object> values)
    {
        if (values.Count == 0) return [];
        var root = new Nbt.Compound();
        foreach (var (key, value) in values) root.Add(key, Encode(value));
        return Nbt.Encode(root);
    }

    private static object Decode(Nbt.Value value) => value.Type switch
    {
        Nbt.TagType.Byte => value.AsByte(),
        Nbt.TagType.Short => value.AsShort(),
        Nbt.TagType.Int => value.AsInt(),
        Nbt.TagType.Long => value.AsLong(),
        Nbt.TagType.Float => value.AsFloat(),
        Nbt.TagType.Double => value.AsDouble(),
        Nbt.TagType.ByteArray => value.AsByteArray(),
        Nbt.TagType.String => value.AsString(),
        Nbt.TagType.List => value.AsList().Select(Decode).ToArray(),
        Nbt.TagType.Compound => value.AsCompound().ToDictionary(
            entry => entry.Key,
            entry => Decode(entry.Value),
            StringComparer.Ordinal),
        Nbt.TagType.IntArray => value.AsIntArray(),
        Nbt.TagType.LongArray => value.AsLongArray(),
        _ => throw new InvalidDataException($"unsupported stack value tag {value.Type}"),
    };

    private static Nbt.Value Encode(object value)
    {
        ArgumentNullException.ThrowIfNull(value);
        return value switch
        {
            bool boolean => Nbt.Value.Byte(boolean ? (byte)1 : (byte)0),
            byte number => Nbt.Value.Byte(number),
            sbyte number => Nbt.Value.Byte(unchecked((byte)number)),
            short number => Nbt.Value.Short(number),
            int number => Nbt.Value.Int(number),
            long number => Nbt.Value.Long(number),
            float number => Nbt.Value.Float(number),
            double number => Nbt.Value.Double(number),
            string text => Nbt.Value.String(text),
            byte[] values => Nbt.Value.ByteArray(values),
            int[] values => Nbt.Value.IntArray(values),
            long[] values => Nbt.Value.LongArray(values),
            IDictionary values => EncodeCompound(values),
            IEnumerable values => EncodeList(values),
            _ => throw new ArgumentException($"stack value type {value.GetType()} cannot be encoded as NBT", nameof(value)),
        };
    }

    private static Nbt.Value EncodeCompound(IDictionary values)
    {
        var result = new Nbt.Compound();
        foreach (DictionaryEntry entry in values)
        {
            if (entry.Key is not string key || entry.Value is null)
                throw new ArgumentException("stack value compounds require string keys and non-null values", nameof(values));
            result.Add(key, Encode(entry.Value));
        }
        return Nbt.Value.Compound(result);
    }

    private static Nbt.Value EncodeList(IEnumerable values)
    {
        var encoded = values.Cast<object?>().Select(value => value is null
            ? throw new ArgumentException("stack value lists cannot contain null", nameof(values))
            : Encode(value)).ToArray();
        if (encoded.Length == 0) return Nbt.Value.List(Nbt.TagType.End);
        return Nbt.Value.List(encoded[0].Type, encoded);
    }
}
