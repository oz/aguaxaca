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
	"strings"
)

const DefaultCSVPrompt = `Perform OCR on this image and extract the schedules (like "matutino" or "nocturno"), list of locations, and location types (like COLONIA or FRACCIONAMIENTOS, but always in singular form and downcased) from the text content.

You will output the information using the CSV format, with the following columns: "date," "schedule," "location_type," "location_name".

For the date column, use this format "YYYY-MM-DD" (for example "2025-03-14" for "14 de marzo de 2025"). Use correct quoting for attributes that may contain commas.

Do not include anything else apart from the CSV output.`

type Extractor interface {
	Extract(ctx context.Context, filePath string) (string, error)
}

// stripCSVFence removes the markdown code fence some LLMs wrap around the
// CSV output (despite being told not to).
func stripCSVFence(text string) string {
	if !strings.HasPrefix(text, "```csv\n") {
		return text
	}
	return strings.TrimSuffix(strings.TrimPrefix(text, "```csv\n"), "\n```")
}
