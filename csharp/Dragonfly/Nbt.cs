using System.Buffers;
using System.Buffers.Binary;
using System.Text;

namespace Dragonfly;

// Internal fixed-width little-endian NBT used by Dragonfly item codecs. Keep this
// out of the plugin API: Public item types expose Dragonfly values, not raw NBT.
internal static class Nbt
{
    internal const int MaxBytes = 16 << 20;
    internal const int MaxDepth = 64;
    internal const int MaxCollectionLength = 1 << 20;
    internal const int MaxNodeCount = 1 << 16;
    private const int MaxStringBytes = ushort.MaxValue;
    private static readonly UTF8Encoding Utf8 = new(false, true);

    internal enum TagType : byte
    {
        End,
        Byte,
        Short,
        Int,
        Long,
        Float,
        Double,
        ByteArray,
        String,
        List,
        Compound,
        IntArray,
        LongArray,
    }

    internal readonly struct Value
    {
        private readonly object _data;

        private Value(TagType type, object data)
        {
            Type = type;
            _data = data;
        }

        internal TagType Type { get; }

        internal static Value Byte(byte value) => new(TagType.Byte, value);
        internal static Value Short(short value) => new(TagType.Short, value);
        internal static Value Int(int value) => new(TagType.Int, value);
        internal static Value Long(long value) => new(TagType.Long, value);
        internal static Value Float(float value) => new(TagType.Float, value);
        internal static Value Double(double value) => new(TagType.Double, value);
        internal static Value ByteArray(byte[] value)
        {
            ArgumentNullException.ThrowIfNull(value);
            return new(TagType.ByteArray, (byte[])value.Clone());
        }

        internal static Value String(string value)
        {
            ArgumentNullException.ThrowIfNull(value);
            return new(TagType.String, value);
        }

        internal static Value List(TagType elementType, params Value[] values) => new(
            TagType.List,
            new ListValue(elementType, values));
        internal static Value List(ListValue value)
        {
            ArgumentNullException.ThrowIfNull(value);
            return new(TagType.List, value);
        }

        internal static Value Compound(Compound value)
        {
            ArgumentNullException.ThrowIfNull(value);
            return new(TagType.Compound, value);
        }

        internal static Value IntArray(int[] value)
        {
            ArgumentNullException.ThrowIfNull(value);
            return new(TagType.IntArray, (int[])value.Clone());
        }

        internal static Value LongArray(long[] value)
        {
            ArgumentNullException.ThrowIfNull(value);
            return new(TagType.LongArray, (long[])value.Clone());
        }

        internal byte AsByte() => Get<byte>(TagType.Byte);
        internal short AsShort() => Get<short>(TagType.Short);
        internal int AsInt() => Get<int>(TagType.Int);
        internal long AsLong() => Get<long>(TagType.Long);
        internal float AsFloat() => Get<float>(TagType.Float);
        internal double AsDouble() => Get<double>(TagType.Double);
        internal byte[] AsByteArray() => (byte[])Get<byte[]>(TagType.ByteArray).Clone();
        internal string AsString() => Get<string>(TagType.String);
        internal ListValue AsList() => Get<ListValue>(TagType.List);
        internal Compound AsCompound() => Get<Compound>(TagType.Compound);
        internal int[] AsIntArray() => (int[])Get<int[]>(TagType.IntArray).Clone();
        internal long[] AsLongArray() => (long[])Get<long[]>(TagType.LongArray).Clone();

        private T Get<T>(TagType expected)
        {
            if (Type != expected) throw new InvalidOperationException($"NBT value is {Type}, not {expected}");
            return (T)_data;
        }

        private object RawData() => _data;

        internal static object Data(Value value) => value.RawData();
    }

    internal sealed class Compound : Dictionary<string, Value>
    {
        internal Compound() : base(StringComparer.Ordinal)
        {
        }

        internal Compound(IEnumerable<KeyValuePair<string, Value>> values) : base(StringComparer.Ordinal)
        {
            ArgumentNullException.ThrowIfNull(values);
            foreach (var (name, value) in values) Add(name, value);
        }
    }

    internal sealed class ListValue : IReadOnlyList<Value>
    {
        private readonly Value[] _values;

        internal ListValue(TagType elementType, IEnumerable<Value> values)
        {
            if (elementType is < TagType.End or > TagType.LongArray)
                throw new ArgumentOutOfRangeException(nameof(elementType));
            ArgumentNullException.ThrowIfNull(values);
            _values = values.ToArray();
            if (_values.Length > MaxCollectionLength)
                throw new ArgumentOutOfRangeException(nameof(values), "NBT list is too large");
            if (elementType == TagType.End && _values.Length != 0)
                throw new ArgumentException("TAG_End is only valid for an empty NBT list", nameof(elementType));
            if (_values.Any(value => value.Type != elementType))
                throw new ArgumentException("NBT list elements must have one tag type", nameof(values));
            ElementType = elementType;
        }

        internal TagType ElementType { get; }
        public int Count => _values.Length;
        public Value this[int index] => _values[index];
        public IEnumerator<Value> GetEnumerator() => ((IEnumerable<Value>)_values).GetEnumerator();
        System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => _values.GetEnumerator();
    }

    internal static byte[] Encode(Compound root)
    {
        ArgumentNullException.ThrowIfNull(root);
        var writer = new Writer();
        writer.WriteRoot(root);
        return writer.ToArray();
    }

    internal static Compound Decode(ReadOnlySpan<byte> data)
    {
        if (data.Length > MaxBytes) throw new InvalidDataException("NBT exceeds maximum size");
        var reader = new Reader(data);
        var root = reader.ReadRoot();
        if (!reader.AtEnd) throw new InvalidDataException("Trailing data after NBT root");
        return root;
    }

    internal static bool TryDecode(ReadOnlySpan<byte> data, out Compound? root)
    {
        try
        {
            root = Decode(data);
            return true;
        }
        catch (Exception exception) when (exception is InvalidDataException or DecoderFallbackException)
        {
            root = null;
            return false;
        }
    }

    internal static bool Equivalent(ReadOnlySpan<byte> left, ReadOnlySpan<byte> right)
    {
        if (left.SequenceEqual(right)) return true;
        return TryDecode(left, out var leftRoot) && TryDecode(right, out var rightRoot) &&
            Equivalent(leftRoot!, rightRoot!);
    }

    private static bool Equivalent(Compound left, Compound right)
    {
        if (left.Count != right.Count) return false;
        foreach (var (key, value) in left)
        {
            if (!right.TryGetValue(key, out var other) || !Equivalent(value, other)) return false;
        }
        return true;
    }

    private static bool Equivalent(Value left, Value right)
    {
        if (left.Type != right.Type) return false;
        return left.Type switch
        {
            TagType.Byte => left.AsByte() == right.AsByte(),
            TagType.Short => left.AsShort() == right.AsShort(),
            TagType.Int => left.AsInt() == right.AsInt(),
            TagType.Long => left.AsLong() == right.AsLong(),
            TagType.Float => left.AsFloat() == right.AsFloat(),
            TagType.Double => left.AsDouble() == right.AsDouble(),
            TagType.ByteArray => left.AsByteArray().AsSpan().SequenceEqual(right.AsByteArray()),
            TagType.String => left.AsString() == right.AsString(),
            TagType.List => Equivalent(left.AsList(), right.AsList()),
            TagType.Compound => Equivalent(left.AsCompound(), right.AsCompound()),
            TagType.IntArray => left.AsIntArray().AsSpan().SequenceEqual(right.AsIntArray()),
            TagType.LongArray => left.AsLongArray().AsSpan().SequenceEqual(right.AsLongArray()),
            _ => false,
        };
    }

    private static bool Equivalent(ListValue left, ListValue right)
    {
        if (left.ElementType != right.ElementType || left.Count != right.Count) return false;
        for (var index = 0; index < left.Count; index++)
        {
            if (!Equivalent(left[index], right[index])) return false;
        }
        return true;
    }

    private sealed class Writer
    {
        private readonly ArrayBufferWriter<byte> _buffer = new(256);
        private int _nodes;

        internal byte[] ToArray() => _buffer.WrittenSpan.ToArray();

        internal void WriteRoot(Compound root)
        {
            WriteByte((byte)TagType.Compound);
            WriteString(string.Empty);
            WriteCompound(root, 0);
        }

        private void WriteNamed(string name, Value value, int depth)
        {
            ArgumentNullException.ThrowIfNull(name);
            CountNode();
            ValidateType(value.Type, false);
            WriteByte((byte)value.Type);
            WriteString(name);
            WritePayload(value, depth);
        }

        private void WritePayload(Value value, int depth)
        {
            CheckDepth(depth);
            switch (value.Type)
            {
                case TagType.Byte:
                    WriteByte((byte)Value.Data(value));
                    break;
                case TagType.Short:
                    WriteInt16((short)Value.Data(value));
                    break;
                case TagType.Int:
                    WriteInt32((int)Value.Data(value));
                    break;
                case TagType.Long:
                    WriteInt64((long)Value.Data(value));
                    break;
                case TagType.Float:
                    WriteInt32(BitConverter.SingleToInt32Bits((float)Value.Data(value)));
                    break;
                case TagType.Double:
                    WriteInt64(BitConverter.DoubleToInt64Bits((double)Value.Data(value)));
                    break;
                case TagType.ByteArray:
                    WriteByteArray((byte[])Value.Data(value));
                    break;
                case TagType.String:
                    WriteString((string)Value.Data(value));
                    break;
                case TagType.List:
                    WriteList((ListValue)Value.Data(value), depth);
                    break;
                case TagType.Compound:
                    WriteCompound((Compound)Value.Data(value), depth);
                    break;
                case TagType.IntArray:
                    WriteIntArray((int[])Value.Data(value));
                    break;
                case TagType.LongArray:
                    WriteLongArray((long[])Value.Data(value));
                    break;
                default:
                    throw new InvalidDataException("TAG_End has no payload");
            }
        }

        private void WriteList(ListValue list, int depth)
        {
            CheckDepth(depth);
            ValidateCollectionLength(list.Count);
            ValidateType(list.ElementType, list.Count == 0);
            WriteByte((byte)list.ElementType);
            WriteInt32(list.Count);
            foreach (var value in list)
            {
                CountNode();
                if (value.Type != list.ElementType)
                    throw new InvalidDataException("NBT list element type changed");
                WritePayload(value, depth + 1);
            }
        }

        private void WriteCompound(Compound compound, int depth)
        {
            CheckDepth(depth);
            ValidateCollectionLength(compound.Count);
            foreach (var (name, value) in compound) WriteNamed(name, value, depth + 1);
            WriteByte((byte)TagType.End);
        }

        private void WriteByteArray(byte[] values)
        {
            if (values.Length > MaxBytes) throw new InvalidDataException("NBT byte array is too large");
            WriteInt32(values.Length);
            Write(values);
        }

        private void WriteIntArray(int[] values)
        {
            ValidateCollectionLength(values.Length);
            WriteInt32(values.Length);
            foreach (var value in values) WriteInt32(value);
        }

        private void WriteLongArray(long[] values)
        {
            ValidateCollectionLength(values.Length);
            WriteInt32(values.Length);
            foreach (var value in values) WriteInt64(value);
        }

        private void WriteString(string value)
        {
            ArgumentNullException.ThrowIfNull(value);
            var length = Utf8.GetByteCount(value);
            if (length > MaxStringBytes) throw new InvalidDataException("NBT string is too long");
            WriteUInt16(checked((ushort)length));
            EnsureCapacity(length);
            var span = _buffer.GetSpan(length)[..length];
            Utf8.GetBytes(value, span);
            _buffer.Advance(length);
        }

        private void WriteByte(byte value)
        {
            EnsureCapacity(1);
            _buffer.GetSpan(1)[0] = value;
            _buffer.Advance(1);
        }

        private void WriteInt16(short value)
        {
            EnsureCapacity(sizeof(short));
            BinaryPrimitives.WriteInt16LittleEndian(_buffer.GetSpan(sizeof(short)), value);
            _buffer.Advance(sizeof(short));
        }

        private void WriteUInt16(ushort value)
        {
            EnsureCapacity(sizeof(ushort));
            BinaryPrimitives.WriteUInt16LittleEndian(_buffer.GetSpan(sizeof(ushort)), value);
            _buffer.Advance(sizeof(ushort));
        }

        private void WriteInt32(int value)
        {
            EnsureCapacity(sizeof(int));
            BinaryPrimitives.WriteInt32LittleEndian(_buffer.GetSpan(sizeof(int)), value);
            _buffer.Advance(sizeof(int));
        }

        private void WriteInt64(long value)
        {
            EnsureCapacity(sizeof(long));
            BinaryPrimitives.WriteInt64LittleEndian(_buffer.GetSpan(sizeof(long)), value);
            _buffer.Advance(sizeof(long));
        }

        private void Write(ReadOnlySpan<byte> value)
        {
            EnsureCapacity(value.Length);
            value.CopyTo(_buffer.GetSpan(value.Length));
            _buffer.Advance(value.Length);
        }

        private void EnsureCapacity(int length)
        {
            if (length < 0 || _buffer.WrittenCount > MaxBytes - length)
                throw new InvalidDataException("NBT exceeds maximum size");
        }

        private void CountNode()
        {
            if (++_nodes > MaxNodeCount) throw new InvalidDataException("NBT contains too many values");
        }
    }

    private ref struct Reader
    {
        private readonly ReadOnlySpan<byte> _data;
        private int _offset;
        private int _nodes;

        internal Reader(ReadOnlySpan<byte> data) => _data = data;
        internal readonly bool AtEnd => _offset == _data.Length;

        internal Compound ReadRoot()
        {
            var type = ReadType(false);
            if (type != TagType.Compound) throw new InvalidDataException("NBT root must be a compound");
            _ = ReadString();
            return ReadCompound(0);
        }

        private Value ReadPayload(TagType type, int depth)
        {
            CheckDepth(depth);
            CountNode();
            return type switch
            {
                TagType.Byte => Value.Byte(ReadByte()),
                TagType.Short => Value.Short(ReadInt16()),
                TagType.Int => Value.Int(ReadInt32()),
                TagType.Long => Value.Long(ReadInt64()),
                TagType.Float => Value.Float(BitConverter.Int32BitsToSingle(ReadInt32())),
                TagType.Double => Value.Double(BitConverter.Int64BitsToDouble(ReadInt64())),
                TagType.ByteArray => Value.ByteArray(ReadByteArray()),
                TagType.String => Value.String(ReadString()),
                TagType.List => Value.List(ReadList(depth)),
                TagType.Compound => Value.Compound(ReadCompound(depth)),
                TagType.IntArray => Value.IntArray(ReadIntArray()),
                TagType.LongArray => Value.LongArray(ReadLongArray()),
                _ => throw new InvalidDataException("TAG_End has no payload"),
            };
        }

        private ListValue ReadList(int depth)
        {
            CheckDepth(depth);
            var elementType = ReadType(true);
            var length = ReadLength();
            if (elementType == TagType.End && length != 0)
                throw new InvalidDataException("TAG_End is only valid for an empty NBT list");
            EnsureRemaining(checked(length * MinimumPayloadSize(elementType)));
            var values = new Value[length];
            for (var index = 0; index < length; index++) values[index] = ReadPayload(elementType, depth + 1);
            return new ListValue(elementType, values);
        }

        private Compound ReadCompound(int depth)
        {
            CheckDepth(depth);
            var compound = new Compound();
            while (true)
            {
                var type = ReadType(true);
                if (type == TagType.End) return compound;
                var name = ReadString();
                if (compound.Count >= MaxCollectionLength)
                    throw new InvalidDataException("NBT compound is too large");
                if (!compound.TryAdd(name, ReadPayload(type, depth + 1)))
                    throw new InvalidDataException("NBT compound contains a duplicate name");
            }
        }

        private byte[] ReadByteArray()
        {
            var length = ReadInt32();
            if (length is < 0 or > MaxBytes)
                throw new InvalidDataException("Invalid NBT byte array length");
            return Read(length).ToArray();
        }

        private int[] ReadIntArray()
        {
            var length = ReadLength();
            EnsureRemaining(checked(length * sizeof(int)));
            var values = new int[length];
            for (var index = 0; index < length; index++) values[index] = ReadInt32();
            return values;
        }

        private long[] ReadLongArray()
        {
            var length = ReadLength();
            EnsureRemaining(checked(length * sizeof(long)));
            var values = new long[length];
            for (var index = 0; index < length; index++) values[index] = ReadInt64();
            return values;
        }

        private string ReadString()
        {
            var length = ReadUInt16();
            return Utf8.GetString(Read(length));
        }

        private int ReadLength()
        {
            var length = ReadInt32();
            if (length is < 0 or > MaxCollectionLength)
                throw new InvalidDataException("Invalid NBT collection length");
            return length;
        }

        private TagType ReadType(bool allowEnd)
        {
            var type = (TagType)ReadByte();
            ValidateType(type, allowEnd);
            return type;
        }

        private byte ReadByte() => Read(1)[0];
        private short ReadInt16() => BinaryPrimitives.ReadInt16LittleEndian(Read(sizeof(short)));
        private ushort ReadUInt16() => BinaryPrimitives.ReadUInt16LittleEndian(Read(sizeof(ushort)));
        private int ReadInt32() => BinaryPrimitives.ReadInt32LittleEndian(Read(sizeof(int)));
        private long ReadInt64() => BinaryPrimitives.ReadInt64LittleEndian(Read(sizeof(long)));

        private ReadOnlySpan<byte> Read(int length)
        {
            EnsureRemaining(length);
            var value = _data.Slice(_offset, length);
            _offset += length;
            return value;
        }

        private readonly void EnsureRemaining(int length)
        {
            if (length < 0 || _offset > _data.Length - length)
                throw new InvalidDataException("Unexpected end of NBT data");
        }

        private void CountNode()
        {
            if (++_nodes > MaxNodeCount) throw new InvalidDataException("NBT contains too many values");
        }

        private static int MinimumPayloadSize(TagType type) => type switch
        {
            TagType.End => 0,
            TagType.Byte => 1,
            TagType.Short => 2,
            TagType.Int or TagType.Float => 4,
            TagType.Long or TagType.Double => 8,
            TagType.String => 2,
            TagType.List or TagType.ByteArray or TagType.IntArray or TagType.LongArray => 4,
            TagType.Compound => 1,
            _ => throw new InvalidDataException("Invalid NBT tag type"),
        };
    }

    private static void ValidateType(TagType type, bool allowEnd)
    {
        if (type > TagType.LongArray || (!allowEnd && type == TagType.End))
            throw new InvalidDataException("Invalid NBT tag type");
    }

    private static void ValidateCollectionLength(int length)
    {
        if (length is < 0 or > MaxCollectionLength)
            throw new InvalidDataException("Invalid NBT collection length");
    }

    private static void CheckDepth(int depth)
    {
        if (depth >= MaxDepth) throw new InvalidDataException("NBT maximum depth exceeded");
    }
}
