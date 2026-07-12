package native

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func nativeArtifacts(t testing.TB) (string, string) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	extension := ".so"
	if runtime.GOOS == "darwin" {
		extension = ".dylib"
	} else if runtime.GOOS == "windows" {
		extension = ".dll"
	}
	library := filepath.Join(root, "build", "lib", "libdragonfly_plugin_runtime"+extension)
	plugins := filepath.Join(root, "build", "plugins")
	if _, err := os.Stat(library); err != nil {
		t.Skipf("native runtime not built: run make build-native (%v)", err)
	}
	return library, plugins
}

func openTestRuntime(t testing.TB) *Runtime {
	t.Helper()
	library, plugins := nativeArtifacts(t)
	runtime, err := Open(library, plugins)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(runtime.Close)
	t.Cleanup(runtime.Disable)
	return runtime
}

func TestMovementGuard(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.PluginCount() != 5 {
		t.Fatalf("plugin count = %d, want 5", runtime.PluginCount())
	}
	if runtime.Subscriptions()&PlayerMoveSubscription == 0 {
		t.Fatal("movement subscription missing")
	}

	input := PlayerMoveInput{NewPosition: Vec3{X: 10, Y: 64, Z: 10}}
	cancelled, err := runtime.HandlePlayerMove(input, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("valid movement cancelled")
	}

	input.NewPosition.Y = -65
	cancelled, err = runtime.HandlePlayerMove(input, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("movement below world was not cancelled")
	}
}

func TestCommand(t *testing.T) {
	runtime := openTestRuntime(t)
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 2 || commands[0].Name != "hello" || commands[1].Name != "ping" {
		t.Fatalf("commands = %#v, want hello and ping", commands)
	}
	last := commands[0].Overloads[len(commands[0].Overloads)-1]
	if len(last.Parameters) != 2 || !last.Parameters[1].Optional {
		t.Fatalf("optional overload = %#v", last)
	}
	output, err := runtime.HandleCommand(commands[0].Index, CommandInput{
		Source:    "Danick",
		Arguments: "say excited dragonfly plugins rock",
	})
	if err != nil {
		t.Fatal(err)
	}
	if output.Failed || output.Message != "HELLO, DANICK! DRAGONFLY PLUGINS ROCK" {
		t.Fatalf("output = %#v", output)
	}
}

func TestPingCommandUsesPlayerLatency(t *testing.T) {
	runtime := openTestRuntime(t)
	commands, err := runtime.Commands()
	if err != nil {
		t.Fatal(err)
	}
	ping := commands[1]
	id := PlayerID{Generation: 9}
	id.UUID[0] = 1
	input := CommandInput{
		Source:       "Danick",
		SourcePlayer: &id,
		OnlinePlayers: []CommandPlayer{{
			Player:              id,
			Name:                "Danick",
			LatencyMilliseconds: 37,
		}},
	}
	output, err := runtime.HandleCommand(ping.Index, input)
	if err != nil {
		t.Fatal(err)
	}
	if output.Failed || output.Message != "Danick's ping: 37ms" {
		t.Fatalf("output = %#v", output)
	}
}

func TestDynamicCommandEnum(t *testing.T) {
	runtime := openTestRuntime(t)
	options, err := runtime.CommandEnumOptions(0, 6, 1, "Danick", []string{"Danick", "RestartFU"})
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 3 || options[0] != "Danick" || options[1] != "RestartFU" || options[2] != "everyone" {
		t.Fatalf("options = %#v", options)
	}
}

func TestChatFilter(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerChatSubscription == 0 {
		t.Fatal("chat subscription missing")
	}

	output, err := runtime.HandlePlayerChat(PlayerChatInput{Message: "foo fighters"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled {
		t.Fatal("ordinary chat cancelled")
	}
	if output.Replacement == nil || *output.Replacement != "bar fighters" {
		t.Fatalf("replacement = %v, want bar fighters", output.Replacement)
	}

	output, err = runtime.HandlePlayerChat(PlayerChatInput{Message: "blocked"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !output.Cancelled {
		t.Fatal("blocked chat was not cancelled")
	}
}

func TestPlayerJoinAndQuit(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerJoinSubscription == 0 || runtime.Subscriptions()&PlayerQuitSubscription == 0 {
		t.Fatal("join or quit subscription missing")
	}
	id := PlayerID{Generation: 12}
	cancelled, err := runtime.HandlePlayerJoin(PlayerJoinInput{Player: id, Name: "Danick"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("lifecycle logger cancelled join")
	}
	if err := runtime.HandlePlayerQuit(PlayerQuitInput{Player: id, Name: "Danick"}); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerHurtAndHeal(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerHurtSubscription == 0 || runtime.Subscriptions()&PlayerHealSubscription == 0 {
		t.Fatal("hurt or heal subscription missing")
	}
	hurt, err := runtime.HandlePlayerHurt(PlayerHurtInput{
		Damage:         4,
		AttackImmunity: 500 * time.Millisecond,
		Source:         "testDamageSource",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if hurt.Cancelled || hurt.Damage != 4 || hurt.AttackImmunity != 500*time.Millisecond {
		t.Fatalf("hurt = %+v", hurt)
	}
	heal, err := runtime.HandlePlayerHeal(PlayerHealInput{Health: 2, Source: "testHealingSource"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if heal.Cancelled || heal.Health != 2 {
		t.Fatalf("heal = %+v", heal)
	}
}

func TestPlayerBlockBreakAndPlace(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerBlockBreakSubscription == 0 || runtime.Subscriptions()&PlayerBlockPlaceSubscription == 0 {
		t.Fatal("block-break or block-place subscription missing")
	}
	broken, err := runtime.HandlePlayerBlockBreak(PlayerBlockBreakInput{
		Position:   BlockPos{X: 1, Y: 2, Z: 3},
		Block:      "minecraft:stone",
		Experience: 4,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if broken.Cancelled || broken.Experience != 4 {
		t.Fatalf("block break = %+v", broken)
	}
	cancelled, err := runtime.HandlePlayerBlockPlace(PlayerBlockPlaceInput{
		Position: BlockPos{X: 4, Y: 5, Z: 6},
		Block:    "minecraft:dirt",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("block place cancelled")
	}
}

func TestPlayerFoodLossAndDeath(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerFoodLossSubscription == 0 || runtime.Subscriptions()&PlayerDeathSubscription == 0 {
		t.Fatal("food-loss or death subscription missing")
	}
	food, err := runtime.HandlePlayerFoodLoss(PlayerFoodLossInput{From: 10, To: 9}, false)
	if err != nil {
		t.Fatal(err)
	}
	if food.Cancelled || food.To != 9 {
		t.Fatalf("food loss = %+v", food)
	}
	keep, err := runtime.HandlePlayerDeath(PlayerDeathInput{Source: "testDamageSource"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if keep {
		t.Fatal("lifecycle logger changed keep inventory")
	}
}

func TestPlayerToggleSprintAndSneak(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerToggleSprintSubscription == 0 || runtime.Subscriptions()&PlayerToggleSneakSubscription == 0 {
		t.Fatal("sprint-toggle or sneak-toggle subscription missing")
	}
	for name, handle := range map[string]func(PlayerToggleInput, bool) (bool, error){
		"sprint": runtime.HandlePlayerToggleSprint,
		"sneak":  runtime.HandlePlayerToggleSneak,
	} {
		cancelled, err := handle(PlayerToggleInput{After: true}, false)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cancelled {
			t.Fatalf("%s toggle cancelled", name)
		}
	}
}

func TestPlayerJumpAndTeleport(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerJumpSubscription == 0 || runtime.Subscriptions()&PlayerTeleportSubscription == 0 {
		t.Fatal("jump or teleport subscription missing")
	}
	if err := runtime.HandlePlayerJump(PlayerID{}); err != nil {
		t.Fatal(err)
	}
	cancelled, err := runtime.HandlePlayerTeleport(PlayerTeleportInput{Position: Vec3{X: 1, Y: 64, Z: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("teleport cancelled")
	}
}

func TestPlayerExperienceGainAndPunchAir(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerExperienceGainSubscription == 0 || runtime.Subscriptions()&PlayerPunchAirSubscription == 0 {
		t.Fatal("experience-gain or punch-air subscription missing")
	}
	output, err := runtime.HandlePlayerExperienceGain(PlayerID{}, 5, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.Amount != 5 {
		t.Fatalf("experience gain = %+v", output)
	}
	cancelled, err := runtime.HandlePlayerPunchAir(PlayerID{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("punch air cancelled")
	}
}

func TestPlayerHeldSlotChangeAndSleep(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerHeldSlotChangeSubscription == 0 || runtime.Subscriptions()&PlayerSleepSubscription == 0 {
		t.Fatal("held-slot-change or sleep subscription missing")
	}
	cancelled, err := runtime.HandlePlayerHeldSlotChange(PlayerHeldSlotChangeInput{From: 1, To: 2}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("held-slot change cancelled")
	}
	output, err := runtime.HandlePlayerSleep(PlayerID{}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || !output.SendReminder {
		t.Fatalf("sleep = %+v", output)
	}
}

func TestPlayerBlockPickAndLecternPageTurn(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerBlockPickSubscription == 0 || runtime.Subscriptions()&PlayerLecternPageTurnSubscription == 0 {
		t.Fatal("block-pick or lectern-page subscription missing")
	}
	cancelled, err := runtime.HandlePlayerBlockPick(PlayerBlockPickInput{Block: "minecraft:stone"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("block pick cancelled")
	}
	output, err := runtime.HandlePlayerLecternPageTurn(PlayerLecternPageTurnInput{OldPage: 1, NewPage: 2}, false)
	if err != nil {
		t.Fatal(err)
	}
	if output.Cancelled || output.NewPage != 2 {
		t.Fatalf("lectern page = %+v", output)
	}
}

func TestPlayerSignEditAndItemUse(t *testing.T) {
	runtime := openTestRuntime(t)
	if runtime.Subscriptions()&PlayerSignEditSubscription == 0 || runtime.Subscriptions()&PlayerItemUseSubscription == 0 {
		t.Fatal("sign-edit or item-use subscription missing")
	}
	cancelled, err := runtime.HandlePlayerSignEdit(PlayerSignEditInput{OldText: "old", NewText: "new"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("sign edit cancelled")
	}
	cancelled, err = runtime.HandlePlayerItemUse(PlayerID{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("item use cancelled")
	}
}

func TestCancellationIsMonotonic(t *testing.T) {
	runtime := openTestRuntime(t)
	cancelled, err := runtime.HandlePlayerMove(PlayerMoveInput{NewPosition: Vec3{Y: 64}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("plugin cleared existing cancellation")
	}
}

func TestLifecycleControlsDispatch(t *testing.T) {
	runtime := openTestRuntime(t)
	runtime.Disable()
	cancelled, err := runtime.HandlePlayerMove(PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("disabled plugin handled movement")
	}
	if err := runtime.Enable(); err != nil {
		t.Fatal(err)
	}
	cancelled, err = runtime.HandlePlayerMove(PlayerMoveInput{NewPosition: Vec3{Y: -65}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("enabled plugin did not handle movement")
	}
}

//go:noinline
func rawGoMovement(input PlayerMoveInput, cancelled *bool) {
	if input.NewPosition.Y < 0 {
		*cancelled = true
	}
}

func BenchmarkRawGoMovement(b *testing.B) {
	input := PlayerMoveInput{NewPosition: Vec3{Y: 64}}
	for b.Loop() {
		cancelled := false
		rawGoMovement(input, &cancelled)
	}
}

func BenchmarkNativeRustMovement(b *testing.B) {
	runtime := openTestRuntime(b)
	input := PlayerMoveInput{NewPosition: Vec3{Y: 64}}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := runtime.HandlePlayerMove(input, false); err != nil {
			b.Fatal(err)
		}
	}
}
