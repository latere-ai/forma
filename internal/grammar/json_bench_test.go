// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package grammar

import (
	"fmt"
	"strings"
	"testing"
)

// benchSchema builds a response_format schema with n object properties, each a
// described string, integer, enum or array, so the compiler walks every
// keyword family it supports.
func benchSchema(n int) string {
	props := make([]string, 0, n)
	req := make([]string, 0, n)
	for i := range n {
		name := fmt.Sprintf("field_%03d", i)
		var p string
		switch i % 4 {
		case 0:
			p = `{"type":"string","description":"A short label for this field.","maxLength":40}`
		case 1:
			p = `{"type":"integer"}`
		case 2:
			p = `{"enum":["open","closed","blocked"]}`
		default:
			p = `{"type":"array","items":{"type":"string"},"maxItems":4}`
		}
		props = append(props, fmt.Sprintf("%q:%s", name, p))
		req = append(req, fmt.Sprintf("%q", name))
	}
	return `{"type":"object","properties":{` + strings.Join(props, ",") +
		`},"required":[` + strings.Join(req, ",") + `],"additionalProperties":false}`
}

// BenchmarkJSONSchemaCompileDecode is the schema parse and compile a structured
// output request pays before its first token: the JSON Schema is decoded and
// walked into the grammar the sampler is constrained by.
func BenchmarkJSONSchemaCompileDecode(b *testing.B) {
	for _, sz := range []struct {
		name  string
		props int
	}{{"small", 3}, {"medium", 12}} {
		b.Run(sz.name, func(b *testing.B) {
			schema := []byte(benchSchema(sz.props))
			v := full()
			if _, err := Compile(schema, v, Options{}); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.SetBytes(int64(len(schema)))
			for b.Loop() {
				if _, err := Compile(schema, v, Options{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
