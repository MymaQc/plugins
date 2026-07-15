package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

var titleSignatures = map[string]goSignature{
	"New":                 {Parameters: "...any", Results: "Title"},
	"Text":                {Results: "string"},
	"WithSubtitle":        {Parameters: "...any", Results: "Title"},
	"Subtitle":            {Results: "string"},
	"WithActionText":      {Parameters: "...any", Results: "Title"},
	"ActionText":          {Results: "string"},
	"Duration":            {Results: "time.Duration"},
	"WithDuration":        {Parameters: "time.Duration", Results: "Title"},
	"WithFadeInDuration":  {Parameters: "time.Duration", Results: "Title"},
	"FadeInDuration":      {Results: "time.Duration"},
	"WithFadeOutDuration": {Parameters: "time.Duration", Results: "Title"},
	"FadeOutDuration":     {Results: "time.Duration"},
}

func inspectTitle(titlePath, playerPath string) error {
	file, err := parser.ParseFile(token.NewFileSet(), titlePath, nil, 0)
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
		} else if valueReceiver(function, "Title") {
			found[function.Name.Name] = function
		}
	}
	for name, signature := range titleSignatures {
		function := found[name]
		if function == nil {
			return fmt.Errorf("Dragonfly title.Title has no %s method", name)
		}
		if got := goFunctionSignature(function); got != signature {
			return fmt.Errorf("Dragonfly title.%s signature changed: %+v", name, got)
		}
	}

	file, err = parser.ParseFile(token.NewFileSet(), playerPath, nil, 0)
	if err != nil {
		return err
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && playerMethod(function) && function.Name.Name == "SendTitle" {
			if got := goFunctionSignature(function); got != (goSignature{Parameters: "title.Title"}) {
				return fmt.Errorf("Dragonfly player.Player.SendTitle signature changed: %+v", got)
			}
			return nil
		}
	}
	return fmt.Errorf("Dragonfly player.Player has no SendTitle method")
}

func generateTitle() []byte {
	var output bytes.Buffer
	output.WriteString(`// Code generated from Dragonfly server/player/title/title.go and server/player/player.go Go AST. DO NOT EDIT.
#nullable enable
using System;

namespace Dragonfly;

public readonly record struct Title
{
    private readonly string? _text;
    private readonly string? _subtitle;
    private readonly string? _actionText;
    private readonly TimeSpan _fadeInDuration;
    private readonly TimeSpan _duration;
    private readonly TimeSpan _fadeOutDuration;

    private Title(string text, string subtitle, string actionText, TimeSpan fadeInDuration, TimeSpan duration, TimeSpan fadeOutDuration) =>
        (_text, _subtitle, _actionText, _fadeInDuration, _duration, _fadeOutDuration) =
        (text, subtitle, actionText, fadeInDuration, duration, fadeOutDuration);

    public static Title New(params object?[] text) => new(
        GoText.Format(text), string.Empty, string.Empty,
        TimeSpan.FromTicks(500000), TimeSpan.FromSeconds(2), TimeSpan.FromTicks(500000));

    public string Text() => _text ?? string.Empty;
    public Title WithSubtitle(params object?[] text) => new(
        Text(), GoText.Format(text), ActionText(), FadeInDuration(), Duration(), FadeOutDuration());
    public string Subtitle() => _subtitle ?? string.Empty;
    public Title WithActionText(params object?[] text) => new(
        Text(), Subtitle(), GoText.Format(text), FadeInDuration(), Duration(), FadeOutDuration());
    public string ActionText() => _actionText ?? string.Empty;
    public TimeSpan Duration() => _duration;
    public Title WithDuration(TimeSpan d) => new(Text(), Subtitle(), ActionText(), FadeInDuration(), d, FadeOutDuration());
    public Title WithFadeInDuration(TimeSpan d) => new(Text(), Subtitle(), ActionText(), d, Duration(), FadeOutDuration());
    public TimeSpan FadeInDuration() => _fadeInDuration;
    public Title WithFadeOutDuration(TimeSpan d) => new(Text(), Subtitle(), ActionText(), FadeInDuration(), Duration(), d);
    public TimeSpan FadeOutDuration() => _fadeOutDuration;
}

public sealed partial class Player
{
    public void SendTitle(Title t) => PluginBridge.Host.SendPlayerTitle(_invocation, Id, t);
}
`)
	return output.Bytes()
}
