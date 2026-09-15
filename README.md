# tamago

**AW YAML → WF YAML compiler**

`tango` が「考える頭」なら、`tamago` はその計画を実行可能な Workflow に孵化させるコンパイラ。

```text
*.aw.yaml
    │
    ├─ schema validation
    ├─ prompt construction
    ├─ examples / cases
    ├─ LLM planner
    │
    ├─ structured WF JSON
    ├─ WF validation
    └─ repair loop (planned)
          │
          ▼
     *.wf.yaml
          │
          ▼
       Runtime
```

## Why not just send it to an LLM?

LLM は「どう達成するか」の計画を作る役割だけにする。

- AW schema: 入力契約
- examples/cases: 過去の正しい分解を与える
- LLM: planner
- WF validation: 出力契約
- Runtime: 検証済み WF の実行

つまり **LLM output = source of truth ではない**。

## Quick start

```bash
go run ./cmd/tamago validate examples/research.aw.yaml
go run ./cmd/tamago compile examples/research.aw.yaml --dry-run
go run ./cmd/tamago compile examples/research.aw.yaml -o examples/research.generated.wf.yaml
```

実 LLM を使う場合は OpenAI-compatible endpoint を指定できる。LM Studio をそのまま使える。

```bash
export TAMAGO_BASE_URL=http://127.0.0.1:1234/v1
export TAMAGO_MODEL=your-local-model
go run ./cmd/tamago compile examples/research.aw.yaml -o research.wf.yaml
```

Hosted API でも同じ adapter を利用できる。

## Positioning

```text
bonsai/aw
  └─ AW / WF specification, schemas, generators

bonsai/tango
  └─ planner / agent team / LLM thinking

bonsai/tamago
  └─ AW → WF compiler

bonsai/plego
  └─ tools / hands
```

GitHub Actions だけに閉じず、LM Studio / OpenAI-compatible API / 他 runtime へ持ち出せることを重視する。
