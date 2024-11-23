package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
)

type Mattermost struct {
	webhookURL string
}

func NewMattermost(
	webhookURL string,
) *Mattermost {
	return &Mattermost{
		webhookURL: webhookURL,
	}
}

func (m *Mattermost) SendMessage(
	_ context.Context,
	message string,
	enableNotification bool,
) error {
	if enableNotification {
		message += "@channel"
	}

	data, err := json.Marshal(map[string]interface{}{
		"text":     message,
		"username": "db-backup",
	})
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := http.Post(m.webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return errors.WithStack(err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		return errors.New("unexpected error code from mattermost")
	}

	return nil
}
