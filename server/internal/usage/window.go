package usage

import (
	"fmt"
	"math/big"
	"time"

	"on-my-interview/server/internal/llm"
)

const AsiaShanghai = "Asia/Shanghai"

type Window struct {
	StartAt  time.Time
	EndAt    time.Time
	Timezone string
}

func WindowForTime(now time.Time) (Window, error) {
	location, err := time.LoadLocation(AsiaShanghai)
	if err != nil {
		return Window{}, err
	}

	localNow := now.In(location)
	startHour := 0
	if localNow.Hour() >= 12 {
		startHour = 12
	}

	localStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), startHour, 0, 0, 0, location)
	localEnd := localStart.Add(12 * time.Hour)

	return Window{
		StartAt:  localStart.UTC(),
		EndAt:    localEnd.UTC(),
		Timezone: AsiaShanghai,
	}, nil
}

func CalculateDeepSeekChatCostCNY(usage llm.Usage) (string, error) {
	if usage.PromptCacheHitTokens < 0 || usage.PromptCacheMissTokens < 0 || usage.CompletionTokens < 0 {
		return "", fmt.Errorf("usage tokens cannot be negative")
	}

	total := new(big.Rat)
	total.Add(total, perMillion(usage.PromptCacheHitTokens, "0.2"))
	total.Add(total, perMillion(usage.PromptCacheMissTokens, "2"))
	total.Add(total, perMillion(usage.CompletionTokens, "3"))
	return total.FloatString(8), nil
}

func perMillion(tokens int64, price string) *big.Rat {
	rate, _ := new(big.Rat).SetString(price)
	tokenCount := big.NewRat(tokens, 1)
	value := new(big.Rat).Mul(tokenCount, rate)
	return value.Quo(value, big.NewRat(1_000_000, 1))
}
