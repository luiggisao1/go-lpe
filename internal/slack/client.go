package slack

import (
	"context"

	"github.com/slack-go/slack"
)

type Client struct {
	client *slack.Client
}

func NewClient(token string) *Client {
	return &Client{
		client: slack.New(token),
	}
}

func (c *Client) PostMessage(ctx context.Context, channel, threadTS, text string) error {
	opts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
	}
	if threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(threadTS))
	}
	_, _, err := c.client.PostMessageContext(ctx, channel, opts...)
	return err
}

func (c *Client) GetThreadMessages(ctx context.Context, channel, threadTS string) ([]slack.Message, error) {
	params := &slack.GetConversationRepliesParameters{
		ChannelID: channel,
		Timestamp: threadTS,
	}
	messages, _, _, err := c.client.GetConversationRepliesContext(ctx, params)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
