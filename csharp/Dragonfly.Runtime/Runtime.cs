using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;
using Dragonfly.Native;

namespace Dragonfly.Runtime;

internal unsafe sealed class PluginLibrary : IDisposable
{
    internal PluginApi* Api;
    internal void* Instance;
    internal volatile bool Enabled;
    internal volatile bool Enabling;

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
        Api = null;
        Instance = null;
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

internal readonly unsafe struct RuntimeEntityType(
    PluginLibrary plugin,
    ulong localType,
    EntityTypeDescriptorV2 descriptor)
{
    internal PluginLibrary Plugin { get; } = plugin;
    internal ulong LocalType { get; } = localType;
    internal EntityTypeDescriptorV2 Descriptor { get; } = descriptor;
}

internal readonly unsafe struct RuntimeEntityInstance(
    PluginLibrary plugin,
    ulong localType,
    ulong localInstance)
{
    internal PluginLibrary Plugin { get; } = plugin;
    internal ulong LocalType { get; } = localType;
    internal ulong LocalInstance { get; } = localInstance;
}

internal readonly unsafe struct RuntimeEntityOpen(
    PluginLibrary plugin,
    ulong localType,
    ulong localOpen)
{
    internal PluginLibrary Plugin { get; } = plugin;
    internal ulong LocalType { get; } = localType;
    internal ulong LocalOpen { get; } = localOpen;
}

internal unsafe sealed class RuntimeState : IDisposable
{
    internal readonly object Gate = new();
    internal readonly List<PluginLibrary> Plugins = [];
    internal readonly List<RuntimeCommand> Commands = [];
    internal readonly List<RuntimeEntityType> EntityTypes = [];
    internal readonly Dictionary<ulong, RuntimeEntityInstance> EntityInstances = [];
    internal readonly Dictionary<ulong, RuntimeEntityOpen> EntityOpens = [];
    internal ulong Subscriptions;
    private volatile bool _running;
    private volatile bool _enabling;
    private int _activeCalls;
    private long _nextEntityInstance;
    private long _nextEntityOpen;

    internal static RuntimeState Load(string directory, void* host)
    {
        if (host is null) throw new InvalidOperationException("null host API");
        var hostHeader = (HostHeader*)host;
        if (hostHeader->Version != Abi.HostVersion || hostHeader->Size < (uint)sizeof(HostApi))
            throw new InvalidOperationException("incompatible host API");

        var runtime = new RuntimeState();
        try
        {
            foreach (var path in Directory.EnumerateFiles(directory, NativeExtension()).Order())
                runtime.Add(path, host);
            runtime.PublishEntityTypes();
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
        // NativeAOT libraries install process-wide signal handlers while loading. Unloading one
        // leaves those handlers pointing into unmapped code, so the process owns every successful
        // load until exit even when validation or later server startup fails.
        var library = NativeLibrary.Load(path);
        var entry = (delegate* unmanaged[Cdecl]<PluginApi*>)NativeLibrary.GetExport(library, "df_plugin_entry_v9");
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
        Plugins.Add(new PluginLibrary { Api = api, Instance = instance });
        Subscriptions |= api->Header.Subscriptions;
    }

    private void PublishEntityTypes()
    {
        var identifiers = new HashSet<string>(StringComparer.Ordinal);
        foreach (var plugin in Plugins)
        {
            if (plugin.Api->EntityTypeCount is null) continue;
            var count = ((delegate* unmanaged[Cdecl]<void*, ulong>)plugin.Api->EntityTypeCount)(plugin.Instance);
            if (count > 1 << 16)
                throw new InvalidOperationException($"plugin {Utf8(plugin.Api->Id)} returned too many entity types");
            if (count != 0 && (plugin.Api->EntityTypeAt is null || plugin.Api->HandleEntity is null))
                throw new InvalidOperationException($"plugin {Utf8(plugin.Api->Id)} has incomplete entity callbacks");
            var localTypes = new HashSet<ulong>();
            for (var index = 0UL; index < count; index++)
            {
                EntityTypeDescriptorV2 descriptor;
                var read = (delegate* unmanaged[Cdecl]<void*, ulong, EntityTypeDescriptorV2*, int>)plugin.Api->EntityTypeAt;
                if (read(plugin.Instance, index, &descriptor) != Abi.Ok || descriptor.TypeKey == 0 ||
                    descriptor.SaveId.Data is null || descriptor.SaveId.Length == 0 ||
                    descriptor.NetworkId.Data is null || descriptor.NetworkId.Length == 0 ||
                    !localTypes.Add(descriptor.TypeKey))
                    throw new InvalidOperationException($"plugin {Utf8(plugin.Api->Id)} returned an invalid entity type");
                var identifier = Utf8(descriptor.SaveId);
                if (!identifiers.Add(identifier))
                    throw new InvalidOperationException($"duplicate entity type {identifier}");
                EntityTypes.Add(new RuntimeEntityType(plugin, descriptor.TypeKey, descriptor));
            }
        }
    }

    internal void Enable(StringBuffer* error)
    {
        lock (Gate)
        {
            if (_running) return;
            _enabling = true;
            try
            {
                for (var index = 0; index < Plugins.Count; index++)
                {
                    var plugin = Plugins[index];
                    plugin.Enabling = true;
                    int status;
                    try { status = plugin.Api->Enable == null ? Abi.Ok : plugin.Api->Enable(plugin.Instance, error); }
                    finally { plugin.Enabling = false; }
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
            finally { _enabling = false; }
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

    internal ulong EntityTypeCount()
    {
        lock (Gate) return (ulong)EntityTypes.Count;
    }

    internal int EntityTypeAt(ulong index, EntityTypeDescriptorV2* output)
    {
        if (output is null) return Abi.Error;
        lock (Gate)
        {
            if (index >= (ulong)EntityTypes.Count) return Abi.Error;
            *output = EntityTypes[checked((int)index)].Descriptor;
            output->TypeKey = index + 1;
            return Abi.Ok;
        }
    }

    internal int EntityAdopt(ulong type, ulong opaque, ulong* output)
    {
        if (type == 0 || output is null) return Abi.Error;
        RuntimeEntityType route;
        lock (Gate)
        {
            if (type > (ulong)EntityTypes.Count) return Abi.Error;
            route = EntityTypes[checked((int)type - 1)];
        }
        return Adopt(route, opaque, output);
    }

    internal int EntityAdoptLocal(ulong pluginToken, ulong localType, ulong opaque, ulong* output)
    {
        if (pluginToken == 0 || localType == 0 || output is null) return Abi.Error;
        RuntimeEntityType route;
        lock (Gate)
        {
            var found = EntityTypes.FirstOrDefault(candidate =>
                (ulong)(nuint)candidate.Plugin.Api == pluginToken && candidate.LocalType == localType);
            if (found.Plugin is null) return Abi.Error;
            route = found;
        }
        return Adopt(route, opaque, output);
    }

    private int Adopt(RuntimeEntityType route, ulong opaque, ulong* output)
    {
        Interlocked.Increment(ref _activeCalls);
        try
        {
            ulong localInstance = 0;
            var handle = (delegate* unmanaged[Cdecl]<void*, ulong, uint, ulong, void*, void*, int>)
                route.Plugin.Api->HandleEntity;
            if (handle(route.Plugin.Instance, route.LocalType, Abi.EntityOperationAdopt, 0, &opaque, &localInstance) != Abi.Ok ||
                localInstance == 0) return Abi.Error;
            var raw = Interlocked.Increment(ref _nextEntityInstance);
            if (raw <= 0)
            {
                _ = handle(route.Plugin.Instance, route.LocalType, Abi.EntityOperationDestroy, localInstance, null, null);
                return Abi.Error;
            }
            var instance = checked((ulong)raw);
            lock (Gate) EntityInstances.Add(instance, new RuntimeEntityInstance(route.Plugin, route.LocalType, localInstance));
            *output = instance;
            return Abi.Ok;
        }
        finally { Interlocked.Decrement(ref _activeCalls); }
    }

    internal int EntityDestroy(ulong instance)
    {
        RuntimeEntityInstance route;
        lock (Gate)
        {
            if (!EntityInstances.Remove(instance, out route)) return Abi.Error;
        }
        Interlocked.Increment(ref _activeCalls);
        try
        {
            var handle = (delegate* unmanaged[Cdecl]<void*, ulong, uint, ulong, void*, void*, int>)
                route.Plugin.Api->HandleEntity;
            return handle(route.Plugin.Instance, 0, Abi.EntityOperationDestroy, route.LocalInstance, null, null);
        }
        finally { Interlocked.Decrement(ref _activeCalls); }
    }

    internal int EntityDecodeNbt(ulong type, EntityExactInput* input, EntityExactState* state)
    {
        if (type == 0 || input is null || state is null) return Abi.Error;
        RuntimeEntityType route;
        lock (Gate)
        {
            if (type > (ulong)EntityTypes.Count) return Abi.Error;
            route = EntityTypes[checked((int)type - 1)];
        }
        Interlocked.Increment(ref _activeCalls);
        try
        {
            var handle = (delegate* unmanaged[Cdecl]<void*, ulong, uint, ulong, void*, void*, int>)
                route.Plugin.Api->HandleEntity;
            state->Instance = 0;
            if (handle(route.Plugin.Instance, route.LocalType, Abi.EntityOperationDecodeNbt, 0, input, state) != Abi.Ok ||
                state->Instance == 0) return Abi.Error;
            var local = state->Instance;
            var raw = Interlocked.Increment(ref _nextEntityInstance);
            if (raw <= 0)
            {
                _ = handle(route.Plugin.Instance, route.LocalType, Abi.EntityOperationDestroy, local, null, null);
                return Abi.Error;
            }
            var global = checked((ulong)raw);
            lock (Gate) EntityInstances.Add(global, new RuntimeEntityInstance(route.Plugin, route.LocalType, local));
            state->Instance = global;
            return Abi.Ok;
        }
        finally { Interlocked.Decrement(ref _activeCalls); }
    }

    internal int EntityCall(ulong identity, uint operation, EntityExactInput* input, EntityExactState* state)
    {
        if (identity == 0) return Abi.Error;
        if (operation == Abi.EntityOperationEncodeNbt || operation == Abi.EntityOperationOpen)
        {
            RuntimeEntityInstance route;
            lock (Gate)
            {
                if (!EntityInstances.TryGetValue(identity, out route)) return Abi.Error;
            }
            var result = CallEntity(route.Plugin, route.LocalType, operation, route.LocalInstance, input, state);
            if (result != Abi.Ok || operation != Abi.EntityOperationOpen) return result;
            if (state is null || state->Instance == 0) return Abi.Error;
            var localOpen = state->Instance;
            var raw = Interlocked.Increment(ref _nextEntityOpen);
            if (raw <= 0)
            {
                _ = CallEntity(route.Plugin, route.LocalType, Abi.EntityOperationReleaseOpen, localOpen, null, null);
                return Abi.Error;
            }
            var globalOpen = checked((ulong)raw);
            lock (Gate) EntityOpens.Add(globalOpen, new RuntimeEntityOpen(route.Plugin, route.LocalType, localOpen));
            state->Instance = globalOpen;
            return Abi.Ok;
        }

        RuntimeEntityOpen open;
        lock (Gate)
        {
            if (operation == Abi.EntityOperationReleaseOpen)
            {
                if (!EntityOpens.Remove(identity, out open)) return Abi.Error;
            }
            else if (!EntityOpens.TryGetValue(identity, out open)) return Abi.Error;
        }
        return CallEntity(open.Plugin, open.LocalType, operation, open.LocalOpen, input, state);
    }

    private int CallEntity(
        PluginLibrary plugin,
        ulong localType,
        uint operation,
        ulong localIdentity,
        EntityExactInput* input,
        EntityExactState* state)
    {
        Interlocked.Increment(ref _activeCalls);
        try
        {
            var handle = (delegate* unmanaged[Cdecl]<void*, ulong, uint, ulong, void*, void*, int>)plugin.Api->HandleEntity;
            return handle(plugin.Instance, localType, operation, localIdentity, input, state);
        }
        finally { Interlocked.Decrement(ref _activeCalls); }
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
        if (!_running || input is null || state is null) return Abi.Error;
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

    internal int HandleScheduled(ulong pluginToken, ulong callback, ulong invocation, byte execute)
    {
        if (execute > 1) return Abi.Error;
        Interlocked.Increment(ref _activeCalls);
        try
        {
            var plugin = Plugins.FirstOrDefault(candidate => (ulong)(nuint)candidate.Api == pluginToken);
            if (plugin is null || plugin.Api->HandleScheduled == null ||
                execute != 0 && !(plugin.Enabling || (_enabling || _running) && plugin.Enabled))
                return Abi.Error;
            return plugin.Api->HandleScheduled(plugin.Instance, callback, invocation, execute);
        }
        finally
        {
            Interlocked.Decrement(ref _activeCalls);
        }
    }

    public void Dispose()
    {
        Disable();
        foreach (var open in EntityOpens.Keys.ToArray())
            _ = EntityCall(open, Abi.EntityOperationReleaseOpen, null, null);
        foreach (var instance in EntityInstances.Keys.ToArray()) _ = EntityDestroy(instance);
        for (var index = Plugins.Count - 1; index >= 0; index--) Plugins[index].Dispose();
        Commands.Clear();
        EntityTypes.Clear();
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
    public static ulong EntityTypeCount(void* runtime)
    {
        try { return State(runtime).EntityTypeCount(); }
        catch { return 0; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_type_at", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityTypeAt(void* runtime, ulong index, void* output)
    {
        try { return State(runtime).EntityTypeAt(index, (EntityTypeDescriptorV2*)output); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_adopt", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityAdopt(void* runtime, ulong type, ulong opaque, ulong* output)
    {
        try { return State(runtime).EntityAdopt(type, opaque, output); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_adopt_local", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityAdoptLocal(
        void* runtime,
        ulong plugin,
        ulong type,
        ulong opaque,
        ulong* output)
    {
        try { return State(runtime).EntityAdoptLocal(plugin, type, opaque, output); }
        catch { return Abi.Error; }
    }

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
    public static int EntityDestroy(void* runtime, ulong instance)
    {
        try { return State(runtime).EntityDestroy(instance); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_decode_nbt", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityDecodeNbt(void* runtime, ulong type, EntityExactInput* input, EntityExactState* state)
    {
        try { return State(runtime).EntityDecodeNbt(type, input, state); }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_entity_call", CallConvs = [typeof(CallConvCdecl)])]
    public static int EntityCall(
        void* runtime,
        ulong identity,
        uint operation,
        EntityExactInput* input,
        EntityExactState* state)
    {
        try { return State(runtime).EntityCall(identity, operation, input, state); }
        catch { return Abi.Error; }
    }

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

    [UnmanagedCallersOnly(EntryPoint = "df_runtime_handle_scheduled", CallConvs = [typeof(CallConvCdecl)])]
    public static int HandleScheduled(void* runtime, ulong plugin, ulong callback, ulong invocation, byte execute)
    {
        try { return State(runtime).HandleScheduled(plugin, callback, invocation, execute); }
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
