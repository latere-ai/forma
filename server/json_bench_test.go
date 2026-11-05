// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"latere.ai/x/pkg/llmdialect"
	"latere.ai/x/pkg/llmdialect/anthropic"
	"latere.ai/x/pkg/llmdialect/ir"
	"latere.ai/x/pkg/llmdialect/openaichat"
	"latere.ai/x/pkg/llmdialect/openairesp"
)

// These benchmarks measure the JSON cost of the request path: the body decode
// every route runs before a session is built, the error body, and the response
// and stream frames each dialect writes. They drive the real handlers over the
// scripted engine in fake_test.go, so no device or weights are involved.

// benchSizes are the history sizes a chat request really carries: a short
// prompt, a working conversation, and a long agent transcript.
var benchSizes = []struct {
	name   string
	target int // approximate request body size in bytes
}{
	{"small", 512},
	{"medium", 10 << 10},
	{"large", 200 << 10},
}

// benchSentence is fixed prose, so a payload of a given size is the same bytes
// on every run.
const benchSentence = "The scheduler admits a request only when its KV reservation fits, " +
	"otherwise it queues the request and reports the wait in the response headers. "

// benchText returns n bytes of deterministic prose.
func benchText(n int) string {
	var b strings.Builder
	for b.Len() < n {
		b.WriteString(benchSentence)
	}
	return b.String()[:n]
}

// benchTurns is how many messages a body of the given size is split into.
func benchTurns(target int) int {
	return max(1, target/400)
}

// openAIBody builds a chat completions request of about target bytes.
func openAIBody(target int) []byte {
	turns := benchTurns(target)
	per := max(32, target/turns-60)
	var b strings.Builder
	fmt.Fprintf(&b, `{"model":%q,"max_tokens":256,"temperature":0.2,"top_p":0.9,"seed":7,"stream":false,"messages":[`, fakeName)
	b.WriteString(`{"role":"system","content":"You are a terse assistant."}`)
	for i := range turns {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		fmt.Fprintf(&b, `,{"role":%q,"content":%q}`, role, benchText(per))
	}
	b.WriteString(`,{"role":"user","content":"Summarize the above."}]}`)
	return []byte(b.String())
}

// anthropicBody builds a messages request of about target bytes, with the
// block-array content form and a tool definition.
func anthropicBody(target int) []byte {
	turns := benchTurns(target)
	per := max(32, target/turns-90)
	var b strings.Builder
	fmt.Fprintf(&b, `{"model":%q,"max_tokens":256,"temperature":0.2,"system":"You are a terse assistant.",`, fakeName)
	b.WriteString(`"tools":[{"name":"lookup","description":"Look a record up by id.",` +
		`"input_schema":{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}}],`)
	b.WriteString(`"messages":[`)
	for i := range turns {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"role":%q,"content":[{"type":"text","text":%q}]}`, role, benchText(per))
	}
	b.WriteString(`,{"role":"user","content":"Summarize the above."}]}`)
	return []byte(b.String())
}

// newBenchServer builds a server over the scripted engine.
func newBenchServer(b *testing.B, eng *fakeEngine) *Server {
	b.Helper()
	s, err := New(eng, WithNotice(&strings.Builder{}))
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	return s
}

// forget drops the sessions the fake engine records, so a long run does not
// grow the engine's bookkeeping and charge it to the request.
func forget(e *fakeEngine) {
	e.mu.Lock()
	e.sessions = nil
	e.mu.Unlock()
}

func benchDecode(b *testing.B, front llmdialect.Frontend, body []byte) {
	s := newBenchServer(b, &fakeEngine{script: text("ok"), prompt: 4})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		r.Body = io.NopCloser(bytes.NewReader(body))
		if _, aerr := s.decode(front, w, r); aerr != nil {
			b.Fatalf("decode: %v", aerr.msg)
		}
	}
}

// BenchmarkJSONChatCompletionsRequestDecode is the body decode of
// POST /v1/chat/completions: the top-level split, the raw-member refusals, the
// dialect frontend's parse, the extras parse and the IR-to-chat mapping.
func BenchmarkJSONChatCompletionsRequestDecode(b *testing.B) {
	for _, sz := range benchSizes {
		b.Run(sz.name, func(b *testing.B) {
			benchDecode(b, openaichat.NewFrontend(), openAIBody(sz.target))
		})
	}
}

// BenchmarkJSONMessagesRequestDecode is the same decode for POST /v1/messages,
// with block-array content and a tool definition.
func BenchmarkJSONMessagesRequestDecode(b *testing.B) {
	for _, sz := range benchSizes {
		b.Run(sz.name, func(b *testing.B) {
			benchDecode(b, anthropic.NewFrontend(), anthropicBody(sz.target))
		})
	}
}

// BenchmarkJSONResponsesRequestDecode is the body decode of POST /v1/responses.
func BenchmarkJSONResponsesRequestDecode(b *testing.B) {
	body := []byte(`{"model":"` + fakeName + `","max_output_tokens":256,"temperature":0.2,` +
		`"instructions":"You are a terse assistant.","input":[` +
		`{"role":"user","content":[{"type":"input_text","text":` + fmt.Sprintf("%q", benchText(400)) + `}]}]}`)
	benchDecode(b, openairesp.NewFrontend(), body)
}

// BenchmarkJSONCompletionsRequestDecode is the body decode of the legacy
// POST /v1/completions route, which parses the body once more for its prompt
// and stop members.
func BenchmarkJSONCompletionsRequestDecode(b *testing.B) {
	body := []byte(`{"model":"` + fakeName + `","max_tokens":128,"temperature":0.7,` +
		`"stop":["\n\n","END"],"prompt":` + fmt.Sprintf("%q", benchText(450)) + `}`)
	benchDecode(b, newLegacyFrontend(), body)
}

func benchServe(b *testing.B, route int, extra string, eng *fakeEngine) {
	s := newBenchServer(b, eng)
	rt := routes[route]
	body := []byte(rt.body(extra))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, rt.path, nil)
	// One run outside the loop fixes the size of what a client receives.
	r.Body = io.NopCloser(bytes.NewReader(body))
	s.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		b.Fatalf("%s: status %d: %s", rt.name, w.Code, w.Body.String())
	}
	out := w.Body.Len()
	b.ReportAllocs()
	b.SetBytes(int64(out))
	for b.Loop() {
		w = httptest.NewRecorder()
		r.Body = io.NopCloser(bytes.NewReader(body))
		s.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			b.Fatalf("%s: status %d", rt.name, w.Code)
		}
		forget(eng)
	}
}

// words returns n short scripted deltas, about six bytes each.
func words(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf(" w%04d", i%10000)
	}
	return out
}

// BenchmarkJSONNonStreamingResponseRoundTrip serves a whole non-streaming
// request on each route: body decode, the scripted generation, and the
// dialect's response encode. SetBytes counts the response a client reads.
func BenchmarkJSONNonStreamingResponseRoundTrip(b *testing.B) {
	for _, n := range []struct {
		name  string
		words int
	}{{"small", 40}, {"medium", 1700}} {
		for i, rt := range routes {
			b.Run(n.name+"/"+rt.name, func(b *testing.B) {
				benchServe(b, i, "", &fakeEngine{script: text(words(n.words)...), prompt: 12})
			})
		}
	}
}

// BenchmarkJSONStreamingResponseEncode serves a streaming request on each
// route. Every scripted delta becomes one SSE frame, each with its own JSON
// body, which is the per-token cost of a streamed completion.
func BenchmarkJSONStreamingResponseEncode(b *testing.B) {
	for _, n := range []struct {
		name   string
		deltas int
	}{{"small", 4}, {"medium", 50}, {"large", 1700}} {
		for i, rt := range routes {
			b.Run(n.name+"/"+rt.name, func(b *testing.B) {
				benchServe(b, i, `,"stream":true`,
					&fakeEngine{script: text(words(n.deltas)...), prompt: 12})
			})
		}
	}
}

// BenchmarkJSONAPIErrorBodyMarshal is the body of a refusal in each dialect,
// which every 4xx and 5xx answer writes.
func BenchmarkJSONAPIErrorBodyMarshal(b *testing.B) {
	for _, d := range []ir.Dialect{ir.DialectOpenAIChat, ir.DialectAnthropicMessages} {
		b.Run(string(d), func(b *testing.B) {
			e := badRequest("forma: the request asked for %q and this server serves %q", "other-model", fakeName)
			e.field = "model"
			n := len(e.body(d))
			b.ReportAllocs()
			b.SetBytes(int64(n))
			for b.Loop() {
				if len(e.body(d)) == 0 {
					b.Fatal("empty error body")
				}
			}
		})
	}
}

// BenchmarkJSONModelsListEncode is GET /v1/models, which load balancers and
// SDK clients poll, through writeJSON.
func BenchmarkJSONModelsListEncode(b *testing.B) {
	s := newBenchServer(b, &fakeEngine{script: text("ok"), prompt: 4})
	r := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		b.Fatalf("status %d", w.Code)
	}
	b.ReportAllocs()
	b.SetBytes(int64(w.Body.Len()))
	for b.Loop() {
		w = httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			b.Fatalf("status %d", w.Code)
		}
	}
}
