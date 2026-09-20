---
title: "Forma: one name for the repository, module and command"
status: complete
layer: all
depends_on:
  - 000-decisions.md
  - 013-distribution.md
---

# Forma naming migration

Rename the inference framework from tgo to Forma. Its source repository becomes
`latere-ai/forma`, its Go module becomes `latere.ai/x/forma`, its root package
becomes `forma`, and its command is installed from `latere.ai/x/forma/cmd/forma`.

## Scope

- Update imports, package declarations, examples, source links, CLI help and
  errors, documentation, specs, and CI configuration together.
- Rename `cmd/tgo` to `cmd/forma`. Use `FORMA_CACHE`, `FORMA_MODEL`, and
  `FORMA_REQUIRE_METAL`, a `forma` cache directory, `X-Forma-Loss`,
  `forma.bench/1`, and `forma_*` metrics. Document the names a caller must migrate.
  Existing checkpoint
  bytes remain usable by pointing `FORMA_CACHE` at the old cache directory.
- Rename the GitHub repository in place, preserving history, issues and tags.
  Verify the existing generic vanity route resolves `latere.ai/x/forma`.
- Update references in the organization profile, shared tooling, and sibling
  documentation. Preserve unrelated working-tree changes and copyright notices.
- Do not change inference algorithms, model formats, dependencies, or deployment
  policy. Do not publish a release tag as part of this rename.

## Verification

Run the Go suite, coverage gate, cgo-free and dependency gates, spec lint,
formatting, and cross-compilation. Exercise the renamed CLI and imports from a
separate consumer module. Check the cache configuration and HTTP loss header
through their existing behavioral tests. After publication, check repository
metadata, vanity metadata, and installation from the new module path.

## Decision record

| id | decision | rejected | consequence |
| --- | --- | --- | --- |
| 031-D1 | Adopt Forma across source and public interfaces in one migration | Rename only the GitHub display name | Consumers update imports and the command together; documentation has one current name |
| 031-D2 | Document explicit migration of cache and protocol names | Silently move cached checkpoints or retain two permanent interfaces | No user data is moved; existing caches can be reused through `FORMA_CACHE` |

## Outcome

The module, package, command directory, imports, CI configuration, and public
names use Forma. `docs/migration.md` documents consumer changes and cache reuse.
The existing Go suite passes and all 14 coverage-gated packages exceed 90%.
A separate consumer imports the new module, and the built `forma` command runs.
The dependency gate and builds pass on all ten configured platforms; formatting,
spec lint, and the cgo-free gate pass. GitHub is renamed to `latere-ai/forma`,
the vanity metadata resolves it, and installing
`latere.ai/x/forma/cmd/forma@main` from the published repository succeeds.
Reference updates are published in the organization profile, specs, pkg, ci,
ci-gate, fornax, and accel repositories. Historical snapshots and original
copyright notices retain their original wording.

**Not built.** Nothing in this naming migration. A release tag and new hardware
or real-checkpoint measurements are outside its scope.
