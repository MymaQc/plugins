package host

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

var commandTestID atomic.Uint64

type commandRuntimeStub struct {
	commands  []native.Command
	input     native.CommandInput
	handleErr error
}

func (r *commandRuntimeStub) Commands() ([]native.Command, error) { return r.commands, nil }
func (r *commandRuntimeStub) HandleCommand(_ uint64, input native.CommandInput) (native.CommandOutput, error) {
	r.input = input
	if r.handleErr != nil {
		return native.CommandOutput{}, r.handleErr
	}
	return native.CommandOutput{Message: "ok"}, nil
}
func (r *commandRuntimeStub) CommandEnumOptions(_, _, _ uint64, _ native.CommandEnumContext) ([]string, error) {
	return []string{"LiveOne", "LiveTwo"}, nil
}

type commandSourceStub struct {
	output *cmd.Output
}

func (commandSourceStub) Position() mgl64.Vec3               { return mgl64.Vec3{} }
func (s *commandSourceStub) SendCommandOutput(o *cmd.Output) { s.output = o }
func (commandSourceStub) Name() string                       { return "tester" }

func TestStructuredCommandTransportsArbitraryParametersAndExactOverload(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		players.Register(player, 77)
		runtime := &commandRuntimeStub{}
		command := native.Command{
			Index: 3,
			Name:  "everything",
			Overloads: []native.CommandOverload{
				{Parameters: []native.CommandParameter{{Kind: native.CommandParameterSubcommand, Name: "other"}}},
				{Parameters: []native.CommandParameter{
					{Kind: native.CommandParameterSubcommand, Name: "run"},
					{Kind: native.CommandParameterEnum, Name: "style", Values: []string{"plain", "Excited"}},
					{Kind: native.CommandParameterString, Name: "word"},
					{Kind: native.CommandParameterInteger, Name: "count"},
					{Kind: native.CommandParameterFloat, Name: "scale"},
					{Kind: native.CommandParameterBool, Name: "enabled"},
					{Kind: native.CommandParameterDynamicEnum, Name: "live"},
					{Kind: native.CommandParameterPlayer, Name: "player"},
					{Kind: native.CommandParameterVector, Name: "position", Suffix: "xyz"},
					{Kind: native.CommandParameterRawText, Name: "message"},
				}},
			},
		}
		runnables, err := commandRunnables(runtime, players, command)
		if err != nil {
			t.Fatal(err)
		}
		if len(runnables) != 2 {
			t.Fatalf("runnables = %d, want 2", len(runnables))
		}

		registered := cmd.New(command.Name, "", nil, runnables...)
		params := registered.Params(&commandSourceStub{})
		if len(params) != 2 || len(params[1]) != 10 {
			t.Fatalf("params = %#v", params)
		}
		if params[1][8].Suffix != "xyz" {
			t.Fatalf("vector suffix = %q", params[1][8].Suffix)
		}

		source := &commandSourceStub{}
		registered.Execute("run excited hello 42 1.5 TRUE liveone TestPlayer 1 2 3 dragonfly plugins rock", source, player.Tx())
		if source.output == nil || source.output.ErrorCount() != 0 || source.output.MessageCount() != 1 {
			t.Fatalf("output = %#v", source.output)
		}
		if runtime.input.Overload != 1 {
			t.Fatalf("overload = %d, want 1", runtime.input.Overload)
		}
		if len(runtime.input.Arguments) != 10 {
			t.Fatalf("arguments = %#v", runtime.input.Arguments)
		}
		if got := runtime.input.Arguments[0]; got != "run" {
			t.Fatalf("subcommand = %q", got)
		}
		if got := runtime.input.Arguments[1]; got != "Excited" {
			t.Fatalf("enum = %q", got)
		}
		if got := runtime.input.Arguments[5]; got != "true" {
			t.Fatalf("bool = %q", got)
		}
		if got := runtime.input.Arguments[7]; !strings.Contains(got, ":77:") || !strings.HasSuffix(got, ":TestPlayer") {
			t.Fatalf("player = %q", got)
		}
		if got := runtime.input.Arguments[8]; got != "1 2 3" {
			t.Fatalf("vector = %q", got)
		}
		if got := runtime.input.Arguments[9]; got != "dragonfly plugins rock" {
			t.Fatalf("varargs = %q", got)
		}
	})
}

func TestOptionalArgumentsAreAbsent(t *testing.T) {
	runtime := &commandRuntimeStub{}
	command := native.Command{Index: 4, Name: "optional", Overloads: []native.CommandOverload{{Parameters: []native.CommandParameter{
		{Kind: native.CommandParameterString, Name: "required"},
		{Kind: native.CommandParameterInteger, Name: "optional", Optional: true},
	}}}}
	runnables, err := commandRunnables(runtime, NewPlayers(), command)
	if err != nil {
		t.Fatal(err)
	}
	cmd.New(command.Name, "", nil, runnables...).Execute("value", &commandSourceStub{}, nil)
	if len(runtime.input.Arguments) != 1 || runtime.input.Arguments[0] != "value" {
		t.Fatalf("arguments = %#v", runtime.input.Arguments)
	}
}

func TestNoArgumentOverloadRejectsInput(t *testing.T) {
	runtime := &commandRuntimeStub{}
	runnables, err := commandRunnables(runtime, NewPlayers(), native.Command{Index: 7, Name: "empty"})
	if err != nil {
		t.Fatal(err)
	}
	source := &commandSourceStub{}
	cmd.New("empty", "", nil, runnables...).Execute("unexpected", source, nil)
	if runtime.input.Source != "" {
		t.Fatalf("command dispatched with input %#v", runtime.input)
	}
	if source.output == nil || source.output.ErrorCount() == 0 {
		t.Fatalf("output = %#v", source.output)
	}
}

func TestRegisterCommandsRegistersAliases(t *testing.T) {
	id := commandTestID.Add(1)
	name := fmt.Sprintf("plugin-host-alias-registration-test-%d", id)
	alias := fmt.Sprintf("plugin-host-alias-test-%d", id)
	runtime := &commandRuntimeStub{commands: []native.Command{{Index: 5, Name: name, Aliases: []string{alias}}}}
	if err := RegisterCommands(runtime, NewPlayers()); err != nil {
		t.Fatal(err)
	}
	registered, ok := cmd.ByAlias(alias)
	if !ok || registered.Name() != name {
		t.Fatalf("alias resolved to %q, found=%v", registered.Name(), ok)
	}
}

func TestNativeCommandFailureIsGeneric(t *testing.T) {
	runtime := &commandRuntimeStub{handleErr: errors.New("private native detail")}
	runnables, err := commandRunnables(runtime, NewPlayers(), native.Command{Index: 6, Name: "failure"})
	if err != nil {
		t.Fatal(err)
	}
	source := &commandSourceStub{}
	cmd.New("failure", "", nil, runnables...).Execute("", source, nil)
	if source.output == nil || source.output.ErrorCount() != 1 {
		t.Fatalf("output = %#v", source.output)
	}
	message := source.output.Errors()[0].Error()
	if message != "Command failed." || strings.Contains(message, "private native detail") {
		t.Fatalf("public error = %q", message)
	}
}

func TestInvalidDescriptorLayoutRejected(t *testing.T) {
	tests := []struct {
		name       string
		parameters []native.CommandParameter
	}{
		{name: "required after optional", parameters: []native.CommandParameter{
			{Kind: native.CommandParameterString, Name: "first", Optional: true},
			{Kind: native.CommandParameterString, Name: "second"},
		}},
		{name: "raw text not last", parameters: []native.CommandParameter{
			{Kind: native.CommandParameterRawText, Name: "text"},
			{Kind: native.CommandParameterString, Name: "after"},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := commandRunnables(&commandRuntimeStub{}, NewPlayers(), native.Command{
				Name: "invalid", Overloads: []native.CommandOverload{{Parameters: test.parameters}},
			})
			if err == nil {
				t.Fatal("expected descriptor error")
			}
		})
	}
}

func TestPlayerEnumOptionsAreLowercase(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		players.Register(player, 1)
		options := (describedEnum{players: players, playerNames: true}).Options(&commandSourceStub{})
		if len(options) != 1 || options[0] != "testplayer" {
			t.Fatalf("options = %#v", options)
		}
	})
}

func TestCommandParametersReachAvailableCommandsPacket(t *testing.T) {
	runtime := &commandRuntimeStub{}
	command := native.Command{Index: 8, Name: "gamemode", Overloads: []native.CommandOverload{{Parameters: []native.CommandParameter{
		{Kind: native.CommandParameterEnum, Name: "mode", Values: []string{"survival", "creative", "adventure", "spectator"}},
		{Kind: native.CommandParameterPlayer, Name: "target", Optional: true},
	}}}}
	runnables, err := commandRunnables(runtime, NewPlayers(), command)
	if err != nil {
		t.Fatal(err)
	}
	registered := cmd.New(command.Name, "", nil, runnables...)
	packet, _ := session.BuildAvailableCommands(
		map[string]cmd.Command{command.Name: registered},
		&commandSourceStub{},
		nil,
	)
	if len(packet.Commands) != 1 || len(packet.Commands[0].Overloads) != 1 || len(packet.Commands[0].Overloads[0].Parameters) != 2 {
		t.Fatalf("available command = %#v", packet.Commands)
	}
	parameters := packet.Commands[0].Overloads[0].Parameters
	if parameters[0].Name != "mode" || parameters[1].Name != "target" || !parameters[1].Optional {
		t.Fatalf("parameters = %#v", parameters)
	}
	const enumMask = protocol.CommandArgValid | protocol.CommandArgEnum
	if parameters[0].Type&enumMask != enumMask {
		t.Fatalf("mode parameter type = %#x, want enum flags %#x", parameters[0].Type, enumMask)
	}
	enumIndex := parameters[0].Type &^ enumMask
	if enumIndex >= uint32(len(packet.Enums)) {
		t.Fatalf("mode enum index = %d, enums = %#v", enumIndex, packet.Enums)
	}
	mode := packet.Enums[enumIndex]
	values := make([]string, len(mode.ValueIndices))
	for index, valueIndex := range mode.ValueIndices {
		if valueIndex >= uint32(len(packet.EnumValues)) {
			t.Fatalf("mode value index = %d, values = %#v", valueIndex, packet.EnumValues)
		}
		values[index] = packet.EnumValues[valueIndex]
	}
	if mode.Type != "gamemode_mode" || !reflect.DeepEqual(values, []string{"survival", "creative", "adventure", "spectator"}) {
		t.Fatalf("mode enum = %#v, values = %#v", mode, values)
	}
}

func TestPlayerSourceIsResolvedByUUID(t *testing.T) {
	withPlayer(t, func(player *player.Player) {
		players := NewPlayers()
		players.Register(player, 3)
		runtime := &commandRuntimeStub{}
		(pluginCommandBase{runtime: runtime, players: players}).dispatchActive(0, player, &cmd.Output{}, nil)
		if runtime.input.SourcePlayer == nil || runtime.input.SourcePlayer.Generation != 3 {
			t.Fatalf("source player = %#v", runtime.input.SourcePlayer)
		}
		if runtime.input.SourceKind != native.CommandSourcePlayer {
			t.Fatalf("source kind = %v, want player", runtime.input.SourceKind)
		}
	})
}
