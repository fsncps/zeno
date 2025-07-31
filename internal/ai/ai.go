package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func SummarizeKeywordsAndTitle(title, code string) (formattedTitle, description string, keywords []string, err error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", "", nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	client := openai.NewClient(apiKey)

	prompt := fmt.Sprintf(`You are helping build a cheatsheet database.
Given:

Title: %s
Code:
%s

Output 3 lines:
1. A cleaned and nicely capitalized title (only a few words, Title Case).
2. A short description of 20-50 words (single line).
3. A comma-separated list of 5-8 keywords (no extra text).`,
		title, code)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: prompt},
			},
		},
	)
	if err != nil {
		return "", "", nil, err
	}

	text := strings.TrimSpace(resp.Choices[0].Message.Content)
	lines := strings.SplitN(text, "\n", 3)

	// Defensive defaults
	formattedTitle = title
	if len(lines) >= 1 && strings.TrimSpace(lines[0]) != "" {
		formattedTitle = strings.TrimSpace(lines[0])
	}
	if len(lines) >= 2 {
		description = strings.TrimSpace(lines[1])
	}
	if len(lines) >= 3 {
		kwLine := strings.TrimSpace(lines[2])
		keywords = strings.Split(kwLine, ",")
		for i := range keywords {
			keywords[i] = strings.TrimSpace(keywords[i])
		}
	}

	return formattedTitle, description, keywords, nil
}
