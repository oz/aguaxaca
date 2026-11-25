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
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// AnthropicExtractor implements Extractor using Anthropic's API.
type AnthropicExtractor struct {
	client anthropic.Client
	prompt string
}

// NewAnthropicExtractor creates an AnthropicExtractor with the given prompt.
// The API key is set with the ANTHROPIC_API_KEY environment variable.
func NewAnthropicExtractor(prompt string) *AnthropicExtractor {
	return &AnthropicExtractor{client: anthropic.NewClient(), prompt: prompt}
}

func (p *AnthropicExtractor) Extract(ctx context.Context, filePath string) (string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("can't read file %s: %w", filePath, err)
	}
	encodedData := base64.StdEncoding.EncodeToString(file)

	res, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 20000,
		Model:     anthropic.ModelClaudeSonnet5,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(p.prompt),
				anthropic.NewImageBlockBase64("image/jpeg", encodedData),
			),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic error for %s: %w", filePath, err)
	}

	text := res.Content[len(res.Content)-1].Text
	return stripCSVFence(text), nil
}
