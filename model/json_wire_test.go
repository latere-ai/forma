// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package model

import "testing"

// TestConfigIsReadLeniently pins that a config.json written by another tool is
// read with a repeated member (the last wins) and with member names in any case.
func TestConfigIsReadLeniently(t *testing.T) {
	for _, raw := range []string{
		`{"architectures":["x"],"architectures":["y"]}`,
		`{"Architectures":["y"]}`,
	} {
		got, err := architecture([]byte(raw))
		if err != nil || got != "y" {
			t.Errorf("architecture(%s) = %q, %v, want y", raw, got, err)
		}
	}
}
