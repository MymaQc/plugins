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
	"github.com/google/uuid"
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
	base := pluginCommandBase{runtime: runtime, players: players, index: command.Index}
	if len(command.Overloads) == 0 {
		return []cmd.Runnable{pluginCommand{pluginCommandBase: base}}, nil
	}
	runnables := make([]cmd.Runnable, 0, len(command.Overloads))
	typeCounts := map[string]int{}
	for _, overload := range command.Overloads {
		for _, parameter := range overload.Parameters {
			typeCounts[command.Name+"_"+parameter.Name]++
		}
	}
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
				enumType:   commandEnumType(command.Name, parameter.Name, overloadIndex, typeCounts),
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

func commandEnumType(command, parameter string, overload int, counts map[string]int) string {
	name := command + "_" + parameter
	if counts[name] > 1 {
		return fmt.Sprintf("%s_%d", name, overload)
	}
	return name
}

type pluginCommandBase struct {
	runtime commandRuntime
	players *Players
	index   uint64
}

func (c pluginCommandBase) dispatch(source cmd.Source, output *cmd.Output, arguments string, tx *world.Tx) {
	c.players.WithInvocation(tx, func(invocation native.InvocationID) { c.dispatchActive(invocation, source, output, arguments) })
}

func (c pluginCommandBase) dispatchActive(invocation native.InvocationID, source cmd.Source, output *cmd.Output, arguments string) {
	sourceName := fmt.Sprintf("%T", source)
	if named, ok := source.(cmd.NamedTarget); ok {
		sourceName = named.Name()
	}
	input := native.CommandInput{
		Invocation:    invocation,
		Source:        sourceName,
		Arguments:     arguments,
		SourceKind:    native.CommandSourceConsole,
		OnlinePlayers: c.players.CommandSnapshots(),
	}
	if sourcePlayer, ok := source.(interface{ UUID() uuid.UUID }); ok {
		playerUUID := sourcePlayer.UUID()
		if id, found := c.players.ResolveUUID(playerUUID); found {
			input.SourcePlayer = &id
		}
	}
	result, err := c.runtime.HandleCommand(c.index, input)
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

func (c pluginCommand) Run(source cmd.Source, output *cmd.Output, tx *world.Tx) {
	c.dispatch(source, output, string(c.Arguments), tx)
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

func (p pluginParameter) transport(invocation native.InvocationID) string {
	if p.descriptor.Optional && p.selected == "" {
		return ""
	}
	if p.descriptor.Kind != native.CommandParameterPlayer {
		return p.selected
	}
	id, ok := p.players.ResolveName(p.selected)
	if !ok {
		return "invalid"
	}
	connected, ok := p.players.ResolveID(id, invocation)
	if !ok {
		return "invalid"
	}
	return hex.EncodeToString(id.UUID[:]) + ":" + strconv.FormatUint(id.Generation, 10) + ":" + strconv.FormatInt(connected.Latency().Milliseconds(), 10) + ":" + connected.Name()
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
			return lowercaseOptions(e.players.Names())
		}
		return lowercaseOptions(e.options)
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
	return lowercaseOptions(options)
}

func lowercaseOptions(options []string) []string {
	normalized := make([]string, len(options))
	for index, option := range options {
		normalized[index] = strings.ToLower(option)
	}
	return normalized
}

func describe(parameters ...pluginParameter) []cmd.ParamInfo {
	result := make([]cmd.ParamInfo, 0, len(parameters))
	for _, parameter := range parameters {
		before := len(result)
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
		if len(result) != before {
			result[len(result)-1].Optional = parameter.descriptor.Optional
		}
	}
	return result
}

type pluginCommand1 struct {
	pluginCommandBase `cmd:"-"`
	P1                pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand1) Run(source cmd.Source, output *cmd.Output, tx *world.Tx) {
	c.players.WithInvocation(tx, func(invocation native.InvocationID) {
		c.dispatchActive(invocation, source, output, joinedArguments([]string{c.P1.transport(invocation)}, c.Arguments))
	})
}
func (c pluginCommand1) DescribeParams(cmd.Source) []cmd.ParamInfo { return describe(c.P1) }

type pluginCommand2 struct {
	pluginCommandBase `cmd:"-"`
	P1, P2            pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand2) Run(source cmd.Source, output *cmd.Output, tx *world.Tx) {
	c.players.WithInvocation(tx, func(invocation native.InvocationID) {
		c.dispatchActive(invocation, source, output, joinedArguments([]string{c.P1.transport(invocation), c.P2.transport(invocation)}, c.Arguments))
	})
}
func (c pluginCommand2) DescribeParams(cmd.Source) []cmd.ParamInfo { return describe(c.P1, c.P2) }

type pluginCommand3 struct {
	pluginCommandBase `cmd:"-"`
	P1, P2, P3        pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand3) Run(source cmd.Source, output *cmd.Output, tx *world.Tx) {
	c.players.WithInvocation(tx, func(invocation native.InvocationID) {
		c.dispatchActive(invocation, source, output, joinedArguments([]string{c.P1.transport(invocation), c.P2.transport(invocation), c.P3.transport(invocation)}, c.Arguments))
	})
}
func (c pluginCommand3) DescribeParams(cmd.Source) []cmd.ParamInfo {
	return describe(c.P1, c.P2, c.P3)
}

type pluginCommand4 struct {
	pluginCommandBase `cmd:"-"`
	P1, P2, P3, P4    pluginParameter
	Arguments         cmd.Varargs `cmd:"arguments"`
}

func (c pluginCommand4) Run(source cmd.Source, output *cmd.Output, tx *world.Tx) {
	c.players.WithInvocation(tx, func(invocation native.InvocationID) {
		c.dispatchActive(invocation, source, output, joinedArguments([]string{c.P1.transport(invocation), c.P2.transport(invocation), c.P3.transport(invocation), c.P4.transport(invocation)}, c.Arguments))
	})
}
func (c pluginCommand4) DescribeParams(cmd.Source) []cmd.ParamInfo {
	return describe(c.P1, c.P2, c.P3, c.P4)
}
