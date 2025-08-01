package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// SummarizeAndKeywords asks the model to produce a clean title (Title Case),
// an encyclopaedic description, and a list of keywords for a command snippet.
func SummarizeAndKeywords(title, code string) (string, string, []string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", "", nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	client := openai.NewClient(apiKey)

	prompt := fmt.Sprintf(`You are supplied a command or snippet of code with a title.
You must produce for a reference work:

• A concise, encyclopaedic description using a maximum of 100 words.

• A list of 5-8 keywords, separated by commas.

Do not prefix lines with bullets, numbers, colons, or any other symbols.
Use encyclopedic language, prefer subjectless, elliptical sentences.
Output must be exactly two sections separated by a blank line.

BAD EXAMPLE:
This code creates a simple static file server using Node.js...
node.js, static file server, http, readFile

GOOD EXAMPLE:
Creates simple static file server using Node.js serving files from current directory, returns 404 for missing files, listens on port 8080.

node.js, static file server, http, readfile, error handling, web development

Now produce the same structure for the following:

Title: %s
Code:
%s`, title, code)

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

	output := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Split into three sections
	sections := strings.SplitN(output, "\n\n", 3)
	if len(sections) < 3 {
		return "", "", nil, fmt.Errorf("unexpected AI output:\n%s", output)
	}

	refinedTitle := strings.TrimSpace(sections[0])
	description := strings.TrimSpace(sections[1])
	keywordStr := strings.TrimSpace(sections[2])

	// Parse keywords
	kwParts := strings.Split(keywordStr, ",")
	var keywords []string
	for _, kw := range kwParts {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			keywords = append(keywords, kw)
		}
	}

	return refinedTitle, description, keywords, nil
}
