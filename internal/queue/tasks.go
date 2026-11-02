package queue

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/luiggisao/auto-golpe/internal/gemini"
	"github.com/luiggisao/auto-golpe/internal/slack"
	"google.golang.org/genai"
)

const TypeSlackProcessEvent = "slack:process_event"

type SlackEventTaskPayload struct {
	Channel  string `json:"channel"`
	User     string `json:"user"`
	Text     string `json:"text"`
	ThreadTS string `json:"thread_ts"`
}

func NewSlackProcessEventTask(channel, user, text, threadTS string) (*asynq.Task, error) {
	payload := SlackEventTaskPayload{
		Channel:  channel,
		User:     user,
		Text:     text,
		ThreadTS: threadTS,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSlackProcessEvent, payloadBytes), nil
}

func HandleSlackProcessEventTask(
	ctx context.Context,
	t *asynq.Task,
	geminiClient *gemini.Client,
	slackClient *slack.Client,
) error {
	var payload SlackEventTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	slog.Info("processing slack event in background",
		"channel", payload.Channel,
		"user", payload.User,
		"text", payload.Text,
		"thread_ts", payload.ThreadTS,
	)
	var history []*genai.Content
	if payload.ThreadTS != "" {
		msgs, err := slackClient.GetThreadMessages(ctx, payload.Channel, payload.ThreadTS)
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			if msg.BotID != "" || msg.SubType == "bot_message" {
				history = append(history, genai.NewContentFromText(msg.Text, genai.RoleModel))

			} else {
				if msg.Text != "" {
					history = append(history, genai.NewContentFromText(msg.Text, genai.RoleUser))
				}
			}

		}
	} else {
		history = append(history, genai.NewContentFromText(payload.Text, genai.RoleUser))
	}

	response, err := geminiClient.GenerateText(ctx, history)
	if err != nil {
		return err
	}

	slog.Info("gemini response generated", "response", response)
	return slackClient.PostMessage(ctx, payload.Channel, payload.ThreadTS, response)
}
