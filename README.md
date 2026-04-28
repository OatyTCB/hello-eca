# hello-eca

A minimal ECA example: two modules, one cell, working Go code.

Demonstrates how gateway calls greeter through a generated client without ever importing greeter's source.

Greeter is at **v2.0.0** — the repo ships in the post-migration state from a v1→v2 breaking change. See [docs/BREAKING-CHANGE-DEMO.md](docs/BREAKING-CHANGE-DEMO.md) to reproduce the v1→v2 contract diff in eca-web.

## What's here

```
hello-eca/
├── main.go                              ← Wires both modules into one cell and starts HTTP on :8081
├── eca.yaml                             ← registry: http://localhost:8080/hello-eca
├── topology.yaml                        ← Authored topology
├── eca-cells/main/cell.yaml             ← Cell definition
├── docs/
│   ├── BREAKING-CHANGE-DEMO.md          ← How to reproduce the v1→v2 diff in eca-web
│   └── v1-before-break/                 ← Pre-migration artifacts (greeter v1 contract + module + usage)
└── services/
    ├── greeter/                         ← Currently at v2.0.0
    │   ├── eca-contracts/
    │   │   ├── module.yaml
    │   │   └── greeter.yaml
    │   └── handler/handler.go
    └── gateway/                         ← Migrated to consume greeter v2
        ├── eca-contracts/
        │   ├── module.yaml
        │   ├── gateway.yaml
        │   └── usage/greeter.usage.yaml
        ├── eca-gen/greeter/             ← Generated client (gitignored)
        │   ├── types.go
        │   └── client.go
        └── handler/handler.go
```

## The key idea

**gateway never imports greeter's source.** It calls greeter through a generated client (`eca-gen/greeter/client.go`). That client delegates to an **adapter** — in this example, an in-process adapter since both modules share the same cell.

Split them into separate cells later and the adapter changes to gRPC. **No module code changes.**

## Run it

Start an eca-web server and register the project:

```bash
eca registry serve --store ./data &
```

Then from this directory:

```bash
eca registry publish --modules          # Publish both module contracts
eca registry publish --cells            # Publish the cell
eca registry publish --topology         # Publish topology

(cd services/gateway && eca generate)   # Generate the greeter client
go run main.go
```

```bash
curl http://localhost:8081/welcome/alice
# {"title":"Welcome, alice","greeting":"Hello, alice!"}

curl 'http://localhost:8081/welcome/alice?locale=fr'
# {"title":"Welcome, alice","greeting":"Bonjour, alice !"}
```

## Validate

```bash
eca doctor              # Health check — should be all green
eca topology validate   # Run the topology pipeline
eca graph build         # See the dependency graph
eca module list         # List published modules
```

## Publish a change

```bash
cd services/gateway
eca registry publish    # diffs against the server; fails on unacknowledged breaking changes

# Platform team publishes cells + topology from the project root:
cd ../..
eca registry publish --cells
eca registry publish --topology
```

Each publish POSTs to the eca-web server, which validates the new contract against the previously-published one and rejects breaking changes unless `--breaking` is passed.

## How the pieces connect

1. **Greeter team** writes `services/greeter/eca-contracts/greeter.yaml` (the contract) and `services/greeter/handler/handler.go` (the implementation). They publish with `eca registry publish`.

2. **Gateway team** declares `consumes: [greeter >=1.0.0]` in their module.yaml, runs `eca generate` to get a typed Go client in `eca-gen/greeter/`, then calls `greeter.NewGreeterClient(adapter)` like any Go interface.

3. **Platform team** defines cells and topology. Both modules live in one cell (`eca-cells/main/cell.yaml`), so calls are in-process. To split them:
   - Create a second cell containing greeter
   - Move greeter's ref from the main cell to the new one
   - Run `eca topology validate` — ECA detects the cross-cell edge and resolves the transport

The module teams change nothing.

## What to look at

| File | Why it matters |
|------|----------------|
| `services/gateway/handler/handler.go` | Shows how a module calls a dependency through eca-gen |
| `services/gateway/eca-gen/greeter/client.go` | Generated client — adapter-backed, no hardcoded transport |
| `services/gateway/eca-contracts/usage/greeter.usage.yaml` | Scoped usage — only these fields affect gateway on breaking changes |
| `main.go` | Wires modules together at cell startup |
| `eca-cells/main/cell.yaml` | Both modules in one cell = in-process calls |
| `eca.yaml` | Points at the local eca-web server |
