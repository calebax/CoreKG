package reranksearch

import (
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
)

func TestJoinNeighborChunksPreservesMatchedChunkReference(t *testing.T) {
	wrapper := &RerankSearchWrapper{
		conf: &SearchConfig{NeighborSize: 1},
	}
	matchedLocation := ragtypes.Location{44, 1, 2, 3, 4}
	leftLocation := ragtypes.Location{43, 1, 2, 3, 4}

	chunks := []*SearchType{
		{
			FileID:      11559,
			Sequence:    44,
			Description: "matched 2.4 income chunk",
			Location:    matchedLocation,
		},
	}
	neighbors := []*SearchType{
		{
			FileID:      11559,
			Sequence:    43,
			Description: "left 2.3 chunk",
			Location:    leftLocation,
		},
		{
			FileID:      11559,
			Sequence:    45,
			Description: "right income standard chunk",
		},
	}

	got := wrapper.JoinNeighborChunks(chunks, neighbors)
	if len(got) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(got))
	}
	if got[0].Sequence != 44 {
		t.Fatalf("sequence should keep the matched chunk, got %d", got[0].Sequence)
	}
	if got[0].Location != matchedLocation {
		t.Fatalf("location should keep the matched chunk, got %v", got[0].Location)
	}
	for _, want := range []string{"left 2.3 chunk", "matched 2.4 income chunk", "right income standard chunk"} {
		if !strings.Contains(got[0].Description, want) {
			t.Fatalf("joined description missing %q: %s", want, got[0].Description)
		}
	}
	matchedIndex := strings.Index(got[0].Description, "matched 2.4 income chunk")
	leftIndex := strings.Index(got[0].Description, "left 2.3 chunk")
	rightIndex := strings.Index(got[0].Description, "right income standard chunk")
	if !(matchedIndex >= 0 && leftIndex > matchedIndex && rightIndex > leftIndex) {
		t.Fatalf("joined description should be matched, left, right; got %s", got[0].Description)
	}
}

func TestBuildChunkWindowSequencesIncludesMatchedChunk(t *testing.T) {
	got := buildChunkWindowSequences(44, 2)
	want := []int{44, 45, 43, 46, 42}
	if len(got) != len(want) {
		t.Fatalf("expected %d sequences, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sequence[%d] = %d, want %d; all=%v", i, got[i], want[i], got)
		}
	}
}

func TestSelectKeywordFallbackChunksPrefersKeywordEvidenceOverRerankScore(t *testing.T) {
	chunks := []*SearchType{
		{
			FileID:      11559,
			Sequence:    8,
			Description: "2.9.1 超标准缴存住房公积金 2.9.2 违规承担职工个人应承担的部分支出",
			rawContent:  "2.9.1 超标准缴存住房公积金 2.9.2 违规承担职工个人应承担的部分支出",
			Score:       9.0,
			RerankScore: 0.44,
		},
		{
			FileID:      11559,
			Sequence:    44,
			Description: "## ◆2.4 未按规定确认收入 - 2.4.1 通过虚开商品销售发票虚增收入",
			rawContent:  "## ◆2.4 未按规定确认收入 - 2.4.1 通过虚开商品销售发票虚增收入",
			Score:       7.5,
			RerankScore: 0.04,
		},
		{
			FileID:      11559,
			Sequence:    46,
			Description: "第四十八条 企业发生下列情形之一... 常见表现 表现形式 按规定 未按",
			rawContent:  "第四十八条 企业发生下列情形之一，违反会计法和企业会计准则。",
			Score:       8.2,
			RerankScore: 0.12,
		},
	}

	got, candidates := selectKeywordFallbackChunksByKeywords(chunks, []string{"未按规定", "确认", "收入", "常见", "表现形式"}, 5)
	if len(got) != 1 {
		t.Fatalf("expected 1 keyword fallback chunk, got %d; candidates=%v", len(got), candidates)
	}
	if got[0].Sequence != 44 {
		t.Fatalf("selected sequence = %d, want 44", got[0].Sequence)
	}
}

func TestFilterKeywordFallbackTokensKeepsLongerTerms(t *testing.T) {
	got := filterKeywordFallbackTokens([]string{"收入", "确认收入", "规定", "规定"})
	want := []string{"确认收入", "规定"}
	if len(got) != len(want) {
		t.Fatalf("expected %d keywords, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("keyword[%d] = %q, want %q; all=%v", i, got[i], want[i], got)
		}
	}
}
