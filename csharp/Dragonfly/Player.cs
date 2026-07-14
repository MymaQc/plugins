using Dragonfly.Native;

namespace Dragonfly;

public sealed partial class Player
{
    internal Player(PlayerId id, string name = "")
    {
        Id = id;
        Name = name;
    }

    internal PlayerId Id { get; }
    public string Name { get; }

    public sealed class Context
    {
        private bool _cancelled;

        internal Context(Player player, bool cancelled)
        {
            Player = player;
            _cancelled = cancelled;
        }

        public Player Player { get; }
        public bool Cancelled() => _cancelled;
        public void Cancel() => _cancelled = true;
    }
}
