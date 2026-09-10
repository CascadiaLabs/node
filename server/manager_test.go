package main

import "testing"

// Конфиг без inbound'ов — ничего не слушает, не биндит порты (безопасно для тестов).
const noBindConfig = `{"log":{"level":"error"},"outbounds":[{"type":"direct","tag":"direct"}]}`

func TestManagerApplyAndGet(t *testing.T) {
	m := NewManager("")
	if err := m.Apply([]byte(noBindConfig)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	t.Cleanup(m.Close)

	if got := string(m.Current()); got != noBindConfig {
		t.Fatalf("Current() mismatch:\n got=%s", got)
	}
	st := m.Status()
	if !st.Running {
		t.Fatal("want running=true")
	}
	if st.Inbounds != 0 || st.Outbounds != 1 {
		t.Fatalf("counts: in=%d out=%d (want 0/1)", st.Inbounds, st.Outbounds)
	}
	if st.LastUpdateUnix == 0 {
		t.Fatal("want LastUpdateUnix set")
	}
}

func TestManagerApplyInvalidKeepsPrevious(t *testing.T) {
	m := NewManager("")
	if err := m.Apply([]byte(noBindConfig)); err != nil {
		t.Fatalf("apply valid: %v", err)
	}
	t.Cleanup(m.Close)

	if err := m.Apply([]byte("{ not valid json")); err == nil {
		t.Fatal("expected error on invalid config")
	}
	// Прежний конфиг должен остаться рабочим.
	if got := string(m.Current()); got != noBindConfig {
		t.Fatalf("previous config not retained: %s", got)
	}
	if !m.Status().Running {
		t.Fatal("want still running after failed apply")
	}
}
