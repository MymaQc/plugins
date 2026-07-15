// Code generated from Dragonfly server/allower.go and gophertunnel login AST. DO NOT EDIT.
#nullable enable
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Dragonfly;

public sealed partial class Server
{
    public interface Allower
    {
        (string Message, bool Allowed) Allow(Net.Addr addr, Login.IdentityData d, Login.ClientData c);
    }
}

public abstract partial class Plugin : Server.Allower
{
    public virtual (string Message, bool Allowed) Allow(Net.Addr addr, Login.IdentityData d, Login.ClientData c) => (string.Empty, true);
}

public static partial class Login
{
    [JsonConverter(typeof(DeviceIDJsonConverter))]
    public readonly record struct DeviceID
    {
        private readonly string? _value;

        public DeviceID(string value) => _value = value ?? string.Empty;
        public string Value => _value ?? string.Empty;

        public override string ToString() => Value;
        public static implicit operator DeviceID(string value) => new(value);
        public static implicit operator string(DeviceID value) => value.Value;
    }

    internal sealed class DeviceIDJsonConverter : JsonConverter<DeviceID>
    {
        public override DeviceID Read(ref Utf8JsonReader reader, Type type, JsonSerializerOptions options) =>
            new(reader.GetString() ?? string.Empty);
        public override void Write(Utf8JsonWriter writer, DeviceID value, JsonSerializerOptions options) =>
            writer.WriteStringValue(value.Value);
    }

    public sealed class IdentityData
    {
        public string XUID { get; init; } = string.Empty;
        public string Identity { get; init; } = string.Empty;
        public string DisplayName { get; init; } = string.Empty;
        public string TitleID { get; init; } = string.Empty;
        public string PlayFabTitleID { get; init; } = string.Empty;
        public string PlayFabID { get; init; } = string.Empty;
    }

    public sealed class ClientData
    {
        public SkinAnimation[] AnimatedImageData { get; init; } = [];
        public string CapeData { get; init; } = string.Empty;
        public string CapeID { get; init; } = string.Empty;
        public int CapeImageHeight { get; init; }
        public int CapeImageWidth { get; init; }
        public bool CapeOnClassicSkin { get; init; }
        public long ClientRandomID { get; init; }
        public int CurrentInputMode { get; init; }
        public int DefaultInputMode { get; init; }
        public string DeviceModel { get; init; } = string.Empty;
        public Protocol.DeviceOS DeviceOS { get; init; }
        public DeviceID DeviceID { get; init; }
        public string GameVersion { get; init; } = string.Empty;
        public int GUIScale { get; init; }
        public bool IsEditorMode { get; init; }
        public string LanguageCode { get; init; } = string.Empty;
        public bool PersonaSkin { get; init; }
        public string PlatformOfflineID { get; init; } = string.Empty;
        public string PlatformOnlineID { get; init; } = string.Empty;
        public string PlatformUserID { get; init; } = string.Empty;
        public bool PremiumSkin { get; init; }
        public string SelfSignedID { get; init; } = string.Empty;
        public string ServerAddress { get; init; } = string.Empty;
        public string SkinAnimationData { get; init; } = string.Empty;
        public string SkinData { get; init; } = string.Empty;
        public string SkinGeometry { get; init; } = string.Empty;
        public string SkinGeometryVersion { get; init; } = string.Empty;
        public string SkinID { get; init; } = string.Empty;
        public string PlayFabID { get; init; } = string.Empty;
        public int SkinImageHeight { get; init; }
        public int SkinImageWidth { get; init; }
        public string SkinResourcePatch { get; init; } = string.Empty;
        public string SkinColour { get; init; } = string.Empty;
        public string ArmSize { get; init; } = string.Empty;
        public PersonaPiece[] PersonaPieces { get; init; } = [];
        public PersonaPieceTintColour[] PieceTintColours { get; init; } = [];
        public string ThirdPartyName { get; init; } = string.Empty;
        public bool? ThirdPartyNameOnly { get; init; }
        public int UIProfile { get; init; }
        public bool TrustedSkin { get; init; }
        public bool OverrideSkin { get; init; }
        public bool CompatibleWithClientSideChunkGen { get; init; }
        public int MaxViewDistance { get; init; }
        public int MemoryTier { get; init; }
        public int PlatformType { get; init; }
        public int GraphicsMode { get; init; }
        public string PartyID { get; init; } = string.Empty;
        public bool PartyLeader { get; init; }
    }

    public sealed class PersonaPiece
    {
        public bool Default { get; init; }
        public string PackID { get; init; } = string.Empty;
        public string PieceID { get; init; } = string.Empty;
        public string PieceType { get; init; } = string.Empty;
        public string ProductID { get; init; } = string.Empty;
    }

    public sealed class PersonaPieceTintColour
    {
        public string[] Colours { get; init; } = [];
        public string PieceType { get; init; } = string.Empty;
    }

    public sealed class SkinAnimation
    {
        public double Frames { get; init; }
        public string Image { get; init; } = string.Empty;
        public int ImageHeight { get; init; }
        public int ImageWidth { get; init; }
        public int Type { get; init; }
        public int AnimationExpression { get; init; }
    }
}

public static partial class Protocol
{
    public enum DeviceOS
    {
        Android = 1,
        IOS = 2,
        OSX = 3,
        FireOS = 4,
        GearVR = 5,
        Hololens = 6,
        Win10 = 7,
        Win32 = 8,
        Dedicated = 9,
        TVOS = 10,
        Orbis = 11,
        NX = 12,
        XBOX = 13,
        WP = 14,
        Linux = 15,
    }
}
