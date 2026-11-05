// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"encoding/json/v2"
	"testing"
)

// TestMergesReadInBothSpellings pins the two merge spellings in circulation and
// that a repeated or differently cased member of the file is accepted.
func TestMergesReadInBothSpellings(t *testing.T) {
	var m modelJSON
	in := `{"Type":"BPE","type":"BPE","merges":["a b",["c","d"]]}`
	if err := json.Unmarshal([]byte(in), &m, foreignJSON); err != nil {
		t.Fatal(err)
	}
	if m.Type != "BPE" || len(m.Merges) != 2 || m.Merges[0] != (mergeJSON{"a", "b"}) || m.Merges[1] != (mergeJSON{"c", "d"}) {
		t.Fatalf("model = %+v", m)
	}
	var bad mergeJSON
	if err := json.Unmarshal([]byte(`"nospace"`), &bad, foreignJSON); err == nil {
		t.Error("a merge with no separating space was accepted")
	}
	if err := json.Unmarshal([]byte(`7`), &bad, foreignJSON); err == nil {
		t.Error("a number was accepted as a merge")
	}
}
