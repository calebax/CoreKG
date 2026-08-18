package extractor

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/webfetch/models/domain"
)

func BenchmarkHTMLExtractor(b *testing.B) {
	extractor, err := NewHTMLExtractor()
	if err != nil {
		b.Fatal(err)
	}
	for _, size := range []struct {
		name  string
		bytes int
	}{
		{name: "lower-4KiB", bytes: 4 << 10},
		{name: "typical-64KiB", bytes: 64 << 10},
		{name: "upper-1MiB", bytes: 1 << 20},
	} {
		body := benchmarkArticle(size.bytes)
		resource := domain.Resource{URL: "https://go.dev/blog/context", FinalURL: "https://go.dev/blog/context", ContentType: "text/html", Body: body}
		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(body)))
			b.ResetTimer()
			for iteration := 0; iteration < b.N; iteration++ {
				if _, extractErr := extractor.Extract(context.Background(), resource); extractErr != nil {
					b.Fatal(extractErr)
				}
			}
		})
	}
}

func benchmarkArticle(targetBytes int) []byte {
	const paragraph = `<p>这是用于性能测试的真实文章段落结构，包含自然语言正文、<a href="/reference">参考链接</a>以及用于正文提取和清洗的标点内容。</p>`
	count := targetBytes/len(paragraph) + 1
	body := fmt.Sprintf(`<!doctype html><html lang="zh-CN"><head><title>Benchmark Article</title><meta name="author" content="CoreKG"></head><body><nav>首页 产品 登录</nav><main><article><h1>性能测试文章</h1>%s</article></main><script>window.secret="removed";</script></body></html>`, strings.Repeat(paragraph, count))
	return []byte(body)
}
