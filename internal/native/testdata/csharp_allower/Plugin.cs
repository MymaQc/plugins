using System.Linq;
using Dragonfly;

public sealed class AllowerProbe : Plugin
{
    public override (string Message, bool Allowed) Allow(
        Net.Addr addr,
        Login.IdentityData identity,
        Login.ClientData client)
    {
        var emptyDevice = default(Login.DeviceID);
        if (emptyDevice.Value != "" || emptyDevice.ToString() != "" ||
            addr is not Net.UDPAddr udp || !ValidAddress(udp) ||
            identity.XUID != "123" || identity.Identity != "identity" ||
            identity.DisplayName != "Unicode玩家" || identity.TitleID != "title" ||
            identity.PlayFabTitleID != "playfab-title" || identity.PlayFabID != "playfab" ||
            client.AnimatedImageData is not [{ Frames: 2, Image: "animation", ImageHeight: 3,
                ImageWidth: 4, Type: 1, AnimationExpression: 2 }] ||
            client.CapeData != "cape" || client.CapeID != "cape-id" ||
            client.CapeImageHeight != 5 || client.CapeImageWidth != 6 || !client.CapeOnClassicSkin ||
            client.ClientRandomID != 7 || client.CurrentInputMode != 8 || client.DefaultInputMode != 9 ||
            client.DeviceModel != "model" || client.DeviceOS != Protocol.DeviceOS.Linux ||
            client.DeviceID != "device" || client.GameVersion != "1.26.30" || client.GUIScale != -1 ||
            !client.IsEditorMode || client.LanguageCode != "en_US" || !client.PersonaSkin ||
            client.PlatformOfflineID != "offline" || client.PlatformOnlineID != "online" ||
            client.PlatformUserID != "user" || !client.PremiumSkin || client.SelfSignedID != "self" ||
            client.ServerAddress != "server:19132" || client.SkinAnimationData != "skin-animation" ||
            client.SkinData != "skin" || client.SkinGeometry != "geometry" ||
            client.SkinGeometryVersion != "geometry-version" || client.SkinID != "skin-id" ||
            client.PlayFabID != "skin-playfab" || client.SkinImageHeight != 10 ||
            client.SkinImageWidth != 11 || client.SkinResourcePatch != "resource" ||
            client.SkinColour != "#010203" || client.ArmSize != "slim" ||
            client.PersonaPieces is not [{ Default: true, PackID: "pack", PieceID: "piece",
                PieceType: "eyes", ProductID: "product" }] ||
            client.PieceTintColours.Length != 1 || client.PieceTintColours[0].PieceType != "eyes" ||
            !client.PieceTintColours[0].Colours.SequenceEqual(["one", "two", "three", "four"]) ||
            client.ThirdPartyName != "third" || client.ThirdPartyNameOnly != true ||
            client.UIProfile != 12 || !client.TrustedSkin || !client.OverrideSkin ||
            !client.CompatibleWithClientSideChunkGen || client.MaxViewDistance != 13 ||
            client.MemoryTier != 14 || client.PlatformType != 15 || client.GraphicsMode != 16 ||
            client.PartyID != "party" || !client.PartyLeader)
            return ("snapshot-invalid", false);
        return ("snapshot-ok", false);
    }

    private static bool ValidAddress(Net.UDPAddr addr) => addr.Port switch
    {
        19132 => addr.Network() == "udp" && addr.Zone == "eth0" &&
            addr.IP.SequenceEqual(new byte[] { 0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1 }) &&
            addr.String() == "[fe80::1%eth0]:19132",
        19133 => addr.Network() == "udp" && addr.Zone == "" &&
            addr.IP.SequenceEqual(new byte[] { 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 192, 0, 2, 1 }) &&
            addr.String() == "192.0.2.1:19133",
        _ => false,
    };
}
