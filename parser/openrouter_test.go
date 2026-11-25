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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestImage(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.jpg")
	require.NoError(t, os.WriteFile(path, []byte("fake jpeg bytes"), 0644))
	return path
}

func messagesResponse(text string) string {
	return fmt.Sprintf(`{
		"id": "msg_test",
		"model": "%s",
		"role": "assistant",
		"type": "message",
		"content": [{"type": "text", "text": %s}]
	}`, openRouterModel, mustJSON(text))
}

func mustJSON(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func newTestExtractor(baseURL string) *OpenRouterExtractor {
	return &OpenRouterExtractor{
		prompt:   "extract this",
		apiKey:   "test-key",
		baseURL:  baseURL,
		client:   &http.Client{Timeout: 5 * time.Second},
		attempts: 3,
		delay:    time.Millisecond,
	}
}

func TestExtractSuccess(t *testing.T) {
	var gotAuth, gotContentType, gotPath string
	var gotReq openRouterRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		decodeErr := json.NewDecoder(r.Body).Decode(&gotReq)
		if decodeErr != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, messagesResponse("date,schedule\n2025-03-14,matutino\n"))
	}))
	defer server.Close()

	extractor := newTestExtractor(server.URL)
	text, err := extractor.Extract(t.Context(), writeTestImage(t))

	require.NoError(t, err)
	assert.Equal(t, "date,schedule\n2025-03-14,matutino\n", text)

	// Request shape
	assert.Equal(t, "/messages", gotPath)
	assert.Equal(t, "Bearer test-key", gotAuth)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, openRouterModel, gotReq.Model)
	assert.Equal(t, openRouterMaxTokens, gotReq.MaxTokens)
	require.Len(t, gotReq.Messages, 1)
	assert.Equal(t, "user", gotReq.Messages[0].Role)
	require.Len(t, gotReq.Messages[0].Content, 2)
	assert.Equal(t, "text", gotReq.Messages[0].Content[0].Type)
	assert.Equal(t, "extract this", gotReq.Messages[0].Content[0].Text)
	assert.Equal(t, "image", gotReq.Messages[0].Content[1].Type)
	require.NotNil(t, gotReq.Messages[0].Content[1].Source)
	assert.Equal(t, "base64", gotReq.Messages[0].Content[1].Source.Type)
	assert.Equal(t, "image/jpeg", gotReq.Messages[0].Content[1].Source.MediaType)
	expectedData := base64.StdEncoding.EncodeToString([]byte("fake jpeg bytes"))
	assert.Equal(t, expectedData, gotReq.Messages[0].Content[1].Source.Data)
}

func TestExtractStripsCSVFence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, messagesResponse("```csv\ndate,schedule\n2025-03-14,matutino\n```"))
	}))
	defer server.Close()

	extractor := newTestExtractor(server.URL)
	text, err := extractor.Extract(t.Context(), writeTestImage(t))

	require.NoError(t, err)
	assert.Equal(t, "date,schedule\n2025-03-14,matutino", text)
}

func TestExtractRetriesOnServerError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, `{"error": {"type": "api_error", "message": "boom"}}`)
			return
		}
		_, _ = fmt.Fprint(w, messagesResponse("date,schedule\n2025-03-14,matutino\n"))
	}))
	defer server.Close()

	extractor := newTestExtractor(server.URL)
	text, err := extractor.Extract(t.Context(), writeTestImage(t))

	require.NoError(t, err)
	assert.Equal(t, "date,schedule\n2025-03-14,matutino\n", text)
	assert.EqualValues(t, 3, calls.Load())
}

func TestExtractRetriesOnRateLimit(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprint(w, `{"error": {"type": "rate_limit_error", "message": "slow down"}}`)
			return
		}
		_, _ = fmt.Fprint(w, messagesResponse("ok"))
	}))
	defer server.Close()

	extractor := newTestExtractor(server.URL)
	text, err := extractor.Extract(t.Context(), writeTestImage(t))

	require.NoError(t, err)
	assert.Equal(t, "ok", text)
	assert.EqualValues(t, 2, calls.Load())
}

func TestExtractDoesNotRetryClientError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"error": {"type": "authentication_error", "message": "Invalid API key"}}`)
	}))
	defer server.Close()

	extractor := newTestExtractor(server.URL)
	_, err := extractor.Extract(t.Context(), writeTestImage(t))

	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	assert.Equal(t, "authentication_error", apiErr.Type)
	assert.Equal(t, "Invalid API key", apiErr.Message)
	assert.EqualValues(t, 1, calls.Load())
}

func TestExtractFailsAfterRetriesExhausted(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `{"error": {"type": "overloaded_error", "message": "overloaded"}}`)
	}))
	defer server.Close()

	extractor := newTestExtractor(server.URL)
	_, err := extractor.Extract(context.Background(), writeTestImage(t))

	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusServiceUnavailable, apiErr.StatusCode)
	assert.EqualValues(t, 3, calls.Load())
}

func TestExtractMissingFile(t *testing.T) {
	extractor := newTestExtractor("http://127.0.0.1:0")
	_, err := extractor.Extract(context.Background(), filepath.Join(t.TempDir(), "missing.jpg"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't read file")
}
