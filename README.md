# TAMAGO

**TAMA-GO — Total Agent Management Architecture, implemented in Go.**

TAMAGO owns the agent lifecycle:

```text
Agent Type
    │
    ▼
  EGG spawn
    │
    ▼
  Agent hatch
    │
    ▼
Lifecycle management
    │
    ├─ ready
    ├─ running
    ├─ paused
    └─ stopped
```

## Responsibility

TAMAGO defines reusable **Agent Types**, creates immutable **EGGs** from those types, and hatches concrete **Agent Instances**. It also manages the in-memory lifecycle state of those instances.

```text
TANGO
  └─ Think / Decide
        │
        ▼
      TAKT
  └─ Team / Assign / Handoff / Checkpoint
        │
        ▼
     TAMAGO
  └─ Type → EGG → Hatch → Manage
        │
        ▼
      PLEGO
  └─ Tools / Execution
        │
        ▼
     Worklog
  └─ History / Outcome
```

### Boundary

| Component | Responsibility |
|---|---|
| **AW** | Task / Plan / WF specification and compilation |
| **TANGO** | Thinking, planning and decisions |
| **TAKT** | Coordination topology and assignment |
| **TAMAGO** | Agent types, EGG spawn, hatch, lifecycle management |
| **PLEGO** | Concrete tools and execution |
| **Worklog** | Execution history and outcomes |

TAMAGO is **not** the AW → WF compiler. Workflow specification and compilation belong to AW. Agent reasoning belongs to TANGO; coordination belongs to TAKT; concrete execution belongs to PLEGO.

## Go API

```go
m := agent.NewManager()

m.RegisterType(agent.Type{
    ID: "researcher",
    Name: "Researcher",
    Capabilities: []string{"search", "summarize"},
})

egg, _ := m.Spawn("researcher", "rx-01", nil)
instance, _ := m.Hatch(egg.ID)
_, _ = m.SetStatus(instance.ID, agent.StatusRunning)
```

The manager is intentionally in-memory. Persistence/history belongs to Worklog or the runtime layer; execution belongs to PLEGO.

## Layout

```text
agent/
  agent.go       # Type / EGG / Instance / Manager
  agent_test.go  # lifecycle tests
```
