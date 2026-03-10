package usage

import (
	"testing"
	"time"

	"on-my-interview/server/internal/llm"
)

func TestWindowForTimeUsesAsiaShanghaiHalfDays(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 3, 30, 0, 0, time.UTC)

	window, err := WindowForTime(now)
	if err != nil {
		t.Fatalf("WindowForTime: %v", err)
	}

	wantStart := time.Date(2026, 3, 10, 16, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 3, 11, 4, 0, 0, 0, time.UTC)

	if !window.StartAt.Equal(wantStart) {
		t.Fatalf("expected start %s, got %s", wantStart, window.StartAt)
	}
	if !window.EndAt.Equal(wantEnd) {
		t.Fatalf("expected end %s, got %s", wantEnd, window.EndAt)
	}
	if window.Timezone != "Asia/Shanghai" {
		t.Fatalf("unexpected timezone %q", window.Timezone)
	}
}

func TestWindowForTimeUsesNoonBoundary(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 4, 0, 0, 0, time.UTC)

	window, err := WindowForTime(now)
	if err != nil {
		t.Fatalf("WindowForTime: %v", err)
	}

	wantStart := time.Date(2026, 3, 11, 4, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 3, 11, 16, 0, 0, 0, time.UTC)

	if !window.StartAt.Equal(wantStart) {
		t.Fatalf("expected start %s, got %s", wantStart, window.StartAt)
	}
	if !window.EndAt.Equal(wantEnd) {
		t.Fatalf("expected end %s, got %s", wantEnd, window.EndAt)
	}
}

func TestCalculateDeepSeekChatCostCNY(t *testing.T) {
	t.Parallel()

	cost, err := CalculateDeepSeekChatCostCNY(llm.Usage{
		PromptTokens:          230000,
		PromptCacheHitTokens:  100000,
		PromptCacheMissTokens: 130000,
		CompletionTokens:      50000,
		TotalTokens:           280000,
	})
	if err != nil {
		t.Fatalf("CalculateDeepSeekChatCostCNY: %v", err)
	}

	if cost != "0.43000000" {
		t.Fatalf("expected 0.43000000, got %s", cost)
	}
}
