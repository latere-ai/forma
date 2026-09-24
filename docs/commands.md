# The forma command

`forma` fetches a checkpoint, runs it, describes it, serves it, and measures
it. This page covers what each command is for and the flags that decide
behavior. `forma help` prints every flag with its default.

```
forma pull  [flags] <repo-id>
forma run   [flags] <model-dir>
forma info  [flags] <model-dir>
forma serve [flags] <model-dir>
forma bench [flags] <model-dir>
```

**Flags come before the model directory.** Parsing stops at the first argument
that is not a flag, so `forma run ./model --temp 0.7` is refused rather than
run with the flag ignored.

## Flags every model command takes

`run`, `info`, `serve` and `bench` open a model directory and share three
flags:

| flag | default | meaning |
| --- | --- | --- |
| `--device` | `auto` | `auto`, `cpu` or `metal`. `auto` takes Metal where it exists and the CPU backend otherwise. Naming `metal` on a machine without it is an error, not a fallback |
| `--precision` | `auto` | `f16`, `int8`, `int4` or `auto`. `auto` takes the widest form that fits the device and prints what it chose, with the footprints it compared. It chooses int4 only when int8 does not fit |
| `--context` | `4096` | the key/value cache capacity in positions. A request that does not fit is refused, never truncated |

A model directory is a Hugging Face checkpoint: `config.json`, the tokenizer
files, and safetensors weights.

## forma pull

Downloads a checkpoint from Hugging Face and prints the directory it landed
in. Progress goes to stderr and the directory to stdout, so the output can be
captured:

```sh
MODEL=$(forma pull Qwen/Qwen3-0.6B)
```

| flag | default | meaning |
| --- | --- | --- |
| `--revision` | the repository's `main` | a branch, tag or commit |
| `--token` | `$HF_TOKEN`, then `$HUGGING_FACE_HUB_TOKEN` | an access token, for gated repositories |

Checkpoints are cached by the commit the revision resolved to:

```
<cache>/models/<org>/<repo>/<commit>/
```

`<cache>` is `$FORMA_CACHE` when set, else `$XDG_CACHE_HOME/forma`, else
`~/.cache/forma`. Two revisions of one repository sit side by side, and
pulling a moving branch never overwrites an earlier download. A download is
written under a temporary name and renamed when complete, so an interrupted
pull never leaves a file that looks whole, and the next pull resumes it.

## forma run

Generates from one prompt and streams the text to stdout. The model it loaded,
the precision, and the sampling policy are printed to stderr first, and a
token count and rate after.

```sh
forma run --prompt "The capital of France is" --max-tokens 12 "$MODEL"
```

| flag | default | meaning |
| --- | --- | --- |
| `--prompt` | a question about transformers | the prompt |
| `--raw` | off | send the prompt as typed, without the model's chat template |
| `--max-tokens` | `128` | stop after this many generated tokens |
| `--temp` | `0` | sampling temperature; `0` is greedy and deterministic |
| `--top-k`, `--top-p` | off | the usual truncation stages; `0` disables each |
| `--repeat-penalty` | `1` | divisive repetition penalty; `1` is none |
| `--seed` | `0` | the sampler seed. The same seed, prompt and policy give the same output on the same device and build |

## forma info

Prints the architecture, the precision `auto` would choose, and the memory the
weights and the key/value cache will take, without loading any weights. Use it
to decide on `--precision` and `--context` before committing a machine to a
model.

| flag | default | meaning |
| --- | --- | --- |
| `--budget` | what the device reports | the bytes the weights may occupy, to price a machine other than this one |

## forma serve

Serves the model over HTTP on `127.0.0.1:11434`. Before it listens it prints
the model id requests must name, the precision, and the memory arithmetic
behind how many requests it admits. [Serving](serving.md) covers the routes,
the request fields and the metrics.

| flag | default | meaning |
| --- | --- | --- |
| `--addr` | `127.0.0.1:11434` | where to listen |
| `--public` | off | allow an address that is not loopback. The server has no authentication, so this is never a default |
| `--slots` | 4 pooled, or 8 with `--batched` | how many requests generate at once. Fewer are taken if the device holds fewer |
| `--prefix-cache` | `off` | `off`, `session` or `process`. Bare `--prefix-cache` is `session`. See [prompt caching](orientation.md#prompt-caching) |
| `--kv` | slots times context | the size of the shared key/value pool, in positions. Needs `--prefix-cache process` |
| `--batched` | off | put every in-flight request in one forward pass. Implies `--prefix-cache process`. See [batching](orientation.md#running-several-conversations-in-one-pass) |

`--sessions` is the old name for `--slots` and still works.

## forma bench

Measures a model on a synthetic prompt and splits each decode step into four
terms: host (Forma's own code), submit (handing work to the device), device,
and readback (moving the logits to the host). The report goes to stdout as
Markdown, and `--json out.json` also writes a machine-readable record with
schema `forma.bench/1`. Every figure is printed beside the hardware, the Go
version, the accel version, the model, the precision and the sampling policy,
so two records can be compared.

| flag | default | meaning |
| --- | --- | --- |
| `--tokens` | `128` | decode steps to measure |
| `--prompt-tokens` | `128` | synthetic prompt length |
| `--warmup` | `8` | steps to run and discard first |
| `--json` | none | write the record here |
| `--batch` | `1` | sequences in flight. Only `1` is measured today, and any other value is refused rather than reported as a batch it did not run |

## Environment

| variable | read by | meaning |
| --- | --- | --- |
| `FORMA_CACHE` | `pull` | the checkpoint cache directory |
| `XDG_CACHE_HOME` | `pull` | the cache's parent when `FORMA_CACHE` is unset |
| `HF_TOKEN`, `HUGGING_FACE_HUB_TOKEN` | `pull` | a Hugging Face token when `--token` is not given |
