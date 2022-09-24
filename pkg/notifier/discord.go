package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type DiscordChannel struct {
	webhookUrl string
}

func NewDiscordChannel(
	webhookUrl string,
) Channel {
	return &DiscordChannel{
		webhookUrl: webhookUrl,
	}
}

func (d *DiscordChannel) SendMessage(ctx context.Context, message string, enableNotification bool) error {
	if enableNotification {
		message += "@everyone"
	}

	req := struct {
		Content string `json:"content"`
	}{
		Content: message,
	}

	data, err := json.Marshal(req)

	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := http.Post(d.webhookUrl, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return errors.WithStack(err)
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("unexpected error code from discord %v and body [%v]", resp.StatusCode,
			string(body)))
	}

	return nil
}
