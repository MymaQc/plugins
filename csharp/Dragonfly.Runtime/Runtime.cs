using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;
using Dragonfly.Native;

namespace Dragonfly.Runtime;

internal unsafe sealed class PluginLibrary : IDisposable
{
    internal nint Library;
    internal PluginApi* Api;
    internal void* Instance;
    internal volatile bool Enabled;

    internal void Disable()
    {
        if (!Enabled) return;
        Enabled = false;
        if (Api->Disable != null) Api->Disable(Instance);
    }

    public void Dispose()
    {
        Disable();
        if (Api is not null && Api->Destroy != null) Api->Destroy(Instance);
        if (Library != 0) NativeLibrary.Free(Library);
        Api = null;
        Instance = null;
        Library = 0;
    }
}

internal readonly unsafe struct RuntimeCommand(
    PluginLibrary plugin,
    ulong localIndex,
    CommandDescriptor descriptor)
{
    internal PluginLibrary Plugin { get; } = plugin;
    internal ulong LocalIndex { get; } = localIndex;
    internal CommandDescriptor Descriptor { get; } = descriptor;
}

internal unsafe sealed class RuntimeState : IDisposable
{
    internal readonly object Gate = new();
    internal readonly List<PluginLibrary> Plugins = [];
    internal readonly List<RuntimeCommand> Commands = [];
    internal ulong Subscriptions;
    private volatile bool _running;
    private int _activeCalls;

    internal static RuntimeState Load(string directory, void* host)
    {
        if (host is null) throw new InvalidOperationException("null host API");
        var hostHeader = (HostHeader*)host;
        if (hostHeader->Version != Abi.HostVersion || hostHeader->Size < 648)
            throw new InvalidOperationException("incompatible host API");

        var runtime = new RuntimeState();
        try
        {
            foreach (var path in Directory.EnumerateFiles(directory, NativeExtension()).Order())
                runtime.Add(path, host);
            return runtime;
        }
        catch
        {
            runtime.Dispose();
            throw;
        }
    }

    private void Add(string path, void* host)
    {
        var library = NativeLibrary.Load(path);
        try
        {
            var entry = (delegate* unmanaged[Cdecl]<PluginApi*>)NativeLibrary.GetExport(library, "df_plugin_entry_v5");
            var api = entry();
            if (api is null || api->Header.Version != Abi.PluginVersion || api->Header.Size < sizeof(PluginApi))
                throw new InvalidOperationException($"{path} has an incompatible plugin API");
            if (api->Id.Length == 0 || api->Id.Data is null)
                throw new InvalidOperationException($"{path} has an empty plugin ID");
            var id = Utf8(api->Id);
            if (Plugins.Any(plugin => Utf8(plugin.Api->Id) == id))
                throw new InvalidOperationException($"duplicate plugin ID {id}");
            if (api->Header.Subscriptions != 0 && api->HandleEvent == null)
                throw new InvalidOperationException($"plugin {id} has no event handler");
            var instance = api->Create == null ? null : api->Create();
            if (api->Create != null && instance is null)
                throw new InvalidOperationException($"plugin {id} could not be created");
            if (api->SetHost == null || api->SetHost(instance, host) != Abi.Ok)
            {
                if (api->Destroy != null) api->Destroy(instance);
                throw new InvalidOperationException($"plugin {id} rejected the host API");
            }
            Plugins.Add(new PluginLibrary { Library = library, Api = api, Instance = instance });
            Subscriptions |= api->Header.Subscriptions;
        }
        catch
        {
            NativeLibrary.Free(library);
            throw;
        }
    }

    internal void Enable(StringBuffer* error)
    {
        lock (Gate)
        {
            if (_running) return;
            for (var index = 0; index < Plugins.Count; index++)
            {
                var plugin = Plugins[index];
                var status = plugin.Api->Enable == null ? Abi.Ok : plugin.Api->Enable(plugin.Instance, error);
                if (status == Abi.Ok && (error is null || error->Length == 0))
                {
                    plugin.Enabled = true;
                    continue;
                }
                for (var previous = index - 1; previous >= 0; previous--) Plugins[previous].Disable();
                var detail = error is null || error->Length == 0
                    ? string.Empty
                    : Encoding.UTF8.GetString(new ReadOnlySpan<byte>(error->Data, checked((int)error->Length)));
                var suffix = detail.Length == 0 ? string.Empty : $": {detail}";
                throw new InvalidOperationException($"plugin {Utf8(plugin.Api->Id)} failed to enable{suffix}");
            }
            try
            {
                PublishCommands();
            }
            catch
            {
                Commands.Clear();
                for (var index = Plugins.Count - 1; index >= 0; index--) Plugins[index].Disable();
                throw;
            }
            _running = true;
        }
    }

    private void PublishCommands()
    {
        var commands = new List<RuntimeCommand>();
        foreach (var plugin in Plugins)
        {
            if (plugin.Api->Commands == null) continue;
            ulong count = 0;
            var descriptors = plugin.Api->Commands(plugin.Instance, &count);
            if (count == 0) continue;
            var id = Utf8(plugin.Api->Id);
            if (descriptors is null) throw new InvalidOperationException($"plugin {id} returned null command descriptors");
            if (plugin.Api->HandleCommand == null)
                throw new InvalidOperationException($"plugin {id} has commands but no command handler");
            for (var index = 0UL; index < count; index++)
            {
                var descriptor = descriptors[checked((int)index)];
                if (descriptor.Name.Length == 0 || descriptor.Name.Data is null)
                    throw new InvalidOperationException($"plugin {id} returned a command with an empty name");
                commands.Add(new RuntimeCommand(plugin, index, descriptor));
            }
        }
        Commands.AddRange(commands);
    }

    internal void Disable()
    {
        lock (Gate)
        {
            _running = false;
            var spin = new SpinWait();
            while (Volatile.Read(ref _activeCalls) != 0) spin.SpinOnce();
            Commands.Clear();
            for (var index = Plugins.Count - 1; index >= 0; index--) Plugins[index].Disable();
        }
    }

    internal ulong CommandCount()
    {
        lock (Gate) return _running ? (ulong)Commands.Count : 0;
    }

    internal int CommandAt(ulong index, CommandDescriptor* output)
    {
        if (output is null || !_running) return Abi.Error;
        Interlocked.Increment(ref _activeCalls);
        try
        {
            if (!_running || index >= (ulong)Commands.Count) return Abi.Error;
            *output = Commands[checked((int)index)].Descriptor;
            return Abi.Ok;
        }
        finally
        {
            Interlocked.Decrement(ref _activeCalls);
        }
    }

    internal int HandleCommand(ulong index, CommandInput* input, CommandState* state)
    {
        if (input is null || state is null || !_running) return Abi.Error;
        Interlocked.Increment(ref _activeCalls);
        try
        {
            if (!_running || index >= (ulong)Commands.Count) return Abi.Error;
            var command = Commands[checked((int)index)];
            var plugin = command.Plugin;
            if (!plugin.Enabled || plugin.Api->HandleCommand == null) return Abi.Error;
            return plugin.Api->HandleCommand(plugin.Instance, command.LocalIndex, input, state);
        }
        finally
        {
            Interlocked.Decrement(ref _activeCalls);
        }
    }

    internal int CommandEnumOptions(
        ulong index,
        ulong overload,
        ulong parameter,
        CommandEnumContext* context,
        StringBuffer* output)
    {
        if (context is null || output is null || !_running) return Abi.Error;
        Interlocked.Increment(ref _activeCalls);
        try
        {
            if (!_running || index >= (ulong)Commands.Count) return Abi.Error;
            var command = Commands[checked((int)index)];
            var plugin = command.Plugin;
            if (!plugin.Enabled || plugin.Api->CommandEnumOptions == null) return Abi.Error;
            return plugin.Api->CommandEnumOptions(
                plugin.Instance,
                command.LocalIndex,
                overload,
                parameter,
                context,
                output);
        }
        finally
        {
            Interlocked.Decrement(ref _activeCalls);
        }
    }

    internal int HandleEvent(uint eventId, void* input, void* state)
    {
        if (!_running) return Abi.Error;
        Interlocked.Increment(ref _activeCalls);
        try
        {
            if (!_running) return Abi.Error;
            var subscription = eventId is >= 1 and <= 64 ? 1UL << ((int)eventId - 1) : 0;
            foreach (var plugin in Plugins)
            {
                if (!plugin.Enabled || (plugin.Api->Header.Subscriptions & subscription) == 0) continue;
                if (plugin.Api->HandleEvent(plugin.Instance, eventId, input, state) != Abi.Ok) return Abi.Error;
            }
            return Abi.Ok;
        }
        finally
        {
            Interlocked.Decrement(ref _activeCalls);
        }
    }

    public void Dispose()
    {
        Disable();
        for (var index = Plugins.Count - 1; index >= 0; index--) Plugins[index].Dispose();
        Commands.Clear();
        Plugins.Clear();
    }

    internal static string Utf8(StringView value) => value.Length == 0
        ? string.Empty
        : Encoding.UTF8.GetString(new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)));

    private static string NativeExtension() => OperatingSystem.IsWindows() ? "*.dll" : OperatingSystem.IsMacOS() ? "*.dylib" : "*.so";
}

public static unsafe class Exports
{
    [UnmanagedCallersOnly(EntryPoint = "df_runtime_create", CallConvs = [typeof(CallConvCdecl)])]
    public static int Create(RuntimeConfig* config, void** output, byte* error, ulong errorCapacity)
    {
        if (config is null || output is null) { Write(error, errorCapacity, "null runtime config or output"); return Abi.Error; }
        *output = null;
        try
        {
            var runtime = RuntimeState.Load(RuntimeState.Utf8(config->PluginDirectory), config->Host);
            *output = (void*)GCHandle.ToIntPtr(GCHandle.Alloc(runtime));
            return Abi.Ok;
        }
        catch (Exception exception) { Write(error, errorCapacity, exception.Message); return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_enable", CallConvs = [typeof(CallConvCdecl)])]
    public static int Enable(void* runtime, byte* error, ulong errorCapacity)
    {
        try
        {
            var buffer = new StringBuffer { Data = error, Capacity = errorCapacity };
            State(runtime).Enable(&buffer);
            return Abi.Ok;
        }
        catch (Exception exception) { Write(error, errorCapacity, exception.Message); return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_begin_disable", CallConvs = [typeof(CallConvCdecl)])]
    public static void BeginDisable(void* runtime) => DisableState(runtime);

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_finish_disable", CallConvs = [typeof(CallConvCdecl)])]
    public static void FinishDisable(void* runtime) => DisableState(runtime);

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_disable", CallConvs = [typeof(CallConvCdecl)])]
    public static void Disable(void* runtime) => DisableState(runtime);

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_destroy", CallConvs = [typeof(CallConvCdecl)])]
    public static void Destroy(void* runtime)
    {
        if (runtime is null) return;
        var handle = GCHandle.FromIntPtr((nint)runtime);
        try { ((RuntimeState)handle.Target!).Dispose(); } catch { }
        handle.Free();
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_plugin_count", CallConvs = [typeof(CallConvCdecl)])]
    public static ulong PluginCount(void* runtime) => TryState(runtime)?.Plugins.Count is int count ? (ulong)count : 0;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_subscriptions", CallConvs = [typeof(CallConvCdecl)])]
    public static ulong Subscriptions(void* runtime) => TryState(runtime)?.Subscriptions ?? 0;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_type_count", CallConvs = [typeof(CallConvCdecl)])]
    public static ulong EntityTypeCount(void* runtime) => 0;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_type_at", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityTypeAt(void* runtime, ulong index, void* output) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_adopt", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityAdopt(void* runtime, ulong type, ulong opaque, ulong* output) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_load", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityLoad(void* runtime, ulong type, void* input, ulong* output) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_save", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntitySave(void* runtime, ulong instance, void* state) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_tick", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityTick(void* runtime, ulong instance, void* input, void* state) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_hurt", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityHurt(void* runtime, ulong instance, void* input, void* state) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_heal", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityHeal(void* runtime, ulong instance, void* input, void* state) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_death", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityDeath(void* runtime, ulong instance, void* input, void* state) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_destroy", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityDestroy(void* runtime, ulong instance) => Abi.Error;

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_command_count", CallConvs = [typeof(CallConvCdecl)])]
    public static ulong CommandCount(void* runtime)
    {
        try { return State(runtime).CommandCount(); }
        catch { return 0; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_command_at", CallConvs = [typeof(CallConvCdecl)])]
    public static int CommandAt(void* runtime, ulong index, CommandDescriptor* output)
    {
        try { return State(runtime).CommandAt(index, output); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_handle_command", CallConvs = [typeof(CallConvCdecl)])]
    public static int HandleCommand(void* runtime, ulong index, CommandInput* input, CommandState* state)
    {
        try { return State(runtime).HandleCommand(index, input, state); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_command_enum_options", CallConvs = [typeof(CallConvCdecl)])]
    public static int CommandEnumOptions(
        void* runtime,
        ulong index,
        ulong overload,
        ulong parameter,
        CommandEnumContext* context,
        StringBuffer* output)
    {
        try { return State(runtime).CommandEnumOptions(index, overload, parameter, context, output); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_handle_event", CallConvs = [typeof(CallConvCdecl)])]
    public static int HandleEvent(void* runtime, uint eventId, void* input, void* state)
    {
        try { return State(runtime).HandleEvent(eventId, input, state); }
        catch { return Abi.Error; }
    }

    private static RuntimeState State(void* pointer) => TryState(pointer) ?? throw new InvalidOperationException("null runtime");
    private static RuntimeState? TryState(void* pointer) => pointer is null ? null : GCHandle.FromIntPtr((nint)pointer).Target as RuntimeState;
    private static void DisableState(void* runtime) { try { TryState(runtime)?.Disable(); } catch { } }

    private static void Write(byte* output, ulong capacity, string message)
    {
        if (output is null || capacity == 0) return;
        var bytes = Encoding.UTF8.GetBytes(message);
        var length = Math.Min(bytes.Length, checked((int)Math.Min(capacity - 1, int.MaxValue)));
        bytes.AsSpan(0, length).CopyTo(new Span<byte>(output, length));
        output[length] = 0;
    }
}
