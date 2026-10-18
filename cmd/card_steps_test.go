package cmd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/TinoA/claude-desktop-switcher/internal/profile"
)

type recordedSteps []string

func (r *recordedSteps) Step(step string, progress float64) {
	*r = append(*r, fmt.Sprintf("%s %.2f", step, progress))
}

func TestCardStepsFollowASwitch(t *testing.T) {
	events := []string{}
	store := &fakeSwitchStore{events: &events, exists: true, current: "outgoing", health: profile.HealthUsable}
	p := &fakePlatform{events: &events, appData: t.TempDir(), running: true}
	labels := map[string]string{"outgoing": "office@accounts.test", "incoming": "private@accounts.test"}
	var steps recordedSteps

	if err := switchProfileWith("incoming", store, p, cardSteps{card: &steps, label: func(name string) string { return labels[name] }}); err != nil {
		t.Fatal(err)
	}

	want := recordedSteps{
		"closing Claude 0.15",
		"saving office@accounts.test 0.35",
		"restoring private@accounts.test 0.55",
		"opening Claude 0.75",
		"signed in 0.90",
	}
	if !reflect.DeepEqual(steps, want) {
		t.Fatalf("steps = %q, want %q", steps, want)
	}
}

func TestCardStepsWhenClaudeIsClosed(t *testing.T) {
	events := []string{}
	store := &fakeSwitchStore{events: &events, exists: true, current: "outgoing", health: profile.HealthUsable}
	p := &fakePlatform{events: &events, appData: t.TempDir()}
	var steps recordedSteps

	if err := switchProfileWith("incoming", store, p, cardSteps{card: &steps, label: func(name string) string { return name }}); err != nil {
		t.Fatal(err)
	}

	want := recordedSteps{"restoring incoming 0.55", "opening Claude 0.75", "signed in 0.90"}
	if !reflect.DeepEqual(steps, want) {
		t.Fatalf("steps = %q, want %q", steps, want)
	}
}
