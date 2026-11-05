// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package chat

import (
	"fmt"
	"strings"
	"testing"
)

// benchToolSchema is one tool's input schema of realistic shape: a few typed
// properties with descriptions, an enum, a nested object and a required list.
func benchToolSchema(i int) string {
	return fmt.Sprintf(`{"type":"object","properties":{`+
		`"query":{"type":"string","description":"Free-text query for tool %[1]d; supports quoted phrases and field filters like owner:<name>."},`+
		`"limit":{"type":"integer","minimum":1,"maximum":100,"description":"Maximum number of results to return."},`+
		`"order":{"type":"string","enum":["relevance","created","updated"]},`+
		`"filter":{"type":"object","properties":{"since":{"type":"string","description":"RFC 3339 timestamp."},"labels":{"type":"array","items":{"type":"string"}}}}`+
		`},"required":["query"],"additionalProperties":false}`, i)
}

func benchTools(n int) []ToolSpec {
	tools := make([]ToolSpec, n)
	for i := range tools {
		tools[i] = ToolSpec{
			Name:        fmt.Sprintf("search_records_%03d", i),
			Description: fmt.Sprintf("Search the records store %d and return the best matches with their identifiers & scores; use before <create>.", i),
			InputSchema: []byte(benchToolSchema(i)),
		}
	}
	return tools
}

// benchConversation is a short agent exchange: a question, a tool call with
// verbatim arguments, its result and a final question.
func benchConversation() []Message {
	return []Message{
		system("You are a careful research assistant."),
		user("Find the open incidents for the billing service."),
		assistant("I should search first.", "Searching.", &ToolUse{
			ID: "call_0001", Name: "search_records_000",
			Args: []byte(`{"query":"service:billing state:open","limit":20,"order":"updated","filter":{"since":"2026-10-01T00:00:00Z","labels":["incident","p1"]}}`),
		}),
		toolResult("call_0001", strings.Repeat("INC-1042 billing latency regression, opened 2026-10-29. ", 6)),
		user("Which of those is oldest?"),
	}
}

// BenchmarkJSONQwen3RenderToolsEncode is the prompt render a tool-using
// request pays: every tool definition is quoted into the system turn and its
// schema and every call's arguments are checked as JSON.
func BenchmarkJSONQwen3RenderToolsEncode(b *testing.B) {
	for _, sz := range []struct {
		name  string
		tools int
	}{{"small", 1}, {"medium", 20}, {"large", 250}} {
		b.Run(sz.name, func(b *testing.B) {
			r := Qwen3()
			msgs := benchConversation()
			opts := Options{Thinking: true, AddGenerationHint: true, Tools: benchTools(sz.tools)}
			p, err := r.Render(msgs, opts)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.SetBytes(int64(len(p.String())))
			for b.Loop() {
				if _, err := r.Render(msgs, opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
