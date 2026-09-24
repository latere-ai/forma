# Forma documentation

For people running open-weight models with Forma, from the command line or as
a Go library.

## Running models

| page | what it covers |
| --- | --- |
| [Orientation](orientation.md) | what Forma is, what runs where, what a model and its cache cost in memory, prompt reuse, batching, and which models run |
| [Commands](commands.md) | `forma pull`, `run`, `info`, `serve` and `bench`, the flags that matter, and where checkpoints are cached |
| [Serving](serving.md) | the HTTP routes, the request fields Forma honors, structured output, admission and the queue, probes and metrics |

`forma help` prints every flag of every command. The build checks that text
against the flags each command declares, so it does not drift.

## Embedding Forma in a Go program

The [package documentation](https://pkg.go.dev/latere.ai/x/forma) is the
reference for `Model`, `Session`, `Stream`, `Pool`, `Runner` and `Policy`. The
[README](../README.md#use-it-as-a-library) has a complete example.

## Changing Forma

The design, with the decisions and the alternatives that were rejected, is in
[`../specs/`](../specs/README.md). [`../CONTRIBUTING.md`](../CONTRIBUTING.md)
covers the rules a patch has to meet and how to run the gates locally. Neither
is needed to run a model.
