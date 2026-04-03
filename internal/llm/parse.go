package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxAccuseReplyRunes = 3500

func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(s), "json") {
		s = strings.TrimSpace(s[4:])
	}
	if idx := strings.Index(s, "```"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

// extractJSONObject returns the first balanced {...} slice, or falls back to brace trimming.
func extractJSONObject(s string) string {
	s = stripMarkdownFence(strings.TrimSpace(s))
	start := strings.Index(s, "{")
	if start < 0 {
		return s
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if escape {
			escape = false
			continue
		}
		if inString {
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	if j := strings.LastIndex(s, "}"); j > start {
		return s[start : j+1]
	}
	return s[start:]
}

func parseAccuseJSON(raw string) (AccuseResult, error) {
	js := extractJSONObject(raw)
	var out struct {
		Judgment string `json:"judgment"`
		Reply    string `json:"reply"`
	}
	if err := json.Unmarshal([]byte(js), &out); err != nil {
		return AccuseResult{}, fmt.Errorf("parse LLM JSON: %w (snippet: %s)", err, truncate(js, 280))
	}
	jNorm := strings.ToLower(strings.TrimSpace(out.Judgment))
	if jNorm == "" {
		return AccuseResult{}, fmt.Errorf("LLM JSON missing judgment (snippet: %s)", truncate(js, 280))
	}
	j := normalizeJudgment(out.Judgment)
	reply := strings.TrimSpace(out.Reply)
	if reply == "" {
		reply = "No reply text from model."
	}
	reply = clampRunes(reply, maxAccuseReplyRunes)
	return AccuseResult{Judgment: j, Reply: reply}, nil
}

func normalizeJudgment(s string) Judgment {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "hot":
		return Hot
	case "warm":
		return Warm
	case "cold":
		return Cold
	case "stone_cold", "stonecold", "stone-cold", "stone cold":
		return StoneCold
	default:
		return Cold
	}
}

func clampRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	n := utf8.RuneCountInString(s)
	if n <= max {
		return s
	}
	r := []rune(s)
	if len(r) > max {
		r = r[:max]
	}
	return strings.TrimRight(string(r), " \t\n") + "…"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
