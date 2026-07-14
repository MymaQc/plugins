namespace Dragonfly;

// Owns plugin objects used to implement Dragonfly's world.EntityType contract.
// Numeric keys are private transport identities and never enter the public API.
internal sealed class CustomEntityTable : IDisposable
{
    private readonly object _gate = new();
    private readonly World.EntityType[] _types;
    private readonly Dictionary<World.EntityType, ulong> _typeKeys =
        new(ReferenceEqualityComparer.Instance);
    private readonly Dictionary<string, World.EntityType> _typesByIdentifier =
        new(StringComparer.Ordinal);
    private readonly Dictionary<ulong, EntityState> _pending = [];
    private readonly Dictionary<ulong, EntityState> _instances = [];
    private readonly Dictionary<ulong, OpenState> _opened = [];
    private readonly Dictionary<(ulong Value, ulong Generation), OpenState> _openedByHandle = [];
    private readonly Dictionary<(ulong Value, ulong Generation), World.EntityType> _handles = [];
    private readonly Dictionary<(ulong Value, ulong Generation), World.EntityHandle> _canonicalHandles = [];
    private long _nextObject;

    internal CustomEntityTable(IEnumerable<World.EntityType> types)
    {
        ArgumentNullException.ThrowIfNull(types);
        _types = types.ToArray();
        for (var index = 0; index < _types.Length; index++)
        {
            var type = _types[index] ?? throw new InvalidOperationException("an entity type factory returned null");
            if (!_typeKeys.TryAdd(type, checked((ulong)index + 1)))
                throw new InvalidOperationException("an entity type factory returned the same object twice");
            var encoded = type.EncodeEntity();
            if (string.IsNullOrWhiteSpace(encoded))
                throw new InvalidOperationException("an entity type returned an empty encoded identifier");
            if (!_typesByIdentifier.TryAdd(encoded, type))
                throw new InvalidOperationException($"duplicate entity type {encoded}");
        }
    }

    internal int Count => _types.Length;

    internal (ulong Key, World.EntityType Type) TypeAt(int index)
    {
        if ((uint)index >= (uint)_types.Length) throw new ArgumentOutOfRangeException(nameof(index));
        return (checked((ulong)index + 1), _types[index]);
    }

    internal ulong TypeKey(World.EntityType type)
    {
        ArgumentNullException.ThrowIfNull(type);
        lock (_gate)
        {
            if (_typeKeys.TryGetValue(type, out var key)) return key;
        }
        var identifier = type.EncodeEntity();
        lock (_gate)
        {
            if (_typesByIdentifier.TryGetValue(identifier, out var canonical) &&
                _typeKeys.TryGetValue(canonical, out var key)) return key;
        }
        throw new ArgumentException("entity type is not declared by this plugin", nameof(type));
    }

    internal World.EntityType TypeByIdentifier(string identifier)
    {
        ArgumentException.ThrowIfNullOrEmpty(identifier);
        lock (_gate)
        {
            if (_typesByIdentifier.TryGetValue(identifier, out var type)) return type;
        }
        throw new InvalidOperationException($"entity type {identifier} is not declared by this plugin");
    }

    // Applies EntityConfig exactly once before ownership is offered to the host.
    internal ulong Prepare(
        World.EntitySpawnOpts options,
        World.EntityType type,
        World.EntityConfig config)
    {
        ArgumentNullException.ThrowIfNull(options);
        ArgumentNullException.ThrowIfNull(type);
        ArgumentNullException.ThrowIfNull(config);
        var key = TypeKey(type);
        var data = new World.EntityData
        {
            Pos = options.Position,
            Rot = options.Rotation,
            Vel = options.Velocity,
            Name = options.NameTag ?? string.Empty,
        };
        config.Apply(data);
        var opaque = NextObject();
        lock (_gate) _pending.Add(opaque, new EntityState(key, type, data));
        return opaque;
    }

    // Transfers one prepared object to the runtime. Failed adoption leaves the
    // object pending so the caller that offered it can release it explicitly.
    internal ulong Adopt(ulong typeKey, ulong opaque)
    {
        lock (_gate)
        {
            if (!_pending.TryGetValue(opaque, out var state) || state.TypeKey != typeKey)
                throw new InvalidOperationException("unknown pending entity object");
            _pending.Remove(opaque);
            _instances.Add(opaque, state);
            return opaque;
        }
    }

    internal void ReleasePending(ulong opaque)
    {
        lock (_gate) _pending.Remove(opaque);
    }

    internal World.EntityData PreparedData(ulong opaque)
    {
        lock (_gate)
        {
            if (_pending.TryGetValue(opaque, out var state)) return state.Data;
        }
        throw new InvalidOperationException("unknown pending entity object");
    }

    internal World.EntityData Data(ulong instance) => State(instance).Data;
    internal World.EntityData OpenData(ulong open) => Opened(open).EntityState.Data;

    internal World.EntityHandle BindHandle(ulong instance, ulong value, ulong generation)
    {
        var state = State(instance);
        lock (_gate)
        {
            var key = (value, generation);
            if (!_canonicalHandles.TryGetValue(key, out var handle))
            {
                handle = new World.EntityHandle(new Native.EntityHandleId { Value = value, Generation = generation });
                _canonicalHandles.Add(key, handle);
            }
            state.Handle = handle;
            _handles[key] = state.Type;
            return handle;
        }
    }

    internal World.EntityHandle CanonicalHandle(World.EntityHandle handle)
    {
        ArgumentNullException.ThrowIfNull(handle);
        lock (_gate)
        {
            return _canonicalHandles.TryGetValue((handle.Id.Value, handle.Id.Generation), out var canonical)
                ? canonical
                : handle;
        }
    }

    internal World.EntityType HandleType(ulong value, ulong generation)
    {
        lock (_gate)
        {
            if (_handles.TryGetValue((value, generation), out var type)) return type;
        }
        throw new InvalidOperationException("entity handle is not owned by this plugin");
    }

    internal (ulong Open, uint Capabilities) Open(ulong instance, ulong invocation, World.EntityHandle handle)
    {
        ArgumentNullException.ThrowIfNull(handle);
        handle = CanonicalHandle(handle);
        var state = State(instance);
        var lease = new InvocationLease();
        World.Entity opened;
        using (lease.Enter(invocation))
        {
            opened = state.Type.Open(new World.Tx(lease), handle, state.Data) ??
                throw new InvalidOperationException("entity type returned null from Open");
        }
        var id = NextObject();
        lock (_gate)
        {
            var openState = new OpenState(state, lease, opened, handle.Id.Value, handle.Id.Generation);
            _opened.Add(id, openState);
            _openedByHandle[(handle.Id.Value, handle.Id.Generation)] = openState;
        }
        return (id, opened is World.TickerEntity ? 1u : 0u);
    }

    internal World.Entity? OpenedEntity(World.EntityHandle handle)
    {
        ArgumentNullException.ThrowIfNull(handle);
        lock (_gate)
        {
            return _openedByHandle.TryGetValue((handle.Id.Value, handle.Id.Generation), out var state)
                ? state.Entity
                : null;
        }
    }

    internal Cube.BBox BBox(ulong open, ulong invocation)
    {
        var state = Opened(open);
        lock (state.Gate)
        {
            using var scope = state.Invocation.Enter(invocation);
            return state.EntityState.Type.BBox(state.Entity);
        }
    }

    internal Dictionary<string, object?> EncodeNBT(ulong instance)
    {
        var state = State(instance);
        return state.Type.EncodeNBT(state.Data) ?? [];
    }

    internal ulong DecodeNBT(ulong typeKey, Dictionary<string, object?> values, World.EntityData data)
    {
        ArgumentNullException.ThrowIfNull(values);
        ArgumentNullException.ThrowIfNull(data);
        var type = TypeAt(checked((int)typeKey - 1)).Type;
        type.DecodeNBT(values, data);
        var instance = NextObject();
        lock (_gate) _instances.Add(instance, new EntityState(typeKey, type, data));
        return instance;
    }

    internal void DispatchClose(ulong open, ulong invocation)
    {
        var state = Opened(open);
        lock (state.Gate)
        {
            using var scope = state.Invocation.Enter(invocation);
            state.Entity.Close();
        }
    }

    internal World.EntityHandle DispatchHandle(ulong open, ulong invocation)
    {
        var state = Opened(open);
        lock (state.Gate)
        {
            using var scope = state.Invocation.Enter(invocation);
            return state.Entity.H();
        }
    }

    internal Vector3 DispatchPosition(ulong open, ulong invocation)
    {
        var state = Opened(open);
        lock (state.Gate)
        {
            using var scope = state.Invocation.Enter(invocation);
            return state.Entity.Position();
        }
    }

    internal Rotation DispatchRotation(ulong open, ulong invocation)
    {
        var state = Opened(open);
        lock (state.Gate)
        {
            using var scope = state.Invocation.Enter(invocation);
            return state.Entity.Rotation();
        }
    }

    internal void DispatchTick(ulong open, ulong invocation, long current)
    {
        var state = Opened(open);
        lock (state.Gate)
        {
            using var scope = state.Invocation.Enter(invocation);
            if (state.Entity is not World.TickerEntity ticker)
                throw new InvalidOperationException("entity is not a ticking entity");
            ticker.Tick(new World.Tx(state.Invocation), current);
        }
    }

    internal void ReleaseOpen(ulong open)
    {
        lock (_gate)
        {
            if (!_opened.Remove(open, out var state)) return;
            var key = (state.HandleValue, state.HandleGeneration);
            if (_openedByHandle.TryGetValue(key, out var current) && ReferenceEquals(current, state))
                _openedByHandle.Remove(key);
        }
    }

    internal void Destroy(ulong instance)
    {
        lock (_gate)
        {
            if (!_instances.Remove(instance, out var state)) return;
            if (state.Handle is not null)
            {
                _handles.Remove((state.Handle.Id.Value, state.Handle.Id.Generation));
                _canonicalHandles.Remove((state.Handle.Id.Value, state.Handle.Id.Generation));
            }
        }
    }

    public void Dispose()
    {
        lock (_gate)
        {
            _pending.Clear();
            _instances.Clear();
            _opened.Clear();
            _openedByHandle.Clear();
            _handles.Clear();
            _canonicalHandles.Clear();
        }
    }

    private EntityState State(ulong instance)
    {
        lock (_gate)
        {
            if (_instances.TryGetValue(instance, out var state)) return state;
        }
        throw new InvalidOperationException("unknown entity instance");
    }

    private OpenState Opened(ulong open)
    {
        lock (_gate)
        {
            if (_opened.TryGetValue(open, out var state)) return state;
        }
        throw new InvalidOperationException("unknown open entity session");
    }

    private ulong NextObject()
    {
        var value = Interlocked.Increment(ref _nextObject);
        if (value <= 0) throw new InvalidOperationException("entity object identity exhausted");
        return checked((ulong)value);
    }

    private sealed class EntityState(ulong typeKey, World.EntityType type, World.EntityData data)
    {
        internal ulong TypeKey { get; } = typeKey;
        internal World.EntityType Type { get; } = type;
        internal World.EntityData Data { get; } = data;
        internal World.EntityHandle? Handle { get; set; }
    }

    private sealed class OpenState(
        EntityState entityState,
        InvocationLease invocation,
        World.Entity entity,
        ulong handleValue,
        ulong handleGeneration)
    {
        internal object Gate { get; } = new();
        internal EntityState EntityState { get; } = entityState;
        internal InvocationLease Invocation { get; } = invocation;
        internal World.Entity Entity { get; } = entity;
        internal ulong HandleValue { get; } = handleValue;
        internal ulong HandleGeneration { get; } = handleGeneration;
    }
}
