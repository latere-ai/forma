# Migrating to Forma

Forma is the new name of tgo. Model files and inference APIs keep their shape;
the repository, Go imports, command, and project-specific names change together.

| Previous name | Current name |
| --- | --- |
| `github.com/latere-ai/tgo` repository | `github.com/latere-ai/forma` |
| `github.com/latere-ai/tgo` Go module | `latere.ai/x/forma` |
| `tgo` Go package | `forma` |
| `cmd/tgo`, `tgo` command | `cmd/forma`, `forma` command |
| `TGO_CACHE` | `FORMA_CACHE` |
| `TGO_MODEL` | `FORMA_MODEL` |
| `TGO_REQUIRE_METAL` | `FORMA_REQUIRE_METAL` |
| `X-Tgo-Loss` response header | `X-Forma-Loss` |
| `tgo.bench/1` benchmark schema identifier | `forma.bench/1` |
| `tgo_*` Prometheus metrics | `forma_*` |

## Go applications and the command

Replace the old module prefix in every import, including subpackages, and use
`forma` for the root package. Then update your dependencies:

```sh
go get latere.ai/x/forma@main
go mod tidy
go install latere.ai/x/forma/cmd/forma@main
```

Update scripts to call `forma`. The former module path and command are not
aliases for the new release. GitHub redirects the old repository URL, but Go
consumers still need the new module declaration and imports.

## Reuse downloaded models

Forma uses `FORMA_CACHE`, then `$XDG_CACHE_HOME/forma`, then `~/.cache/forma`.
It does not move or delete the old cache. To reuse the usual old cache in place:

```sh
export FORMA_CACHE="${XDG_CACHE_HOME:-$HOME/.cache}/tgo"
```

If you configured `TGO_CACHE`, set `FORMA_CACHE` to that same directory instead.
Existing model directories remain valid arguments to `forma run` and
`forma serve`; no checkpoint conversion or download is required.

## Clients and test runners

Read `X-Forma-Loss` for fields that could not be represented by the inference
engine. Update benchmark readers to accept `forma.bench/1`; the record fields
are unchanged. Update monitoring queries from `tgo_*` to `forma_*` metrics.
Set `FORMA_MODEL` for real-checkpoint tests and
`FORMA_REQUIRE_METAL` for runners that must provide a Metal device.
