package output

import "testing"

func saveAndResetEnvelopes(t *testing.T) {
	t.Helper()
	orig := envelopes
	envelopes = make(map[string]any)
	t.Cleanup(func() { envelopes = orig })
}

func TestRegisterAndRetrieve(t *testing.T) {
	saveAndResetEnvelopes(t)

	type testEnv struct {
		OK bool `json:"ok"`
	}

	RegisterEnvelope("test-cmd", testEnv{})

	got, ok := EnvelopeFor("test-cmd")
	if !ok {
		t.Fatal("expected ok=true for registered command")
	}
	if _, isType := got.(testEnv); !isType {
		t.Fatalf("expected testEnv, got %T", got)
	}
}

func TestEnvelopeFor_Unknown(t *testing.T) {
	saveAndResetEnvelopes(t)

	got, ok := EnvelopeFor("nonexistent")
	if ok {
		t.Fatal("expected ok=false for unregistered command")
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestExistingCommandsRegistered(t *testing.T) {
	for _, name := range []string{"validate", "lookup"} {
		if _, ok := EnvelopeFor(name); !ok {
			t.Errorf("expected %q to be registered", name)
		}
	}
}

func saveAndResetExitCodes(t *testing.T) {
	t.Helper()
	orig := exitCodes
	exitCodes = make(map[string][]int)
	t.Cleanup(func() { exitCodes = orig })
}

func TestRegisterExitCodes_AndRetrieve(t *testing.T) {
	saveAndResetExitCodes(t)

	RegisterExitCodes("validate", []int{0, 1, 2, 65})

	got, ok := ExitCodesFor("validate")
	if !ok {
		t.Fatal("expected ok=true for registered command")
	}
	want := []int{0, 1, 2, 65}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestExitCodesFor_Unknown(t *testing.T) {
	saveAndResetExitCodes(t)

	got, ok := ExitCodesFor("nonexistent")
	if ok {
		t.Fatal("expected ok=false for unregistered command")
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestAllExitCodes_ReturnsCopy(t *testing.T) {
	saveAndResetExitCodes(t)

	RegisterExitCodes("validate", []int{0, 1, 65})
	RegisterExitCodes("lookup", []int{0, 1})

	all := AllExitCodes()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}

	all["schema"] = []int{0}
	delete(all, "validate")

	second := AllExitCodes()
	if len(second) != 2 {
		t.Fatalf("mutation leaked: expected 2 entries, got %d", len(second))
	}
	if _, ok := second["validate"]; !ok {
		t.Fatal("deletion leaked: 'validate' missing from registry")
	}
}

func TestExitCodesFor_ReturnsCopy(t *testing.T) {
	saveAndResetExitCodes(t)

	RegisterExitCodes("test", []int{0, 1, 2})

	got, _ := ExitCodesFor("test")
	got[0] = 99

	got2, _ := ExitCodesFor("test")
	if got2[0] != 0 {
		t.Errorf("mutation leaked: got[0] = %d, want 0", got2[0])
	}
}

func TestExistingExitCodesRegistered(t *testing.T) {
	for _, name := range []string{"validate", "lookup", "extract"} {
		codes, ok := ExitCodesFor(name)
		if !ok {
			t.Errorf("expected %q to have exit codes registered", name)
		}
		if len(codes) == 0 {
			t.Errorf("expected non-empty exit codes for %q", name)
		}
	}
}

func TestAllEnvelopes_ReturnsCopy(t *testing.T) {
	saveAndResetEnvelopes(t)

	type testEnv struct{ OK bool }
	RegisterEnvelope("a", testEnv{})
	RegisterEnvelope("b", testEnv{OK: true})

	all := AllEnvelopes()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}

	all["c"] = testEnv{}
	delete(all, "a")

	second := AllEnvelopes()
	if len(second) != 2 {
		t.Fatalf("mutation leaked: expected 2 entries, got %d", len(second))
	}
	if _, ok := second["a"]; !ok {
		t.Fatal("deletion leaked: 'a' missing from registry")
	}
}
