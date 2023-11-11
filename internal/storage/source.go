package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"rss-bot/internal/model"
	"time"
)

type SourcePostgresStorage struct {
	db *sqlx.DB
}

type dbSource struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	FeedUrl   string    `db:"feed_url"`
	Priority  int       `db:"priority"`
	CreatedAt time.Time `db:"created_at"`
}

func NewSourceStorage(db *sqlx.DB) *SourcePostgresStorage {
	return &SourcePostgresStorage{db: db}
}

func (s *SourcePostgresStorage) Db() *sqlx.DB {
	return s.db
}

func (s *SourcePostgresStorage) SetDb(db *sqlx.DB) {
	s.db = db
}

func (s *SourcePostgresStorage) Sources(ctx context.Context) ([]model.Source, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var sources []dbSource
	// у Conn нет селект контекста. В видео - есть. Вероятно, тут будет ошибка, т.к. возможно,
	// что на данном этапе не используется созданное подключение
	if err := s.db.SelectContext(ctx, &sources, `SELECT * FROM sources`); err != nil {
		return nil, err
	}

	return lo.Map(sources, func(source dbSource, _ int) model.Source {
		return model.Source(source)
	}), nil

}

func (s *SourcePostgresStorage) SourceById(ctx context.Context, id int64) (*model.Source, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	var source dbSource

	if err := s.db.SelectContext(ctx, &source, `SELECT * FROM sources WHERE sources.id = $1`, id); err != nil {
		return nil, err
	}

	return (*model.Source)(&source), nil
}

func (s *SourcePostgresStorage) Add(ctx context.Context, source model.Source) (int64, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	row := conn.QueryRowContext(
		ctx,
		`INSERT INTO sources (name, feed_url, created_at) VALUES($1, $2, $3)`,
		source.Name,
		source.FeedUrl,
		source.CreatedAt,
	)

	if err := row.Err(); err != nil {
		return 0, err
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil

}

func (s *SourcePostgresStorage) Delete(ctx context.Context, id int64) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DELETE FROM sources WHERE id = $1`, id); err != nil {
		return err
	}

	return nil

}
