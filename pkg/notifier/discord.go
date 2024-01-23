package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/samber/lo"

	"github.com/cockroachdb/errors"
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

func (d *DiscordChannel) SendMessage(_ context.Context, fullMessage string, enableNotification bool) error {
	if enableNotification {
		fullMessage += "@everyone"
	}

	var finalErr error
	for _, message := range lo.Chunk([]rune(fullMessage), 2000) {
		req := struct {
			Content string `json:"content"`
		}{
			Content: string(message),
		}

		data, err := json.Marshal(req)

		if err != nil {
			finalErr = errors.Join(finalErr, errors.WithStack(err))
			continue
		}

		resp, err := http.Post(d.webhookUrl, "application/json", bytes.NewBuffer(data))

		if err != nil {
			finalErr = errors.Join(finalErr, errors.WithStack(err))
			continue
		}

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
			finalErr = errors.Join(finalErr, errors.New(fmt.Sprintf("unexpected error code from discord %v and body [%v]", resp.StatusCode,
				string(body))))
			continue
		}
	}

	return finalErr
}
