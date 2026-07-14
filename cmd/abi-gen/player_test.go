package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerCompositeOperationDrivesGeneratedSurfaces(t *testing.T) {
	schema, err := readPlayer(filepath.Join("..", "..", "schema", "player.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	operations := compositePlayerOperations(schema)
	if len(operations) != 1 || operations[0].Name != "experience" || len(operations[0].Args) != 2 {
		t.Fatalf("composite operations = %+v", operations)
	}

	outputs := map[string]string{
		"c abi":        string(generateC(nil, schema)),
		"rust abi":     string(generateRust(nil, schema)),
		"go bridge":    string(generateGoNativePlayerOperations(schema)),
		"c bridge":     string(generateCNativePlayerOperations(schema)),
		"rust wrapper": string(generateRustPlayerStates(schema)),
	}
	wants := map[string][]string{
		"c abi":        {"DfHostPlayerExperienceSetFn", "player_experience_set"},
		"rust abi":     {"DfHostPlayerExperienceSetFn", "player_experience_set"},
		"go bridge":    {"SetPlayerExperience", "levelValue < 0", "math.IsNaN(progressValue)", "progressValue > 1"},
		"c bridge":     {"bg_go_player_experience_set", ".player_experience_set = host_player_experience_set"},
		"rust wrapper": {"pub fn set_experience(&self, level: i32, progress: f64)", "level < 0", "!progress.is_finite()", "contains(&progress)"},
	}
	for name, fragments := range wants {
		for _, fragment := range fragments {
			if !strings.Contains(outputs[name], fragment) {
				t.Fatalf("%s missing %q", name, fragment)
			}
		}
	}
}

func TestReadPlayerRejectsInvalidCompositeOperationSchema(t *testing.T) {
	original := playerTestSchema(t)
	tests := map[string]string{
		"unknown validation":   strings.Replace(original, "validate: unit_interval", "validate: outside_unit_interval", 1),
		"duplicate argument":   strings.Replace(original, "name: progress", "name: level", 1),
		"source and arguments": strings.Replace(original, "    args:\n", "    source: healing\n    args:\n", 1),
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "player.yaml")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readPlayer(path); err == nil {
				t.Fatal("invalid composite operation accepted")
			}
		})
	}
}

func playerTestSchema(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "schema", "player.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
