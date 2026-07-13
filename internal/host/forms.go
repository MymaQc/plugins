package host

import (
	"encoding/json"
	"errors"
	"runtime"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
)

type nativePlayerForm struct {
	id      uint64
	request []byte
	players *Players
}

var _ form.Form = (*nativePlayerForm)(nil)

func (f *nativePlayerForm) MarshalJSON() ([]byte, error) {
	if !json.Valid(f.request) {
		return nil, errors.New("invalid native form JSON")
	}
	return append([]byte(nil), f.request...), nil
}

func (f *nativePlayerForm) SubmitJSON(response []byte, submitter form.Submitter, tx *world.Tx) error {
	connected, ok := submitter.(*player.Player)
	if !ok {
		return errors.New("native form submitter is not a player")
	}
	id, ok := f.players.ID(connected)
	if !ok {
		return errors.New("native form submitter is not registered")
	}
	var accepted bool
	f.players.WithInvocation(tx, func(invocation native.InvocationID) {
		accepted = native.CompletePlayerForm(f.id, invocation, id, response == nil, response)
	})
	runtime.SetFinalizer(f, nil)
	if !accepted {
		return errors.New("native form response was rejected")
	}
	return nil
}
