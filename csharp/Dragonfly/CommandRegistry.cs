using System.Runtime.InteropServices;
using System.Text;
using Dragonfly.Native;

namespace Dragonfly;

internal static unsafe class CommandRegistry
{
    private static readonly object Gate = new();
    private static readonly List<Cmd.Command> Commands = [];
    private static NativeCommandSnapshot? Snapshot;
    private static bool Frozen;

    internal static void Register(Cmd.Command command)
    {
        ArgumentNullException.ThrowIfNull(command);
        lock (Gate)
        {
            if (Frozen) throw new InvalidOperationException("commands are already published");
            var index = Commands.FindIndex(existing => existing.Name() == command.Name());
            if (index < 0) Commands.Add(command);
            else Commands[index] = command;
        }
    }

    internal static void Clear()
    {
        lock (Gate)
        {
            Snapshot?.Dispose();
            Snapshot = null;
            Commands.Clear();
            Frozen = false;
        }
    }

    internal static CommandDescriptor* Native(ulong* count)
    {
        lock (Gate)
        {
            Frozen = true;
            Snapshot ??= new NativeCommandSnapshot([.. Commands]);
            if (count is not null) *count = (ulong)Snapshot.Commands.Length;
            return Snapshot.Descriptors;
        }
    }

    internal static int Execute(ulong commandIndex, CommandInput* input, CommandState* state)
    {
        if (input is null || state is null) return Abi.Error;
        NativeCommandSnapshot snapshot;
        lock (Gate)
        {
            snapshot = Snapshot ?? throw new InvalidOperationException("commands are not published");
        }
        if (commandIndex >= (ulong)snapshot.Commands.Length) return Abi.Error;
        var command = snapshot.Commands[checked((int)commandIndex)];
        if (input->Overload >= (ulong)command.Bindings.Length) return Abi.Error;

        var output = new Cmd.Output();
        try
        {
            using var invocationScope = InvocationContext.Enter(input->Invocation);
            var source = Source(input, output);
            var binding = command.Bindings[checked((int)input->Overload)];
            if (binding.Runnable is Cmd.Allower allower && !allower.Allow(source))
            {
                output.Error("This command cannot be used by this source.");
            }
            else
            {
                var arguments = Strings(input->Arguments, input->ArgumentCount);
                lock (binding.Gate)
                {
                    binding.Bind(arguments, source);
                    binding.Runnable.Run(source, output, input->Invocation == 0 ? null : new World.Tx(input->Invocation));
                }
            }
        }
        catch (ArgumentException exception)
        {
            output.Error(exception.Message);
        }
        catch
        {
            output.Error("Command execution failed.");
        }
        state->Failed = output.ErrorCount() == 0 ? (byte)0 : (byte)1;
        Write(&state->Output, output.Text());
        return Abi.Ok;
    }

    internal static int EnumOptions(
        ulong commandIndex,
        ulong overloadIndex,
        ulong parameterIndex,
        CommandEnumContext* context,
        StringBuffer* output)
    {
        if (context is null || output is null) return Abi.Error;
        try
        {
            NativeCommandSnapshot snapshot;
            lock (Gate) snapshot = Snapshot ?? throw new InvalidOperationException("commands are not published");
            var command = snapshot.Commands[checked((int)commandIndex)];
            var binding = command.Bindings[checked((int)overloadIndex)];
            var field = binding.Fields[checked((int)parameterIndex)];
            if (field.DynamicEnum is null) return Abi.Error;
            var source = Source(context, new Cmd.Output());
            Write(output, string.Join('\n', field.DynamicEnum.Options(source).Select(option => option.ToLowerInvariant())));
            return Abi.Ok;
        }
        catch
        {
            return Abi.Error;
        }
    }

    private static Cmd.Source Source(CommandInput* input, Cmd.Output output)
    {
        var name = Text(input->Source);
        var position = Vector(input->SourcePosition);
        if (input->SourceKind != Abi.CommandSourcePlayer)
            return new CommandSource(name, position, output);
        var latency = TimeSpan.Zero;
        for (var index = 0; index < checked((int)input->OnlinePlayerCount); index++)
        {
            var player = input->OnlinePlayers[index];
            if (Player.SameId(player.Player, input->SourcePlayer))
            {
                latency = TimeSpan.FromMilliseconds(player.LatencyMilliseconds);
                break;
            }
        }
        return new Player(input->SourcePlayer, name, latency, position, output, input->Invocation);
    }

    private static Cmd.Source Source(CommandEnumContext* input, Cmd.Output output)
    {
        var name = Text(input->Source);
        var position = Vector(input->SourcePosition);
        if (input->SourceKind != Abi.CommandSourcePlayer)
            return new CommandSource(name, position, output);
        var latency = TimeSpan.Zero;
        for (var index = 0; index < checked((int)input->OnlinePlayerCount); index++)
        {
            var player = input->OnlinePlayers[index];
            if (Player.SameId(player.Player, input->SourcePlayer))
            {
                latency = TimeSpan.FromMilliseconds(player.LatencyMilliseconds);
                break;
            }
        }
        return new Player(input->SourcePlayer, name, latency, position, output);
    }

    private static string[] Strings(StringView* values, ulong count)
    {
        if (count != 0 && values is null) throw new ArgumentException("null command arguments");
        var result = new string[checked((int)count)];
        for (var index = 0; index < result.Length; index++) result[index] = Text(values[index]);
        return result;
    }

    private static string Text(StringView value) => value.Length == 0
        ? string.Empty
        : Encoding.UTF8.GetString(new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)));

    private static Vector3 Vector(Vec3 value) => new(value.X, value.Y, value.Z);

    private static void Write(StringBuffer* output, string value)
    {
        if (output is null) return;
        var bytes = Encoding.UTF8.GetBytes(value);
        var length = Math.Min(bytes.Length, checked((int)Math.Min(output->Capacity, int.MaxValue)));
        if (length != 0 && output->Data is null) throw new ArgumentException("null command output");
        bytes.AsSpan(0, length).CopyTo(new Span<byte>(output->Data, length));
        output->Length = (ulong)length;
    }
}

internal sealed class CommandSource(string name, Vector3 position, Cmd.Output output) : Cmd.Source, Cmd.NamedTarget
{
    public string Name() => name;
    public Vector3 Position() => position;
    public void SendCommandOutput(Cmd.Output value) => output.Merge(value);
}

internal unsafe sealed class NativeCommandSnapshot : IDisposable
{
    private readonly List<nint> _allocations = [];

    internal NativeCommandSnapshot(Cmd.Command[] commands)
    {
        Commands = commands;
        Descriptors = Allocate<CommandDescriptor>(commands.Length);
        for (var commandIndex = 0; commandIndex < commands.Length; commandIndex++)
        {
            var command = commands[commandIndex];
            var aliases = Allocate<StringView>(command.CommandAliases.Length);
            for (var index = 0; index < command.CommandAliases.Length; index++) aliases[index] = String(command.CommandAliases[index]);

            var overloads = Allocate<CommandOverload>(command.Bindings.Length);
            for (var overloadIndex = 0; overloadIndex < command.Bindings.Length; overloadIndex++)
            {
                var fields = command.Bindings[overloadIndex].Fields;
                var parameters = Allocate<CommandParameter>(fields.Length);
                for (var parameterIndex = 0; parameterIndex < fields.Length; parameterIndex++)
                {
                    var field = fields[parameterIndex];
                    var values = Allocate<StringView>(field.Values.Length);
                    for (var valueIndex = 0; valueIndex < field.Values.Length; valueIndex++) values[valueIndex] = String(field.Values[valueIndex]);
                    parameters[parameterIndex] = new CommandParameter
                    {
                        Kind = field.Kind,
                        Optional = field.Optional ? (byte)1 : (byte)0,
                        Name = String(field.Name),
                        Suffix = String(field.Suffix),
                        Values = values,
                        ValueCount = (ulong)field.Values.Length,
                    };
                }
                overloads[overloadIndex] = new CommandOverload
                {
                    Parameters = parameters,
                    ParameterCount = (ulong)fields.Length,
                };
            }
            Descriptors[commandIndex] = new CommandDescriptor
            {
                Name = String(command.CommandName),
                Description = String(command.CommandDescription),
                Aliases = aliases,
                AliasCount = (ulong)command.CommandAliases.Length,
                Overloads = overloads,
                OverloadCount = (ulong)command.Bindings.Length,
            };
        }
    }

    internal Cmd.Command[] Commands { get; }
    internal CommandDescriptor* Descriptors { get; }

    public void Dispose()
    {
        foreach (var allocation in _allocations) NativeMemory.Free((void*)allocation);
        _allocations.Clear();
    }

    private T* Allocate<T>(int count) where T : unmanaged
    {
        if (count == 0) return null;
        var value = (T*)NativeMemory.AllocZeroed(checked((nuint)(sizeof(T) * count)));
        _allocations.Add((nint)value);
        return value;
    }

    private StringView String(string value)
    {
        if (value.Length == 0) return default;
        var bytes = Encoding.UTF8.GetBytes(value);
        var pointer = (byte*)NativeMemory.Alloc((nuint)bytes.Length);
        bytes.CopyTo(new Span<byte>(pointer, bytes.Length));
        _allocations.Add((nint)pointer);
        return new StringView { Data = pointer, Length = (ulong)bytes.Length };
    }
}
