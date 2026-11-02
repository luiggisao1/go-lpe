package http

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/luiggisao/auto-golpe/internal/queue"
)

type SlackInnerEvent struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Text     string `json:"text"`
	Channel  string `json:"channel"`
	ThreadTS string `json:"thread_ts"`
}

type SlackEventPayload struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge"`
	Event     SlackInnerEvent `json:"event"`
}

func verifySlackSignature(r *http.Request, signingSecret string) error {
	slackRequestTimestamp := r.Header.Get("X-Slack-Request-Timestamp")
	slackSignature := r.Header.Get("X-Slack-Signature")

	if slackRequestTimestamp == "" || slackSignature == "" {
		return errors.New("missing X-Slack-Request-Timestamp or X-Slack-Signature")
	}

	ts, err := strconv.ParseInt(slackRequestTimestamp, 10, 64)
	if err != nil {
		return errors.New("invalid X-Slack-Request-Timestamp")
	}

	diff := time.Now().Unix() - ts
	if diff > 300 || diff < -300 {
		return errors.New("slack request timestamp expired")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	h := hmac.New(sha256.New, []byte(signingSecret))
	baseString := "v0:" + slackRequestTimestamp + ":" + string(bodyBytes)
	h.Write([]byte(baseString))
	actualSignature := "v0=" + hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(slackSignature), []byte(actualSignature)) {
		return errors.New("slack signature mismatch")
	}
	return nil
}

func (s *Server) handleSlackWebhook(w http.ResponseWriter, r *http.Request) {
	if err := verifySlackSignature(r, s.cfg.SlackSigningSecret); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		return
	}

	var payload SlackEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if payload.Type == "url_verification" {
		writeJSON(w, http.StatusOK, map[string]string{"challenge": payload.Challenge})
		return
	}

	if payload.Type == "event_callback" {
		slog.Info("slack event received",
			"type", payload.Event.Type,
			"user", payload.Event.User,
			"channel", payload.Event.Channel,
			"text", payload.Event.Text,
		)
		task, err := queue.NewSlackProcessEventTask(
			payload.Event.Channel,
			payload.Event.User,
			payload.Event.Text,
			payload.Event.ThreadTS,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if _, err := s.queueClient.Enqueue(task); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to enqueue task"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
