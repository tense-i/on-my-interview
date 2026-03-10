package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr           string
	MySQLDSN           string
	NowCoderBaseURL    string
	LLMBaseURL         string
	LLMAPIKey          string
	LLMModel           string
	SchedulerEnabled   bool
	SchedulerInterval  time.Duration
	SchedulerPlatforms []string
	SchedulerKeywords  []string
	SchedulerPages     int
}

func LoadFromEnv() Config {
	loadDotEnv()

	return Config{
		HTTPAddr:           getenv("INTERVIEW_CRAWLER_HTTP_ADDR", ":8080"),
		MySQLDSN:           os.Getenv("INTERVIEW_CRAWLER_MYSQL_DSN"),
		NowCoderBaseURL:    getenv("INTERVIEW_CRAWLER_NOWCODER_BASE_URL", "https://gw-c.nowcoder.com"),
		LLMBaseURL:         getenv("INTERVIEW_CRAWLER_LLM_BASE_URL", "https://api.openai.com"),
		LLMAPIKey:          os.Getenv("INTERVIEW_CRAWLER_LLM_API_KEY"),
		LLMModel:           getenv("INTERVIEW_CRAWLER_LLM_MODEL", "gpt-4.1-mini"),
		SchedulerEnabled:   parseBool(getenv("INTERVIEW_CRAWLER_SCHEDULE_ENABLED", "false")),
		SchedulerInterval:  parseDuration(getenv("INTERVIEW_CRAWLER_SCHEDULE_INTERVAL", "1h")),
		SchedulerPlatforms: parseCSV(getenv("INTERVIEW_CRAWLER_SCHEDULE_PLATFORMS", "nowcoder")),
		SchedulerKeywords:  parseCSV(getenv("INTERVIEW_CRAWLER_SCHEDULE_KEYWORDS", "面经")),
		SchedulerPages:     parseInt(getenv("INTERVIEW_CRAWLER_SCHEDULE_PAGES", "1"), 1),
	}
}

func loadDotEnv() {
	path, err := findDotEnv()
	if err != nil {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if after, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(after)
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
	}
}

func findDotEnv() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, ".env")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return parsed
}

func parseDuration(value string) time.Duration {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return time.Hour
	}
	return parsed
}

func parseCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
