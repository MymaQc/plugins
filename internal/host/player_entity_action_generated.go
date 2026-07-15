// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
package host

import (
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

func runExactPlayerEntityAction(connected *player.Player, entity world.Entity, kind native.PlayerEntityActionKind) (bool, bool) {
	switch kind {
	case native.PlayerEntityActionUseItemOnEntity:
		return connected.UseItemOnEntity(entity), true
	case native.PlayerEntityActionAttackEntity:
		return connected.AttackEntity(entity), true
	default:
		return false, false
	}
}
