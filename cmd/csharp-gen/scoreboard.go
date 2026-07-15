package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var scoreboardSignatures = map[string]goSignature{
	"New":           {Parameters: "...any", Results: "*Scoreboard"},
	"Name":          {Results: "string"},
	"Write":         {Parameters: "[]byte", Results: "int, error"},
	"WriteString":   {Parameters: "string", Results: "int, error"},
	"Set":           {Parameters: "int, string"},
	"Remove":        {Parameters: "int"},
	"RemovePadding": {},
	"Lines":         {Results: "[]string"},
	"Descending":    {Results: "bool"},
	"SetDescending": {},
}

func inspectScoreboard(scoreboardPath, playerPath string) error {
	file, err := parser.ParseFile(token.NewFileSet(), scoreboardPath, nil, 0)
	if err != nil {
		return err
	}
	found := map[string]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if function.Recv == nil && function.Name.Name == "New" {
			found[function.Name.Name] = function
		} else if pointerReceiver(function, "Scoreboard") {
			found[function.Name.Name] = function
		}
	}
	for name, signature := range scoreboardSignatures {
		function := found[name]
		if function == nil {
			return fmt.Errorf("Dragonfly scoreboard.Scoreboard has no %s method", name)
		}
		if got := goFunctionSignature(function); got != signature {
			return fmt.Errorf("Dragonfly scoreboard.%s signature changed: %+v", name, got)
		}
	}

	file, err = parser.ParseFile(token.NewFileSet(), playerPath, nil, 0)
	if err != nil {
		return err
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && playerMethod(function) && function.Name.Name == "SendScoreboard" {
			if got := goFunctionSignature(function); got != (goSignature{Parameters: "*scoreboard.Scoreboard"}) {
				return fmt.Errorf("Dragonfly player.Player.SendScoreboard signature changed: %+v", got)
			}
			return nil
		}
	}
	return fmt.Errorf("Dragonfly player.Player has no SendScoreboard method")
}

func generateScoreboard() []byte {
	var output bytes.Buffer
	output.WriteString(`// Code generated from Dragonfly server/player/scoreboard/scoreboard.go and server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
using System;
using System.Collections.Generic;
using System.Text;

namespace Dragonfly;

public sealed class Scoreboard
{
    private readonly string _name;
    private readonly List<string> _lines = [];
    private bool _padding = true;
    private bool _descending;

    private Scoreboard(string name) => _name = name;

    public static Scoreboard New(params object?[] name) => new(GoText.Format(name));
    public string Name() => _name;
    public int Write(byte[] p)
    {
        ArgumentNullException.ThrowIfNull(p);
        return WriteString(Encoding.UTF8.GetString(p));
    }
    public int WriteString(string s)
    {
        ArgumentNullException.ThrowIfNull(s);
        var lines = s.Split('\n');
        _lines.AddRange(lines);
        if (_lines.Count >= 15)
            throw new InvalidOperationException("write scoreboard: maximum of 15 lines of text exceeded");
        return lines.Length;
    }
    public void Set(int index, string s)
    {
        if (index is < 0 or >= 15) throw new ArgumentOutOfRangeException(nameof(index));
        ArgumentNullException.ThrowIfNull(s);
        var difference = index - (_lines.Count - 1);
        if (difference > 0) for (var i = 0; i < difference; i++) _lines.Add(string.Empty);
        _lines[index] = TrimTwoNewlines(s);
    }
    public void Remove(int index)
    {
        if (index is < 0 or >= 15) throw new ArgumentOutOfRangeException(nameof(index));
        _lines.RemoveAt(index);
    }
    public void RemovePadding() => _padding = false;
    public string[] Lines()
    {
        var lines = _lines.ToArray();
        if (_padding)
        {
            var nameLength = Encoding.UTF8.GetByteCount(_name);
            for (var i = 0; i < lines.Length; i++)
            {
                var difference = nameLength - Encoding.UTF8.GetByteCount(lines[i]) - 2;
                lines[i] = difference <= 0
                    ? " " + lines[i] + " "
                    : " " + lines[i] + new string(' ', difference);
            }
        }
        if (_descending) Array.Reverse(lines);
        return lines;
    }
    public bool Descending() => _descending;
    public void SetDescending() => _descending = true;

    internal string[] RawLines() => _lines.ToArray();
    internal bool Padding() => _padding;

    private static string TrimTwoNewlines(string value)
    {
        if (value.EndsWith('\n')) value = value[..^1];
        if (value.EndsWith('\n')) value = value[..^1];
        return value;
    }
}

public sealed partial class Player
{
    public void SendScoreboard(Scoreboard scoreboard)
    {
        ArgumentNullException.ThrowIfNull(scoreboard);
        PluginBridge.Host.SendPlayerScoreboard(_invocation, Id, scoreboard);
    }
}
`)
	return output.Bytes()
}
