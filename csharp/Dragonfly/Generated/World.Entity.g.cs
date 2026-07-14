// Code generated from Dragonfly server/world/entity.go Go AST. DO NOT EDIT.
#nullable enable
using System;
using System.Collections.Generic;

namespace Dragonfly;

public sealed partial class World
{
    public interface EntityType
    {
        Entity Open(Tx tx, EntityHandle handle, EntityData data);
        string EncodeEntity();
        Cube.BBox BBox(Entity e);
        void DecodeNBT(Dictionary<string, object?> m, EntityData data);
        Dictionary<string, object?> EncodeNBT(EntityData data);
    }

    public interface EntityConfig
    {
        void Apply(EntityData data);
    }

    public interface TickerEntity : Entity
    {
        void Tick(Tx tx, long current);
    }

    public sealed class EntityData
    {
        public Vector3 Pos = default;
        public Vector3 Vel = default;
        public Rotation Rot = default;
        public string Name = string.Empty;
        public TimeSpan FireDuration = default;
        public TimeSpan Age = default;
        public object? Data;
    }

    public sealed class EntitySpawnOpts
    {
        public Vector3 Position = default;
        public Rotation Rotation = default;
        public Vector3 Velocity = default;
        public Guid ID = default;
        public string NameTag = string.Empty;

        public EntityHandle New(EntityType t, EntityConfig conf) =>
            PluginBridge.Host.NewEntity(this, t, conf);
    }

    public static EntityHandle NewEntity(EntityType t, EntityConfig conf) =>
        new EntitySpawnOpts().New(t, conf);

    public interface Entity
    {
        void Close();
        EntityHandle H();
        Vector3 Position();
        Rotation Rotation();
    }

    public sealed partial class EntityHandle
    {
        public (Entity? Entity, bool Ok) Entity(Tx tx) =>
            PluginBridge.Host.EntityHandleEntity(tx.Invocation, this);

        public Guid UUID() =>
            PluginBridge.Host.EntityHandleUuid(this);

        public bool Closed() =>
            PluginBridge.Host.EntityHandleClosed(this);

        public void Close() =>
            PluginBridge.Host.CloseEntityHandle(this);
    }
}
