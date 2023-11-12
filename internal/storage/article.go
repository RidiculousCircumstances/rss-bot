package storage

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"rss-bot/internal/model"
	"time"
)

type ArticlePostgresStorage struct {
	db *sqlx.DB
}

func NewArticleStorage(db *sqlx.DB) *ArticlePostgresStorage {
	return &ArticlePostgresStorage{db: db}
}

func (a *ArticlePostgresStorage) Store(ctx context.Context, article model.Article) error {

	conn, err := a.db.Conn(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `
		INSERT INTO articles (source_id, title, link, summary, published_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING;
`, article.SourceId, article.Title, article.Link, article.Summary, article.PublishedAt); err != nil {
		return nil
	}

	return nil
}

func (a *ArticlePostgresStorage) AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error) {
	conn, err := a.db.Conn(ctx)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var articles []dbArticle

	//if err := a.db.SelectContext(ctx, &articles, `
	//	SELECT * from articles
	//	         WHERE posted_at IS NULL
	//	           AND published_at > $1::timestamp
	//	         ORDER BY published_at DESC
	//	         limit $2
	//`,
	//	since.UTC().Format(time.RFC3339),
	//	limit); err != nil {
	//	return nil, err
	//}

	if err := a.db.SelectContext(ctx, &articles, `
		SELECT * from articles
		         WHERE posted_at IS NULL
		         ORDER BY published_at DESC
		         limit $1
	`,
		limit); err != nil {
		return nil, err
	}

	result := lo.Map(articles, func(article dbArticle, _ int) model.Article {
		return model.Article{
			ID:          article.ID,
			SourceId:    article.SourceId,
			Title:       article.Title,
			Link:        article.Link,
			Summary:     article.Summary.String,
			PublishedAt: article.PublishedAt,
			CreatedAt:   article.CreatedAt,
		}
	})

	return result, nil
}

func (a *ArticlePostgresStorage) MarkAsPosted(ctx context.Context, id int64) error {
	conn, err := a.db.Conn(ctx)

	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `UPDATE articles SET posted_at = $1::timestamp WHERE id = $2`,
		time.Now().UTC().Format(time.RFC3339), id); err != nil {
		return err
	}

	return nil
}

type dbArticle struct {
	ID int64 `db:"id"`
	//SourcePriority int64          `db:"priority"`
	SourceId    int64          `db:"source_id"`
	Title       string         `db:"title"`
	Link        string         `db:"link"`
	Summary     sql.NullString `db:"summary"`
	PublishedAt time.Time      `db:"published_at"`
	PostedAt    sql.NullTime   `db:"posted_at"`
	CreatedAt   time.Time      `db:"created_at"`
}
