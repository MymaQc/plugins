package host

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
)

type commandRuntime interface {
	Commands() ([]native.Command, error)
	HandleCommand(index uint64, input native.CommandInput) (native.CommandOutput, error)
	CommandEnumOptions(index, overload, parameter uint64, sourceName string, onlinePlayers []string) ([]string, error)
}

// RegisterCommands publishes all enabled plugin commands in Dragonfly's command registry.
func RegisterCommands(runtime commandRuntime, players *Players) error {
	commands, err := runtime.Commands()
	if err != nil {
		return err
	}
	for _, command := range commands {
		if _, exists := cmd.ByAlias(command.Name); exists {
			return fmt.Errorf("register plugin command %q: name already registered", command.Name)
		}
		runnables, err := commandRunnables(runtime, players, command)
		if err != nil {
			return fmt.Errorf("register plugin command %q: %w", command.Name, err)
		}
		cmd.Register(cmd.New(command.Name, command.Description, nil, runnables...))
	}
	return nil
}

func commandRunnables(runtime commandRuntime, players *Players, command native.Command) ([]cmd.Runnable, error) {
	base := pluginCommandBase{runtime: runtime, index: command.Index}
	if len(command.Overloads) == 0 {
		return []cmd.Runnable{pluginCommand{pluginCommandBase: base}}, nil
	}
	runnables := make([]cmd.Runnable, 0, len(command.Overloads))
	for overloadIndex, overload := range command.Overloads {
		parameters := make([]pluginParameter, len(overload.Parameters))
		for index, parameter := range overload.Parameters {
			parameters[index] = pluginParameter{
				descriptor: parameter,
				runtime:    runtime,
				players:    players,
				command:    command.Index,
				overload:   uint64(overloadIndex),
				parameter:  uint64(index),
				enumType: fmt.Sprintf(
					"bedrock_gophers_%s_%d_%s",
					command.Name,
					overloadIndex,
					parameter.Name,
				),
			}
		}
		switch len(parameters) {
		case 0:
			runnables = append(runnables, pluginCommand{pluginCommandBase: base})
		case 1:
			runnables = append(runnables, pluginCommand1{pluginCommandBase: base, P1: parameters[0]})
		case 2:
			runnables = append(runnables, pluginCommand2{pluginCommandBase: base, P1: parameters[0], P2: parameters[1]})
		case 3:
			runnables = append(runnables, pluginCommand3{pluginCommandBase: base, P1: parameters[0], P2: parameters[1], P3: parameters[2]})
		case 4:
			runnables = append(runnables, pluginCommand4{pluginCommandBase: base, P1: parameters[0], P2: parameters[1], P3: parameters[2], P4: parameters[3]})
		default:
			return nil, fmt.Errorf("overloads support at most four typed parameters")
		}
	}
	return runnables, nil
}

type pluginCommandBase struct {
	runtime commandRuntime
	index   uint64
}

func (c pluginCommandBase) dispatch(source cmd.Source, output *cmd.Output, arguments string) {
	sourceName := fmt.Sprintf("%T", source)
	if named, ok := source.(cmd.NamedTarget); ok {
		sourceName = named.Name()
	}
	result, err := c.runtime.HandleCommand(c.index, native.CommandInput{
		Source:    sourceName,
		Arguments: arguments,
	})
	if err != nil {
		output.Error(err)
		return
	}
	if result.Message == "" {
		return
	}
	if result.Failed {
		output.Error(result.Message)
	} else {
		output.Print(result.Message)
	}
}

func joinedArguments(parameters []string, trailing cmd.Varargs) string {
	if value := strings.TrimSpace(string(trailing)); value != "" {
		parameters = append(parameters, value)
	}
	return strings.TrimSpace(strings.Join(parameters, " "))
}

type pluginCommand struct {
	pluginCommandBase `cmd:"-"`
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand) Run(source cmd.Source, output *cmd.Output, _ *world.Tx) {
	c.dispatch(source, output, string(c.Arguments))
}

type pluginParameter struct {
	descriptor native.CommandParameter
	runtime    commandRuntime
	players    *Players
	command    uint64
	overload   uint64
	parameter  uint64
	enumType   string
	selected   string
}

func (p pluginParameter) Parse(line *cmd.Line, value reflect.Value) error {
	argument, ok := line.Next()
	if !ok {
		p.selected = ""
		value.Set(reflect.ValueOf(p))
		return nil
	}
	if p.descriptor.Kind < native.CommandParameterSubcommand || p.descriptor.Kind > native.CommandParameterRawText {
		return fmt.Errorf("unknown plugin parameter kind %d", p.descriptor.Kind)
	}
	p.selected = argument
	value.Set(reflect.ValueOf(p))
	return nil
}

func (p pluginParameter) Type() string { return p.descriptor.Name }

func (p pluginParameter) transport() string {
	if p.descriptor.Kind != native.CommandParameterPlayer {
		return p.selected
	}
	id, ok := p.players.ResolveName(p.selected)
	if !ok {
		return "invalid"
	}
	return hex.EncodeToString(id.UUID[:]) + ":" + strconv.FormatUint(id.Generation, 10)
}

type describedEnum struct {
	typeName  string
	options   []string
	parameter *pluginParameter
	players   *Players
}

func (e describedEnum) Type() string { return e.typeName }
func (e describedEnum) Options(source cmd.Source) []string {
	if e.parameter == nil {
		if e.players != nil {
			return e.players.Names()
		}
		return e.options
	}
	sourceName := fmt.Sprintf("%T", source)
	if named, ok := source.(cmd.NamedTarget); ok {
		sourceName = named.Name()
	}
	options, err := e.parameter.runtime.CommandEnumOptions(
		e.parameter.command,
		e.parameter.overload,
		e.parameter.parameter,
		sourceName,
		e.parameter.players.Names(),
	)
	if err != nil {
		return nil
	}
	return options
}

func describe(parameters ...pluginParameter) []cmd.ParamInfo {
	result := make([]cmd.ParamInfo, 0, len(parameters))
	for _, parameter := range parameters {
		switch parameter.descriptor.Kind {
		case native.CommandParameterSubcommand:
			result = append(result, cmd.ParamInfo{Name: parameter.descriptor.Name, Value: cmd.SubCommand{}})
		case native.CommandParameterEnum, native.CommandParameterDynamicEnum:
			var dynamic *pluginParameter
			if parameter.descriptor.Kind == native.CommandParameterDynamicEnum {
				copy := parameter
				dynamic = &copy
			}
			result = append(result, cmd.ParamInfo{
				Name: parameter.descriptor.Name,
				Value: describedEnum{
					typeName:  parameter.enumType,
					options:   parameter.descriptor.Values,
					parameter: dynamic,
				},
			})
		case native.CommandParameterString:
			result = append(result, cmd.ParamInfo{Name: parameter.descriptor.Name, Value: ""})
		case native.CommandParameterInteger:
			result = append(result, cmd.ParamInfo{Name: parameter.descriptor.Name, Value: int64(0)})
		case native.CommandParameterFloat:
			result = append(result, cmd.ParamInfo{Name: parameter.descriptor.Name, Value: float64(0)})
		case native.CommandParameterBool:
			result = append(result, cmd.ParamInfo{Name: parameter.descriptor.Name, Value: false})
		case native.CommandParameterPlayer:
			result = append(result, cmd.ParamInfo{
				Name: parameter.descriptor.Name,
				Value: describedEnum{
					typeName: parameter.enumType,
					players:  parameter.players,
				},
			})
		case native.CommandParameterRawText:
			result = append(result, cmd.ParamInfo{Name: parameter.descriptor.Name, Value: cmd.Varargs("")})
		}
	}
	return result
}

type pluginCommand1 struct {
	pluginCommandBase `cmd:"-"`
	P1                pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand1) Run(source cmd.Source, output *cmd.Output, _ *world.Tx) {
	c.dispatch(source, output, joinedArguments([]string{c.P1.transport()}, c.Arguments))
}
func (c pluginCommand1) DescribeParams(cmd.Source) []cmd.ParamInfo { return describe(c.P1) }

type pluginCommand2 struct {
	pluginCommandBase `cmd:"-"`
	P1, P2            pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand2) Run(source cmd.Source, output *cmd.Output, _ *world.Tx) {
	c.dispatch(source, output, joinedArguments([]string{c.P1.transport(), c.P2.transport()}, c.Arguments))
}
func (c pluginCommand2) DescribeParams(cmd.Source) []cmd.ParamInfo { return describe(c.P1, c.P2) }

type pluginCommand3 struct {
	pluginCommandBase `cmd:"-"`
	P1, P2, P3        pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand3) Run(source cmd.Source, output *cmd.Output, _ *world.Tx) {
	c.dispatch(source, output, joinedArguments([]string{c.P1.transport(), c.P2.transport(), c.P3.transport()}, c.Arguments))
}
func (c pluginCommand3) DescribeParams(cmd.Source) []cmd.ParamInfo {
	return describe(c.P1, c.P2, c.P3)
}

type pluginCommand4 struct {
	pluginCommandBase `cmd:"-"`
	P1, P2, P3, P4    pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand4) Run(source cmd.Source, output *cmd.Output, _ *world.Tx) {
	c.dispatch(source, output, joinedArguments([]string{c.P1.transport(), c.P2.transport(), c.P3.transport(), c.P4.transport()}, c.Arguments))
}
func (c pluginCommand4) DescribeParams(cmd.Source) []cmd.ParamInfo {
	return describe(c.P1, c.P2, c.P3, c.P4)
}
