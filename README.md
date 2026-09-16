# TAMAGO

**TAMA-GO — Total Agent Management Architecture, implemented in Go.**

TAMAGO is the **agent-based plugin and team-building layer** of the bonsai agent ecosystem.

It answers:

> **Who can work together?**

```text
Agent Plugins
      │
      ▼
 Team Template
      │
      ▼
 EGG × N
      │
      ▼
    Hatch
      │
      ▼
  Team Instance
      │
      ▼
Lifecycle management
```

## Responsibility

TAMAGO defines reusable **Agent Types / Agent Plugins**, composes them into team templates, creates EGGs for concrete team members, hatches agent instances, and manages their lifecycle.

TAMAGO does **not** decide the semantic task, reason about the task, compile AW into WF, or execute concrete work.

```text
AW
  └─ Task / Plan / WF specification
        │
        ▼
TANGO
  └─ Think / Decide
        │
        ▼
TAKT
  └─ Topology / Team / Assign / Handoff / Checkpoint
        │
        ▼
TAMAGO
  └─ Agent Plugin / Team Building
        │
        ├─ Agent Type
        ├─ Team Template
        ├─ EGG spawn
        ├─ Hatch
        └─ Lifecycle
        │
        ▼
PLEGO
  └─ Work Plugin / Execution
        │
        ▼
Worklog
  └─ History / Outcome
```

## Boundary

| Component | Responsibility |
|---|---|
| **AW** | Task / Plan / WF specification and compilation |
| **TANGO** | Thinking, planning and decisions |
| **TAKT** | Coordination topology and assignment |
| **TAMAGO** | Agent plugins, team building, spawning, hatching, lifecycle |
| **PLEGO** | Pluggable work execution and concrete tools |
| **Worklog** | Execution history and outcomes |

A useful boundary is:

```text
TAMAGO = Who can work together?
PLEGO  = What can they do?
```

## Agent Plugin

An Agent Plugin is a reusable agent capability/type that can participate in one or more team templates.

```go
type Type struct {
    ID           string
    Name         string
    Description  string
    Capabilities []string
    Config       map[string]string
}
```

Examples:

```text
researcher
coder
reviewer
verifier
publisher
monitor
```

## Team Building

A team is a composition of Agent Types rather than a single agent definition.

```text
research-team
├─ researcher
├─ verifier
└─ publisher

build-team
├─ planner
├─ coder
└─ reviewer
```

TAKT owns the coordination topology and assignments; TAMAGO supplies the concrete agent members.

## Agent lifecycle

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
Lifecycle
 ├─ ready
 ├─ running
 ├─ paused
 └─ stopped
```

The current Go manager provides type registration, EGG spawning, hatching, status changes, lookup and listing. fileciteturn74file0

The manager is intentionally in-memory. Persistence/history belongs to Worklog or the runtime layer; concrete execution belongs to PLEGO.

## Go API

```go
m := agent.NewManager()

_ = m.RegisterType(agent.Type{
    ID: "researcher",
    Name: "Researcher",
    Capabilities: []string{"search", "summarize"},
})

egg, _ := m.Spawn("researcher", "rx-01", nil)
instance, _ := m.Hatch(egg.ID)
_, _ = m.SetStatus(instance.ID, agent.StatusRunning)
```

## Layout

```text
agent/
  agent.go       # Agent Type / EGG / Instance / Manager
  agent_test.go  # lifecycle tests
```
