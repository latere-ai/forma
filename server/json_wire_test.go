// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json/v2"
	"strings"
	"testing"

	"latere.ai/x/pkg/llmdialect/ir"
)

// TestRequestBodiesFromOtherClientsAreReadLeniently pins that a repeated member
// (the last wins) and differently cased member names are accepted.
func TestRequestBodiesFromOtherClientsAreReadLeniently(t *testing.T) {
	top, err := topLevel([]byte(`{"seed":1,"seed":2,"Top_K":5}`))
	if err != nil {
		t.Fatal(err)
	}
	seed, err := jsonUint(top, "seed")
	if err != nil || seed == nil || *seed != 2 {
		t.Fatalf("seed = %v, %v, want 2", seed, err)
	}
	var wire struct {
		TopK int `json:"top_k"`
	}
	if err := json.Unmarshal(top["Top_K"], &wire.TopK, foreignJSON); err != nil || wire.TopK != 5 {
		t.Fatalf("top_k = %d, %v", wire.TopK, err)
	}
	var cased struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal([]byte(`{"MODEL":"m"}`), &cased, foreignJSON); err != nil || cased.Model != "m" {
		t.Fatalf("model = %q, %v", cased.Model, err)
	}
}

// TestResponsesCarryModelBytesThatAreNotUTF8 pins that text a model produced
// is written with U+FFFD for invalid bytes and the body still encodes.
func TestResponsesCarryModelBytesThatAreNotUTF8(t *testing.T) {
	body, err := (&legacyFrontend{}).EncodeResponse(&ir.Response{
		Blocks: []ir.Block{{Type: ir.BlockText, Text: "a\xffb"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "\"text\":\"a�b\"") {
		t.Errorf("body = %s", body)
	}
}
