package middleware

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"rss-bot/internal/botkit"
	"rss-bot/internal/config"
)

func AdminOnly(next botkit.ViewFunc) botkit.ViewFunc {

	channelId := config.Get().TelegramChannelID

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {

		admins, err := bot.GetChatAdministrators(
			tgbotapi.ChatAdministratorsConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: channelId,
				},
			},
		)
		if err != nil {
			return err
		}

		for _, admin := range admins {
			if admin.User.ID == update.Message.From.ID {
				return next(ctx, bot, update)
			}
		}

		if _, err := bot.Send(tgbotapi.NewMessage(
			update.FromChat().ID,
			"У вас нет прав на выполнение этой команды",
		)); err != nil {
			return err
		}

		return nil
	}
}
