namespace Dragonfly;

public readonly record struct ImageBounds(int Width, int Height);

public sealed class Skin
{
    private readonly ImageBounds _bounds;

    public Skin(int width, int height)
    {
        _bounds = CheckedBounds(width, height);
        Pix = new byte[checked(width * height * 4)];
    }

    internal Skin(int width, int height, byte[] pix)
    {
        _bounds = CheckedBounds(width, height);
        Pix = pix;
    }

    public bool Persona { get; set; }
    public string PlayFabID { get; set; } = "";
    public string FullID { get; set; } = "";
    public byte[] Pix { get; set; }
    public SkinModelConfig ModelConfig { get; set; } = new();
    public byte[] Model { get; set; } = [];
    public SkinCape Cape { get; set; } = SkinCape.Empty;
    public SkinAnimation[] Animations { get; set; } = [];

    public ImageBounds Bounds() => _bounds;

    private static ImageBounds CheckedBounds(int width, int height)
    {
        if (width < 0 || height < 0 || width > 4096 || height > 4096 ||
            (width == 0) != (height == 0))
            throw new ArgumentOutOfRangeException(nameof(width), "invalid skin bounds");
        return new ImageBounds(width, height);
    }

}

public sealed class SkinModelConfig
{
    public string Default { get; set; } = "";
    public string AnimatedFace { get; set; } = "";
}

public sealed class SkinCape
{
    private readonly ImageBounds _bounds;

    public SkinCape(int width, int height)
    {
        _bounds = CheckedBounds(width, height);
        Pix = new byte[checked(width * height * 4)];
    }

    internal SkinCape(int width, int height, byte[] pix)
    {
        _bounds = CheckedBounds(width, height);
        Pix = pix;
    }

    public static SkinCape Empty => new(0, 0, []);
    public byte[] Pix { get; set; }
    public ImageBounds Bounds() => _bounds;

    private static ImageBounds CheckedBounds(int width, int height)
    {
        if (width < 0 || height < 0 || width > 4096 || height > 4096 ||
            (width == 0) != (height == 0))
            throw new ArgumentOutOfRangeException(nameof(width), "invalid cape bounds");
        return new ImageBounds(width, height);
    }
}

public enum SkinAnimationType
{
    Head,
    Body32x32,
    Body128x128,
}

public sealed class SkinAnimation
{
    private readonly ImageBounds _bounds;
    private readonly SkinAnimationType _type;

    public SkinAnimation(int width, int height, int expression, SkinAnimationType type)
    {
        _bounds = CheckedBounds(width, height);
        _type = type;
        Pix = new byte[checked(width * height * 4)];
        FrameCount = 1;
        AnimationExpression = expression;
    }

    internal SkinAnimation(
        int width,
        int height,
        SkinAnimationType type,
        byte[] pix,
        int frameCount,
        int expression)
    {
        _bounds = CheckedBounds(width, height);
        _type = type;
        Pix = pix;
        FrameCount = frameCount;
        AnimationExpression = expression;
    }

    public byte[] Pix { get; set; }
    public int FrameCount { get; set; }
    public int AnimationExpression { get; set; }
    public SkinAnimationType Type() => _type;
    public ImageBounds Bounds() => _bounds;

    private static ImageBounds CheckedBounds(int width, int height)
    {
        if (width <= 0 || height <= 0 || width > 4096 || height > 4096)
            throw new ArgumentOutOfRangeException(nameof(width), "invalid animation bounds");
        return new ImageBounds(width, height);
    }
}
