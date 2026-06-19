// Package llm generates match narratives via OpenAI-compatible chat APIs
// (NVIDIA NIM with nvapi- keys, or DeepSeek / other compatible endpoints).
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

// NarrativeInput carries quantitative signals for the narrative prompt.
type NarrativeInput struct {
	HomeTeam       string
	AwayTeam       string
	HomeWinProb    float64
	DrawProb       float64
	AwayWinProb    float64
	HomeELO        float64
	AwayELO        float64
	HomeGDP        float64
	AwayGDP        float64
	HomeLambda     float64
	AwayLambda     float64
	W1             float64
	W2             float64
	W3             float64
	VenueLabel     string
	PoissonFavors  string
	FinalFavors    string
	SignalConflict bool
}

// NarrativeResult is the structured narrative output.
type NarrativeResult struct {
	Narrative  string
	Confidence float64
}

// Config holds LLM endpoint settings (from env / config).
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

type narrativeJSON struct {
	Narrative  string  `json:"narrative"`
	Confidence float64 `json:"confidence"`
}

const (
	systemPrompt = `You are a quantitative football analyst. You MUST respond with ONLY valid JSON in this exact format:
{"narrative":"<your analysis in Traditional Chinese, max 200 characters>","confidence":<float between 0.0 and 1.0>}
Do not include any text outside the JSON object.`

	fallbackNarrative = "分析服務暫時不可用"
	httpTimeout       = 60 * time.Second
)

// ResolveConfig picks base URL and model from env overrides or API key prefix.
func ResolveConfig(apiKey, baseURL, model string) Config {
	cfg := Config{APIKey: strings.TrimSpace(apiKey)}
	if baseURL != "" {
		cfg.BaseURL = strings.TrimRight(baseURL, "/")
	} else if strings.HasPrefix(cfg.APIKey, "nvapi-") {
		cfg.BaseURL = "https://integrate.api.nvidia.com/v1"
	} else {
		cfg.BaseURL = "https://api.deepseek.com"
	}

	if model != "" {
		cfg.Model = model
	} else if strings.HasPrefix(cfg.APIKey, "nvapi-") {
		cfg.Model = "meta/llama-3.1-70b-instruct"
	} else {
		cfg.Model = "deepseek-chat"
	}
	return cfg
}

func buildUserPrompt(in NarrativeInput) string {
	conflictNote := ""
	if in.SignalConflict {
		conflictNote = "\nNote: Poisson (W2) and final blended probabilities favor different sides — mention this tension explicitly; do not overstate ELO gap."
	}
	return fmt.Sprintf(
		"Analyze match: %s vs %s.\n"+
			"Venue: %s.\n"+
			"ELO: %.0f vs %.0f. GDP per capita: %.0f vs %.0f USD.\n"+
			"Expected goals (Poisson λ): %.2f vs %.2f.\n"+
			"Model weights: W3(macro)=%.2f, W2(poisson)=%.2f, W1(micro)=%.2f.\n"+
			"Poisson win side: %s. Final win side: %s.\n"+
			"Final win probabilities: Home=%.3f, Draw=%.3f, Away=%.3f.%s",
		in.HomeTeam, in.AwayTeam,
		in.VenueLabel,
		in.HomeELO, in.AwayELO,
		in.HomeGDP, in.AwayGDP,
		in.HomeLambda, in.AwayLambda,
		in.W3, in.W2, in.W1,
		in.PoissonFavors, in.FinalFavors,
		in.HomeWinProb, in.DrawProb, in.AwayWinProb,
		conflictNote,
	)
}

func clampConfidence(v float64) float64 {
	return math.Max(0.0, math.Min(1.0, v))
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func callOnce(ctx context.Context, cfg Config, userMsg string) (*NarrativeResult, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm: API key is empty")
	}

	body, err := json.Marshal(chatRequest{
		Model: cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMsg},
		},
		MaxTokens:   512,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}

	url := cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req) //nolint:gosec // URL from configured base
	if err != nil {
		return nil, fmt.Errorf("llm: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm: status %d: %s", resp.StatusCode, string(raw))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("llm: decode response: %w", err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("llm: API error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("llm: empty choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	// Strip markdown code fences if the model wraps JSON.
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var narr narrativeJSON
	if err := json.Unmarshal([]byte(content), &narr); err != nil {
		return nil, fmt.Errorf("llm: unmarshal narrative JSON: %w (raw: %q)", err, content)
	}

	return &NarrativeResult{
		Narrative:  narr.Narrative,
		Confidence: clampConfidence(narr.Confidence),
	}, nil
}

// GenerateNarrative calls the configured LLM and retries once on failure.
func GenerateNarrative(ctx context.Context, cfg Config, in NarrativeInput) (NarrativeResult, error) {
	if cfg.APIKey == "" {
		log.Println("llm: API key not set — returning fallback narrative")
		return NarrativeResult{Narrative: fallbackNarrative, Confidence: 0.0}, nil
	}

	userMsg := buildUserPrompt(in)
	result, err := callOnce(ctx, cfg, userMsg)
	if err == nil {
		return *result, nil
	}
	log.Printf("llm: first attempt failed (%s/%s): %v — retrying", cfg.BaseURL, cfg.Model, err)

	result, err = callOnce(ctx, cfg, userMsg)
	if err == nil {
		return *result, nil
	}
	log.Printf("llm: second attempt failed: %v — returning fallback", err)

	return NarrativeResult{Narrative: fallbackNarrative, Confidence: 0.0}, nil
}
