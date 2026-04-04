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
	repair   bool
	// contactWire enables LLM-generated hint sysadmin/network/archivist/senior (see ContactWire).
	contactWire bool
}

type completionMode int

const (
	modePlain completionMode = iota
	modeAccuseJSON
	modeDebrief
	modeContactWire
)

const accuseSystemPrompt = `You are a Kubernetes incident examiner in a noir training game.
You MUST respond with exactly one JSON object and nothing else: no markdown fences, no preamble, no trailing commentary.
Keys: "judgment" (string: hot | warm | cold | stone_cold) and "reply" (string: short in-character feedback to the detective, under 12 sentences).`

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
		kind:        kind,
		baseURL:     base,
		apiKey:      apiKey,
		model:       model,
		client:      &http.Client{Timeout: 120 * time.Second},
		fallback:    s.LLMFallbackMock,
		repair:      s.LLMRepairAccuse,
		contactWire: s.LLMContactWire,
	}, nil
}

func (h *HTTP) EvaluateAccusation(ctx context.Context, def *scenario.Definition, hypothesis string) (AccuseResult, error) {
	prompt := buildAccusePrompt(def, hypothesis)
	text, err := h.complete(ctx, prompt, modeAccuseJSON)
	if err != nil && h.fallback {
		fmt.Fprintf(os.Stderr, "pod-noir: LLM HTTP error (accuse): %v — using mock\n", err)
		return h.mock.EvaluateAccusation(ctx, def, hypothesis)
	}
	if err != nil {
		return AccuseResult{}, err
	}
	res, perr := parseAccuseJSON(text)
	if perr != nil && h.repair {
		fixPrompt := accuseRepairPrompt(prompt, text)
		text2, err2 := h.complete(ctx, fixPrompt, modeAccuseJSON)
		if err2 == nil {
			res, perr = parseAccuseJSON(text2)
		}
	}
	if perr != nil && h.fallback {
		fmt.Fprintf(os.Stderr, "pod-noir: LLM parse error (accuse): %v — using mock\n", perr)
		return h.mock.EvaluateAccusation(ctx, def, hypothesis)
	}
	if perr != nil {
		return AccuseResult{}, perr
	}
	return res, nil
}

func accuseRepairPrompt(originalUserPrompt, badModelOutput string) string {
	return strings.TrimSpace(fmt.Sprintf(`Your previous answer was not valid JSON with keys "judgment" and "reply".
Output ONLY one JSON object. No markdown.

Original task:
%s

Broken output (do not repeat verbatim; fix it):
%s`,
		originalUserPrompt,
		truncate(badModelOutput, 1200)))
}

func (h *HTTP) Debrief(ctx context.Context, def *scenario.Definition) (string, error) {
	prompt := buildDebriefPrompt(def)
	text, err := h.complete(ctx, prompt, modeDebrief)
	if err != nil && h.fallback {
		fmt.Fprintf(os.Stderr, "pod-noir: LLM HTTP error (debrief): %v — using mock\n", err)
		return h.mock.Debrief(ctx, def)
	}
	if err != nil {
		return "", err
	}
	out := strings.TrimSpace(text)
	out = stripMarkdownFence(out)
	if out == "" && h.fallback {
		return h.mock.Debrief(ctx, def)
	}
	out = clampRunes(out, 12000)
	return out, nil
}

func buildAccusePrompt(def *scenario.Definition, hypothesis string) string {
	hyp := strings.TrimSpace(hypothesis)
	hyp = clampRunes(hyp, 2000)
	var b strings.Builder
	fmt.Fprintf(&b, "You are the incident commander in a Kubernetes noir training game.\n")
	fmt.Fprintf(&b, "Scenario ID: %s\nTitle: %s\nNamespace: %s\n", def.ID, def.Title, def.Namespace)
	fmt.Fprintf(&b, "Main workload for this file: deployment/%s\n", def.SolveDeployment)
	if strings.TrimSpace(def.VictoryMode) == "endpoints" && strings.TrimSpace(def.VictoryService) != "" {
		fmt.Fprintf(&b, "Victory for debrief includes endpoints on Service %q.\n", def.VictoryService)
	}
	fmt.Fprintf(&b, "Teaching hints (judge depth; do not quote verbatim):\n  hot keywords: %v\n  warm: %v\n", def.HotHints, def.WarmHints)
	fmt.Fprintf(&b, "Detective hypothesis: %q\n\n", hyp)
	fmt.Fprintf(&b, "Reply with JSON ONLY (no markdown):\n")
	fmt.Fprintf(&b, `{"judgment":"hot|warm|cold|stone_cold","reply":"<short in-character feedback>"}`+"\n")
	fmt.Fprintf(&b, "Judgment rubric: hot=nailed root cause; warm=right layer; cold=weak; stone_cold=nonsense or empty theory.\n")
	return b.String()
}

func buildDebriefPrompt(def *scenario.Definition) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Write the closing case debrief for Kubernetes training scenario %s (%s).\n\n", def.ID, def.Title)
	fmt.Fprintf(&b, "Constraints:\n")
	fmt.Fprintf(&b, "- Plain text or ASCII box drawing — NOT JSON.\n")
	fmt.Fprintf(&b, "- Noir voice; precise; no hallucinating tools the player didn't have.\n")
	fmt.Fprintf(&b, "- Teach real root cause + 1–3 concrete kubectl examples in namespace %q.\n\n", def.Namespace)
	fmt.Fprintf(&b, "Theme map: 001 rollout/missing file; 002 Secret/secretKeyRef; 003 image pull; 004 probes; 005 OOM/limits/tmpfs; 006 Service selector/endpoints.\n")
	fmt.Fprintf(&b, "The live cluster already passed a health check — describe the fix as if the burden is now explained and filed.\n")
	fmt.Fprintf(&b, "Keep under ~45 short lines.\n")
	return b.String()
}

func (h *HTTP) complete(ctx context.Context, userPrompt string, mode completionMode) (string, error) {
	switch h.kind {
	case "anthropic":
		return h.anthropic(ctx, userPrompt, mode)
	case "openai":
		return h.openai(ctx, userPrompt, mode)
	case "ollama":
		return h.ollama(ctx, userPrompt, mode)
	default:
		return "", fmt.Errorf("unknown kind %q", h.kind)
	}
}

func (h *HTTP) anthropic(ctx context.Context, userPrompt string, mode completionMode) (string, error) {
	maxTok := 4096
	system := ""
	switch mode {
	case modeAccuseJSON:
		maxTok = 1024
		system = accuseSystemPrompt
	case modeDebrief:
		maxTok = 4096
		system = "You write clear operational debriefs with a noir tone. Never output JSON for debriefs."
	case modeContactWire:
		maxTok = 2048
		system = contactWireSystemPrompt
	}
	body := map[string]any{
		"model":      h.model,
		"max_tokens": maxTok,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}
	if system != "" {
		body["system"] = system
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

func (h *HTTP) openai(ctx context.Context, userPrompt string, mode completionMode) (string, error) {
	url := h.baseURL + "/chat/completions"
	body := map[string]any{
		"model": h.model,
		"messages": []map[string]string{
			{"role": "system", "content": "Follow the user instructions exactly. For JSON requests, output only valid JSON."},
			{"role": "user", "content": userPrompt},
		},
	}
	if mode == modeAccuseJSON {
		body["response_format"] = map[string]any{"type": "json_object"}
	}
	if mode == modeDebrief {
		delete(body, "response_format")
		body["messages"] = []map[string]string{
			{"role": "system", "content": "You write operational Kubernetes debriefs in plain text. Never use JSON."},
			{"role": "user", "content": userPrompt},
		}
	}
	if mode == modeContactWire {
		delete(body, "response_format")
		body["messages"] = []map[string]string{
			{"role": "system", "content": contactWireSystemPrompt},
			{"role": "user", "content": userPrompt},
		}
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

func (h *HTTP) ollama(ctx context.Context, userPrompt string, mode completionMode) (string, error) {
	url := h.baseURL + "/api/chat"
	body := map[string]any{
		"model": h.model,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
		"stream": false,
	}
	if mode == modeAccuseJSON {
		body["format"] = "json"
		body["messages"] = []map[string]string{
			{"role": "system", "content": accuseSystemPrompt},
			{"role": "user", "content": userPrompt},
		}
	}
	if mode == modeDebrief {
		body["messages"] = []map[string]string{
			{"role": "system", "content": "Write plain-text operational debriefs. Do not output JSON."},
			{"role": "user", "content": userPrompt},
		}
		delete(body, "format")
	}
	if mode == modeContactWire {
		body["messages"] = []map[string]string{
			{"role": "system", "content": contactWireSystemPrompt},
			{"role": "user", "content": userPrompt},
		}
		delete(body, "format")
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
