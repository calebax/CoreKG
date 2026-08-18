package duckduckgo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	empty := readBenchmarkFixture(b, "empty.html")
	normal := readBenchmarkFixture(b, "normal.html")
	cases := []struct {
		name  string
		body  []byte
		limit int
	}{
		{name: "lower-empty", body: empty, limit: 10},
		{name: "typical-20-results", body: expandResults(normal, 20), limit: 20},
		{name: "upper-1000-results", body: expandResults(normal, 1000), limit: 20},
	}
	for _, test := range cases {
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(test.body)))
			b.ResetTimer()
			for iteration := 0; iteration < b.N; iteration++ {
				if _, err := Parse(test.body, test.limit); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func readBenchmarkFixture(b *testing.B, name string) []byte {
	b.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "duckduckgo", name))
	if err != nil {
		b.Fatal(err)
	}
	return body
}

func expandResults(base []byte, count int) []byte {
	const marker = "</div>\n  </body>"
	const result = `<div class="result"><h2 class="result__title"><a class="result__a" href="https://go.dev/doc/article/%d">Result %d</a></h2><div class="result__snippet">A representative search result snippet with enough text to exercise normalization.</div></div>`
	items := make([]string, count)
	for index := range items {
		items[index] = fmt.Sprintf(result, index, index)
	}
	return []byte(strings.Replace(string(base), marker, strings.Join(items, "")+marker, 1))
}
