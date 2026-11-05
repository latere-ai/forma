// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package chat

import (
	"encoding/json/jsontext"
	"testing"
)

// TestJSONStringWritesInvalidUTF8AsReplacement pins that a tool description
// that is not UTF-8 renders instead of collapsing to an empty string.
func TestJSONStringWritesInvalidUTF8AsReplacement(t *testing.T) {
	if got, want := jsonString("a\xffb<&>"), "\"a�b<&>\""; got != want {
		t.Errorf("jsonString = %q, want %q", got, want)
	}
}

// TestToolArgumentsMayRepeatAKey pins the leniency toward model-written
// arguments: a repeated key is still valid input that is copied through.
func TestToolArgumentsMayRepeatAKey(t *testing.T) {
	v := jsontext.Value(`{"a":1,"a":2}`)
	if v.IsValid() {
		t.Fatal("the default check accepts a repeated key; this test no longer pins anything")
	}
	if !v.IsValid(foreignJSON) {
		t.Error("a repeated key is refused")
	}
}
