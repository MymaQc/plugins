package host

import (
	"context"
	"math"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bedrock-gophers/plugins/internal/native"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Players owns stable native IDs for the lifetime of connected Dragonfly players.
type Players struct {
	mu             sync.RWMutex
	entries        map[*world.EntityHandle]*playerEntry
	byID           map[native.PlayerID]*playerEntry
	entities       *Entities
	departedWorlds map[*world.EntityHandle]native.WorldID
	invocations    map[native.InvocationID]*world.Tx
	invocationEnds map[native.InvocationID][]func()
	nextInvocation native.InvocationID
}

type playerEntry struct {
	id       native.PlayerID
	handle   *world.EntityHandle
	name     string
	xuid     string
	latency  uint64
	position mgl64.Vec3
}

func NewPlayers() *Players {
	return &Players{
		entries:        map[*world.EntityHandle]*playerEntry{},
		byID:           map[native.PlayerID]*playerEntry{},
		entities:       NewEntities(),
		departedWorlds: map[*world.EntityHandle]native.WorldID{},
		invocations:    map[native.InvocationID]*world.Tx{},
		invocationEnds: map[native.InvocationID][]func(){},
	}
}

func (p *Players) Register(player *player.Player, generation uint64) native.PlayerID {
	entityID := p.entities.registerHandle(player.H(), generation)
	id := native.PlayerID{UUID: entityID.UUID, Generation: entityID.Generation}
	entry := &playerEntry{
		id: id, handle: player.H(), name: player.Name(), xuid: player.XUID(),
		latency:  uint64(max(player.Latency().Milliseconds(), 0)),
		position: player.Position(),
	}
	p.mu.Lock()
	p.entries[player.H()] = entry
	p.byID[id] = entry
	p.mu.Unlock()
	return id
}

// registerTemporary exposes a player through the synchronous callback that
// first observes its entity, without publishing it as an accepted player.
// cleanup leaves the shared entity identity intact for the later Register.
func (p *Players) registerTemporary(connected *player.Player, id native.EntityID) (cleanup func()) {
	if p == nil || connected == nil || id.Generation == 0 {
		return func() {}
	}
	entityID := p.entities.registerHandle(connected.H(), id.Generation)
	if entityID != id {
		return func() {}
	}
	playerID := native.PlayerID{UUID: id.UUID, Generation: id.Generation}
	entry := &playerEntry{
		id: playerID, handle: connected.H(), name: connected.Name(), xuid: connected.XUID(),
		latency: uint64(max(connected.Latency().Milliseconds(), 0)), position: connected.Position(),
	}
	p.mu.Lock()
	if _, exists := p.entries[connected.H()]; exists {
		p.mu.Unlock()
		return func() {}
	}
	p.entries[connected.H()] = entry
	p.byID[playerID] = entry
	p.mu.Unlock()
	return func() {
		p.mu.Lock()
		if p.entries[connected.H()] == entry {
			delete(p.entries, connected.H())
			if p.byID[playerID] == entry {
				delete(p.byID, playerID)
			}
		}
		p.mu.Unlock()
	}
}

func (p *Players) Unregister(player *player.Player) {
	if player != nil {
		p.UnregisterHandle(player.H())
	}
}

// UnregisterHandle expires the player entry for an exact handle. It is used
// when a transfer loses both worlds and no transaction-scoped Player remains.
func (p *Players) UnregisterHandle(handle *world.EntityHandle) {
	if handle == nil {
		return
	}
	p.mu.Lock()
	var id native.PlayerID
	var registered bool
	if entry, ok := p.entries[handle]; ok {
		id, registered = entry.id, true
		delete(p.byID, entry.id)
		delete(p.entries, handle)
	}
	delete(p.departedWorlds, handle)
	p.mu.Unlock()
	if registered {
		p.entities.unregisterHandle(handle)
		native.CancelPlayerForms(id)
	}
}

// recordWorldDeparture retains the source world handle until Dragonfly emits
// HandleChangeWorld on the player's first tick in the destination world.
func (p *Players) recordWorldDeparture(connected *player.Player, id native.WorldID) {
	if p == nil || connected == nil || id == 0 {
		return
	}
	p.mu.Lock()
	p.departedWorlds[connected.H()] = id
	p.mu.Unlock()
}

func (p *Players) takeWorldDeparture(connected *player.Player) (native.WorldID, bool) {
	if p == nil || connected == nil {
		return 0, false
	}
	p.mu.Lock()
	id, ok := p.departedWorlds[connected.H()]
	delete(p.departedWorlds, connected.H())
	p.mu.Unlock()
	return id, ok
}

func (p *Players) forgetWorldDeparture(handle *world.EntityHandle) {
	if p == nil || handle == nil {
		return
	}
	p.mu.Lock()
	delete(p.departedWorlds, handle)
	p.mu.Unlock()
}

// ForgetWorldDeparture clears transfer bookkeeping after a failed handoff is
// restored to its source world.
func (p *Players) ForgetWorldDeparture(handle *world.EntityHandle) {
	p.forgetWorldDeparture(handle)
}

// EntityRegistry returns the generic entity registry shared by player IDs.
func (p *Players) EntityRegistry() *Entities { return p.entities }

func (p *Players) SendPlayerForm(invocation native.InvocationID, id native.PlayerID, value native.PlayerForm) bool {
	f := &nativePlayerForm{id: value.ID, request: append([]byte(nil), value.RequestJSON...), players: p}
	if !p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.SendForm(f) }) {
		return false
	}
	// Arm the finalizer only after Dragonfly accepted the form. A synchronous
	// send failure is terminated by the native host and must not race a second
	// drop callback from an unreachable wrapper.
	runtime.SetFinalizer(f, func(form *nativePlayerForm) { native.CancelPlayerForm(form.id) })
	return true
}

func (p *Players) ClosePlayerForm(invocation native.InvocationID, id native.PlayerID) bool {
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.CloseForm() })
}

func (p *Players) ID(player *player.Player) (native.PlayerID, bool) {
	p.mu.RLock()
	entry, ok := p.entries[player.H()]
	p.mu.RUnlock()
	if !ok {
		return native.PlayerID{}, false
	}
	p.mu.Lock()
	entry.name, entry.latency, entry.position = player.Name(), uint64(max(player.Latency().Milliseconds(), 0)), player.Position()
	p.mu.Unlock()
	return entry.id, true
}

func (p *Players) Names() []string {
	p.mu.RLock()
	names := make([]string, 0, len(p.entries))
	for _, entry := range p.entries {
		names = append(names, entry.name)
	}
	p.mu.RUnlock()
	slices.Sort(names)
	return names
}

func (p *Players) CommandSnapshots() []native.CommandPlayer {
	type cachedPlayer struct {
		id       native.PlayerID
		name     string
		latency  uint64
		position mgl64.Vec3
	}
	p.mu.RLock()
	entries := make([]cachedPlayer, 0, len(p.entries))
	for _, entry := range p.entries {
		entries = append(entries, cachedPlayer{id: entry.id, name: entry.name, latency: entry.latency, position: entry.position})
	}
	p.mu.RUnlock()
	snapshots := make([]native.CommandPlayer, 0, len(entries))
	for _, entry := range entries {
		name, latency := entry.name, entry.latency
		snapshots = append(snapshots, native.CommandPlayer{
			Player: entry.id, Name: name, LatencyMilliseconds: latency,
			Position: native.Vec3{X: entry.position[0], Y: entry.position[1], Z: entry.position[2]},
		})
	}
	slices.SortFunc(snapshots, func(left, right native.CommandPlayer) int {
		return strings.Compare(left.Name, right.Name)
	})
	return snapshots
}

func (p *Players) ResolveName(name string) (native.PlayerID, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, entry := range p.entries {
		if strings.EqualFold(entry.name, name) {
			return entry.id, true
		}
	}
	return native.PlayerID{}, false
}

func (p *Players) ResolveUUID(uuid [16]byte) (native.PlayerID, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, entry := range p.entries {
		if entry.id.UUID == uuid {
			return entry.id, true
		}
	}
	return native.PlayerID{}, false
}

// PlayerXUID returns the exact Dragonfly XUID for a player borrowed by a live
// invocation. XUID is immutable for the registered player lifetime.
func (p *Players) PlayerXUID(invocation native.InvocationID, id native.PlayerID) (string, bool) {
	p.mu.RLock()
	entry, registered := p.byID[id]
	_, invoked := p.invocations[invocation]
	p.mu.RUnlock()
	if invocation == 0 || !registered || !invoked {
		return "", false
	}
	return entry.xuid, true
}

// Handle returns the exact live Dragonfly handle registered for id. The full
// player generation is checked so a stale ID cannot address a later session.
func (p *Players) Handle(id native.PlayerID) (*world.EntityHandle, bool) {
	p.mu.RLock()
	entry, ok := p.byID[id]
	p.mu.RUnlock()
	if !ok || entry.handle == nil || entry.handle.Closed() {
		return nil, false
	}
	return entry.handle, true
}

func (p *Players) ResolveID(id native.PlayerID, invocation native.InvocationID) (*player.Player, bool) {
	p.mu.RLock()
	entry, ok := p.byID[id]
	tx, invoked := p.invocations[invocation]
	p.mu.RUnlock()
	if !ok || !invoked || invocation == 0 {
		return nil, false
	}
	return playerInTransaction(entry.handle, tx)
}

func (p *Players) ResolveEntityID(id native.EntityID, invocation native.InvocationID) (world.Entity, bool) {
	tx, ok := p.InvocationTx(invocation)
	if !ok {
		return nil, false
	}
	return p.entities.Resolve(id, tx)
}

// BeginInvocation registers tx for one synchronous native invocation. The returned end function is idempotent.
func (p *Players) BeginInvocation(tx *world.Tx) (native.InvocationID, func()) {
	if tx == nil {
		return 0, func() {}
	}
	p.mu.Lock()
	if p.nextInvocation == native.InvocationID(^uint64(0)) {
		p.mu.Unlock()
		return 0, func() {}
	}
	p.nextInvocation++
	id := p.nextInvocation
	p.invocations[id] = tx
	p.mu.Unlock()
	var once sync.Once
	return id, func() {
		once.Do(func() { p.EndInvocation(id) })
	}
}

// WithInvocation registers tx for the duration of function.
func (p *Players) WithInvocation(tx *world.Tx, function func(native.InvocationID)) {
	id, end := p.BeginInvocation(tx)
	defer end()
	function(id)
}

// InvocationTx resolves a live invocation to its exact owner transaction.
func (p *Players) InvocationTx(id native.InvocationID) (*world.Tx, bool) {
	if id == 0 {
		return nil, false
	}
	p.mu.RLock()
	tx, ok := p.invocations[id]
	p.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if _, ok := txWorld(tx); !ok {
		return nil, false
	}
	return tx, true
}

// EndInvocation removes id. It is safe to call for an unknown or already-ended invocation.
func (p *Players) EndInvocation(id native.InvocationID) {
	if id == 0 {
		return
	}
	p.mu.Lock()
	delete(p.invocations, id)
	callbacks := p.invocationEnds[id]
	delete(p.invocationEnds, id)
	p.mu.Unlock()
	for _, callback := range callbacks {
		callback()
	}
}

// OnInvocationEnd registers cleanup that runs exactly once when id ends.
func (p *Players) OnInvocationEnd(id native.InvocationID, callback func()) bool {
	if id == 0 || callback == nil {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.invocations[id]; !ok {
		return false
	}
	p.invocationEnds[id] = append(p.invocationEnds[id], callback)
	return true
}

func txWorld(tx *world.Tx) (value *world.World, ok bool) {
	if tx == nil {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			value, ok = nil, false
		}
	}()
	return tx.World(), true
}

func playerInTransaction(handle *world.EntityHandle, tx *world.Tx) (connected *player.Player, ok bool) {
	if handle == nil || tx == nil {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			connected, ok = nil, false
		}
	}()
	entity, ok := handle.Entity(tx)
	if !ok {
		return nil, false
	}
	connected, ok = entity.(*player.Player)
	return connected, ok
}

func (p *Players) playerEntry(id native.PlayerID) (*playerEntry, bool) {
	p.mu.RLock()
	entry, ok := p.byID[id]
	p.mu.RUnlock()
	return entry, ok
}

func (p *Players) mutatePlayer(invocation native.InvocationID, id native.PlayerID, function func(*player.Player)) bool {
	if connected, ok := p.ResolveID(id, invocation); ok {
		function(connected)
		return true
	}
	if invocation != 0 {
		if _, ok := p.InvocationTx(invocation); !ok {
			return false
		}
	}
	entry, ok := p.playerEntry(id)
	if !ok {
		return false
	}
	task := world.NewEntityRef[*player.Player](entry.handle).Do(func(_ *world.Tx, connected *player.Player) {
		function(connected)
	})
	return task.Err() == nil
}

func readPlayer[T any](p *Players, invocation native.InvocationID, id native.PlayerID, function func(*player.Player) T) (T, bool) {
	return callPlayer(p, invocation, id, func(_ *world.Tx, connected *player.Player) T { return function(connected) })
}

func callPlayer[T any](p *Players, invocation native.InvocationID, id native.PlayerID, function func(*world.Tx, *player.Player) T) (T, bool) {
	var zero T
	if invocation != 0 {
		tx, ok := p.InvocationTx(invocation)
		if !ok {
			return zero, false
		}
		entry, ok := p.playerEntry(id)
		if !ok {
			return zero, false
		}
		connected, ok := playerInTransaction(entry.handle, tx)
		if !ok {
			return zero, false
		}
		return function(tx, connected), true
	}
	entry, ok := p.playerEntry(id)
	if !ok {
		return zero, false
	}
	value, err := world.CallRef(context.Background(), world.NewEntityRef[*player.Player](entry.handle), func(tx *world.Tx, connected *player.Player) (T, error) {
		return function(tx, connected), nil
	})
	return value, err == nil
}

func (p *Players) SendPlayerText(invocation native.InvocationID, id native.PlayerID, kind native.PlayerTextKind, message string) bool {
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { sendPlayerText(connected, kind, message) })
}

func (p *Players) SendPlayerTitle(invocation native.InvocationID, id native.PlayerID, value native.PlayerTitle) bool {
	t := title.New(value.Text).
		WithSubtitle(value.Subtitle).
		WithActionText(value.ActionText).
		WithFadeInDuration(value.FadeIn).
		WithDuration(value.Duration).
		WithFadeOutDuration(value.FadeOut)
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.SendTitle(t) })
}

func (p *Players) SendPlayerScoreboard(invocation native.InvocationID, id native.PlayerID, value native.PlayerScoreboard) bool {
	if len(value.Lines) > 15 {
		return false
	}
	board := scoreboard.New(value.Name)
	for index, line := range value.Lines {
		board.Set(index, line)
	}
	if !value.Padding {
		board.RemovePadding()
	}
	if value.Descending {
		board.SetDescending()
	}
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.SendScoreboard(board) })
}

func (p *Players) RemovePlayerScoreboard(invocation native.InvocationID, id native.PlayerID) bool {
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.RemoveScoreboard() })
}

func (p *Players) TransformPlayer(invocation native.InvocationID, id native.PlayerID, kind native.PlayerTransformKind, vector native.Vec3, yaw, pitch float64) bool {
	if kind > native.PlayerTransformDisplace {
		return false
	}
	v := mgl64.Vec3{vector.X, vector.Y, vector.Z}
	return p.mutatePlayer(invocation, id, func(connected *player.Player) {
		switch kind {
		case native.PlayerTransformTeleport:
			connected.Teleport(v)
		case native.PlayerTransformMove:
			connected.Move(v, yaw, pitch)
		case native.PlayerTransformVelocity:
			connected.SetVelocity(v)
		case native.PlayerTransformDisplace:
			connected.Displace(v)
		}
	})
}

func (p *Players) PlayerKinematics(invocation native.InvocationID, id native.PlayerID) (native.PlayerKinematics, bool) {
	return readPlayer(p, invocation, id, func(connected *player.Player) native.PlayerKinematics {
		position, velocity, rotation := connected.Position(), connected.Velocity(), connected.Rotation()
		return native.PlayerKinematics{
			Position: native.Vec3{X: position.X(), Y: position.Y(), Z: position.Z()},
			Velocity: native.Vec3{X: velocity.X(), Y: velocity.Y(), Z: velocity.Z()},
			Rotation: native.Rotation{Yaw: rotation.Yaw(), Pitch: rotation.Pitch()},
		}
	})
}

func (p *Players) KnockBackPlayer(invocation native.InvocationID, id native.PlayerID, source native.Vec3, force, height float64) bool {
	src := mgl64.Vec3{source.X, source.Y, source.Z}
	return p.mutatePlayer(invocation, id, func(connected *player.Player) {
		connected.KnockBack(src, force, height)
	})
}

// EntityPlayer materializes an exact registered player entity as a fresh
// transaction-safe player snapshot.
func (p *Players) EntityPlayer(invocation native.InvocationID, id native.EntityID) (native.PlayerSnapshot, bool) {
	playerID := native.PlayerID{UUID: id.UUID, Generation: id.Generation}
	return readPlayer(p, invocation, playerID, func(connected *player.Player) native.PlayerSnapshot {
		position := connected.Position()
		return native.PlayerSnapshot{
			Player:              playerID,
			Name:                connected.Name(),
			LatencyMilliseconds: uint64(max(connected.Latency().Milliseconds(), 0)),
			Position:            native.Vec3{X: position.X(), Y: position.Y(), Z: position.Z()},
		}
	})
}

func (p *Players) SetPlayerState(invocation native.InvocationID, id native.PlayerID, kind native.PlayerStateKind, value native.PlayerStateValue) bool {
	changed := false
	ok := p.mutatePlayer(invocation, id, func(connected *player.Player) {
		changed = setPlayerState(connected, kind, value)
	})
	return ok && changed
}

func (p *Players) HealPlayer(invocation native.InvocationID, id native.PlayerID, health float64, source native.HealingSource) (float64, bool) {
	return readPlayer(p, invocation, id, func(connected *player.Player) float64 {
		return connected.Heal(health, healingSource(source))
	})
}

func (p *Players) HurtPlayer(invocation native.InvocationID, id native.PlayerID, damage float64, source native.DamageSource) (native.PlayerHurtResult, bool) {
	if !finite(damage) {
		return native.PlayerHurtResult{}, false
	}
	type result struct {
		value native.PlayerHurtResult
		ok    bool
	}
	resolved, ok := callPlayer(p, invocation, id, func(tx *world.Tx, connected *player.Player) result {
		damageSource, ok := p.damageSource(tx, source)
		if !ok {
			return result{}
		}
		dealt, vulnerable := connected.Hurt(damage, damageSource)
		return result{value: native.PlayerHurtResult{Damage: dealt, Vulnerable: vulnerable}, ok: true}
	})
	return resolved.value, ok && resolved.ok
}

func (p *Players) FinalPlayerDamage(invocation native.InvocationID, id native.PlayerID, damage float64, source native.DamageSource) (float64, bool) {
	type result struct {
		value float64
		ok    bool
	}
	resolved, ok := callPlayer(p, invocation, id, func(tx *world.Tx, connected *player.Player) result {
		damageSource, ok := p.damageSource(tx, source)
		if !ok {
			return result{}
		}
		return result{value: connected.FinalDamageFrom(damage, damageSource), ok: true}
	})
	return resolved.value, ok && resolved.ok
}

func (p *Players) PlayerUsingItem(invocation native.InvocationID, id native.PlayerID) (bool, bool) {
	return readPlayer(p, invocation, id, (*player.Player).UsingItem)
}

func (p *Players) PlayerSleeping(invocation native.InvocationID, id native.PlayerID) (native.BlockPos, bool, bool) {
	type result struct {
		position native.BlockPos
		sleeping bool
	}
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) result {
		position, sleeping := connected.Sleeping()
		return result{
			position: native.BlockPos{X: int32(position.X()), Y: int32(position.Y()), Z: int32(position.Z())},
			sleeping: sleeping,
		}
	})
	return value.position, value.sleeping, ok
}

func (p *Players) PlayerDeathPosition(invocation native.InvocationID, id native.PlayerID) (native.Vec3, native.WorldDimensionView, bool, bool) {
	type result struct {
		position  native.Vec3
		dimension native.WorldDimensionView
		found     bool
		valid     bool
	}
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) result {
		position, dimension, found := connected.DeathPosition()
		if !found {
			return result{valid: true}
		}
		view, valid := WorldDimensionView(dimension)
		return result{
			position:  native.Vec3{X: position.X(), Y: position.Y(), Z: position.Z()},
			dimension: view, found: true, valid: valid,
		}
	})
	return value.position, value.dimension, value.found, ok && value.valid
}

func (p *Players) PlayerState(invocation native.InvocationID, id native.PlayerID, kind native.PlayerStateKind) (native.PlayerStateValue, bool) {
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) struct {
		value native.PlayerStateValue
		ok    bool
	} {
		value, ok := readPlayerState(connected, kind)
		return struct {
			value native.PlayerStateValue
			ok    bool
		}{value, ok}
	})
	return value.value, ok && value.ok
}

func (p *Players) PlayerAction(invocation native.InvocationID, id native.PlayerID, kind native.PlayerActionKind, input native.PlayerStateValue) (native.PlayerStateValue, bool) {
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) struct {
		value native.PlayerStateValue
		ok    bool
	} {
		value, ok := runPlayerAction(connected, kind, input)
		return struct {
			value native.PlayerStateValue
			ok    bool
		}{value, ok}
	})
	return value.value, ok && value.ok
}

func (p *Players) PlayerBlockAction(invocation native.InvocationID, id native.PlayerID, kind native.PlayerBlockActionKind, position native.BlockPos, face int32, clickPosition native.Vec3) bool {
	valid, ok := readPlayer(p, invocation, id, func(connected *player.Player) bool {
		return runPlayerBlockAction(connected, kind, position, face, clickPosition)
	})
	return ok && valid
}

func (p *Players) PlayerString(invocation native.InvocationID, id native.PlayerID, kind native.PlayerStringKind) (string, bool) {
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) struct {
		value string
		ok    bool
	} {
		value, ok := playerString(connected, kind)
		return struct {
			value string
			ok    bool
		}{value, ok}
	})
	return value.value, ok && value.ok
}

func (p *Players) SendPlayerToast(invocation native.InvocationID, id native.PlayerID, title, message string) bool {
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.SendToast(title, message) })
}

func (p *Players) ChangePlayerEffect(invocation native.InvocationID, id native.PlayerID, operation native.PlayerEffectOperation, value native.PlayerEffect) bool {
	effectType, ok := effect.ByID(int(value.Type))
	if !ok {
		return false
	}
	if operation == native.PlayerEffectRemove {
		return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.RemoveEffect(effectType) })
	}
	if operation != native.PlayerEffectAdd || value.Level <= 0 || value.Duration < 0 {
		return false
	}
	applied, ok := playerEffectFromNative(effectType, value)
	if !ok {
		return false
	}
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.AddEffect(applied) })
}

const maximumPlayerEffectTick = 1_000_000

func playerEffectFromNative(effectType effect.Type, value native.PlayerEffect) (effect.Effect, bool) {
	if value.Tick < 0 || value.Tick > maximumPlayerEffectTick || value.Ambient && value.Infinite {
		return effect.Effect{}, false
	}
	lasting, isLasting := effectType.(effect.LastingType)
	if !isLasting {
		if value.Duration != 0 || value.Ambient || value.Infinite || value.Tick != 0 {
			return effect.Effect{}, false
		}
		reconstructed := effect.NewInstantWithPotency(effectType, int(value.Level), value.Potency)
		if value.ParticlesHidden {
			reconstructed = reconstructed.WithoutParticles()
		}
		return reconstructed, true
	}
	duration := value.Duration
	if value.Infinite {
		if duration != 0 {
			return effect.Effect{}, false
		}
	} else {
		tickDuration := time.Duration(value.Tick) * (time.Second / 20)
		if duration > time.Duration(math.MaxInt64)-tickDuration {
			return effect.Effect{}, false
		}
		duration += tickDuration
	}
	var reconstructed effect.Effect
	switch {
	case value.Infinite:
		reconstructed = effect.NewInfinite(lasting, int(value.Level))
	case value.Ambient:
		reconstructed = effect.NewAmbient(lasting, int(value.Level), duration)
	default:
		reconstructed = effect.New(lasting, int(value.Level), duration)
	}
	for range value.Tick {
		reconstructed = reconstructed.TickDuration()
	}
	if value.ParticlesHidden {
		reconstructed = reconstructed.WithoutParticles()
	}
	return reconstructed, true
}

func (p *Players) PlayerEffects(invocation native.InvocationID, id native.PlayerID) ([]native.PlayerEffect, bool) {
	type result struct {
		values []native.PlayerEffect
		ok     bool
	}
	resolved, ok := readPlayer(p, invocation, id, func(connected *player.Player) result {
		active := connected.Effects()
		values := make([]native.PlayerEffect, 0, len(active))
		for _, current := range active {
			value, valid := snapshotPlayerEffect(current)
			if !valid {
				return result{}
			}
			values = append(values, value)
		}
		return result{values: values, ok: true}
	})
	return resolved.values, ok && resolved.ok
}

func snapshotPlayerEffect(current effect.Effect) (native.PlayerEffect, bool) {
	effectType := current.Type()
	level := current.Level()
	if effectType == nil || level <= 0 || int64(level) > math.MaxInt32 {
		return native.PlayerEffect{}, false
	}
	id, ok := effect.ID(effectType)
	if !ok || int64(id) < math.MinInt32 || int64(id) > math.MaxInt32 {
		return native.PlayerEffect{}, false
	}
	tick := current.Tick()
	if tick < 0 {
		return native.PlayerEffect{}, false
	}
	if current.Infinite() {
		if current.Duration() != 0 {
			return native.PlayerEffect{}, false
		}
	}
	if current.Ambient() && current.Infinite() {
		return native.PlayerEffect{}, false
	}
	potency := float64(1)
	if _, lasting := effectType.(effect.LastingType); lasting {
		potency = 0
	}
	return native.PlayerEffect{
		Type: native.EffectType(id), Level: int32(level), Duration: current.Duration(), Potency: potency,
		Ambient: current.Ambient(), ParticlesHidden: current.ParticlesHidden(), Infinite: current.Infinite(), Tick: int64(tick),
	}, true
}

func (p *Players) ClearPlayerEffects(invocation native.InvocationID, id native.PlayerID) bool {
	cleared, ok := readPlayer(p, invocation, id, func(connected *player.Player) bool {
		types := make([]effect.Type, 0, len(connected.Effects()))
		for _, current := range connected.Effects() {
			// Dragonfly flushes every pending initial effect on the first RemoveEffect.
			// Validate even non-lasting entries before that flush can call EffectManager.Add.
			if current.Level() <= 0 || current.Duration() < 0 {
				return false
			}
			if _, lasting := current.Type().(effect.LastingType); !lasting {
				continue
			}
			if _, valid := snapshotPlayerEffect(current); !valid {
				return false
			}
			types = append(types, current.Type())
		}
		for _, effectType := range types {
			connected.RemoveEffect(effectType)
		}
		return true
	})
	return ok && cleared
}

func (p *Players) SetPlayerEntityVisible(invocation native.InvocationID, viewerID native.PlayerID, entityID native.EntityID, visible bool) bool {
	return p.runPlayerEntityAction(invocation, viewerID, entityID, func(viewer *player.Player, entity world.Entity) bool {
		setPlayerEntityVisible(viewer, entity, visible)
		return true
	})
}

func (p *Players) PlayerViewLayer(invocation native.InvocationID, viewerID native.PlayerID, entityID native.EntityID, kind native.PlayerViewLayerKind, text string, visibility uint8) bool {
	return p.runPlayerEntityAction(invocation, viewerID, entityID, func(viewer *player.Player, entity world.Entity) bool {
		return runPlayerViewLayer(viewer, entity, kind, text, visibility)
	})
}

func (p *Players) PlayerEntityAction(invocation native.InvocationID, playerID native.PlayerID, entityID native.EntityID, kind native.PlayerEntityActionKind) (bool, bool) {
	return p.runPlayerEntityResult(invocation, playerID, entityID, func(connected *player.Player, entity world.Entity) (bool, bool) {
		return runExactPlayerEntityAction(connected, entity, kind)
	})
}

func (p *Players) runPlayerEntityAction(invocation native.InvocationID, viewerID native.PlayerID, entityID native.EntityID, action func(*player.Player, world.Entity) bool) bool {
	viewer, ok := p.ResolveID(viewerID, invocation)
	if ok {
		entity, ok := p.ResolveEntityID(entityID, invocation)
		if !ok {
			return false
		}
		return action(viewer, entity)
	}
	if invocation != 0 {
		if _, ok := p.InvocationTx(invocation); !ok {
			return false
		}
	}
	viewerEntry, ok := p.playerEntry(viewerID)
	if !ok {
		return false
	}
	entityHandle, ok := p.entities.Handle(entityID)
	if !ok {
		return false
	}
	accepted := false
	task := world.NewEntityRef[*player.Player](viewerEntry.handle).Do(func(tx *world.Tx, connected *player.Player) {
		entity, ok := entityHandle.Entity(tx)
		if ok {
			accepted = action(connected, entity)
		}
	})
	return task.Err() == nil && accepted
}

func (p *Players) runPlayerEntityResult(invocation native.InvocationID, playerID native.PlayerID, entityID native.EntityID, action func(*player.Player, world.Entity) (bool, bool)) (bool, bool) {
	connected, ok := p.ResolveID(playerID, invocation)
	if ok {
		entity, ok := p.ResolveEntityID(entityID, invocation)
		if !ok {
			return false, false
		}
		return action(connected, entity)
	}
	if invocation != 0 {
		if _, ok := p.InvocationTx(invocation); !ok {
			return false, false
		}
	}
	playerEntry, ok := p.playerEntry(playerID)
	if !ok {
		return false, false
	}
	entityHandle, ok := p.entities.Handle(entityID)
	if !ok {
		return false, false
	}
	var result, completed bool
	task := world.NewEntityRef[*player.Player](playerEntry.handle).Do(func(tx *world.Tx, connected *player.Player) {
		entity, ok := entityHandle.Entity(tx)
		if ok {
			result, completed = action(connected, entity)
		}
	})
	return result, task.Err() == nil && completed
}

func setPlayerEntityVisible(viewer *player.Player, entity world.Entity, visible bool) {
	if visible {
		viewer.ShowEntity(entity)
	} else {
		viewer.HideEntity(entity)
	}
}

const (
	maxSkinDimension  = 4096
	maxSkinAnimations = 64
	maxSkinDataBytes  = 64 << 20
	maxSkinIDBytes    = 4096
)

func (p *Players) PlayerSkin(invocation native.InvocationID, id native.PlayerID) (native.PlayerSkin, bool) {
	value, ok := readPlayer(p, invocation, id, func(connected *player.Player) struct {
		skin native.PlayerSkin
		ok   bool
	} {
		value, ok := playerSkinToNative(connected.Skin())
		return struct {
			skin native.PlayerSkin
			ok   bool
		}{value, ok}
	})
	return value.skin, ok && value.ok
}

func (p *Players) SetPlayerSkin(invocation native.InvocationID, id native.PlayerID, value native.PlayerSkin) bool {
	converted, ok := playerSkinFromNative(value)
	if !ok {
		return false
	}
	return p.mutatePlayer(invocation, id, func(connected *player.Player) { connected.SetSkin(converted) })
}

func playerSkinToNative(value skin.Skin) (native.PlayerSkin, bool) {
	bounds := value.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	capeBounds := value.Cape.Bounds()
	if !validSkinDimensions(width, height, value.Pix, true) ||
		!validSkinDimensions(capeBounds.Dx(), capeBounds.Dy(), value.Cape.Pix, true) ||
		len(value.Animations) > maxSkinAnimations ||
		!validSkinStrings(value.PlayFabID, value.FullID, value.ModelConfig.Default, value.ModelConfig.AnimatedFace) {
		return native.PlayerSkin{}, false
	}
	total := uint64(len(value.Pix)) + uint64(len(value.Model)) + uint64(len(value.Cape.Pix)) +
		uint64(len(value.PlayFabID)) + uint64(len(value.FullID)) +
		uint64(len(value.ModelConfig.Default)) + uint64(len(value.ModelConfig.AnimatedFace))
	for _, animation := range value.Animations {
		animationBounds := animation.Bounds()
		if !validSkinDimensions(animationBounds.Dx(), animationBounds.Dy(), animation.Pix, false) ||
			animation.FrameCount <= 0 ||
			animation.Type() < skin.AnimationHead || animation.Type() > skin.AnimationBody128x128 {
			return native.PlayerSkin{}, false
		}
		total += uint64(len(animation.Pix))
		if total > maxSkinDataBytes {
			return native.PlayerSkin{}, false
		}
	}
	if total > maxSkinDataBytes {
		return native.PlayerSkin{}, false
	}
	animations := make([]native.SkinAnimation, len(value.Animations))
	for i, animation := range value.Animations {
		animationBounds := animation.Bounds()
		animations[i] = native.SkinAnimation{
			Width:      uint32(animationBounds.Dx()),
			Height:     uint32(animationBounds.Dy()),
			Type:       uint32(animation.Type()),
			FrameCount: int64(animation.FrameCount),
			Expression: int64(animation.AnimationExpression),
			Pixels:     slices.Clone(animation.Pix),
		}
	}
	return native.PlayerSkin{
		Width:             uint32(width),
		Height:            uint32(height),
		Persona:           value.Persona,
		PlayFabID:         value.PlayFabID,
		FullID:            value.FullID,
		Pixels:            slices.Clone(value.Pix),
		ModelDefault:      value.ModelConfig.Default,
		ModelAnimatedFace: value.ModelConfig.AnimatedFace,
		Model:             slices.Clone(value.Model),
		CapeWidth:         uint32(capeBounds.Dx()),
		CapeHeight:        uint32(capeBounds.Dy()),
		CapePixels:        slices.Clone(value.Cape.Pix),
		Animations:        animations,
	}, true
}

func playerSkinFromNative(value native.PlayerSkin) (skin.Skin, bool) {
	if !validNativeSkinDimensions(value.Width, value.Height, value.Pixels, true) ||
		!validNativeSkinDimensions(value.CapeWidth, value.CapeHeight, value.CapePixels, true) ||
		len(value.Animations) > maxSkinAnimations ||
		!validSkinStrings(value.PlayFabID, value.FullID, value.ModelDefault, value.ModelAnimatedFace) {
		return skin.Skin{}, false
	}
	total := uint64(len(value.Pixels)) + uint64(len(value.Model)) + uint64(len(value.CapePixels)) +
		uint64(len(value.PlayFabID)) + uint64(len(value.FullID)) +
		uint64(len(value.ModelDefault)) + uint64(len(value.ModelAnimatedFace))
	for _, animation := range value.Animations {
		if !validNativeSkinDimensions(animation.Width, animation.Height, animation.Pixels, false) ||
			animation.Type > uint32(skin.AnimationBody128x128) || animation.FrameCount <= 0 ||
			animation.FrameCount > int64(math.MaxInt) ||
			animation.Expression < int64(math.MinInt) || animation.Expression > int64(math.MaxInt) {
			return skin.Skin{}, false
		}
		total += uint64(len(animation.Pixels))
		if total > maxSkinDataBytes {
			return skin.Skin{}, false
		}
	}
	if total > maxSkinDataBytes {
		return skin.Skin{}, false
	}
	converted := skin.New(int(value.Width), int(value.Height))
	converted.Persona = value.Persona
	converted.PlayFabID = value.PlayFabID
	converted.FullID = value.FullID
	copy(converted.Pix, value.Pixels)
	converted.ModelConfig = skin.ModelConfig{Default: value.ModelDefault, AnimatedFace: value.ModelAnimatedFace}
	converted.Model = slices.Clone(value.Model)
	converted.Cape = skin.NewCape(int(value.CapeWidth), int(value.CapeHeight))
	copy(converted.Cape.Pix, value.CapePixels)
	converted.Animations = make([]skin.Animation, len(value.Animations))
	for i, animation := range value.Animations {
		convertedAnimation := skin.NewAnimation(int(animation.Width), int(animation.Height), int(animation.Expression), skin.AnimationType(animation.Type))
		copy(convertedAnimation.Pix, animation.Pixels)
		convertedAnimation.FrameCount = int(animation.FrameCount)
		converted.Animations[i] = convertedAnimation
	}
	return converted, true
}

func validSkinStrings(values ...string) bool {
	for _, value := range values {
		if len(value) > maxSkinIDBytes {
			return false
		}
	}
	return true
}

func validNativeSkinDimensions(width, height uint32, pixels []byte, empty bool) bool {
	if width > maxSkinDimension || height > maxSkinDimension {
		return false
	}
	return validSkinDimensions(int(width), int(height), pixels, empty)
}

func validSkinDimensions(width, height int, pixels []byte, empty bool) bool {
	if width < 0 || height < 0 || width > maxSkinDimension || height > maxSkinDimension {
		return false
	}
	if width == 0 || height == 0 {
		return empty && width == 0 && height == 0 && len(pixels) == 0
	}
	return uint64(width)*uint64(height)*4 == uint64(len(pixels))
}

func healingSource(source native.HealingSource) world.HealingSource {
	switch source.Kind {
	case native.HealingSourceFood:
		return entity.FoodHealingSource{QuickRegeneration: source.Data}
	case native.HealingSourceInstant:
		return effect.InstantHealingSource{}
	case native.HealingSourceRegeneration:
		return effect.RegenerationHealingSource{}
	default:
		return pluginHealingSource{name: source.Name}
	}
}

func (p *Players) damageSource(tx *world.Tx, source native.DamageSource) (world.DamageSource, bool) {
	switch source.Kind {
	case native.DamageSourceAttack:
		var attacker world.Entity
		if source.Entity.Generation != 0 {
			var ok bool
			attacker, ok = p.entities.Resolve(source.Entity, tx)
			if !ok {
				return nil, false
			}
		}
		return entity.AttackDamageSource{Attacker: attacker}, true
	case native.DamageSourceBlock:
		if source.Block == nil {
			return nil, false
		}
		properties, ok := DecodeBlockProperties(source.Block.PropertiesNBT)
		if !ok {
			return nil, false
		}
		resolved, ok := tx.World().BlockRegistry().BlockByName(source.Block.Identifier, properties)
		return block.DamageSource{Block: resolved}, ok
	case native.DamageSourceDrowning:
		return entity.DrowningDamageSource{}, true
	case native.DamageSourceExplosion:
		return entity.ExplosionDamageSource{}, true
	case native.DamageSourceFall:
		return entity.FallDamageSource{}, true
	case native.DamageSourceFire:
		return block.FireDamageSource{}, true
	case native.DamageSourceGlide:
		return entity.GlideDamageSource{}, true
	case native.DamageSourceInstant:
		return effect.InstantDamageSource{}, true
	case native.DamageSourceLava:
		return block.LavaDamageSource{}, true
	case native.DamageSourceLightning:
		return entity.LightningDamageSource{}, true
	case native.DamageSourceMagma:
		return block.MagmaDamageSource{}, true
	case native.DamageSourcePoison:
		return effect.PoisonDamageSource{Fatal: source.Data}, true
	case native.DamageSourceProjectile:
		var projectile world.Entity
		if source.Entity.Generation != 0 {
			var ok bool
			projectile, ok = p.entities.Resolve(source.Entity, tx)
			if !ok {
				return nil, false
			}
		}
		var owner world.Entity
		var ok bool
		if source.SecondaryEntity.Generation != 0 {
			owner, ok = p.entities.Resolve(source.SecondaryEntity, tx)
			if !ok {
				return nil, false
			}
		}
		return entity.ProjectileDamageSource{Projectile: projectile, Owner: owner}, true
	case native.DamageSourceStarvation:
		return player.StarvationDamageSource{}, true
	case native.DamageSourceSuffocation:
		return entity.SuffocationDamageSource{}, true
	case native.DamageSourceThorns:
		var owner world.Entity
		if source.Entity.Generation != 0 {
			var ok bool
			owner, ok = p.entities.Resolve(source.Entity, tx)
			if !ok {
				return nil, false
			}
		}
		return enchantment.ThornsDamageSource{Owner: owner}, true
	case native.DamageSourceVoid:
		return entity.VoidDamageSource{}, true
	case native.DamageSourceWither:
		return effect.WitherDamageSource{}, true
	default:
		return pluginDamageSource{source: source}, true
	}
}

type pluginHealingSource struct{ name string }

func (pluginHealingSource) HealingSource() {}
func (s pluginHealingSource) Name() string { return s.name }

type pluginDamageSource struct{ source native.DamageSource }

func (s pluginDamageSource) ReducedByArmour() bool     { return s.source.ReducedByArmour }
func (s pluginDamageSource) ReducedByResistance() bool { return s.source.ReducedByResistance }
func (s pluginDamageSource) Fire() bool                { return s.source.Fire }
func (s pluginDamageSource) IgnoreTotem() bool         { return s.source.IgnoresTotem }
func (s pluginDamageSource) Name() string              { return s.source.Name }
func (s pluginDamageSource) AffectedByEnchantment(value item.EnchantmentType) bool {
	switch value {
	case enchantment.FireProtection:
		return s.source.FireProtection
	case enchantment.FeatherFalling:
		return s.source.FeatherFalling
	case enchantment.BlastProtection:
		return s.source.BlastProtection
	case enchantment.ProjectileProtection:
		return s.source.ProjectileProtection
	default:
		return false
	}
}

func finite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
