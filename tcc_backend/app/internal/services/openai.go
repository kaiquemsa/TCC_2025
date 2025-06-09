package services

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client *openai.Client
}

func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		client: openai.NewClient(apiKey),
	}
}

func (s *OpenAIService) GenerateEmbedding(text string) ([]float32, error) {
	resp, err := s.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: []string{text},
		},
	)
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}

func (s *OpenAIService) GenerateSQL(question string, situation string) (string, error) {
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Você é um especialista em SQL. Gere consultas SQL baseadas em perguntas em linguagem natural e no contexto fornecido.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Contexto: " + situation + "\n\nPergunta: " + question + "\n\nGere uma consulta SQL que responda esta pergunta.",
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
