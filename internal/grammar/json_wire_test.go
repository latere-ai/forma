// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package grammar

import (
	"strings"
	"testing"
)

// TestEncodeStringSpellsLiteralsWithoutEscapingHTML pins the literal bytes the
// automaton admits for a property name: angle brackets and ampersands stay as
// typed, and bytes that are not UTF-8 become U+FFFD instead of an error.
func TestEncodeStringSpellsLiteralsWithoutEscapingHTML(t *testing.T) {
	for in, want := range map[string]string{
		"a<b&c>": `"a<b&c>"`,
		"a\xffb": "\"a�b\"",
	} {
		got, err := encodeString(in)
		if err != nil || got != want {
			t.Errorf("encodeString(%q) = %q, %v, want %q", in, got, err, want)
		}
	}
}

// TestFieldsKeepsDocumentOrderAndNamesDuplicates pins that member order is the
// document's and that a repeated key is refused by name, not by a decoder error.
func TestFieldsKeepsDocumentOrderAndNamesDuplicates(t *testing.T) {
	order, m, err := fields([]byte(`{"b":1,"a":2}`), "")
	if err != nil || len(order) != 2 || order[0] != "b" || order[1] != "a" || string(m["a"]) != "2" {
		t.Fatalf("fields = %v, %v, %v", order, m, err)
	}
	_, _, err = fields([]byte(`{"a":1,"a":2}`), "")
	if err == nil || err.Error() == "" {
		t.Fatal("a repeated key was accepted")
	}
	if got := err.Error(); !strings.Contains(got, `duplicate key "a"`) {
		t.Fatalf("error = %q, want it to name the duplicate key", got)
	}
}
