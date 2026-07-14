package native

import (
	"sync"
	"sync/atomic"
	"testing"
)

type formHost struct {
	noopHost
	send func(InvocationID, PlayerID, PlayerForm) bool
}

func (h formHost) SendPlayerForm(invocation InvocationID, player PlayerID, form PlayerForm) bool {
	return h.send(invocation, player, form)
}

func TestSendPlayerFormFailureDropsContextExactlyOnce(t *testing.T) {
	var sent PlayerForm
	host := registerHost(formHost{send: func(_ InvocationID, _ PlayerID, form PlayerForm) bool {
		sent = form
		return false
	}})
	t.Cleanup(func() { unregisterHost(host) })

	var responses, drops atomic.Int32
	ok := sendPlayerForm(host, 7, PlayerID{Generation: 1}, []byte(`{"type":"form"}`),
		func(InvocationID, PlayerSnapshot, bool, []byte) bool {
			responses.Add(1)
			return true
		}, func() { drops.Add(1) })
	if ok {
		t.Fatal("failed host send reported success")
	}
	CancelPlayerForm(sent.ID)
	if got := responses.Load(); got != 0 {
		t.Fatalf("responses = %d, want 0", got)
	}
	if got := drops.Load(); got != 1 {
		t.Fatalf("drops = %d, want 1", got)
	}
}

func TestSendPlayerFormUnavailableHostDropsContext(t *testing.T) {
	var drops atomic.Int32
	if sendPlayerForm(^uint64(0), 0, PlayerID{}, []byte(`{}`),
		func(InvocationID, PlayerSnapshot, bool, []byte) bool { return true },
		func() { drops.Add(1) }) {
		t.Fatal("send through unavailable host reported success")
	}
	if got := drops.Load(); got != 1 {
		t.Fatalf("drops = %d, want 1", got)
	}
}

func TestSendPlayerFormFailureAfterResponseDoesNotDropTwice(t *testing.T) {
	player := PlayerID{Generation: 4}
	var responses, drops atomic.Int32
	host := registerHost(formHost{send: func(_ InvocationID, _ PlayerID, form PlayerForm) bool {
		if !CompletePlayerForm(form.ID, 8, PlayerSnapshot{Player: player, Name: "Inline"}, false, []byte("0")) {
			t.Fatal("inline response rejected")
		}
		return false
	}})
	t.Cleanup(func() { unregisterHost(host) })

	if sendPlayerForm(host, 7, player, []byte(`{}`),
		func(InvocationID, PlayerSnapshot, bool, []byte) bool {
			responses.Add(1)
			return true
		}, func() { drops.Add(1) }) {
		t.Fatal("failed host send reported success")
	}
	if responses.Load() != 1 || drops.Load() != 0 {
		t.Fatalf("terminal callbacks = responses %d drops %d, want 1/0", responses.Load(), drops.Load())
	}
}

func TestCompletePlayerFormCarriesSnapshotAndOwnsTerminalCallback(t *testing.T) {
	var sent PlayerForm
	host := registerHost(formHost{send: func(_ InvocationID, _ PlayerID, form PlayerForm) bool {
		sent = form
		return true
	}})
	t.Cleanup(func() { unregisterHost(host) })

	player := PlayerID{UUID: [16]byte{1, 2, 3}, Generation: 9}
	want := PlayerSnapshot{
		Player: player, Name: "SnapshotPlayer", LatencyMilliseconds: 42,
		Position: Vec3{X: 1.5, Y: 64, Z: -2.25},
	}
	var got PlayerSnapshot
	var gotInvocation InvocationID
	var gotClosed bool
	var gotResponse string
	var responses, drops atomic.Int32
	if !sendPlayerForm(host, 3, player, []byte(`{"type":"form"}`),
		func(invocation InvocationID, snapshot PlayerSnapshot, closed bool, response []byte) bool {
			responses.Add(1)
			gotInvocation, got, gotClosed, gotResponse = invocation, snapshot, closed, string(response)
			return true
		}, func() { drops.Add(1) }) {
		t.Fatal("form send rejected")
	}
	if !CompletePlayerForm(sent.ID, 17, want, false, []byte("0")) {
		t.Fatal("form response rejected")
	}
	CancelPlayerForm(sent.ID)
	if gotInvocation != 17 || got != want || gotClosed || gotResponse != "0" {
		t.Fatalf("response = invocation %d snapshot %+v closed %t body %q", gotInvocation, got, gotClosed, gotResponse)
	}
	if responses.Load() != 1 || drops.Load() != 0 {
		t.Fatalf("terminal callbacks = responses %d drops %d, want 1/0", responses.Load(), drops.Load())
	}
}

func TestPlayerFormResponseAndDropRaceHasOneWinner(t *testing.T) {
	var sent PlayerForm
	host := registerHost(formHost{send: func(_ InvocationID, _ PlayerID, form PlayerForm) bool {
		sent = form
		return true
	}})
	t.Cleanup(func() { unregisterHost(host) })

	player := PlayerID{Generation: 21}
	var responses, drops atomic.Int32
	if !sendPlayerForm(host, 0, player, []byte(`{}`),
		func(InvocationID, PlayerSnapshot, bool, []byte) bool {
			responses.Add(1)
			return true
		}, func() { drops.Add(1) }) {
		t.Fatal("form send rejected")
	}

	start := make(chan struct{})
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		<-start
		CompletePlayerForm(sent.ID, 1, PlayerSnapshot{Player: player}, true, nil)
	}()
	go func() {
		defer group.Done()
		<-start
		CancelPlayerForm(sent.ID)
	}()
	close(start)
	group.Wait()
	if got := responses.Load() + drops.Load(); got != 1 {
		t.Fatalf("terminal callback count = %d (responses %d, drops %d), want 1", got, responses.Load(), drops.Load())
	}
}
