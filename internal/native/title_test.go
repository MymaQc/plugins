package native

import (
	"testing"
	"time"
)

func TestPlayerTitlePreservesSignedNanoseconds(t *testing.T) {
	got := playerTitle("title", "subtitle", "action", -1, int64(2*time.Second)+37, -int64(time.Hour))
	want := PlayerTitle{
		Text: "title", Subtitle: "subtitle", ActionText: "action",
		FadeIn: -1, Duration: 2*time.Second + 37, FadeOut: -time.Hour,
	}
	if got != want {
		t.Fatalf("playerTitle() = %+v, want %+v", got, want)
	}
}
