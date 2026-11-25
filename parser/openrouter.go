// This file is part of Aguaxaca.
// Copyright (C) 2026 Arnaud Berthomier.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package parser

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/avast/retry-go/v4"
)

const (
	openRouterBaseURL   = "https://openrouter.ai/api/v1"
	openRouterModel     = "thinkingmachines/inkling-2-small"
	openRouterMaxTokens = 20000
)

type APIError struct {
	StatusCode int
	Type       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("openrouter error %d (%s): %s", e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("openrouter error %d: %s", e.StatusCode, e.Message)
}

type OpenRouterExtractor struct {
	prompt  string
	apiKey  string
	baseURL string
	client  *http.Client

	// retries
	attempts uint
	delay    time.Duration
}

// NewOpenRouterExtractor creates an OpenRouterExtractor with the given
// prompt. The API key is read from the OPENROUTER_API_KEY environment
// variable. The model is hardcoded to thinkingmachines/inkling-2-small.
func NewOpenRouterExtractor(prompt string) *OpenRouterExtractor {
	return &OpenRouterExtractor{
		prompt:   prompt,
		apiKey:   os.Getenv("OPENROUTER_API_KEY"),
		baseURL:  openRouterBaseURL,
		client:   &http.Client{Timeout: 5 * time.Minute},
		attempts: 3,
		delay:    200 * time.Millisecond,
	}
}

// Extract queries OpenRouter with a file attachment and the extractor's
// prompt, and returns the resulting text.
func (e *OpenRouterExtractor) Extract(ctx context.Context, filePath string) (string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("can't read file %s: %w", filePath, err)
	}
	encodedData := base64.StdEncoding.EncodeToString(file)

	var text string
	op := func() error {
		var err error
		text, err = e.extract(ctx, encodedData)
		if err == nil {
			return nil
		}
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode != http.StatusTooManyRequests && apiErr.StatusCode < 500 {
			return retry.Unrecoverable(err)
		}
		return err
	}

	err = retry.Do(op,
		retry.Attempts(e.attempts),
		retry.Delay(e.delay),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(ctx),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		return "", fmt.Errorf("openrouter error for %s: %w", filePath, err)
	}
	return text, nil
}

func (e *OpenRouterExtractor) extract(ctx context.Context, encodedImage string) (string, error) {
	body, err := json.Marshal(openRouterRequest{
		Model:     openRouterModel,
		MaxTokens: openRouterMaxTokens,
		Messages: []openRouterMessage{
			{
				Role: "user",
				Content: []openRouterContentBlock{
					{Type: "text", Text: e.prompt},
					{
						Type: "image",
						Source: &openRouterImageSource{
							Type:      "base64",
							MediaType: "image/jpeg",
							Data:      encodedImage,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	//nolint:errcheck // unchecked
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		var errResp openRouterErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			apiErr.Type = errResp.Error.Type
			apiErr.Message = errResp.Error.Message
		} else {
			apiErr.Message = truncate(string(respBody), 512)
		}
		return "", apiErr
	}

	var parsed openRouterResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	for _, v := range slices.Backward(parsed.Content) {
		if v.Type == "text" {
			return stripCSVFence(v.Text), nil
		}
	}
	return "", fmt.Errorf("no text content in response")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

type openRouterRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	Messages  []openRouterMessage `json:"messages"`
}

type openRouterMessage struct {
	Role    string                   `json:"role"`
	Content []openRouterContentBlock `json:"content"`
}

type openRouterContentBlock struct {
	Type   string                 `json:"type"`
	Text   string                 `json:"text,omitempty"`
	Source *openRouterImageSource `json:"source,omitempty"`
}

type openRouterImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type openRouterResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type openRouterErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
