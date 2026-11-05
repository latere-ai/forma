// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package bench_test

import (
	"encoding/json/v2"
	"strings"
	"testing"
	"time"

	"latere.ai/x/forma/bench"
)

// TestQuantilesWireIsIntegerNanoseconds pins the bytes: durations are integer
// nanoseconds under their _ns names, in the order the struct declares them.
func TestQuantilesWireIsIntegerNanoseconds(t *testing.T) {
	q := bench.Quantiles{P50: 1500 * time.Nanosecond, P90: 2 * time.Microsecond, P99: time.Millisecond, N: 3}
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"p50_ns":1500,"p90_ns":2000,"p99_ns":1000000,"n":3}`
	if string(b) != want {
		t.Fatalf("got %s, want %s", b, want)
	}
	var back bench.Quantiles
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back != q {
		t.Fatalf("round trip = %+v, want %+v", back, q)
	}
}

// TestPhaseStatsSortsShareKeysWithoutOptions pins that the breakdown map is
// written in key order even when the caller did not ask for it.
func TestPhaseStatsSortsShareKeysWithoutOptions(t *testing.T) {
	p := bench.PhaseStats{ShareOfStep: map[string]float64{"submit": 0.25, "host": 0.25, "device": 0.25, "readback": 0.25}}
	for range 20 {
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		const want = `"share_of_step":{"device":0.25,"host":0.25,"readback":0.25,"submit":0.25}`
		if !strings.Contains(string(b), want) {
			t.Fatalf("share_of_step is not sorted in %s", b)
		}
	}
}
