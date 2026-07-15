package framework

import (
	"log/slog"
	"net"

	"github.com/df-mc/dragonfly/server"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

const pluginAdmissionFailureMessage = "Connection rejected by a plugin."

type pluginAdmissionRuntime interface {
	Allow(net.Addr, login.IdentityData, login.ClientData) (string, bool, error)
}

type pluginAllower struct {
	base    server.Allower
	runtime pluginAdmissionRuntime
	log     *slog.Logger
}

func (a pluginAllower) Allow(addr net.Addr, identity login.IdentityData, client login.ClientData) (string, bool) {
	if a.base != nil {
		if message, allowed := a.base.Allow(addr, identity, client); !allowed {
			return message, false
		}
	}
	message, allowed, err := a.runtime.Allow(addr, identity, client)
	if err != nil {
		a.log.Error("plugin connection allower failed", "address", addr.String(), "error", err)
		return pluginAdmissionFailureMessage, false
	}
	return message, allowed
}

func composeAllower(base server.Allower, runtime pluginAdmissionRuntime, log *slog.Logger) server.Allower {
	return pluginAllower{base: base, runtime: runtime, log: log}
}
