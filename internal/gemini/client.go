package gemini

import (
	"context"
	"errors"

	"google.golang.org/genai"
)

type Client struct {
	client *genai.Client
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      apiKey,
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return nil, errors.New("failed to create genai client")
	}
	return &Client{client: client}, nil
}

func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.Models.GenerateContent(ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", errors.New("failed to generate text")
	}
	respText := resp.Text()
	return respText, nil
}
