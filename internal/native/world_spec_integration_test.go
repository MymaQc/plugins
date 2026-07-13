package native

import (
	"reflect"
	"testing"
	"time"
)

type worldSpecRecordingHost struct {
	noopHost
	calls int
	name  string
	spec  WorldOpenSpec
	id    WorldID
}

func (host *worldSpecRecordingHost) OpenWorldSpec(_ InvocationID, name string, spec WorldOpenSpec) (WorldID, bool) {
	host.calls++
	host.name, host.spec = name, spec
	return host.id, host.id != 0
}

func TestWorldSpecCrossesRustAndCHostBoundary(t *testing.T) {
	library, plugins := nativeArtifacts(t)
	host := &worldSpecRecordingHost{id: 73}
	runtime, err := OpenWithHost(library, plugins, host)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	player := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 11}
	output, err := runtime.HandleCommand(commandNamed(t, commands, "world").Index, CommandInput{
		Source: "Builder", SourceKind: CommandSourcePlayer, SourcePlayer: &player,
		OnlinePlayers: []CommandPlayer{{Player: player, Name: "Builder"}},
		Arguments:     "open-spec example:specified",
	})
	if err != nil || output.Failed {
		t.Fatalf("output=%+v error=%v", output, err)
	}
	want := WorldOpenSpec{
		ProviderPath: "examples/managed", Dimension: WorldDimensionOverworld,
		OpenMode: WorldOpenOrCreate, Save: WorldSaveManual,
		RandomTicks: WorldRandomTicksDisabled, Time: WorldTimeFixed, FixedTime: 6000,
		Weather: WorldWeatherClear, ChunkUnload: WorldChunkUnloadAfter,
		ChunkUnloadAfter: 2 * time.Minute,
	}
	if host.calls != 1 || host.name != "example:specified" || !reflect.DeepEqual(host.spec, want) {
		t.Fatalf("calls=%d name=%q spec=%#v, want %#v", host.calls, host.name, host.spec, want)
	}
}
