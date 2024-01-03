package repository

import (
	"context"
	"database/sql"
	"query-sqlc/query"
	"query-sqlc/utils"
)

type Repository interface {
	GetAllAuthor(ctx context.Context) ([]query.Author, error)
}

type repository struct {
	db           *sql.DB
	query        *query.Queries
	queryBuilder *query.Queries
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db:           db,
		query:        query.New(db),
		queryBuilder: query.New(utils.QueryWrap(db)),
	}
}

func (r *repository) GetAllAuthor(ctx context.Context) ([]query.Author, error) {
	return r.queryBuilder.GetAllAuthor(utils.QueryBuild(ctx, func(b *utils.Builder) {
		b.Where("LOWER(email) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) ", "%example.com", "%example.org")

		if true {
			b.Where("LOWER(first_name) LIKE LOWER(?)", "%an%")
		}
	}))
}
