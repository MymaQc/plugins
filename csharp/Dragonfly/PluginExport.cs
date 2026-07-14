using System.ComponentModel;
using Dragonfly.Native;

namespace Dragonfly;

[EditorBrowsable(EditorBrowsableState.Never)]
public static unsafe class PluginExport<T> where T : Plugin, new()
{
    [EditorBrowsable(EditorBrowsableState.Never)]
    public static PluginApi* Api(
        ulong subscriptions,
        Func<World.EntityType[]> entityTypes) => PluginBridge.Initialize(
        static () => new T(),
        typeof(T).Assembly.GetName().Name ?? typeof(T).Name,
        subscriptions,
        entityTypes);
}
