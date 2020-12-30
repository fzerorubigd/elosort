package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func createKeyboard(items []string, align int) tgbotapi.ReplyKeyboardMarkup {
	if align < 1 {
		align = 1
	}

	total := len(items)/align + 1
	var all [][]tgbotapi.KeyboardButton
	for i := 0; i < total; i++ {
		current := make([]tgbotapi.KeyboardButton, align)
		for j := 0; j < align; j++ {
			idx := i*align + j
			if idx >= len(items) {
				break
			}
			current[j] = tgbotapi.NewKeyboardButton(items[idx])
		}
		all = append(all, tgbotapi.NewKeyboardButtonRow(current...))
	}

	return tgbotapi.NewReplyKeyboard(all...)
}
