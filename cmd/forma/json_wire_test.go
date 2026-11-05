// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json/v2"
	"strings"
	"testing"
	"time"
)

// TestRecordDurationsAreIntegerNanoseconds pins the bytes of the types in the
// benchmark record that hold a time.Duration.
func TestRecordDurationsAreIntegerNanoseconds(t *testing.T) {
	cold := coldFacts{Open: 900 * time.Millisecond, FirstToken: 940 * time.Millisecond}
	b, err := json.Marshal(cold)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"open_ns":900000000,"first_token_ns":940000000}`; string(b) != want {
		t.Fatalf("coldFacts = %s, want %s", b, want)
	}
	var back coldFacts
	if err := json.Unmarshal(b, &back); err != nil || back != cold {
		t.Fatalf("coldFacts round trip = %+v, %v", back, err)
	}

	g := genResult{PromptTokens: 1, CompletionTokens: 2, TTFT: 30 * time.Millisecond, Stop: "eos"}
	b, err = json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"prompt_tokens":1,"completion_tokens":2,"ttft_ns":30000000,"stop":"eos"}`; string(b) != want {
		t.Fatalf("genResult = %s, want %s", b, want)
	}

	p := batchPoint{Batch: 1, Tokens: 12, Wall: 2 * time.Second}
	b, err = json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	const head = `{"batch":1,"cold":{"open_ns":0,"first_token_ns":0},"resident_bytes":0,"generated_tokens":12,"wall_ns":2000000000,`
	if !strings.HasPrefix(string(b), head) {
		t.Fatalf("batchPoint = %s, want prefix %s", b, head)
	}
	var pb batchPoint
	if err := json.Unmarshal(b, &pb); err != nil || pb.Wall != p.Wall || pb.Tokens != 12 {
		t.Fatalf("batchPoint round trip = %+v, %v", pb, err)
	}
}
