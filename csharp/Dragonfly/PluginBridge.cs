using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;
using Dragonfly.Native;

namespace Dragonfly;

internal static unsafe class PluginBridge
{
    private static Func<Plugin>? Factory;
    private static PluginApi* Descriptor;

    internal static PluginApi* Initialize(Func<Plugin> factory, string id, ulong subscriptions)
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
                Subscriptions = subscriptions,
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
        return header->Version == Abi.HostVersion && header->Size >= 496 ? Abi.Ok : Abi.Error;
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
            var plugin = Get(instance);
            switch (eventId)
            {
                case Abi.PlayerMoveEvent:
                {
                    var value = (PlayerMoveInput*)input;
                    var result = (PlayerMoveState*)state;
                    var context = new Player.Context(new Player(value->Player), result->Cancelled != 0);
                    plugin.HandleMove(
                        context,
                        new Vector3(value->NewPosition.X, value->NewPosition.Y, value->NewPosition.Z),
                        new Rotation(value->Rotation.Yaw, value->Rotation.Pitch));
                    if (context.Cancelled()) result->Cancelled = 1;
                    return Abi.Ok;
                }
                case Abi.PlayerQuitEvent:
                {
                    var value = (PlayerQuitInput*)input;
                    plugin.HandleQuit(new Player(value->Player, Utf8(value->Name)));
                    return Abi.Ok;
                }
                default:
                    return Abi.Ok;
            }
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
