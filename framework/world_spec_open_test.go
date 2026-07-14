package framework

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bedrock-gophers/plugins/internal/host"
	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
)

func persistentTestManager(t *testing.T) (*WorldManager, string) {
	t.Helper()
	root := t.TempDir()
	manager, err := NewPersistentWorldManager(root, nil, host.NewPlayers())
	if err != nil {
		t.Fatal(err)
	}
	core := world.Config{Synchronous: true}.New()
	if err := manager.RegisterCore(OverworldID, core); err != nil {
		_ = core.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = manager.CloseCustom()
		_ = core.Close()
	})
	return manager, root
}

func TestWorldManagerConcurrentIdenticalSpecsShareOneWorld(t *testing.T) {
	manager, _ := persistentTestManager(t)
	spec := validWorldSpec("arenas/shared")
	const callers = 8
	ids := make(chan native.WorldID, callers)
	errs := make(chan error, callers)
	var start sync.WaitGroup
	start.Add(1)
	var callersDone sync.WaitGroup
	callersDone.Add(callers)
	for range callers {
		go func() {
			defer callersDone.Done()
			start.Wait()
			id, err := manager.OpenSpec("example:shared", spec)
			ids <- id
			errs <- err
		}()
	}
	start.Done()
	callersDone.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var want native.WorldID
	for id := range ids {
		if want == 0 {
			want = id
		} else if id != want {
			t.Fatalf("world ID = %d, want %d", id, want)
		}
	}
	if got := manager.IDs(); len(got) != 2 {
		t.Fatalf("managed IDs = %v", got)
	}
}

func TestWorldManagerRejectsMismatchedDuplicateSpec(t *testing.T) {
	manager, _ := persistentTestManager(t)
	spec := validWorldSpec("arenas/one")
	if _, err := manager.OpenSpec("example:arena", spec); err != nil {
		t.Fatal(err)
	}
	spec.Time, spec.FixedTime = WorldTimeFixed, 6000
	if _, err := manager.OpenSpec("example:arena", spec); err == nil {
		t.Fatal("mismatched duplicate accepted")
	}
}

func TestWorldManagerRejectsProviderPathAlias(t *testing.T) {
	manager, _ := persistentTestManager(t)
	spec := validWorldSpec("arenas/one")
	if _, err := manager.OpenSpec("example:one", spec); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.OpenSpec("example:two", spec); err == nil {
		t.Fatal("provider path alias accepted")
	}
}

func TestWorldManagerRejectsProviderPathSymlink(t *testing.T) {
	manager, root := persistentTestManager(t)
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "alias")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := manager.OpenSpec("example:alias", validWorldSpec("alias/world")); err == nil {
		t.Fatal("symlink path accepted")
	}
}

func TestWorldManagerOpenExistingRequiresMCDBArtifacts(t *testing.T) {
	manager, root := persistentTestManager(t)
	spec := validWorldSpec("existing/world")
	spec.OpenMode = WorldOpenExisting
	if _, err := manager.OpenSpec("example:missing", spec); err == nil {
		t.Fatal("missing provider accepted")
	}
	createMCDBFixture(t, filepath.Join(root, "existing", "world"))
	if _, err := manager.OpenSpec("example:existing", spec); err != nil {
		t.Fatalf("valid provider rejected: %v", err)
	}
}

func TestWorldManagerCreateNewRejectsExistingPath(t *testing.T) {
	manager, root := persistentTestManager(t)
	if err := os.MkdirAll(filepath.Join(root, "existing"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := validWorldSpec("existing")
	spec.OpenMode = WorldCreateNew
	if _, err := manager.OpenSpec("example:new", spec); err == nil {
		t.Fatal("existing path accepted")
	}
}

func TestWorldManagerFailedOpenReleasesReservations(t *testing.T) {
	manager, root := persistentTestManager(t)
	path := filepath.Join(root, "retry")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	spec := validWorldSpec("retry")
	spec.OpenMode = WorldCreateNew
	if _, err := manager.OpenSpec("example:retry", spec); err == nil {
		t.Fatal("existing path accepted")
	}
	if err := os.RemoveAll(path); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.OpenSpec("example:retry", spec); err != nil {
		t.Fatalf("reservation was not released: %v", err)
	}
}

func TestWorldManagerUnloadReleasesProviderPathAfterClose(t *testing.T) {
	manager, _ := persistentTestManager(t)
	spec := validWorldSpec("reusable")
	first, err := manager.OpenSpec("example:first", spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Unload("example:first"); err != nil {
		t.Fatal(err)
	}
	second, err := manager.OpenSpec("example:second", spec)
	if err != nil {
		t.Fatalf("provider path was not released: %v", err)
	}
	if second == first {
		t.Fatalf("world handle was reused: %d", second)
	}
}

func TestWorldManagerCloseCustomWaitsForOpeningsAndRejectsNewOpens(t *testing.T) {
	manager, _ := persistentTestManager(t)
	opening := &worldOpening{done: make(chan struct{})}
	manager.mu.Lock()
	manager.openings["example:opening"] = opening
	manager.providerPaths["synthetic"] = "example:opening"
	manager.mu.Unlock()

	closed := make(chan error, 1)
	go func() { closed <- manager.CloseCustom() }()
	deadline := time.Now().Add(2 * time.Second)
	for {
		manager.mu.RLock()
		closing := manager.closing
		manager.mu.RUnlock()
		if closing {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("CloseCustom did not enter closing state")
		}
		time.Sleep(time.Millisecond)
	}
	if _, err := manager.OpenSpec("example:late", validWorldSpec("late")); err == nil {
		t.Fatal("open accepted after shutdown started")
	}
	select {
	case err := <-closed:
		t.Fatalf("CloseCustom returned before opening completed: %v", err)
	default:
	}
	manager.mu.Lock()
	delete(manager.openings, "example:opening")
	delete(manager.providerPaths, "synthetic")
	close(opening.done)
	manager.mu.Unlock()
	if err := <-closed; err != nil {
		t.Fatal(err)
	}
}

func TestWorldManagerRetainsProviderPathWhenCloseFails(t *testing.T) {
	manager := NewWorldManager()
	spec := normalizedWorldSpec{absoluteProviderPath: "/reserved"}
	entry := &managedWorld{id: 1, name: "example:failing", spec: &spec}
	manager.worlds[entry.name] = entry
	manager.handles[entry.id] = entry
	manager.providerPaths[spec.absoluteProviderPath] = entry.name
	want := errors.New("close failed")
	if err := manager.finishWorldClose(entry, func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("close error = %v, want %v", err, want)
	}
	if owner, ok := manager.providerPaths[spec.absoluteProviderPath]; !ok || owner != entry.name {
		t.Fatalf("provider reservation = %q, %v", owner, ok)
	}
}

type retryCloseProvider struct {
	world.NopProvider
	calls int
}

func (p *retryCloseProvider) Close() error {
	p.calls++
	if p.calls == 1 {
		return errors.New("close failed")
	}
	return nil
}

func TestWorldManagerRetriesFailedProviderClose(t *testing.T) {
	manager := NewWorldManager()
	spec := normalizedWorldSpec{absoluteProviderPath: "/reserved"}
	provider := &retryCloseProvider{}
	entry := &managedWorld{
		id: 1, name: "example:failing", spec: &spec,
		provider: &recordingProvider{Provider: provider}, unloading: true,
	}
	manager.worlds[entry.name] = entry
	manager.handles[entry.id] = entry
	manager.providerPaths[spec.absoluteProviderPath] = entry.name
	if err := manager.finishWorldClose(entry, entry.provider.Close); err == nil {
		t.Fatal("first provider close unexpectedly succeeded")
	}
	if entry.unloading {
		t.Fatal("failed close cannot be retried")
	}
	if err := manager.Unload(entry.name); err != nil {
		t.Fatalf("retry provider close: %v", err)
	}
	if provider.calls != 2 {
		t.Fatalf("provider close calls = %d, want 2", provider.calls)
	}
	if _, ok := manager.providerPaths[spec.absoluteProviderPath]; ok {
		t.Fatal("successful retry retained provider reservation")
	}
	if _, ok := manager.handles[entry.id]; ok {
		t.Fatal("successful retry retained world handle")
	}
}

func TestWorldManagerCloseCustomRetriesFailedProviderClose(t *testing.T) {
	manager := NewWorldManager()
	spec := normalizedWorldSpec{absoluteProviderPath: "/reserved"}
	provider := &retryCloseProvider{}
	entry := &managedWorld{
		id: 1, name: "example:failing", spec: &spec, closed: true,
		provider: &recordingProvider{Provider: provider},
	}
	manager.worlds[entry.name] = entry
	manager.handles[entry.id] = entry
	manager.providerPaths[spec.absoluteProviderPath] = entry.name
	if err := manager.CloseCustom(); err == nil {
		t.Fatal("first close unexpectedly succeeded")
	}
	if err := manager.CloseCustom(); err != nil {
		t.Fatalf("retry CloseCustom: %v", err)
	}
	if provider.calls != 2 {
		t.Fatalf("provider close calls = %d, want 2", provider.calls)
	}
	if _, ok := manager.providerPaths[spec.absoluteProviderPath]; ok {
		t.Fatal("successful retry retained provider reservation")
	}
}

func createMCDBFixture(t *testing.T, path string) {
	t.Helper()
	provider, err := (mcdb.Config{}).Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
}
