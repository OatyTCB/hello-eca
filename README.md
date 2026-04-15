# hello-eca

A minimal ECA example: two modules, one cell, working Go code, **backed by a
git-based registry** at https://github.com/OatyTCB/hello-eca.

The repo itself is the registry. Publishing commits to it; consumers clone it.
No server, no hosted registry, no extra infrastructure.

## What's here

```
hello-eca/
├── main.go                              ← Wires both modules into one cell and starts HTTP
├── eca.yaml                             ← registry: github URL;  modules: services
├── topology.yaml                        ← Authored topology — also what GitStore reads
├── eca-cells/main/cell.yaml             ← Authored cell definition
├── services/
│   ├── greeter/
│   │   ├── eca-contracts/               ← Contract + manifest (what greeter exposes)
│   │   │   ├── module.yaml
│   │   │   └── greeter.yaml
│   │   └── handler/handler.go           ← Implementation (what the greeter team writes)
│   └── gateway/
│       ├── eca-contracts/               ← Contract + manifest (what gateway exposes)
│       │   ├── module.yaml
│       │   ├── gateway.yaml
│       │   └── usage/greeter.usage.yaml  ← Scoped usage declaration
│       ├── eca-gen/greeter/             ← Generated client for calling greeter (gitignored)
│       │   ├── types.go                 ← Greeting struct
│       │   └── client.go                ← GreeterClient with SayHello()
│       └── handler/handler.go           ← Implementation (uses eca-gen/greeter client)
└── modules/  cells/                     ← Created on first `eca registry publish`,
                                           committed to the repo root alongside authoring dirs
```

**Authoring vs published.** `eca-cells/main/cell.yaml` is the authored form;
`cells/main/cell.yaml` is what `eca registry publish --cells` writes back and
what GitStore reads. Same for topology. The `modules/` tree at the repo root
is populated by each `eca registry publish`.

## The key idea

**gateway never imports greeter's source code.** It calls greeter through a generated client (`eca-gen/greeter/client.go`). That client delegates to an **adapter** — in this example, an in-process adapter since both modules share the same cell.

If you later split them into separate cells, the adapter changes to gRPC. **No module code changes.**

## Run it

```bash
cd examples/hello-eca
# eca-gen/ is gitignored — regenerate from the git registry before first run.
# GitStore clones the registry repo into ~/.eca/cache/ on first use.
(cd services/gateway && eca generate)
go run main.go
```

```bash
curl http://localhost:8080/welcome/alice
# {"title":"Welcome, alice","greeting":"Hello, alice!"}
```

## Validate with ECA

```bash
eca doctor              # Health check — should be all green
eca topology validate   # Run the 6-stage topology pipeline
eca graph build         # See the dependency graph
eca module list         # List published modules (from the git registry)
```

## Publish a change

```bash
# From the gateway service directory:
cd services/gateway
eca registry publish    # diff vs previous, then git commit + git push

# Platform team publishes cells + topology from the project root:
cd ../..
eca registry publish --cells
eca registry publish --topology
```

Each publish writes into the local clone at `~/.eca/cache/<hash>/`, runs
`git add . && git commit -m "publish <module>@<version>" && git push`. No
other infrastructure involved — the git repo is the registry.

## How the pieces connect

1. **Greeter team** writes `services/greeter/eca-contracts/greeter.yaml` (the contract) and `services/greeter/handler/handler.go` (the implementation). They publish with `eca registry publish`.

2. **Gateway team** declares `consumes: [greeter >=1.0.0]` in their module.yaml, then runs `eca generate` to get a typed Go client in `eca-gen/greeter/`. They call `greeter.NewGreeterClient(adapter)` and use it like any Go interface.

3. **Platform team** defines cells and topology. In this example, both modules are in one cell (`cells/main/cell.yaml`), so calls are in-process. To split them:
   - Create a second cell with greeter
   - Move greeter's ref from the main cell to the new one
   - Run `eca topology validate` — ECA detects the cross-cell edge and resolves the transport

The module teams change nothing.

## What to look at

| File | Why it matters |
|------|----------------|
| `services/gateway/handler/handler.go` | Shows how a module calls a dependency through eca-gen |
| `services/gateway/eca-gen/greeter/client.go` | The generated client — adapter-backed, no hardcoded transport |
| `services/gateway/eca-contracts/usage/greeter.usage.yaml` | Scoped usage declaration — only these fields affect gateway on breaking changes |
| `main.go` | Shows how modules are wired together at cell startup |
| `eca-cells/main/cell.yaml` | Both modules in one cell = in-process calls |
| `eca.yaml` | Points at the github repo; `modules: services` tells publish-all where to look |
