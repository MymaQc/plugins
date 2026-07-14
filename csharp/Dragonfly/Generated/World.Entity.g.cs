// Code generated from Dragonfly server/world/entity.go Go AST. DO NOT EDIT.
#nullable enable
using System;

namespace Dragonfly;

public sealed partial class World
{
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
