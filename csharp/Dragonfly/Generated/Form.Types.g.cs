// Code generated from Dragonfly server/player/form Go AST. DO NOT EDIT.
#nullable enable
using System;
using System.Collections.Generic;

namespace Dragonfly;

public static partial class Form
{
    public interface Value
    {
        byte[] MarshalJSON();
        void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx);
    }

    public interface Element
    {
        byte[] MarshalJSON();
    }

    public interface MenuElement
    {
        byte[] MarshalJSON();
    }

    public interface Submittable
    {
        void Submit(Submitter submitter, World.Tx tx);
    }

    public interface MenuSubmittable
    {
        void Submit(Submitter submitter, Button pressed, World.Tx tx);
    }

    public interface ModalSubmittable : MenuSubmittable { }

    public interface Closer
    {
        void Close(Submitter submitter, World.Tx tx);
    }

    public interface Submitter
    {
        void SendForm(Value form);
        void CloseForm();
    }

    public struct Divider : Element, MenuElement
    {

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);
    }


    public struct Header : Element, MenuElement
    {
        public string Text;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);
    }

    public static Header NewHeader(string text) => new()
    {
        Text = text,
    };

    public struct Label : Element, MenuElement
    {
        public string Text;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);
    }

    public static Label NewLabel(string text) => new()
    {
        Text = text,
    };

    public struct Input : Element
    {
        public string Text;
        public string Default;
        public string Placeholder;
        public string Tooltip;
        private string? _value;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);

        public readonly Input WithTooltip(string tooltip)
        {
            var copy = this;
            copy.Tooltip = tooltip;
            return copy;
        }

        public readonly string Value() => _value ?? string.Empty;

        internal readonly Input WithValue(string value)
        {
            var copy = this;
            copy._value = value;
            return copy;
        }
    }

    public static Input NewInput(string text, string defaultValue, string placeholder) => new()
    {
        Text = text,
        Default = defaultValue,
        Placeholder = placeholder,
    };

    public struct Toggle : Element
    {
        public string Text;
        public bool Default;
        public string Tooltip;
        private bool _value;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);

        public readonly Toggle WithTooltip(string tooltip)
        {
            var copy = this;
            copy.Tooltip = tooltip;
            return copy;
        }

        public readonly bool Value() => _value;

        internal readonly Toggle WithValue(bool value)
        {
            var copy = this;
            copy._value = value;
            return copy;
        }
    }

    public static Toggle NewToggle(string text, bool defaultValue) => new()
    {
        Text = text,
        Default = defaultValue,
    };

    public struct Slider : Element
    {
        public string Text;
        public double Min;
        public double Max;
        public double StepSize;
        public double Default;
        public string Tooltip;
        private double _value;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);

        public readonly Slider WithTooltip(string tooltip)
        {
            var copy = this;
            copy.Tooltip = tooltip;
            return copy;
        }

        public readonly double Value() => _value;

        internal readonly Slider WithValue(double value)
        {
            var copy = this;
            copy._value = value;
            return copy;
        }
    }

    public static Slider NewSlider(string text, double min, double max, double stepSize, double defaultValue) => new()
    {
        Text = text,
        Min = min,
        Max = max,
        StepSize = stepSize,
        Default = defaultValue,
    };

    public struct Dropdown : Element
    {
        public string Text;
        public string[] Options;
        public int DefaultIndex;
        public string Tooltip;
        private int _value;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);

        public readonly Dropdown WithTooltip(string tooltip)
        {
            var copy = this;
            copy.Tooltip = tooltip;
            return copy;
        }

        public readonly int Value() => _value;

        internal readonly Dropdown WithValue(int value)
        {
            var copy = this;
            copy._value = value;
            return copy;
        }
    }

    public static Dropdown NewDropdown(string text, string[] options, int defaultIndex) => new()
    {
        Text = text,
        Options = options,
        DefaultIndex = defaultIndex,
    };

    public struct StepSlider : Element
    {
        public string Text;
        public string[] Options;
        public int DefaultIndex;
        public string Tooltip;
        private int _value;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeElement(this);

        public readonly StepSlider WithTooltip(string tooltip)
        {
            var copy = this;
            copy.Tooltip = tooltip;
            return copy;
        }

        public readonly int Value() => _value;

        internal readonly StepSlider WithValue(int value)
        {
            var copy = this;
            copy._value = value;
            return copy;
        }
    }

    public static StepSlider NewStepSlider(string text, string[] options, int defaultIndex) => new()
    {
        Text = text,
        Options = options,
        DefaultIndex = defaultIndex,
    };

    public struct Button : MenuElement
    {
        public string Text;
        public string Image;

        public readonly byte[] MarshalJSON() => FormCodec.EncodeMenuElement(this);
    }

    public static Button NewButton(string text, string image) => new()
    {
        Text = text,
        Image = image,
    };

    public readonly struct Custom : Value
    {
        internal Custom(Submittable submittable, string title) =>
            (Submittable, FormTitle) = (submittable, title);

        internal Submittable Submittable { get; }
        internal string FormTitle { get; }

        public string Title() => FormTitle;
        public IReadOnlyList<Element> Elements() => FormCodec.Elements(this);
        public byte[] MarshalJSON() => FormCodec.Encode(this);
        public void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx) =>
            FormCodec.Respond(this, submitter, tx, response is null, response ?? Array.Empty<byte>());
    }

    public readonly struct Menu : Value
    {
        internal Menu(MenuSubmittable submittable, string title, string body, MenuElement[] elements) =>
            (Submittable, FormTitle, FormBody, ExtraElements) = (submittable, title, body, elements);

        internal MenuSubmittable Submittable { get; }
        internal string FormTitle { get; }
        internal string FormBody { get; }
        internal MenuElement[] ExtraElements { get; }

        public Menu WithBody(params object?[] body) =>
            new(Submittable, FormTitle, FormCodec.Format(body), ExtraElements);

        public Menu AddButton(Button button) => WithElements(button);
        public Menu AddDivider(Divider divider) => WithElements(divider);
        public Menu AddHeader(Header header) => WithElements(header);
        public Menu AddLabel(Label label) => WithElements(label);

        public Menu WithButtons(params Button[] buttons)
        {
            var elements = new MenuElement[buttons.Length];
            for (var index = 0; index < buttons.Length; index++) elements[index] = buttons[index];
            return WithElements(elements);
        }

        public Menu WithElements(params MenuElement[] elements)
        {
            var combined = new MenuElement[ExtraElements.Length + elements.Length];
            ExtraElements.CopyTo(combined, 0);
            elements.CopyTo(combined, ExtraElements.Length);
            return new Menu(Submittable, FormTitle, FormBody, combined);
        }

        public string Title() => FormTitle;
        public string Body() => FormBody;
        public IReadOnlyList<Button> Buttons() => FormCodec.Buttons(this);
        public IReadOnlyList<MenuElement> Elements() => FormCodec.Elements(this);
        public byte[] MarshalJSON() => FormCodec.Encode(this);
        public void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx) =>
            FormCodec.Respond(this, submitter, tx, response is null, response ?? Array.Empty<byte>());
    }

    public readonly struct Modal : Value
    {
        internal Modal(ModalSubmittable submittable, string title, string body) =>
            (Submittable, FormTitle, FormBody) = (submittable, title, body);

        internal ModalSubmittable Submittable { get; }
        internal string FormTitle { get; }
        internal string FormBody { get; }

        public Modal WithBody(params object?[] body) =>
            new(Submittable, FormTitle, FormCodec.Format(body));

        public string Title() => FormTitle;
        public string Body() => FormBody;
        public IReadOnlyList<Button> Buttons() => FormCodec.Buttons(this);
        public byte[] MarshalJSON() => FormCodec.Encode(this);
        public void SubmitJSON(byte[]? response, Submitter submitter, World.Tx tx) =>
            FormCodec.Respond(this, submitter, tx, response is null, response ?? Array.Empty<byte>());
    }

    public static Custom New(Submittable submittable, params object?[] title)
    {
        ArgumentNullException.ThrowIfNull(submittable);
        FormCodec.VerifyCustom(submittable);
        return new Custom(submittable, FormCodec.Format(title));
    }

    public static Menu NewMenu(MenuSubmittable submittable, params object?[] title)
    {
        ArgumentNullException.ThrowIfNull(submittable);
        FormCodec.VerifyMenu(submittable);
        return new Menu(submittable, FormCodec.Format(title), string.Empty, Array.Empty<MenuElement>());
    }

    public static Modal NewModal(ModalSubmittable submittable, params object?[] title)
    {
        ArgumentNullException.ThrowIfNull(submittable);
        FormCodec.VerifyModal(submittable);
        return new Modal(submittable, FormCodec.Format(title), string.Empty);
    }

    public static Button YesButton() => NewButton("gui.yes", string.Empty);
    public static Button NoButton() => NewButton("gui.no", string.Empty);
}
