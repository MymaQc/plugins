package host

import (
	"strings"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/go-gl/mathgl/mgl64"
)

type commandRuntimeStub struct {
	input native.CommandInput
}

func TestPlayerCommandResolvesStableHandle(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		players.Register(player, 77)
		runtime := &commandRuntimeStub{}
		command := native.Command{
			Index: 4,
			Name:  "direct",
			Overloads: []native.CommandOverload{{Parameters: []native.CommandParameter{
				{Kind: native.CommandParameterSubcommand, Name: "send"},
				{Kind: native.CommandParameterPlayer, Name: "player"},
			}}},
		}
		runnables, err := commandRunnables(runtime, players, command)
		if err != nil {
			t.Fatal(err)
		}
		runnable := runnables[0].(pluginCommand2)
		if runnable.P2.enumType != "direct_player" {
			t.Fatalf("enum type = %q", runnable.P2.enumType)
		}
		cmd.New("direct", "", nil, runnables...).Execute("send TestPlayer", &commandSourceStub{}, player.Tx())
		if !strings.HasPrefix(runtime.input.Arguments, "send ") || !strings.Contains(runtime.input.Arguments, ":77:") {
			t.Fatalf("arguments = %q", runtime.input.Arguments)
		}
	})
}

func (r *commandRuntimeStub) Commands() ([]native.Command, error) { return nil, nil }
func (r *commandRuntimeStub) HandleCommand(_ uint64, input native.CommandInput) (native.CommandOutput, error) {
	r.input = input
	return native.CommandOutput{Message: "ok"}, nil
}
func (r *commandRuntimeStub) CommandEnumOptions(_, _, _ uint64, _ string, _ []string) ([]string, error) {
	return []string{"one", "two"}, nil
}

type commandSourceStub struct {
	output *cmd.Output
}

func (commandSourceStub) Position() mgl64.Vec3               { return mgl64.Vec3{} }
func (s *commandSourceStub) SendCommandOutput(o *cmd.Output) { s.output = o }
func (commandSourceStub) Name() string                       { return "tester" }

func TestStructuredCommandParsesSubcommandAndEnum(t *testing.T) {
	runtime := &commandRuntimeStub{}
	command := native.Command{
		Index: 3,
		Name:  "hello",
		Overloads: []native.CommandOverload{{Parameters: []native.CommandParameter{
			{Kind: native.CommandParameterSubcommand, Name: "say"},
			{Kind: native.CommandParameterEnum, Name: "style", Values: []string{"plain", "excited"}},
		}}},
	}
	runnables, err := commandRunnables(runtime, NewPlayers(), command)
	if err != nil {
		t.Fatal(err)
	}
	source := &commandSourceStub{}
	cmd.New("hello", "", nil, runnables...).Execute("say excited", source, nil)
	if runtime.input.Source != "tester" || runtime.input.Arguments != "say excited" {
		t.Fatalf("input = %#v", runtime.input)
	}
	if source.output == nil || source.output.ErrorCount() != 0 || source.output.MessageCount() != 1 {
		t.Fatalf("output = %#v", source.output)
	}
}

func TestPlayerEnumOptionsAreLowercase(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		players.Register(player, 1)
		options := (describedEnum{players: players}).Options(&commandSourceStub{})
		if len(options) != 1 || options[0] != "testplayer" {
			t.Fatalf("options = %#v", options)
		}
	})
}

func TestPlayerSourceIsResolvedByUUID(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		players.Register(player, 3)
		runtime := &commandRuntimeStub{}
		(pluginCommandBase{runtime: runtime, players: players}).dispatch(player, &cmd.Output{}, "", player.Tx())
		if runtime.input.SourcePlayer == nil || runtime.input.SourcePlayer.Generation != 3 {
			t.Fatalf("source player = %#v", runtime.input.SourcePlayer)
		}
		if runtime.input.SourceKind != native.CommandSourcePlayer {
			t.Fatalf("source kind = %v, want player", runtime.input.SourceKind)
		}
	})
}

func TestMissingOptionalPlayerStaysAbsent(t *testing.T) {
	parameter := pluginParameter{
		descriptor: native.CommandParameter{
			Kind:     native.CommandParameterPlayer,
			Optional: true,
		},
		players: NewPlayers(),
	}
	if value := parameter.transport(0); value != "" {
		t.Fatalf("transport = %q, want empty", value)
	}
}
