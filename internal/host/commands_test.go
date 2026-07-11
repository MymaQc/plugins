package host

import (
	"testing"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/go-gl/mathgl/mgl64"
)

type commandRuntimeStub struct {
	input native.CommandInput
}

func (r *commandRuntimeStub) Commands() ([]native.Command, error) { return nil, nil }
func (r *commandRuntimeStub) HandleCommand(_ uint64, input native.CommandInput) (native.CommandOutput, error) {
	r.input = input
	return native.CommandOutput{Message: "ok"}, nil
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
	runnables, err := commandRunnables(runtime, command)
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
