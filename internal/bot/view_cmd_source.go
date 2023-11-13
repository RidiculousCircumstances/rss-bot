package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"rss-bot/internal/botkit"
	"rss-bot/internal/model"
	"time"
)

type SourceStorage interface {
	Add(ctx context.Context, source model.Source) (int64, error)
}

func ViewCmdAddSource(sourceStorage SourceStorage) botkit.ViewFunc {

	type addSourceArgs struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Priority int    `json:"priority"`
	}

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {

		args, err := botkit.ParseJSON[addSourceArgs](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		source := model.Source{
			Name:      args.Name,
			FeedUrl:   args.URL,
			Priority:  args.Priority,
			CreatedAt: time.Time{},
		}

		sourceId, err := sourceStorage.Add(ctx, source)
		if err != nil {
			return err
		}

		var (
			msgText = fmt.Sprintf("Источние с ID `%d` успешно добавлен\\. Используйте этот id для управления источником\\.", sourceId)
		)

		replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		replyMsg.ParseMode = tgbotapi.ModeMarkdownV2

		if _, err := bot.Send(replyMsg); err != nil {
			return err
		}

		return nil
	}
}
