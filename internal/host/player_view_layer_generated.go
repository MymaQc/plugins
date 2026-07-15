// Code generated from Dragonfly player.go and visibility_level.go Go AST. DO NOT EDIT.
package host

import (
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

func runPlayerViewLayer(viewer *player.Player, entity world.Entity, kind native.PlayerViewLayerKind, text string, visibility uint8) bool {
	switch kind {
	case native.PlayerViewLayerViewNameTag:
		viewer.ViewNameTag(entity, text)
		return true
	case native.PlayerViewLayerViewPublicNameTag:
		viewer.ViewPublicNameTag(entity)
		return true
	case native.PlayerViewLayerViewScoreTag:
		viewer.ViewScoreTag(entity, text)
		return true
	case native.PlayerViewLayerViewPublicScoreTag:
		viewer.ViewPublicScoreTag(entity)
		return true
	case native.PlayerViewLayerViewVisibility:
		var level world.VisibilityLevel
		switch visibility {
		case 0:
			level = world.PublicVisibility()
		case 1:
			level = world.EnforceInvisible()
		case 2:
			level = world.EnforceVisible()
		default:
			return false
		}
		viewer.ViewVisibility(entity, level)
		return true
	case native.PlayerViewLayerRemoveViewLayer:
		viewer.RemoveViewLayer(entity)
		return true
	default:
		return false
	}
}
