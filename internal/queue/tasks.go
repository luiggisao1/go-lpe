package queue

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
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

func HandleSlackProcessEventTask(ctx context.Context, t *asynq.Task) error {
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

	// Process the Slack event
	return nil
}
