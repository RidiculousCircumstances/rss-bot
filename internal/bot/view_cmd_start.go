package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"rss-bot/internal/botkit"
)

func ViewCmdStart() botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ready to work")

		if _, err := bot.Send(msg); err != nil {
			return err
		}
		return nil
	}
}
