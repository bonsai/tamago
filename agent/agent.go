package agent

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Status is the lifecycle state of an agent instance.
type Status string

const (
	StatusEgg     Status = "egg"
	StatusReady   Status = "ready"
	StatusRunning Status = "running"
	StatusPaused  Status = "paused"
	StatusStopped Status = "stopped"
)

// Type describes a reusable kind of agent. TAMAGO owns agent types.
type Type struct {
	ID           string            `json:"id" yaml:"id"`
	Name         string            `json:"name" yaml:"name"`
	Description  string            `json:"description,omitempty" yaml:"description,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Config       map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
}

// Egg is a spawn request. It becomes an Instance when hatched.
type Egg struct {
	ID        string            `json:"id" yaml:"id"`
	TypeID    string            `json:"type_id" yaml:"type_id"`
	Name      string            `json:"name,omitempty" yaml:"name,omitempty"`
	Config    map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	CreatedAt time.Time         `json:"created_at" yaml:"created_at"`
}

// Instance is a concrete manageable agent.
type Instance struct {
	ID        string            `json:"id" yaml:"id"`
	EggID     string            `json:"egg_id" yaml:"egg_id"`
	TypeID    string            `json:"type_id" yaml:"type_id"`
	Name      string            `json:"name" yaml:"name"`
	Status    Status            `json:"status" yaml:"status"`
	Config    map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	CreatedAt time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"`
}

// Manager owns agent types, EGGs, and concrete agent instances.
// Persistence and execution remain outside TAMAGO (Worklog / PLEGO).
type Manager struct {
	mu        sync.RWMutex
	types     map[string]Type
	eggs      map[string]Egg
	instances map[string]Instance
	seq       uint64
}

func NewManager() *Manager {
	return &Manager{types: map[string]Type{}, eggs: map[string]Egg{}, instances: map[string]Instance{}}
}

func (m *Manager) RegisterType(t Type) error {
	if t.ID == "" { return errors.New("agent type id is required") }
	if t.Name == "" { return errors.New("agent type name is required") }
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.types[t.ID]; ok { return fmt.Errorf("agent type already exists: %s", t.ID) }
	m.types[t.ID] = cloneType(t)
	return nil
}

func (m *Manager) Type(id string) (Type, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.types[id]
	return cloneType(t), ok
}

// Spawn creates an EGG for an existing agent type.
func (m *Manager) Spawn(typeID, name string, config map[string]string) (Egg, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.types[typeID]; !ok { return Egg{}, fmt.Errorf("unknown agent type: %s", typeID) }
	m.seq++
	e := Egg{ID: fmt.Sprintf("egg-%06d", m.seq), TypeID: typeID, Name: name, Config: cloneMap(config), CreatedAt: time.Now().UTC()}
	m.eggs[e.ID] = e
	return e, nil
}

// Hatch turns an EGG into a concrete agent instance.
func (m *Manager) Hatch(eggID string) (Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.eggs[eggID]
	if !ok { return Instance{}, fmt.Errorf("unknown egg: %s", eggID) }
	if _, ok := m.types[e.TypeID]; !ok { return Instance{}, fmt.Errorf("unknown agent type: %s", e.TypeID) }
	m.seq++
	now := time.Now().UTC()
	name := e.Name
	if name == "" { name = fmt.Sprintf("%s-%06d", e.TypeID, m.seq) }
	i := Instance{ID: fmt.Sprintf("agent-%06d", m.seq), EggID: e.ID, TypeID: e.TypeID, Name: name, Status: StatusReady, Config: cloneMap(e.Config), CreatedAt: now, UpdatedAt: now}
	m.instances[i.ID] = i
	return i, nil
}

func (m *Manager) SetStatus(id string, status Status) (Instance, error) {
	if !validStatus(status) { return Instance{}, fmt.Errorf("invalid agent status: %s", status) }
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.instances[id]
	if !ok { return Instance{}, fmt.Errorf("unknown agent: %s", id) }
	i.Status, i.UpdatedAt = status, time.Now().UTC()
	m.instances[id] = i
	return i, nil
}

func (m *Manager) Get(id string) (Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i, ok := m.instances[id]
	if !ok { return Instance{}, false }
	i.Config = cloneMap(i.Config)
	return i, true
}

func (m *Manager) List() []Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Instance, 0, len(m.instances))
	for _, i := range m.instances { i.Config = cloneMap(i.Config); out = append(out, i) }
	return out
}

func validStatus(s Status) bool {
	switch s {
	case StatusEgg, StatusReady, StatusRunning, StatusPaused, StatusStopped:
		return true
	default:
		return false
	}
}

func cloneMap(in map[string]string) map[string]string {
	if in == nil { return nil }
	out := make(map[string]string, len(in))
	for k, v := range in { out[k] = v }
	return out
}

func cloneType(t Type) Type {
	t.Capabilities = append([]string(nil), t.Capabilities...)
	t.Config = cloneMap(t.Config)
	return t
}
