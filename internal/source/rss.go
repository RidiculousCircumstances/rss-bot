package source

import (
	"context"
	"github.com/SlyMarbo/rss"
	"github.com/samber/lo"
	"rss-bot/internal/model"
)

type RSSSource struct {
	URL        string
	SourceId   int64
	SourceName string
}

func NewRSSSourceFromModel(model model.Source) *RSSSource {
	return &RSSSource{
		URL:        model.FeedUrl,
		SourceName: model.Name,
		SourceId:   model.ID,
	}
}

func (source *RSSSource) Fetch(ctx context.Context) ([]model.Item, error) {
	feed, err := source.loadFeed(ctx, source.URL)
	if err != nil {
		return nil, err
	}

	return lo.Map(feed.Items, func(item *rss.Item, _ int) model.Item {
		return model.Item{
			Title:      item.Title,
			Categories: item.Categories,
			Link:       item.Link,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: source.SourceName,
		}
	}), nil
}

func (source *RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	var (
		chFeed = make(chan *rss.Feed)
		chErr  = make(chan error)
	)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			chErr <- err
			return
		}
		chFeed <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-chErr:
		return nil, err
	case feed := <-chFeed:
		return feed, nil
	}

}

func (rss *RSSSource) ID() int64 {
	return rss.SourceId
}

func (rss *RSSSource) Name() string {
	return rss.SourceName
}
