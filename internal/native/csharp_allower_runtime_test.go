package native

import (
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestCSharpAllowerReceivesCompleteSnapshots(t *testing.T) {
	library, _ := csharpArtifacts(t)
	pluginDirectory := publishCSharpAllowerProbe(t)
	pluginRuntime, err := Open(library, pluginDirectory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pluginRuntime.Close)
	if err := pluginRuntime.Enable(); err != nil {
		t.Fatal(err)
	}

	addresses := []*net.UDPAddr{
		{IP: net.ParseIP("fe80::1"), Port: 19132, Zone: "eth0"},
		{IP: net.ParseIP("192.0.2.1"), Port: 19133},
	}
	for _, address := range addresses {
		message, allowed, err := pluginRuntime.Allow(address, allowerIdentityProbe(), allowerClientProbe())
		if err != nil {
			t.Fatal(err)
		}
		if allowed || message != "snapshot-ok" {
			t.Fatalf("Allow(%s) = (%q, %v)", address, message, allowed)
		}
	}
}

func publishCSharpAllowerProbe(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	osName := map[string]string{"linux": "linux", "darwin": "osx", "windows": "win"}[runtime.GOOS]
	arch := map[string]string{"amd64": "x64", "arm64": "arm64"}[runtime.GOARCH]
	if osName == "" || arch == "" {
		t.Skipf("no .NET RID for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	output := t.TempDir()
	project := filepath.Join(root, "internal", "native", "testdata", "csharp_allower", "AllowerProbe.csproj")
	command := exec.Command("dotnet", "publish", project, "-c", "Release", "-r", osName+"-"+arch,
		"--self-contained", "true", "-o", output)
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("publish C# allower probe: %v\n%s", err, data)
	}
	return output
}

func allowerIdentityProbe() login.IdentityData {
	return login.IdentityData{
		XUID: "123", Identity: "identity", DisplayName: "Unicode玩家", TitleID: "title",
		PlayFabTitleID: "playfab-title", PlayFabID: "playfab",
	}
}

func allowerClientProbe() login.ClientData {
	thirdPartyOnly := true
	return login.ClientData{
		AnimatedImageData: []login.SkinAnimation{{
			Frames: 2, Image: "animation", ImageHeight: 3, ImageWidth: 4, Type: 1, AnimationExpression: 2,
		}},
		CapeData: "cape", CapeID: "cape-id", CapeImageHeight: 5, CapeImageWidth: 6,
		CapeOnClassicSkin: true, ClientRandomID: 7, CurrentInputMode: 8, DefaultInputMode: 9,
		DeviceModel: "model", DeviceOS: protocol.DeviceLinux, DeviceID: "device",
		GameVersion: "1.26.30", GUIScale: -1, IsEditorMode: true, LanguageCode: "en_US",
		PersonaSkin: true, PlatformOfflineID: "offline", PlatformOnlineID: "online",
		PlatformUserID: "user", PremiumSkin: true, SelfSignedID: "self", ServerAddress: "server:19132",
		SkinAnimationData: "skin-animation", SkinData: "skin", SkinGeometry: "geometry",
		SkinGeometryVersion: "geometry-version", SkinID: "skin-id", PlayFabID: "skin-playfab",
		SkinImageHeight: 10, SkinImageWidth: 11, SkinResourcePatch: "resource",
		SkinColour: "#010203", ArmSize: "slim",
		PersonaPieces: []login.PersonaPiece{{
			Default: true, PackID: "pack", PieceID: "piece", PieceType: "eyes", ProductID: "product",
		}},
		PieceTintColours: []login.PersonaPieceTintColour{{
			Colours: [4]string{"one", "two", "three", "four"}, PieceType: "eyes",
		}},
		ThirdPartyName: "third", ThirdPartyNameOnly: &thirdPartyOnly, UIProfile: 12,
		TrustedSkin: true, OverrideSkin: true, CompatibleWithClientSideChunkGen: true,
		MaxViewDistance: 13, MemoryTier: 14, PlatformType: 15, GraphicsMode: 16,
		PartyID: "party", PartyLeader: true,
	}
}
