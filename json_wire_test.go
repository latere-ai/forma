// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package forma

import "testing"

// TestTemplateConfigIsReadLeniently pins that a tokenizer_config.json written
// by another tool is read with a repeated member (the last wins) and with the
// member name in any letter case.
func TestTemplateConfigIsReadLeniently(t *testing.T) {
	for _, raw := range []string{
		`{"chat_template":"x","chat_template":"y"}`,
		`{"Chat_Template":"y"}`,
	} {
		got, err := checkpointTemplate([]byte(raw))
		if err != nil || got != "y" {
			t.Errorf("checkpointTemplate(%s) = %q, %v, want y", raw, got, err)
		}
	}
}
