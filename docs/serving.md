# Serving

`forma serve <model-dir>` serves one model over HTTP. This page covers the
routes, the request fields Forma honors, how requests are admitted, and what an
operator can watch. [Commands](commands.md#forma-serve) lists the flags.

## Binding and access

The server listens on `127.0.0.1:11434`, loopback only. `--addr` moves it, and
an address reachable from the network also needs `--public`, because the
server has **no authentication and no rate limiting**. Put something in front
of it that does both before exposing it.

It serves one model and does no routing. Every request names that model by its
id, the last element of the model directory, which is printed at startup and
answered by `GET /v1/models`. A request naming any other model is a 404.

## Routes

| route | speaks |
| --- | --- |
| `POST /v1/chat/completions` | OpenAI Chat Completions |
| `POST /v1/messages` | Anthropic Messages |
| `POST /v1/responses` | OpenAI Responses |
| `POST /v1/completions` | OpenAI legacy completions: the prompt reaches the model as written, with no chat template |
| `GET /v1/models` | the one model id this process serves |
| `GET /livez`, `GET /readyz` | liveness and readiness. A loaded model is a ready one |
| `GET /version` | the build: module version, commit and build time |
| `GET /metrics` | Prometheus text exposition |

The three chat dialects are decoded into one neutral request and the answer is
encoded back in the dialect the client spoke, so a feature behaves the same on
every route. All four stream with `"stream": true`. Errors come back in each
dialect's own error shape, including a failure in the middle of a stream.

## What a request can set

| setting | wire fields |
| --- | --- |
| temperature, top-p | `temperature`, `top_p` |
| top-k | `top_k` |
| penalties | `repetition_penalty`, `presence_penalty`, `frequency_penalty`, `penalty_window` |
| logit bias | `logit_bias`, keyed by token id. An id outside the vocabulary is refused |
| seed | `seed`. The same seed, prompt and policy reproduce the same completion on the same device and build |
| length | `max_tokens`, `max_completion_tokens` or `max_output_tokens`, per dialect |
| stop sequences | `stop` or `stop_sequences`, per dialect |
| JSON schema | `response_format` (Chat Completions), `output_format` (Messages), `text.format` (Responses) |
| log probabilities | `logprobs` and `top_logprobs` on Chat Completions; `logprobs: N` on legacy completions |
| thinking | Qwen3 thinks before it answers. `thinking: {"type": "disabled"}`, `reasoning_effort: "none"` or `reasoning: {"effort": "none"}` turns it off |
| cache isolation | `cache_salt`, see below |

Tool definitions and tool calls are passed through the model's chat template.

### Refused, or accepted and reported

Forma never drops a request field silently. A field that would change the
answer and cannot be honored is refused by name with a 400: more than one
completion (`n` above 1), an image, a `logit_bias` id outside the vocabulary,
or a JSON schema keyword the constraint compiler cannot express. A field that
cannot change the answer, such as a caller identity or a cache breakpoint, is
accepted and named in the `X-Forma-Loss` response header, and counted in
`forma_request_loss_total`.

### JSON schema output

A schema is compiled into a constraint applied at every step: a token that
could not continue a document matching the schema gets probability zero, so the
output parses and matches without a retry loop. A keyword the compiler cannot
turn into a constraint, such as `minimum`, is refused with the keyword and the
reason when the request arrives, not discovered in the output. Three
narrowings are deliberate: properties are emitted in the schema's order, objects
are closed, and `integer` admits the plain spelling. A number's magnitude is
not constrained, so check bounds after decoding.

## Prompt caching and cache_salt

With `--prefix-cache`, a request that begins with tokens the server already
scored reuses that work and prefills only what is new. `session` scope reuses
within one conversation; `process` scope shares across conversations that send
the same `cache_salt`, so their common system prompt is prefilled once. A reused prefix gives the same answer
in distribution, not byte for byte, because floating point addition is not
associative. [Orientation](orientation.md#prompt-caching) explains the scopes
and what each costs.

A cache hit is observable in timing: a request whose prompt was cached answers
faster, so a caller who can share another caller's cache can test what that
caller sent. `cache_salt` is the boundary. A request carrying a salt reuses
only work done for requests carrying the same salt.

What a request carrying no salt shares depends on the scope:

- **`process`, on either engine.** The request is given a random salt of its
  own and shares with nothing, not even the earlier turns of its own
  conversation. `--batched` always runs in this scope. A client that wants its
  next turn to reuse the conversation sends a salt, and a single-tenant
  deployment that wants its requests to share a system prompt sends the same
  salt on every request.
- **`session`, pooled.** Unsalted requests share the pool's sessions with each
  other: a request can reuse the cache an earlier unsalted request left in its
  session. That is how a conversation's next turn reuses its own work without a
  salt, and it also means one unsalted caller's hit can reveal what another
  sent.

Anything that multiplexes several tenants through one server should set a salt
per tenant.

## Admission

`--slots` requests generate at once. Behind them up to 32 wait in a queue; a
request that waits 30 seconds, or arrives to a full queue, is refused with 429
and a `Retry-After` header. Nothing queues without bound.

How the slots are served depends on the engine:

- **Pooled, the default.** Each slot is a session that keeps its own key/value
  cache for the life of the process, so every slot reserves a whole context of
  memory at startup. Concurrent requests interleave, and total throughput is
  close to what one conversation gets.
- **Batched, with `--batched`.** Every in-flight request is in one forward
  pass, so the weights are read once for all of them. Slots draw from one
  shared key/value pool (`--kv`), and admission reserves room for a request's
  prompt and part of its answer together, so an admitted request can finish.

At startup the server prints what it reserved, the memory budget it divided,
and the admission limit that came out, so a small limit reads as arithmetic
rather than a bug. Ctrl-C stops accepting requests and lets in-flight ones
finish.

## Metrics

`GET /metrics` exposes:

| metric | type | meaning |
| --- | --- | --- |
| `forma_requests_in_flight{dialect}` | gauge | requests generating now |
| `forma_queue_depth` | gauge | requests waiting for a slot |
| `forma_queue_wait_seconds` | histogram | time spent waiting for a slot |
| `forma_decode_step_seconds` | histogram | each request's median decode step |
| `forma_logits_readback_seconds` | histogram | each request's median time moving logits to the host |
| `forma_request_loss_total{field}` | counter | advisory fields accepted and not acted on |
| `forma_sessions_rejected_total{reason}` | counter | requests refused rather than run |

`GET /healthz` and `GET /health` are older names for the probes, kept for one
release while callers move.
