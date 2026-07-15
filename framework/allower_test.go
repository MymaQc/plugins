package framework

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/df-mc/dragonfly/server"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

type testAllower struct {
	message string
	allowed bool
	calls   int
}

func (a *testAllower) Allow(net.Addr, login.IdentityData, login.ClientData) (string, bool) {
	a.calls++
	return a.message, a.allowed
}

type testAdmissionRuntime struct {
	message string
	allowed bool
	err     error
	calls   int
}

func (r *testAdmissionRuntime) Allow(net.Addr, login.IdentityData, login.ClientData) (string, bool, error) {
	r.calls++
	return r.message, r.allowed, r.err
}

func TestComposeAllowerKeepsExistingDenialFirst(t *testing.T) {
	base := &testAllower{message: "base", allowed: false}
	runtime := &testAdmissionRuntime{allowed: true}
	allower := composeAllower(base, runtime, slog.New(slog.NewTextHandler(io.Discard, nil)))
	message, allowed := allower.Allow(&net.UDPAddr{}, login.IdentityData{}, login.ClientData{})
	if allowed || message != "base" || base.calls != 1 || runtime.calls != 0 {
		t.Fatalf("got (%q, %v), base calls %d, runtime calls %d", message, allowed, base.calls, runtime.calls)
	}
}

func TestComposeAllowerReturnsPluginDecision(t *testing.T) {
	runtime := &testAdmissionRuntime{message: "banned", allowed: false}
	allower := composeAllower(nil, runtime, slog.New(slog.NewTextHandler(io.Discard, nil)))
	message, allowed := allower.Allow(&net.UDPAddr{}, login.IdentityData{}, login.ClientData{})
	if allowed || message != "banned" || runtime.calls != 1 {
		t.Fatalf("got (%q, %v), runtime calls %d", message, allowed, runtime.calls)
	}
}

func TestComposeAllowerRedactsRuntimeErrors(t *testing.T) {
	runtime := &testAdmissionRuntime{err: errors.New("secret native detail")}
	allower := composeAllower(server.Allower(nil), runtime, slog.New(slog.NewTextHandler(io.Discard, nil)))
	message, allowed := allower.Allow(&net.UDPAddr{}, login.IdentityData{}, login.ClientData{})
	if allowed || message != pluginAdmissionFailureMessage {
		t.Fatalf("got (%q, %v)", message, allowed)
	}
}
