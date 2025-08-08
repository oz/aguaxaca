// This file is part of Aguaxaca.
// Copyright (C) 2025 Arnaud Berthomier.
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

const DefaultCSVPrompt = `Perform OCR on this image and extract the schedules (like "matutino" or "nocturno"), list of locations, and location types (like COLONIA or FRACCIONAMIENTOS, but always in singular form and downcased) from the text content.


You will output the information using the CSV format, with the following columns: "date," "schedule," "location_type," "location_name". For the date column, use this format "YYYY-MM-DD" (for example "2025-03-14" for "14 de marzo de 2025"). Use correct quoting for attributes that may contain commas.


Do not include more details about what the image is about, or other helpful text.`

// ParseFileWithPrompt queries Anthropic with a file attachment, prompting as indicated, and returns the resulting text.
func ParseFileWithPrompt(ctx context.Context, filePath string, prompt string) (string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("can't read file %s: %w", filePath, err)
	}
	encodedData := base64.StdEncoding.EncodeToString(file)

	// API key is set with: os.LookupEnv("ANTHROPIC_API_KEY")
	client := anthropic.NewClient()
	res, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 20000,
		Model:     anthropic.ModelClaude4Sonnet20250514,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(prompt),
				anthropic.NewImageBlockBase64("image/jpeg", encodedData),
			),
		},
	})
	if err != nil {
		return "", fmt.Errorf("Anthropic error for %s: %w", filePath, err)
	}

	return res.Content[len(res.Content)-1].Text, nil
}
