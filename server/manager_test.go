package main

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/google/uuid"
)

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

// TestManagerRedeploySamePort воспроизводит баг повторного деплоя: новый
// инстанс bind-ится на порты старого, ещё живого → "address already in use".
func TestManagerRedeploySamePort(t *testing.T) {
	port := freePort(t)
	uuid1 := uuid.NewString()
	uuid2 := uuid.NewString()
	conf := func(id string) string {
		return fmt.Sprintf(`{"log":{"level":"error"},"inbounds":[{"type":"vless","tag":"in","listen":"127.0.0.1","listen_port":%d,"users":[{"name":"u","uuid":%q}]}],"outbounds":[{"type":"direct","tag":"direct"}]}`,
			port, id)
	}

	m := NewManager("")
	if err := m.Apply([]byte(conf(uuid1))); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	t.Cleanup(m.Close)

	// Тот же порт, другой конфиг — должно примениться без конфликта портов.
	if err := m.Apply([]byte(conf(uuid2))); err != nil {
		t.Fatalf("re-apply on same port: %v", err)
	}
	if got := string(m.Current()); !strings.Contains(got, uuid2) {
		t.Fatalf("new config not applied:\n%s", got)
	}
	if !m.Status().Running {
		t.Fatal("want running after redeploy")
	}
}

// freePort возвращает свободный TCP-порт (закрывает сокет сразу после выделения).
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
