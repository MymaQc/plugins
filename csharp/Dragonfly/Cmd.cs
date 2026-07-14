using System.Globalization;

namespace Dragonfly;

public static partial class Cmd
{
    public sealed class Command
    {
        internal Command(string name, string description, string[] aliases, CommandBinding[] bindings)
        {
            CommandName = name;
            CommandDescription = description;
            CommandAliases = aliases;
            Bindings = bindings;
        }

        internal Command(string name, string description, string usage, string[] aliases)
        {
            CommandName = name;
            CommandDescription = description;
            CommandUsage = usage;
            CommandAliases = aliases;
            Bindings = [];
        }

        internal string CommandName { get; }
        internal string CommandDescription { get; }
        internal string? CommandUsage { get; }
        internal string[] CommandAliases { get; }
        internal CommandBinding[] Bindings { get; }

        public string Name() => CommandName;
        public string Description() => CommandDescription;
        public string Usage() => CommandUsage ??
            string.Join('\n', Bindings.Select(binding => binding.Usage(CommandName)));
        public IReadOnlyList<string> Aliases() => CommandAliases;
    }

    public static Command New(string name, string description, string[] aliases, params Runnable[] runnables)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentNullException.ThrowIfNull(aliases);
        ArgumentNullException.ThrowIfNull(runnables);
        name = name.ToLowerInvariant();
        var normalizedAliases = aliases.Select(alias => alias.ToLowerInvariant()).ToList();
        if (normalizedAliases.Count != 0 && !normalizedAliases.Contains(name)) normalizedAliases.Add(name);
        return new Command(
            name,
            description ?? string.Empty,
            [.. normalizedAliases],
            runnables.Select(CommandBinding.Create).ToArray());
    }

    public static void Register(Command command) => CommandRegistry.Register(command);

    public readonly record struct SubCommand;

    public readonly record struct Varargs(string Value)
    {
        public override string ToString() => Value ?? string.Empty;
        public static implicit operator string(Varargs value) => value.Value ?? string.Empty;
        public static implicit operator Varargs(string value) => new(value);
    }

    public readonly struct Optional<T>
    {
        private readonly T? _value;
        private readonly bool _set;

        public Optional(T value)
        {
            _value = value;
            _set = true;
        }

        public (T? Value, bool Set) Load() => (_value, _set);
        public T LoadOr(T fallback) => _set ? _value! : fallback;
    }

    [AttributeUsage(AttributeTargets.Field)]
    public sealed class TagAttribute(string name = "", string suffix = "") : Attribute
    {
        internal string Name { get; } = name;
        internal string Suffix { get; } = suffix;
    }

    public sealed class Output
    {
        private readonly List<string> _errors = [];
        private readonly List<string> _messages = [];

        public void Error(params object?[] values) => _errors.Add(Join(values));
        public void Errorf(string format, params object?[] values) =>
            _errors.Add(string.Format(CultureInfo.InvariantCulture, format, values));
        public void Print(params object?[] values) => _messages.Add(Join(values));
        public void Printf(string format, params object?[] values) =>
            _messages.Add(string.Format(CultureInfo.InvariantCulture, format, values));

        public IReadOnlyList<string> Errors() => _errors;
        public int ErrorCount() => _errors.Count;
        public IReadOnlyList<string> Messages() => _messages;
        public int MessageCount() => _messages.Count;

        internal void Merge(Output other)
        {
            _messages.AddRange(other._messages);
            _errors.AddRange(other._errors);
        }

        internal string Text() => string.Join('\n', _messages.Concat(_errors));

        private static string Join(object?[] values) => string.Concat(values.Select(value => value?.ToString()));
    }
}
