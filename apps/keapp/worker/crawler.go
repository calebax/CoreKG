package worker

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"golang.org/x/net/html"

	"github.com/insmtx/corekg/apps/keapp/models/web"
	"github.com/ygpkg/yg-go/logs"
)

type CrawlResult struct {
	NewCount     int
	UpdatedCount int
	SkippedCount int
	CrawledCount int
	Err          error
}

func Crawl(ctx context.Context, task *web.KeCrawlTask, rules []*web.KeWebCrawlRule, cancelCheckInterval int) *CrawlResult {
	result := &CrawlResult{}

	if task.SourceURL == "" {
		result.Err = fmt.Errorf("task %d has empty source_url", task.ID)
		return result
	}

	visited := make(map[string]bool)
	queue := []string{task.SourceURL}
	pagesProcessed := 0

	resDao := web.NewWebResourceDao()
	taskDao := web.NewCrawlTaskDao()

	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			result.Err = ctx.Err()
			return result
		default:
		}

		if cancelCheckInterval > 0 && pagesProcessed > 0 && pagesProcessed%cancelCheckInterval == 0 {
			fresh, err := taskDao.GetByID(ctx, task.ID)
			if err != nil {
				logs.ErrorContextf(ctx, "crawl task %d cancel check error: %v", task.ID, err)
			} else if fresh != nil && fresh.Status == web.CrawlTaskCancelled {
				logs.InfoContextf(ctx, "crawl task %d cancelled after %d pages", task.ID, pagesProcessed)
				return result
			}
		}

		currentURL := queue[0]
		queue = queue[1:]

		if visited[currentURL] {
			continue
		}
		visited[currentURL] = true

		if !matchRules(currentURL, rules) {
			result.SkippedCount++
			continue
		}

		rawHTML, title, links, fetchErr := fetchPage(ctx, currentURL)
		if fetchErr != nil {
			logs.ErrorContextf(ctx, "fetch page %s error: %v", currentURL, fetchErr)
			result.SkippedCount++
			pagesProcessed++
			continue
		}

		markdown, err := convertToMarkdown(rawHTML)
		if err != nil {
			logs.ErrorContextf(ctx, "convert to markdown %s error: %v", currentURL, err)
			result.SkippedCount++
			pagesProcessed++
			continue
		}

		hash := computeHash(markdown)

		existing, err := resDao.GetByURL(ctx, task.AppID, currentURL)
		if err != nil {
			logs.ErrorContextf(ctx, "query resource %s error: %v", currentURL, err)
			result.SkippedCount++
			pagesProcessed++
			continue
		}

		now := time.Now()
		metaBytes, _ := json.Marshal(map[string]string{"title": title})

		if existing != nil {
			if existing.ContentHash == hash {
				result.SkippedCount++
			} else {
				existing.RawContent = markdown
				existing.ContentHash = hash
				existing.Title = title
				existing.LastCrawlAt = &now
				existing.CrawlCount = existing.CrawlCount + 1
				existing.Metadata = metaBytes
				if updateErr := resDao.Update(ctx, existing); updateErr != nil {
					logs.ErrorContextf(ctx, "update resource %s error: %v", currentURL, updateErr)
					result.SkippedCount++
				} else {
					result.UpdatedCount++
				}
			}
		} else {
			entity := &web.KeWebResource{
				AppID:       task.AppID,
				SourceURL:   currentURL,
				Title:       title,
				RawContent:  markdown,
				ContentHash: hash,
				LastCrawlAt: &now,
				CrawlCount:  1,
				Metadata:    metaBytes,
				IndexStatus: web.IndexPending,
			}
			if insertErr := resDao.Insert(ctx, entity); insertErr != nil {
				logs.ErrorContextf(ctx, "insert resource %s error: %v", currentURL, insertErr)
				result.SkippedCount++
			} else {
				result.NewCount++
			}
		}

		result.CrawledCount++
		pagesProcessed++

		if task.TaskType == web.CrawlTaskSingle {
			break
		}

		for _, link := range links {
			abs := resolveURL(currentURL, link)
			if abs != "" && !visited[abs] && matchRules(abs, rules) {
				queue = append(queue, abs)
			}
		}
	}

	return result
}

func fetchPage(ctx context.Context, pageURL string) (rawHTML, title string, links []string, err error) {
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", "", nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KeCrawler/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return "", "", nil, fmt.Errorf("read body: %w", err)
	}

	rawHTML = string(bodyBytes)

	doc, parseErr := html.Parse(strings.NewReader(rawHTML))
	if parseErr == nil {
		title = extractTitle(doc)
		links = extractLinks(doc)
	}

	return rawHTML, title, links, nil
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := extractTitle(c); t != "" {
			return t
		}
	}
	return ""
}

func extractLinks(n *html.Node) []string {
	var links []string
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" && attr.Val != "" {
				links = append(links, attr.Val)
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, extractLinks(c)...)
	}
	return links
}

func convertToMarkdown(rawHTML string) (string, error) {
	md, err := htmltomarkdown.ConvertString(rawHTML)
	if err != nil {
		return "", fmt.Errorf("html to markdown: %w", err)
	}
	return strings.TrimSpace(md), nil
}

func computeHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}

func matchRules(pageURL string, rules []*web.KeWebCrawlRule) bool {
	if len(rules) == 0 {
		return true
	}

	for _, rule := range rules {
		if rule.RuleType == web.CrawlRuleExclude {
			if matchPattern(pageURL, rule.Pattern) {
				return false
			}
		}
	}

	hasInclude := false
	for _, rule := range rules {
		if rule.RuleType == web.CrawlRuleInclude {
			hasInclude = true
			if matchPattern(pageURL, rule.Pattern) {
				return true
			}
		}
	}

	return !hasInclude
}

func matchPattern(pageURL, pattern string) bool {
	if pattern == "" {
		return false
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(pageURL, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(pageURL, pattern[:len(pattern)-1])
	}
	return strings.Contains(pageURL, pattern)
}

func resolveURL(base, ref string) string {
	if ref == "" || strings.HasPrefix(ref, "javascript:") || strings.HasPrefix(ref, "mailto:") || strings.HasPrefix(ref, "#") {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	resolved := baseURL.ResolveReference(refURL)
	resolved.Fragment = ""
	u := resolved.String()

	ext := strings.ToLower(path.Ext(resolved.Path))
	if ext != "" && ext != ".html" && ext != ".htm" && ext != ".php" && ext != ".asp" && ext != ".aspx" && ext != ".jsp" && ext != "/" {
		return ""
	}

	return u
}
