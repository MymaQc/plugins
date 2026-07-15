using System.Globalization;

namespace Dragonfly;

internal static class GoText
{
    internal static string Format(object?[] values) => string.Join(" ", values.Select(FormatValue));

    private static string FormatValue(object? value) => value switch
    {
        null => "<nil>",
        bool boolean => boolean ? "true" : "false",
        float number => number.ToString("G", CultureInfo.InvariantCulture).Replace('E', 'e'),
        double number => number.ToString("G", CultureInfo.InvariantCulture).Replace('E', 'e'),
        IFormattable formattable => formattable.ToString(null, CultureInfo.InvariantCulture) ?? string.Empty,
        _ => value.ToString() ?? string.Empty,
    };
}
