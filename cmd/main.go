package main

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"rss-bot/internal/bot"
	"rss-bot/internal/botkit"
	"rss-bot/internal/config"
	"rss-bot/internal/fetcher"
	"rss-bot/internal/notifier"
	"rss-bot/internal/storage"
	"rss-bot/internal/summary"
	"syscall"
)

func main() {
	botApi, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("failed to create bot: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
	}

	defer db.Close()

	var (
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage  = storage.NewSourceStorage(db)

		fetcherInstance = fetcher.New(
			articleStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)

		summariser       = summary.New(config.Get().OpenAIKey, config.Get().OpenAIPrompt)
		notifierInstance = notifier.New(articleStorage,
			summariser, botApi,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
			config.Get().TelegramChannelID)
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	)

	defer cancel()

	newsBot := botkit.NewBot(botApi)
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())

	go func(ctx context.Context) {
		if err := fetcherInstance.Start(ctx); err != nil {

			if !errors.Is(err, context.Canceled) {
				log.Printf("failed to start fetcher: %v", err)
				return
			}

			log.Println("fetcher stopped")
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := notifierInstance.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("failed to start notifier: %v", err)
				return
			}

			log.Println("notifier stopped")
		}
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("[ERROR]: failed to start bot: %v", err)
		}

		fmt.Println("news bot stopped")
	}
}
