package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
	"rss-bot/internal/botkit"
	"rss-bot/internal/botkit/markup"
	"rss-bot/internal/model"
	"strings"
)

type SourceLister interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

func ViewCmdListSources(sourceLister SourceLister) botkit.ViewFunc {

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		sources, err := sourceLister.Sources(ctx)
		if err != nil {
			return err
		}

		var (
			sourceInfos = lo.Map(sources, func(source model.Source, _ int) string {
				return formatSource(source)
			})

			msg = fmt.Sprintf("Список источников \\(всего `%d`\\): \n\n%s",
				len(sourceInfos),
				strings.Join(sourceInfos, "\n\n"))
		)

		replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
		replyMsg.ParseMode = tgbotapi.ModeMarkdownV2

		if _, err := bot.Send(replyMsg); err != nil {
			return err
		}

		return nil

	}

}

func formatSource(source model.Source) string {
	return fmt.Sprintf("*%s*\nID: `%d`\nURL фида: %s",
		markup.EscapeForMarkdown(source.Name),
		source.ID,
		markup.EscapeForMarkdown(source.FeedUrl),
	)
}
