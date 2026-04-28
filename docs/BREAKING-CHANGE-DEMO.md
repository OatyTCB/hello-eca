# Breaking change demo

This repo ships with greeter **v2.0.0** in source, migrated from v1. The migration introduced two breaking changes — preserved in [docs/v1-before-break/](v1-before-break/) so you can reproduce the v1→v2 diff in eca-web.

## What changed

### greeter contract (v1 → v2)

| Change | v1 | v2 | Why it's breaking |
|---|---|---|---|
| Response field rename | `Greeting.message` | `Greeting.text` | Consumers reading `message` get `undefined` |
| Required input added | `SayHello(name)` | `SayHello(name, locale)` | Consumers not sending `locale` get rejected |
| New response field | — | `Greeting.locale` | Additive — non-breaking |

### gateway migration

- `consumes[].version`: `>=1.0.0` → `>=2.0.0 <3.0.0`
- Usage contract rewritten: reads `text` (not `message`), sends `locale`
- Handler sends `locale` to `SayHello` and reads `result.Text`

## Reproduce the diff in eca-web

Start an eca-web server, then publish v1 and v2 sequentially.

```bash
# 1. Start eca-web (if not already running)
eca registry serve --store ./data &

# 2. Publish greeter v1 by temporarily swapping the v1 artifacts in
cp docs/v1-before-break/greeter.yaml         services/greeter/eca-contracts/greeter.yaml
cp docs/v1-before-break/greeter-module.yaml  services/greeter/eca-contracts/module.yaml
(cd services/greeter && eca registry publish)

# 3. Restore the v2 source (the repo's real state)
git checkout services/greeter/eca-contracts/greeter.yaml
git checkout services/greeter/eca-contracts/module.yaml

# 4. Publish greeter v2 — the server will detect a breaking change
(cd services/greeter && eca registry publish --breaking)
```

Open the eca-web UI:

- **Modules → greeter**: lists both v1 and v2 as independent major versions
- **Modules → greeter → v1 contract ↔ v2 contract** (Diff page): the breaking changes are highlighted — field removal (`message`), new required input (`locale`)
- **Scoped impact** (from the step-4 CLI output): gateway's usage contract intersects the diff — `message` is read-removed, `locale` is send-missing

## The Pact parallel

Pact Broker models the same situation with a different vocabulary:

| ECA | Pact Broker |
|---|---|
| Module contract | Provider's pact (by consumer) |
| Usage contract | Consumer's pact |
| Version range in `consumes:` | Consumer version tag |
| `eca registry publish --breaking` | Provider publishes a new version; broker records verification matrix |
| Scoped impact analysis on publish | `can-i-deploy` pre-deploy query |
| Major-version coexistence in the registry | Provider serves multiple consumer versions simultaneously |

In both systems the provider is responsible for honouring every consumer still pinned to an old version. The registry (ECA) or broker (Pact) is the ledger that says *who is pinned to what*.

### Where ECA matches Pact today

- **Majors coexist.** `modules/greeter/v1/` and `modules/greeter/v2/` live side by side in the registry. Consumers on `>=1.0.0 <2.0.0` keep resolving to v1 after v2 is published.
- **Consumer pins are first-class.** `consumes[].version` is analogous to Pact's consumer version tag.
- **Scoped impact** runs on publish — the analysis reports which usage contracts intersect the breaking change, not just that the change is breaking.

### Where ECA still has gaps

- **No `can-i-deploy` yet.** ECA's analysis runs at publish time. Pact also answers "before I deploy consumer@v to prod, is prod-provider compatible?" — a pre-deploy query ECA should add.
- **No pending/advisory mode.** Pact lets consumers publish new expectations that don't fail provider CI until explicitly verified. ECA's publish is binary: succeed or fail with `--breaking`.
- **No environment tags on module versions.** Pact tracks "auth@1.2.3 is deployed in prod, 1.3.0 in staging." ECA publishes are global.
- **No `can-i-remove` query.** "Is any consumer still pinned to greeter@v1?" — useful for safely retiring old majors. The data is there (usage contracts record `provider_version`); it just needs an endpoint + UI.

## Migration sequence in practice

For a real migration (not a demo) the platform + module teams run through this cycle:

1. **Provider publishes v2** with `--breaking`. v1 stays published; v2 lands alongside.
2. **eca-web shows impact** — every consumer's usage contract is evaluated against the diff. Consumers whose declared surface doesn't intersect are unaffected; consumers whose does are flagged.
3. **Consumer teams migrate one at a time** — bump their `consumes:` version range, regenerate client, update handler, republish their own module.
4. **Provider team waits** until `can-i-remove greeter v1` returns "no consumers pinned" (once we add that query), then deletes v1 from the registry.

Provider code during steps 1–3 keeps serving both v1 and v2 — typically by keeping the v1 handler alive on a legacy path alongside the v2 handler. In hello-eca we've shortcut step 3 by migrating gateway at the same commit as step 1; a real codebase would have a period where both versions of the provider's handler run.
