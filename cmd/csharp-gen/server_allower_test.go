package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestServerAllowerUsesGoAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allower.go")
	source := `package server
type Allower interface {
	Allow(addr net.Addr, d login.IdentityData, c login.ClientData) (string, bool)
}`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := inspectServerAllower(path)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Method.Name != "Allow" || spec.Method.ReturnType != "(string Message, bool Allowed)" {
		t.Fatalf("method = %+v", spec.Method)
	}
	for _, replacement := range [][2]string{
		{"addr net.Addr", "addr string"},
		{"d login.IdentityData", "d *login.IdentityData"},
		{"c login.ClientData", "c login.Client"},
		{"(string, bool)", "(bool, string)"},
	} {
		t.Run(replacement[0], func(t *testing.T) {
			changed := strings.Replace(source, replacement[0], replacement[1], 1)
			changedPath := filepath.Join(t.TempDir(), "allower.go")
			if err := os.WriteFile(changedPath, []byte(changed), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := inspectServerAllower(changedPath); err == nil {
				t.Fatal("expected signature drift error")
			}
		})
	}
}

func TestPinnedAllowerAndLoginSurface(t *testing.T) {
	dragonfly := modulePath(t, "github.com/df-mc/dragonfly")
	gophertunnel := modulePath(t, "github.com/sandertv/gophertunnel")
	allower, err := inspectServerAllower(filepath.Join(dragonfly, "server", "allower.go"))
	if err != nil {
		t.Fatal(err)
	}
	loginData, err := inspectLoginData(filepath.Join(gophertunnel, "minecraft", "protocol", "login", "data.go"))
	if err != nil {
		t.Fatal(err)
	}
	deviceOS, err := inspectDeviceOS(filepath.Join(gophertunnel, "minecraft", "protocol", "os.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(loginData.Types) != 5 || len(loginData.Types[0].Fields) != 6 || len(loginData.Types[1].Fields) != 48 {
		t.Fatalf("login types = %+v", loginData.Types)
	}
	if len(deviceOS.Values) != 15 || deviceOS.Values[0] != (deviceOSValue{Name: "Android", Value: 1}) ||
		deviceOS.Values[14] != (deviceOSValue{Name: "Linux", Value: 15}) {
		t.Fatalf("DeviceOS = %+v", deviceOS.Values)
	}
	generated := string(generateServerAllower(allower, loginData, deviceOS))
	for _, expected := range []string{
		"public interface Allower",
		"(string Message, bool Allowed) Allow(Net.Addr addr, Login.IdentityData d, Login.ClientData c)",
		"public abstract partial class Plugin : Server.Allower",
		"public readonly record struct DeviceID",
		"public DeviceID DeviceID { get; init; }",
		"public string PlayFabTitleID { get; init; }",
		"public bool? ThirdPartyNameOnly { get; init; }",
		"public PersonaPieceTintColour[] PieceTintColours { get; init; } = [];",
		"public enum DeviceOS",
		"Linux = 15",
	} {
		if !strings.Contains(generated, expected) {
			t.Fatalf("generated Allower surface missing %q:\n%s", expected, generated)
		}
	}
}

func modulePath(t *testing.T, module string) string {
	t.Helper()
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes.TrimSpace(output))
}
