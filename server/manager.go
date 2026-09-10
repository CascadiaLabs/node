package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

// Manager владеет жизненным циклом sing-box и применяет конфиги на лету.
// Все операции сериализуются мьютексом.
type Manager struct {
	configPath string

	mu         sync.Mutex
	instance   *box.Box
	cancel     context.CancelFunc
	rawConfig  []byte // последний применённый JSON
	startedAt  time.Time
	lastUpdate time.Time
	inbounds   int
	outbounds  int
}

// NewManager создаёт менеджер. configPath — файл персистентности конфига.
func NewManager(configPath string) *Manager {
	return &Manager{configPath: configPath}
}

// Load поднимает ноду со стартовым конфигом: из файла персистентности,
// а если его нет (или он битый) — из дефолтного каскадного конфига.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		// нет файла (или недоступен) — стартуем с дефолта
		return m.Apply([]byte(defaultConfigJSON))
	}
	if err := m.Apply(data); err != nil {
		// файл есть, но битый — не валимся, откатываемся на дефолт
		return m.Apply([]byte(defaultConfigJSON))
	}
	return nil
}

// build парсит и создаёт новый инстанс box из JSON, ничего не запуская.
// Возвращает инстанс, cancel, распарсенные опции. Вызывающий обязан либо
// Start()+сохранить, либо Close()+cancel().
func build(raw []byte) (*box.Box, context.CancelFunc, option.Options, error) {
	ctx, cancel := context.WithCancel(include.Context(context.Background()))
	opts, err := json.UnmarshalExtendedContext[option.Options](ctx, raw)
	if err != nil {
		cancel()
		return nil, nil, option.Options{}, err
	}
	instance, err := box.New(box.Options{Context: ctx, Options: opts})
	if err != nil {
		cancel()
		return nil, nil, option.Options{}, err
	}
	return instance, cancel, opts, nil
}

// Apply валидирует и применяет новый конфиг на лету. При любой ошибке
// текущий работающий инстанс не трогается.
func (m *Manager) Apply(raw []byte) error {
	instance, cancel, opts, err := build(raw)
	if err != nil {
		return err
	}
	if err := instance.Start(); err != nil {
		cancel()
		_ = instance.Close()
		return err
	}

	m.mu.Lock()
	old, oldCancel := m.instance, m.cancel
	m.instance = instance
	m.cancel = cancel
	m.rawConfig = append([]byte(nil), raw...)
	now := time.Now()
	if m.startedAt.IsZero() {
		m.startedAt = now
	}
	m.lastUpdate = now
	m.inbounds = len(opts.Inbounds)
	m.outbounds = len(opts.Outbounds)
	m.mu.Unlock()

	if old != nil {
		oldCancel()
		_ = old.Close()
	}

	m.persist(raw)
	return nil
}

// persist сохраняет конфиг на диск (best-effort, ошибки только логируются вызывающим слоем при желании).
func (m *Manager) persist(raw []byte) {
	if m.configPath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(m.configPath), 0o755)
	tmp := m.configPath + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, m.configPath)
}

// Current возвращает копию последнего применённого конфига.
func (m *Manager) Current() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]byte(nil), m.rawConfig...)
}

// Status — снимок состояния ноды.
type Status struct {
	Running        bool
	SingboxVersion string
	UptimeSeconds  int64
	Inbounds       int
	Outbounds      int
	LastUpdateUnix int64
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Status{
		Running:        m.instance != nil,
		SingboxVersion: constant.Version,
		Inbounds:       m.inbounds,
		Outbounds:      m.outbounds,
	}
	if !m.startedAt.IsZero() {
		s.UptimeSeconds = int64(time.Since(m.startedAt).Seconds())
	}
	if !m.lastUpdate.IsZero() {
		s.LastUpdateUnix = m.lastUpdate.Unix()
	}
	return s
}

// Close останавливает текущий инстанс.
func (m *Manager) Close() {
	m.mu.Lock()
	instance, cancel := m.instance, m.cancel
	m.instance, m.cancel = nil, nil
	m.mu.Unlock()
	if instance != nil {
		cancel()
		_ = instance.Close()
	}
}
