// Code generated from Dragonfly server/player/player.go Go AST. DO NOT EDIT.
package host

import (
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/go-gl/mathgl/mgl64"
)

func runPlayerBlockAction(connected *player.Player, kind native.PlayerBlockActionKind, position native.BlockPos, face int32, clickPosition native.Vec3) bool {
	pos := cube.Pos{int(position.X), int(position.Y), int(position.Z)}
	switch kind {
	case native.PlayerBlockActionBreakBlock:
		connected.BreakBlock(pos)
		return true
	case native.PlayerBlockActionContinueBreaking:
		connected.ContinueBreaking(cube.Face(face))
		return true
	case native.PlayerBlockActionPickBlock:
		connected.PickBlock(pos)
		return true
	case native.PlayerBlockActionSleep:
		connected.Sleep(pos)
		return true
	case native.PlayerBlockActionStartBreaking:
		connected.StartBreaking(pos, cube.Face(face))
		return true
	case native.PlayerBlockActionUseItemOnBlock:
		connected.UseItemOnBlock(pos, cube.Face(face), mgl64.Vec3{clickPosition.X, clickPosition.Y, clickPosition.Z})
		return true
	default:
		return false
	}
}
