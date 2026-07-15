using System.Reflection;
using System.Text.Json;
using System.Diagnostics.CodeAnalysis;

namespace Dragonfly;

internal static class FormCodec
{
    internal static string Format(object?[] values) => GoText.Format(values);

    internal static void VerifyCustom(Form.Submittable submittable)
    {
        foreach (var field in Fields(submittable))
        {
            if (!typeof(Form.Element).IsAssignableFrom(field.FieldType))
                throw new ArgumentException("all public fields must implement Form.Element", nameof(submittable));
            if (field.IsInitOnly)
                throw new ArgumentException("form element fields cannot be readonly", nameof(submittable));
        }
    }

    internal static void VerifyMenu(Form.MenuSubmittable submittable)
    {
        foreach (var field in Fields(submittable))
            if (!typeof(Form.MenuElement).IsAssignableFrom(field.FieldType))
                throw new ArgumentException("all public fields must implement Form.MenuElement", nameof(submittable));
    }

    internal static void VerifyModal(Form.ModalSubmittable submittable)
    {
        var fields = Fields(submittable);
        if (fields.Length != 2 || fields.Any(field => field.FieldType != typeof(Form.Button)))
            throw new ArgumentException("modal forms require exactly two public Form.Button fields", nameof(submittable));
    }

    internal static IReadOnlyList<Form.Element> Elements(Form.Custom form) =>
        Fields(form.Submittable).Select(field => (Form.Element)Read(field, form.Submittable)).ToArray();

    internal static IReadOnlyList<Form.MenuElement> Elements(Form.Menu form)
    {
        var declared = Fields(form.Submittable)
            .Select(field => (Form.MenuElement)Read(field, form.Submittable));
        return declared.Concat(form.ExtraElements).ToArray();
    }

    internal static IReadOnlyList<Form.Button> Buttons(Form.Menu form) =>
        Elements(form).OfType<Form.Button>().ToArray();

    internal static IReadOnlyList<Form.Button> Buttons(Form.Modal form) =>
        Fields(form.Submittable).Select(field => (Form.Button)Read(field, form.Submittable)).ToArray();

    internal static byte[] Encode(Form.Value form)
    {
        using var stream = new MemoryStream();
        using (var writer = new Utf8JsonWriter(stream))
        {
            switch (form)
            {
                case Form.Custom custom:
                    WriteCustom(writer, custom);
                    break;
                case Form.Menu menu:
                    WriteMenu(writer, menu);
                    break;
                case Form.Modal modal:
                    WriteModal(writer, modal);
                    break;
                default:
                    throw new ArgumentException("form type is not registered", nameof(form));
            }
        }
        return stream.ToArray();
    }

    internal static byte[] EncodeElement(Form.Element element)
    {
        using var stream = new MemoryStream();
        using (var writer = new Utf8JsonWriter(stream)) WriteElement(writer, element);
        return stream.ToArray();
    }

    internal static byte[] EncodeMenuElement(Form.MenuElement element)
    {
        using var stream = new MemoryStream();
        using (var writer = new Utf8JsonWriter(stream)) WriteMenuElement(writer, element);
        return stream.ToArray();
    }

    internal static void Respond(
        Form.Value form,
        Form.Submitter submitter,
        World.Tx tx,
        bool closed,
        ReadOnlyMemory<byte> response)
    {
        if (closed)
        {
            if (Submittable(form) is Form.Closer closer) closer.Close(submitter, tx);
            return;
        }
        using var document = JsonDocument.Parse(response);
        switch (form)
        {
            case Form.Custom custom:
                SubmitCustom(custom, submitter, tx, document.RootElement);
                break;
            case Form.Menu menu:
                SubmitMenu(menu, submitter, tx, document.RootElement);
                break;
            case Form.Modal modal:
                SubmitModal(modal, submitter, tx, document.RootElement);
                break;
            default:
                throw new ArgumentException("form type is not registered", nameof(form));
        }
    }

    private static void SubmitCustom(Form.Custom form, Form.Submitter submitter, World.Tx tx, JsonElement response)
    {
        if (response.ValueKind != JsonValueKind.Array)
            throw new ArgumentException("custom form response must be an array", nameof(response));
        var fields = Fields(form.Submittable);
        if (response.GetArrayLength() < fields.Length)
            throw new ArgumentException("custom form response has too few values", nameof(response));
        var index = 0;
        var updates = new List<(FieldInfo Field, object Value)>();
        foreach (var field in fields)
        {
            var element = Read(field, form.Submittable);
            var value = response[index++];
            switch (element)
            {
                case Form.Divider or Form.Header or Form.Label:
                    break;
                case Form.Input input when value.ValueKind == JsonValueKind.String:
                    updates.Add((field, input.WithValue(value.GetString()!)));
                    break;
                case Form.Toggle toggle when value.ValueKind is JsonValueKind.True or JsonValueKind.False:
                    updates.Add((field, toggle.WithValue(value.GetBoolean())));
                    break;
                case Form.Slider slider when value.ValueKind == JsonValueKind.Number &&
                                                   value.TryGetDouble(out var number) &&
                                                   double.IsFinite(number) && number >= slider.Min && number <= slider.Max:
                    updates.Add((field, slider.WithValue(number)));
                    break;
                case Form.Dropdown dropdown when TryIndex(value, dropdown.Options, out var selected):
                    updates.Add((field, dropdown.WithValue(selected)));
                    break;
                case Form.StepSlider slider when TryIndex(value, slider.Options, out var selected):
                    updates.Add((field, slider.WithValue(selected)));
                    break;
                default:
                    throw new ArgumentException("custom form response contains an invalid value", nameof(response));
            }
        }
        foreach (var (field, value) in updates) field.SetValue(form.Submittable, value);
        form.Submittable.Submit(submitter, tx);
    }

    private static void SubmitMenu(Form.Menu form, Form.Submitter submitter, World.Tx tx, JsonElement response)
    {
        if (response.ValueKind != JsonValueKind.Number || !response.TryGetInt32(out var index))
            throw new ArgumentException("menu response must be a button index", nameof(response));
        var buttons = Buttons(form);
        if ((uint)index >= (uint)buttons.Count)
            throw new ArgumentException("menu response button index is out of range", nameof(response));
        form.Submittable.Submit(submitter, buttons[index], tx);
    }

    private static void SubmitModal(Form.Modal form, Form.Submitter submitter, World.Tx tx, JsonElement response)
    {
        if (response.ValueKind is not (JsonValueKind.True or JsonValueKind.False))
            throw new ArgumentException("modal response must be a boolean", nameof(response));
        var buttons = Buttons(form);
        form.Submittable.Submit(submitter, buttons[response.GetBoolean() ? 0 : 1], tx);
    }

    private static object? Submittable(Form.Value form) => form switch
    {
        Form.Custom custom => custom.Submittable,
        Form.Menu menu => menu.Submittable,
        Form.Modal modal => modal.Submittable,
        _ => null,
    };

    private static bool TryIndex(JsonElement value, string[]? options, out int selected)
    {
        selected = 0;
        return options is not null && value.ValueKind == JsonValueKind.Number && value.TryGetInt32(out selected) &&
               (uint)selected < (uint)options.Length;
    }

    [UnconditionalSuppressMessage(
        "Trimming",
        "IL2075",
        Justification = "PluginEntryGenerator roots public fields for every form submittable type.")]
    private static FieldInfo[] Fields(object submittable) => submittable.GetType()
        .GetFields(BindingFlags.Instance | BindingFlags.Public);

    private static object Read(FieldInfo field, object target) =>
        field.GetValue(target) ?? throw new ArgumentException("form element fields cannot be null");

    private static void WriteCustom(Utf8JsonWriter writer, Form.Custom form)
    {
        writer.WriteStartObject();
        writer.WriteString("type", "custom_form");
        writer.WriteString("title", form.FormTitle);
        writer.WritePropertyName("content");
        writer.WriteStartArray();
        foreach (var element in Elements(form)) WriteElement(writer, element);
        writer.WriteEndArray();
        writer.WriteEndObject();
    }

    private static void WriteMenu(Utf8JsonWriter writer, Form.Menu form)
    {
        writer.WriteStartObject();
        writer.WriteString("type", "form");
        writer.WriteString("title", form.FormTitle);
        writer.WriteString("content", form.FormBody);
        writer.WritePropertyName("elements");
        writer.WriteStartArray();
        foreach (var element in Elements(form)) WriteMenuElement(writer, element);
        writer.WriteEndArray();
        writer.WriteEndObject();
    }

    private static void WriteModal(Utf8JsonWriter writer, Form.Modal form)
    {
        var buttons = Buttons(form);
        writer.WriteStartObject();
        writer.WriteString("type", "modal");
        writer.WriteString("title", form.FormTitle);
        writer.WriteString("content", form.FormBody);
        writer.WriteString("button1", buttons[0].Text ?? string.Empty);
        writer.WriteString("button2", buttons[1].Text ?? string.Empty);
        writer.WriteEndObject();
    }

    private static void WriteElement(Utf8JsonWriter writer, Form.Element element)
    {
        switch (element)
        {
            case Form.Divider:
                WriteTextElement(writer, "divider", string.Empty);
                break;
            case Form.Header header:
                WriteTextElement(writer, "header", header.Text);
                break;
            case Form.Label label:
                WriteTextElement(writer, "label", label.Text);
                break;
            case Form.Input input:
                writer.WriteStartObject();
                writer.WriteString("type", "input");
                writer.WriteString("text", input.Text ?? string.Empty);
                writer.WriteString("default", input.Default ?? string.Empty);
                writer.WriteString("placeholder", input.Placeholder ?? string.Empty);
                WriteTooltip(writer, input.Tooltip);
                writer.WriteEndObject();
                break;
            case Form.Toggle toggle:
                writer.WriteStartObject();
                writer.WriteString("type", "toggle");
                writer.WriteString("text", toggle.Text ?? string.Empty);
                writer.WriteBoolean("default", toggle.Default);
                WriteTooltip(writer, toggle.Tooltip);
                writer.WriteEndObject();
                break;
            case Form.Slider slider:
                if (!double.IsFinite(slider.Min) || !double.IsFinite(slider.Max) ||
                    !double.IsFinite(slider.StepSize) || !double.IsFinite(slider.Default))
                    throw new ArgumentException("slider contains a non-finite value");
                writer.WriteStartObject();
                writer.WriteString("type", "slider");
                writer.WriteString("text", slider.Text ?? string.Empty);
                writer.WriteNumber("min", slider.Min);
                writer.WriteNumber("max", slider.Max);
                writer.WriteNumber("step", slider.StepSize);
                writer.WriteNumber("default", slider.Default);
                WriteTooltip(writer, slider.Tooltip);
                writer.WriteEndObject();
                break;
            case Form.Dropdown dropdown:
                WriteOptions(writer, "dropdown", "options", dropdown.Text, dropdown.Options, dropdown.DefaultIndex, dropdown.Tooltip);
                break;
            case Form.StepSlider slider:
                WriteOptions(writer, "step_slider", "steps", slider.Text, slider.Options, slider.DefaultIndex, slider.Tooltip);
                break;
            default:
                throw new ArgumentException("form element type is not registered", nameof(element));
        }
    }

    private static void WriteMenuElement(Utf8JsonWriter writer, Form.MenuElement element)
    {
        if (element is Form.Element custom)
        {
            WriteElement(writer, custom);
            return;
        }
        if (element is not Form.Button button)
            throw new ArgumentException("menu element type is not registered", nameof(element));
        writer.WriteStartObject();
        writer.WriteString("type", "button");
        writer.WriteString("text", button.Text ?? string.Empty);
        if (!string.IsNullOrEmpty(button.Image))
        {
            writer.WritePropertyName("image");
            writer.WriteStartObject();
            writer.WriteString("type", button.Image.StartsWith("http:", StringComparison.Ordinal) ||
                                       button.Image.StartsWith("https:", StringComparison.Ordinal) ? "url" : "path");
            writer.WriteString("data", button.Image);
            writer.WriteEndObject();
        }
        writer.WriteEndObject();
    }

    private static void WriteTextElement(Utf8JsonWriter writer, string type, string? text)
    {
        writer.WriteStartObject();
        writer.WriteString("type", type);
        writer.WriteString("text", text ?? string.Empty);
        writer.WriteEndObject();
    }

    private static void WriteTooltip(Utf8JsonWriter writer, string? tooltip)
    {
        if (!string.IsNullOrEmpty(tooltip)) writer.WriteString("tooltip", tooltip);
    }

    private static void WriteOptions(
        Utf8JsonWriter writer,
        string type,
        string optionsName,
        string? text,
        string[]? options,
        int defaultIndex,
        string? tooltip)
    {
        writer.WriteStartObject();
        writer.WriteString("type", type);
        writer.WriteString("text", text ?? string.Empty);
        writer.WriteNumber("default", defaultIndex);
        writer.WritePropertyName(optionsName);
        if (options is null)
        {
            writer.WriteNullValue();
        }
        else
        {
            writer.WriteStartArray();
            foreach (var option in options) writer.WriteStringValue(option ?? string.Empty);
            writer.WriteEndArray();
        }
        WriteTooltip(writer, tooltip);
        writer.WriteEndObject();
    }
}
