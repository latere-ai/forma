// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package conformance

import (
	"encoding/json/v2"
	"testing"
	"time"
)

// TestDurationsAreIntegerNanoseconds pins the bytes of the two types that hold
// a time.Duration, and that they read back.
func TestDurationsAreIntegerNanoseconds(t *testing.T) {
	r := Readback{Bytes: 4, Share: 0.5, Median: 1500 * time.Nanosecond, Steps: 7}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"vocab":0,"bytes":4,"share":0.5,"median_ns":1500,"steps":7}`; string(b) != want {
		t.Fatalf("Readback = %s, want %s", b, want)
	}
	var rb Readback
	if err := json.Unmarshal(b, &rb); err != nil || rb != r {
		t.Fatalf("Readback round trip = %+v, %v", rb, err)
	}

	k := Bucket{Tokens: 32, Elapsed: 2 * time.Millisecond}
	b, err = json.Marshal(k)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"tokens":32,"elapsed_ns":2000000}`; string(b) != want {
		t.Fatalf("Bucket = %s, want %s", b, want)
	}
	var kb Bucket
	if err := json.Unmarshal(b, &kb); err != nil || kb != k {
		t.Fatalf("Bucket round trip = %+v, %v", kb, err)
	}
}
