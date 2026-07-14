package native

import (
	"os"
	"os/exec"
	"testing"
)

const nativeAOTLifetimeChild = "BEDROCK_GOPHERS_NATIVEAOT_LIFETIME_CHILD"

func TestNativeAOTLibrariesRemainMappedAfterClose(t *testing.T) {
	if os.Getenv(nativeAOTLifetimeChild) == "1" {
		library, plugins := csharpArtifacts(t)
		runtime, err := Open(library, plugins)
		if err != nil {
			t.Fatal(err)
		}
		runtime.Close()

		recovered := false
		func() {
			defer func() { recovered = recover() != nil }()
			var pointer *byte
			_ = *pointer
		}()
		if !recovered {
			t.Fatal("Go SIGSEGV handler did not recover a nil dereference")
		}
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestNativeAOTLibrariesRemainMappedAfterClose$")
	command.Env = append(os.Environ(), nativeAOTLifetimeChild+"=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("process crashed after closing NativeAOT runtime: %v\n%s", err, output)
	}
}
