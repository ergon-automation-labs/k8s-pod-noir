package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"podnoir/internal/scenario"
	"podnoir/internal/settings"
)

// HTTP implements Provider using Anthropic, OpenAI, or Ollama HTTP APIs.
type HTTP struct {
	kind     string
	baseURL  string
	apiKey   string
	model    string
	client   *http.Client
	mock     Mock
	fallback bool
}

// NewHTTP builds an HTTP-backed provider from settings (see settings.FromEnv).
func NewHTTP(s settings.Settings) (*HTTP, error) {
	kind := s.LLMProvider
	base := strings.TrimRight(s.LLMBaseURL, "/")
	apiKey := s.LLMAPIKey
	model := s.LLMModel

	switch kind {
	case "anthropic":
		if base == "" {
			base = "https://api.anthropic.com"
		}
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
		if apiKey == "" {
			return nil, fmt.Errorf("POD_NOIR_LLM_API_KEY is required for anthropic")
		}
	case "openai":
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		if model == "" {
			model = "gpt-4o-mini"
		}
		if apiKey == "" {
			return nil, fmt.Errorf("POD_NOIR_LLM_API_KEY is required for openai")
		}
	case "ollama":
		if base == "" {
			base = "http://127.0.0.1:11434"
		}
		if model == "" {
			model = "llama3.2"
		}
	default:
		return nil, fmt.Errorf("unsupported HTTP LLM kind %q", kind)
	}

	return &HTTP{
		kind:     kind,
		baseURL:  base,
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 120 * time.Second},
		fallback: s.LLMFallbackMock,
	}, nil
}

func (h *HTTP) EvaluateAccusation(ctx context.Context, def *scenario.Definition, hypothesis string) (AccuseResult, error) {
	prompt := buildAccusePrompt(def, hypothesis)
	text, err := h.complete(ctx, prompt)
	if err != nil && h.fallback {
		fmt.Fprintf(os.Stderr, "pod-noir: LLM HTTP error (accuse): %v — using mock\n", err)
		return h.mock.EvaluateAccusation(ctx, def, hypothesis)
	}
	if err != nil {
		return AccuseResult{}, err
	}
	res, perr := parseAccuseJSON(text)
	if perr != nil && h.fallback {
		fmt.Fprintf(os.Stderr, "pod-noir: LLM parse error (accuse): %v — using mock\n", perr)
		return h.mock.EvaluateAccusation(ctx, def, hypothesis)
	}
	if perr != nil {
		return AccuseResult{}, perr
	}
	return res, nil
}

func (h *HTTP) Debrief(ctx context.Context, def *scenario.Definition) (string, error) {
	prompt := buildDebriefPrompt(def)
	text, err := h.complete(ctx, prompt)
	if err != nil && h.fallback {
		fmt.Fprintf(os.Stderr, "pod-noir: LLM HTTP error (debrief): %v — using mock\n", err)
		return h.mock.Debrief(ctx, def)
	}
	if err != nil {
		return "", err
	}
	out := strings.TrimSpace(text)
	if out == "" && h.fallback {
		return h.mock.Debrief(ctx, def)
	}
	return out, nil
}

func buildAccusePrompt(def *scenario.Definition, hypothesis string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are the incident commander in a Kubernetes noir training game.\n")
	fmt.Fprintf(&b, "Scenario ID: %s\nTitle: %s\n", def.ID, def.Title)
	fmt.Fprintf(&b, "Teaching hints (do not quote verbatim; use to judge depth): hot keywords: %v; warm: %v\n", def.HotHints, def.WarmHints)
	fmt.Fprintf(&b, "Player hypothesis: %q\n\n", hypothesis)
	fmt.Fprintf(&b, "Reply with JSON ONLY, no markdown fences, shape:\n")
	fmt.Fprintf(&b, `{"judgment":"hot|warm|cold|stone_cold","reply":"<short in-character feedback>"}`+"\n")
	fmt.Fprintf(&b, "Judgment: hot=cause nailed; warm=right layer; cold=weak; stone_cold=empty/wrong.\n")
	return b.String()
}

func buildDebriefPrompt(def *scenario.Definition) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Write a case debrief for Kubernetes training scenario %s (%s).\n", def.ID, def.Title)
	fmt.Fprintf(&b, "Noir tone; ASCII box optional. Teach the real root cause and give 1–3 valid kubectl fix paths.\n")
	fmt.Fprintf(&b, "Themes: case-001 bad rollout / missing file; case-002 missing Secret / secretKeyRef; case-003 bad image tag / ImagePullBackOff.\n")
	fmt.Fprintf(&b, "Namespace for examples: %s. Keep under ~40 lines.\n", def.Namespace)
	return b.String()
}

func parseAccuseJSON(raw string) (AccuseResult, error) {
	js := extractJSONObject(raw)
	var out struct {
		Judgment string `json:"judgment"`
		Reply    string `json:"reply"`
	}
	if err := json.Unmarshal([]byte(js), &out); err != nil {
		return AccuseResult{}, fmt.Errorf("parse LLM JSON: %w (raw: %s)", err, truncate(raw, 400))
	}
	j := normalizeJudgment(out.Judgment)
	if out.Reply == "" {
		out.Reply = "No reply text from model."
	}
	return AccuseResult{Judgment: j, Reply: out.Reply}, nil
}

func normalizeJudgment(s string) Judgment {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "hot":
		return Hot
	case "warm":
		return Warm
	case "cold":
		return Cold
	case "stone_cold", "stonecold":
		return StoneCold
	default:
		return Cold
	}
}

func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 {
		s = s[:j+1]
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (h *HTTP) complete(ctx context.Context, userPrompt string) (string, error) {
	switch h.kind {
	case "anthropic":
		return h.anthropic(ctx, userPrompt)
	case "openai":
		return h.openai(ctx, userPrompt)
	case "ollama":
		return h.ollama(ctx, userPrompt)
	default:
		return "", fmt.Errorf("unknown kind %q", h.kind)
	}
}

func (h *HTTP) anthropic(ctx context.Context, userPrompt string) (string, error) {
	body := map[string]any{
		"model":      h.model,
		"max_tokens": 2048,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+"/v1/messages", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", h.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("anthropic %s: %s", resp.Status, string(raw))
	}
	var out struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Content) == 0 {
		return "", fmt.Errorf("anthropic: empty content")
	}
	return out.Content[0].Text, nil
}

func (h *HTTP) openai(ctx context.Context, userPrompt string) (string, error) {
	url := h.baseURL + "/chat/completions"
	body := map[string]any{
		"model": h.model,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.apiKey)
	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai %s: %s", resp.Status, string(raw))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai: empty choices")
	}
	return out.Choices[0].Message.Content, nil
}

func (h *HTTP) ollama(ctx context.Context, userPrompt string) (string, error) {
	url := h.baseURL + "/api/chat"
	body := map[string]any{
		"model": h.model,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
		"stream": false,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama %s: %s", resp.Status, string(raw))
	}
	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return out.Message.Content, nil
}
