namespace Dragonfly;

public static class Session
{
    public readonly record struct Diagnostics(
        double AverageFramesPerSecond,
        double AverageServerSimTickTime,
        double AverageClientSimTickTime,
        double AverageBeginFrameTime,
        double AverageInputTime,
        double AverageRenderTime,
        double AverageEndFrameTime,
        double AverageRemainderTimePercent,
        double AverageUnaccountedTimePercent);
}
