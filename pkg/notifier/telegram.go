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

type TelegramChannel struct {
	chatId      string
	accessToken string
}

func NewTelegramChannel(
	chatId string,
	accessToken string,
) Channel {
	return &TelegramChannel{
		chatId:      chatId,
		accessToken: accessToken,
	}
}

func (t TelegramChannel) SendMessage(ctx context.Context, message string, enableNotification bool) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", t.accessToken)

	req := struct {
		ChatId              string `json:"chat_id"`
		Text                string `json:"text"`
		DisableNotification bool   `json:"disable_notification"`
	}{
		ChatId:              t.chatId,
		Text:                message,
		DisableNotification: enableNotification,
	}

	data, err := json.Marshal(req)

	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return errors.WithStack(err)
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("unexpected error code from telegram %v and body [%v]", resp.StatusCode,
			string(body)))
	}

	return nil
}
