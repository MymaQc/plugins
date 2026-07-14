using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;
using Dragonfly.Native;

namespace Dragonfly;

public abstract class Plugin : Player.Handler
{
    public virtual void OnEnable() { }
    public virtual void OnDisable() { }
    public virtual void HandleQuit(Player player) { }
}

public sealed partial class Player
{
    internal Player(PlayerId id, string name)
    {
        Id = id;
        Name = name;
    }

    internal PlayerId Id { get; }
    public string Name { get; }

    public sealed class Context
    {
        internal Context(Player player) => Player = player;
        public Player Player { get; }
        public bool Cancelled { get; set; }
    }
}

public static unsafe class PluginExport<T> where T : Plugin, new()
{
    public static PluginApi* Api => PluginBridge.Initialize(
        static () => new T(),
        typeof(T).Assembly.GetName().Name ?? typeof(T).Name);
}

internal static unsafe class PluginBridge
{
    private static Func<Plugin>? Factory;
    private static PluginApi* Descriptor;

    internal static PluginApi* Initialize(Func<Plugin> factory, string id)
    {
        if (Descriptor is not null) return Descriptor;
        Factory = factory;
        var bytes = Encoding.UTF8.GetBytes(id);
        var idPointer = (byte*)NativeMemory.Alloc((nuint)bytes.Length);
        bytes.CopyTo(new Span<byte>(idPointer, bytes.Length));
        Descriptor = (PluginApi*)NativeMemory.AllocZeroed((nuint)sizeof(PluginApi));
        *Descriptor = new PluginApi
        {
            Header = new AbiHeader
            {
                Version = Abi.PluginVersion,
                Size = (uint)sizeof(PluginApi),
                Subscriptions = Abi.PlayerQuitSubscription,
            },
            Id = new StringView { Data = idPointer, Length = (ulong)bytes.Length },
            Create = &Create,
            Enable = &Enable,
            Disable = &Disable,
            SetHost = &SetHost,
            Destroy = &Destroy,
            HandleEvent = &HandleEvent,
        };
        return Descriptor;
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void* Create()
    {
        try { return (void*)GCHandle.ToIntPtr(GCHandle.Alloc(Factory!())); }
        catch { return null; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int SetHost(void* instance, void* host)
    {
        if (host is null) return Abi.Error;
        var header = (HostHeader*)host;
        if (header->Version != Abi.HostVersion || header->Size < 496) return Abi.Error;
        return Abi.Ok;
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int Enable(void* instance, StringBuffer* error)
    {
        try
        {
            Get(instance).OnEnable();
            if (error is not null) error->Length = 0;
            return Abi.Ok;
        }
        catch (Exception exception)
        {
            Write(error, exception.Message);
            return Abi.Error;
        }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int Disable(void* instance)
    {
        try { Get(instance).OnDisable(); return Abi.Ok; }
        catch { return Abi.Error; }
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static void Destroy(void* instance)
    {
        if (instance is not null) GCHandle.FromIntPtr((nint)instance).Free();
    }

    [UnmanagedCallersOnly(CallConvs = [typeof(CallConvCdecl)])]
    private static int HandleEvent(void* instance, uint eventId, void* input, void* state)
    {
        try
        {
            if (eventId != Abi.PlayerQuitEvent) return Abi.Ok;
            var value = (PlayerQuitInput*)input;
            Get(instance).HandleQuit(new Player(value->Player, Utf8(value->Name)));
            return Abi.Ok;
        }
        catch { return Abi.Error; }
    }

    private static Plugin Get(void* instance) => (Plugin)GCHandle.FromIntPtr((nint)instance).Target!;

    private static string Utf8(StringView value) => value.Length == 0
        ? string.Empty
        : Encoding.UTF8.GetString(new ReadOnlySpan<byte>(value.Data, checked((int)value.Length)));

    private static void Write(StringBuffer* output, string message)
    {
        if (output is null || output->Data is null || output->Capacity == 0) return;
        var bytes = Encoding.UTF8.GetBytes(message);
        var length = Math.Min(bytes.Length, checked((int)output->Capacity));
        bytes.AsSpan(0, length).CopyTo(new Span<byte>(output->Data, length));
        output->Length = (ulong)length;
    }
}
