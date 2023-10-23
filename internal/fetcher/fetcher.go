package fetcher

import (
	"context"
	"github.com/samber/lo"
	"log"
	"rss-bot/internal/model"
	rss "rss-bot/internal/source"
	"sync"
	"time"
)

type ArticleStorageInterface interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceProviderInterface interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type SourceInterface interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles       ArticleStorageInterface
	source         SourceProviderInterface
	fetchInterval  time.Duration
	filterKeywords []string
}

func New(articleStorage ArticleStorageInterface, provider SourceProviderInterface, fetchInterval time.Duration, filterKeywords []string) *Fetcher {
	return &Fetcher{
		articles:       articleStorage,
		source:         provider,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.source.Sources(ctx)

	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, source := range sources {
		wg.Add(1)

		go func(source SourceInterface) {
			defer wg.Done()

			items, err := source.Fetch(ctx)

			if err != nil {
				log.Printf("[ERROR] source.Fetch finished with failure: %q: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR] source.Fetch finished with failure: %q: %v", source.Name(), err)
				return
			}

		}(rss.NewRSSSourceFromModel(source))

	}

	wg.Wait()

	return nil
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)

	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}

}

func (f *Fetcher) processItems(ctx context.Context, source SourceInterface, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.shouldSkipItem(item) {
			continue
		}

		if err := f.articles.Store(ctx, model.Article{
			SourceId:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
			PostedAt:    time.Now().UTC(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (f *Fetcher) shouldSkipItem(item model.Item) bool {
	intersect := lo.Intersect(f.filterKeywords, item.Categories)
	shouldSkip := len(intersect)
	return shouldSkip == 0
}
