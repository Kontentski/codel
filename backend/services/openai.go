package services

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/semanser/ai-coder/assets"
	"github.com/semanser/ai-coder/config"
	"github.com/semanser/ai-coder/templates"
)

var OpenAIclient *openai.Client

func Init() {

	OpenAIclient = openai.NewClient(config.Config.OpenAiKey)

	if config.Config.OpenAiKey == "" {
		log.Fatal("OPEN_AI_KEY is not set")
	}
}

func GetMessageSummary(query string, n int) (string, error) {
	prompt, err := templates.Render(assets.PromptTemplates, "prompts/summary.tmpl", map[string]any{
		"Text": query,
		"N":    n,
	})
	if err != nil {
		return "", err
	}

	req := openai.ChatCompletionRequest{
		Temperature: 0.0,
		Model:       openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
		TopP: 0.2,
		N:    1,
	}

	resp, err := OpenAIclient.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("completion error: %v", err)
	}

	choices := resp.Choices

	if len(choices) == 0 {
		return "", fmt.Errorf("no choices found")
	}

	return choices[0].Message.Content, nil
}

func GetDockerImageName(task string) (string, error) {
	prompt, err := templates.Render(assets.PromptTemplates, "prompts/docker.tmpl", map[string]any{
		"Task": task,
	})
	if err != nil {
		return "", err
	}

	req := openai.ChatCompletionRequest{
		Temperature: 0.0,
		Model:       openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
		TopP: 0.2,
		N:    1,
	}

	resp, err := OpenAIclient.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("completion error: %v", err)
	}

	choices := resp.Choices

	if len(choices) == 0 {
		return "", fmt.Errorf("no choices found")
	}

	return choices[0].Message.Content, nil
}
