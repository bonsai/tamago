package agent

import "testing"

func TestLifecycle(t *testing.T) {
	m := NewManager()
	if err := m.RegisterType(Type{ID: "researcher", Name: "Researcher", Capabilities: []string{"search", "summarize"}}); err != nil {
		t.Fatal(err)
	}
	egg, err := m.Spawn("researcher", "rx-01", map[string]string{"model": "local"})
	if err != nil { t.Fatal(err) }
	if egg.TypeID != "researcher" { t.Fatalf("unexpected type: %s", egg.TypeID) }

	a, err := m.Hatch(egg.ID)
	if err != nil { t.Fatal(err) }
	if a.Status != StatusReady { t.Fatalf("unexpected status: %s", a.Status) }
	if a.EggID != egg.ID { t.Fatalf("unexpected egg: %s", a.EggID) }

	a, err = m.SetStatus(a.ID, StatusRunning)
	if err != nil { t.Fatal(err) }
	if a.Status != StatusRunning { t.Fatalf("unexpected status: %s", a.Status) }
	if len(m.List()) != 1 { t.Fatalf("expected one agent") }
}

func TestUnknownTypeAndEgg(t *testing.T) {
	m := NewManager()
	if _, err := m.Spawn("missing", "", nil); err == nil { t.Fatal("expected unknown type error") }
	if _, err := m.Hatch("missing"); err == nil { t.Fatal("expected unknown egg error") }
}
