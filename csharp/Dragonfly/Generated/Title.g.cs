// Code generated from Dragonfly server/player/title/title.go and server/player/player.go Go AST. DO NOT EDIT.
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
