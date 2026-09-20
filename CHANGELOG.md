# Changelog

Every tag has a section here, and the section is the body of the GitHub
release. A tag without one is refused at the pre-push and fails the release
workflow. Write under `Unreleased` as work lands; `lateregate release vX.Y.Z`
turns that into the tag's section, commits, tags and pushes.

A section says what changed for whoever uses the release, not what was
committed: the commit log already holds that.

## Unreleased

- Rename the framework to Forma, the repository to `latere-ai/forma`, and the Go
  module to `latere.ai/x/forma`. Update imports to package `forma` and install
  `latere.ai/x/forma/cmd/forma`.
- Rename configuration to `FORMA_CACHE`, `FORMA_MODEL`, and `FORMA_REQUIRE_METAL`,
  the default cache directory to `forma`, the loss header to `X-Forma-Loss`, and
  benchmark records to `forma.bench/1`, and metric names to `forma_*`.
  See [migration](docs/migration.md) to reuse
  existing checkpoints and update clients.
