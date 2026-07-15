// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
using System;
using Dragonfly.Native;

namespace Dragonfly;

public static class Language
{
    public readonly record struct Tag
    {
        private readonly string? _value;
        public Tag(string value) => _value = value ?? "und";
        public string String() => _value ?? "und";
        public override string ToString() => String();
    }
}

public sealed partial class Player
{
    public string Name() => PlayerName;
    public Guid UUID() => PluginBridge.Host.PlayerUUID(Id);
    public string XUID() => PluginBridge.Host.PlayerXUID(_invocation, Id);
    public string DeviceID() => PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerStringDeviceID);
    public string DeviceModel() => PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerStringDeviceModel);
    public string SelfSignedID() => PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerStringSelfSignedID);
    public Language.Tag Locale() => new(PluginBridge.Host.PlayerString(_invocation, Id, Abi.PlayerStringLocale));
    public Net.Addr? Addr()
    {
        if (!PluginBridge.Host.TryPlayerString(_invocation, Id, Abi.PlayerStringAddrNetwork, out var network) ||
            !PluginBridge.Host.TryPlayerString(_invocation, Id, Abi.PlayerStringAddrString, out var address)) return null;
        return new Net.AddrSnapshot(network, address);
    }
}
