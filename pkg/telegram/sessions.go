package telegram

import (
	"context"
	"errors"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	sessions        = sync.Map{}
	newMenuFunction func(int64) Menu
)

type sessionData struct {
	m  Menu
	ts time.Time
}

func updateSession(chatID int64, forceNew bool) (Menu, bool) {
	sess, ok := sessions.Load(chatID)
	if !ok || forceNew {
		data := &sessionData{
			m:  newMenuFunction(chatID),
			ts: time.Now(),
		}

		sessions.Store(chatID, data)
		return data.m, true
	}

	sess.(*sessionData).ts = time.Now()

	sessions.Store(chatID, sess)
	return sess.(*sessionData).m, false
}

func cleanUp() {
	sessions.Range(func(k, v interface{}) bool {
		s := v.(*sessionData)
		if time.Since(s.ts) > time.Hour*24 {
			sessions.Delete(k)
		}

		return true
	})
}

func sendMessage(chatID int64, resp Response) ([]tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(chatID, resp.buttonText())
	msg.ParseMode = "html"
	switch t := resp.(type) {
	case *message:
		if t.hideButton {
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}
		return []tgbotapi.MessageConfig{msg}, nil
	case *button:
		//if t.inline() {
		//	msg.ReplyMarkup = createInlineKeyboard(t.keys, 3)
		//	return []tgbotapi.MessageConfig{msg}, nil
		//}
		msg.ReplyMarkup = createKeyboard(t.keys, 1)
		return []tgbotapi.MessageConfig{msg}, nil
	case *multi:
		var result []tgbotapi.MessageConfig
		for i := range t.resp {
			resp, err := sendMessage(chatID, t.resp[i])
			if err != nil {
				return nil, err
			}
			result = append(result, resp...)
		}
		return result, nil
	default:
		return nil, errors.New("invalid message type")
	}
}

func updateMessage(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	chatID := update.Message.Chat.ID
	forceNew := false
	if update.Message.Text == "/reset" {
		forceNew = true
	}
	m, newSession := updateSession(chatID, forceNew)
	if newSession {
		return sendMessage(chatID, m.Reset(ctx))
	}

	return sendMessage(chatID, m.Process(ctx, update.Message.Text))
}

// Update handle updates from telegram
func Update(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	if update.Message != nil {
		return updateMessage(ctx, update)
	}

	return nil, nil
}

// InitLibrary should be called before calling update
func InitLibrary(fn func(int64) Menu) {
	newMenuFunction = fn

	go func() {
		t := time.NewTicker(time.Hour)
		for range t.C {
			cleanUp()
		}
	}()
}
