using Dragonfly.Native;
using System.Globalization;

namespace Dragonfly;

public sealed partial class Player : Cmd.Source, Cmd.NamedTarget, World.Entity
{
    private readonly TimeSpan _latency;
    private readonly Vector3 _position;
    private readonly Cmd.Output? _commandOutput;
    private readonly ulong _invocation;

    internal Player(
        PlayerId id,
        string name = "",
        TimeSpan latency = default,
        Vector3 position = default,
        Cmd.Output? commandOutput = null,
        ulong invocation = 0)
    {
        Id = id;
        PlayerName = name;
        _latency = latency;
        _position = position;
        _commandOutput = commandOutput;
        _invocation = invocation;
    }

    internal PlayerId Id { get; }
    internal string PlayerName { get; }

    public string Name() => PlayerName;
    public TimeSpan Latency() => _latency;
    public World.EntityHandle H() => new(EntityId());
    public void Close() => PluginBridge.Host.CloseEntity(_invocation, EntityId());
    public void SendCommandOutput(Cmd.Output output) => _commandOutput?.Merge(output);

    private void SendText(uint kind, string message) =>
        PluginBridge.Host.SendPlayerText(_invocation, Id, kind, message);

    private static string FormatArguments(object?[] values) => string.Join(
        " ",
        values.Select(FormatArgument));

    private static string FormatArgument(object? value) => value switch
    {
        null => "<nil>",
        bool boolean => boolean ? "true" : "false",
        float number => number.ToString("G", CultureInfo.InvariantCulture).Replace('E', 'e'),
        double number => number.ToString("G", CultureInfo.InvariantCulture).Replace('E', 'e'),
        IFormattable formattable => formattable.ToString(null, CultureInfo.InvariantCulture) ?? "",
        _ => value.ToString() ?? "",
    };

    internal ulong Invocation => _invocation;

    internal static unsafe Player FromCommandArgument(string argument, ulong invocation)
    {
        var parts = argument.Split(':', 7);
        if (parts.Length != 7) throw new ArgumentException("player is no longer available");
        byte[] bytes;
        try
        {
            bytes = Convert.FromHexString(parts[0]);
        }
        catch (FormatException)
        {
            throw new ArgumentException("player is no longer available");
        }
        if (bytes.Length != 16 ||
            !ulong.TryParse(parts[1], NumberStyles.None, CultureInfo.InvariantCulture, out var generation) ||
            !long.TryParse(parts[2], NumberStyles.Integer, CultureInfo.InvariantCulture, out var latency) ||
            !double.TryParse(parts[3], NumberStyles.Float, CultureInfo.InvariantCulture, out var x) ||
            !double.TryParse(parts[4], NumberStyles.Float, CultureInfo.InvariantCulture, out var y) ||
            !double.TryParse(parts[5], NumberStyles.Float, CultureInfo.InvariantCulture, out var z))
            throw new ArgumentException("player is no longer available");

        var id = new PlayerId { Generation = generation };
        for (var index = 0; index < bytes.Length; index++) id.Bytes[index] = bytes[index];
        return new Player(
            id,
            parts[6],
            TimeSpan.FromMilliseconds(Math.Max(latency, 0)),
            new Vector3(x, y, z),
            invocation: invocation);
    }

    internal static unsafe bool SameId(PlayerId left, PlayerId right)
    {
        if (left.Generation != right.Generation) return false;
        for (var index = 0; index < 16; index++)
            if (left.Bytes[index] != right.Bytes[index]) return false;
        return true;
    }

    private unsafe EntityId EntityId()
    {
        var player = Id;
        var entity = new EntityId { Generation = player.Generation };
        for (var index = 0; index < 16; index++) entity.Bytes[index] = player.Bytes[index];
        return entity;
    }

    public sealed class Context : World.Context
    {
        internal Context(Player player, bool cancelled) : base(player.Invocation, cancelled)
        {
            Value = player;
        }

        private Player Value { get; }
        public Player Player() => Value;
    }
}
