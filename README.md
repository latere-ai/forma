<h1 align="center">Forma</h1>

<p align="center">
  <strong>Run open-weight LLMs from Go. No cgo, no Python, no vendor runtime.</strong>
</p>

<p align="center">
  <a href="https://github.com/latere-ai/forma/actions/workflows/ci.yml"><img src="https://github.com/latere-ai/forma/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/latere.ai/x/forma"><img src="https://pkg.go.dev/badge/latere.ai/x/forma.svg" alt="Go Reference"></a>
  <img src="https://img.shields.io/badge/go-1.27+-00ADD8.svg" alt="Go 1.27+">
  <img src="https://img.shields.io/badge/cgo-free-success.svg" alt="cgo-free">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
  <img src="https://img.shields.io/badge/status-early-orange.svg" alt="Status: early">
</p>

Forma runs open-weight language models from Go, as a library or as a command.
It builds to one static binary with `CGO_ENABLED=0`: no C++ runtime, no
Python, no vendor SDK, and nothing to install beside it. You cross-compile it
the way you cross-compile any Go program.

> [!IMPORTANT]
> **Early, and it runs.** Forma loads Qwen3 checkpoints from Hugging Face and
> generates on Apple silicon through Metal. On an 8-core Apple machine,
> Qwen3-0.6B at f16 decodes about 18 tokens a second for one conversation,
> with about 170 ms to the first token. The CPU backend is correct and far too
> slow to serve from. No comparison against vLLM has been run yet.
> [Orientation](docs/orientation.md) says what runs where and what it costs.

## Install

```sh
go install latere.ai/x/forma/cmd/forma@main    # the command
go get latere.ai/x/forma@main                  # the library, imported as package forma
```

There is no tagged release yet, so both follow `main`.

## Run a model

```sh
MODEL=$(forma pull Qwen/Qwen3-0.6B)     # download into the cache and print the directory
forma run --prompt "Why is the sky blue?" "$MODEL"
forma serve "$MODEL"                    # listen on 127.0.0.1:11434
```

`forma serve` answers OpenAI Chat Completions, Anthropic Messages and OpenAI
Responses on the same model, with streaming, so most clients work unchanged.
The model id a request names is the last element of the model directory,
which `forma serve` prints at startup and `GET /v1/models` lists:

```sh
ID=$(basename "$MODEL")
curl -s http://127.0.0.1:11434/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model": "'"$ID"'", "messages": [{"role": "user", "content": "Why is the sky blue?"}]}'
```

`forma info <dir>` prints the architecture, the precision Forma would choose,
and what the weights and the key/value cache will cost before anything is
loaded. [Commands](docs/commands.md) covers all five commands and
[Serving](docs/serving.md) covers the HTTP API.

## Use it as a library

```go
m, err := forma.Open(dir, forma.WithPrecision(forma.Int8))
if err != nil {
	log.Fatal(err)
}
defer m.Close()

s, err := m.NewSession()
if err != nil {
	log.Fatal(err)
}
defer s.Close()

stream, err := s.Chat(ctx, []chat.Message{
	{Role: chat.User, Blocks: []chat.Block{{Type: chat.BlockText, Text: "Why is the sky blue?"}}},
}, forma.Policy{Temperature: 0.7, TopP: 0.8, MaxTokens: 512})
if err != nil {
	log.Fatal(err)
}
for stream.Next() {
	fmt.Print(stream.Text())
}
if err := stream.Err(); err != nil {
	log.Fatal(err)
}
```

`chat` is `latere.ai/x/forma/chat`. A `Model` is shared and safe for
concurrent use; a `Session` is one conversation. `Model.NewPool` keeps
sessions between requests so a conversation's next turn reuses its prefix,
and `Model.NewRunner` puts every in-flight request in one forward pass. The
[package documentation](https://pkg.go.dev/latere.ai/x/forma) is the
reference.

## What it does

| | |
| --- | --- |
| **Models** | Qwen3 dense, 0.6B through 32B, read directly from Hugging Face safetensors. The hybrid-attention `qwen3_5` architecture (Qwen3.5, Qwen3.8) is recognized and refused by name until the compute layer below supports it. Any other architecture is refused with the list of what Forma knows |
| **Precision** | f16, int8 or int4. By default Forma takes the widest that fits the device, prints what it chose, and never picks int4 unless int8 does not fit. Always overridable |
| **Devices** | Metal on Apple silicon, and a CPU backend everywhere. The CPU backend is a correctness reference, not a serving path |
| **APIs** | OpenAI Chat Completions, Anthropic Messages, OpenAI Responses, and legacy OpenAI Completions |
| **Output control** | streaming, seeded reproducible sampling, logprobs, stop sequences, logit bias and penalties, and JSON-schema output that parses every time |
| **Prompt reuse** | with `--prefix-cache`, a conversation's next turn prefills only what is new; with `--prefix-cache process`, conversations that send the same `cache_salt` and share a system prompt prefill it once between them, and a request with no `cache_salt` shares with nothing |
| **Batching** | `forma serve --batched` puts every in-flight request in one forward pass, so the weights are read once for all of them. It is opt-in; the default serves each request from its own pooled session |
| **Refusals** | a request field that would change the answer and cannot be honored is refused by name; one that cannot change it is accepted and listed in the `X-Forma-Loss` response header |

## Why it exists

**Deployment.** One file. Ship a model inside a Go service, run it on a machine
you cannot install a toolchain on, cross-compile it for a platform you do not
build on. No runtime, no version matrix, no container to keep in step with a
driver.

**Speed, as a goal.** The parts of serving that are not matrix multiplication
(scheduling a step, sampling a token, turning it back into text, deciding what
runs next) are overhead on every token, and they are where a compiled language
with no interpreter lock should win. Forma measures where each decode step
goes (host, submit, device, readback) so that claim can be checked. It has not
been compared against vLLM yet, and vLLM is the better choice today if you
already run Python.

**Hardware.** Forma runs wherever its compute layer runs: Metal on Apple silicon
and CPU everywhere today.

## How it is built

Forma does the model; [accel](https://github.com/golang-design/accel) does the
device. Forma contains no GPU code and no kernels. When it needs something accel
cannot do, the gap is reported upstream and recorded here with its reason,
rather than worked around with private device code.

Two consequences for you: Forma gains a backend when accel does, without
changes, and a limit you meet in Forma is a real, recorded limit rather than an
undocumented edge.

## Documentation

- **[Orientation](docs/orientation.md)**: what runs where, what it costs in
  memory, prompt reuse, batching, and which models run.
- **[Commands](docs/commands.md)**: `pull`, `run`, `info`, `serve` and
  `bench`, and where checkpoints are cached.
- **[Serving](docs/serving.md)**: the HTTP routes, request fields, admission,
  probes and metrics.
- **[Package documentation](https://pkg.go.dev/latere.ai/x/forma)**: the Go API.
- **[specs/](specs/README.md)**: the design, written for contributors: what
  was decided, what was rejected, and why.
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: how to work on Forma.

## License

Apache 2.0. See [LICENSE](LICENSE).
