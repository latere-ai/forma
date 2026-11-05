// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package safetensors

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfigMayRepeatAMember pins that a config.json with a repeated member is
// still accepted, and that one that is not JSON is not.
func TestConfigMayRepeatAMember(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, configName), []byte(`{"a":1,"a":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readConfig(dir); err != nil {
		t.Errorf("a repeated member was refused: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, configName), []byte(`{"a":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readConfig(dir); err == nil {
		t.Error("a truncated config was accepted")
	}
}
