package utils

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type QueryDBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

var _ QueryDBTX = (*wrappedDB)(nil)

type wrappedDB struct {
	QueryDBTX
}

func QueryWrap(db QueryDBTX) QueryDBTX {
	return &wrappedDB{db}
}

func (w wrappedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if b, ok := QueryBuilderFrom(ctx); ok {
		query, args = b.Build(query, args...)
	}

	return w.QueryDBTX.ExecContext(ctx, query, args...)
}

func (w wrappedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if b, ok := QueryBuilderFrom(ctx); ok {
		query, args = b.Build(query, args...)
	}

	return w.QueryDBTX.QueryContext(ctx, query, args...)
}

func (w wrappedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if b, ok := QueryBuilderFrom(ctx); ok {
		query, args = b.Build(query, args...)
	}

	return w.QueryDBTX.QueryRowContext(ctx, query, args...)
}

type queryBuilderContextKey struct{}

func QueryWithBuilder(ctx context.Context, b *Builder) context.Context {
	return context.WithValue(ctx, queryBuilderContextKey{}, b)
}

func QueryBuilderFrom(ctx context.Context) (*Builder, bool) {
	b, ok := ctx.Value(queryBuilderContextKey{}).(*Builder)
	return b, ok
}

func QueryBuild(ctx context.Context, f func(builder *Builder)) context.Context {
	b, ok := QueryBuilderFrom(ctx)
	if !ok {
		b = &Builder{}
	} else {
		b = b.clone()
	}

	f(b)
	return QueryWithBuilder(ctx, b)
}

type (
	Builder struct {
		filters       []filter
		order         string
		groupBy       string
		offset, limit int
	}

	filter struct {
		expression string
		args       []interface{}
	}
)

func (b *Builder) clone() *Builder {
	cb := Builder{}
	cb = *b
	return &cb
}

func (b *Builder) Where(query string, args ...interface{}) *Builder {
	b.filters = append(b.filters, filter{
		expression: query,
		args:       args,
	})

	return b
}

func (b *Builder) In(column string, args ...interface{}) *Builder {
	placeholders := make([]string, len(args))
	for i := range args {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ","))
	return b.Where(query, args...)
}

func (b *Builder) Order(cols string) *Builder {
	b.order = cols
	return b
}

func (b *Builder) Offset(x int) *Builder {
	b.offset = x
	return b
}

func (b *Builder) Limit(x int) *Builder {
	b.limit = x
	return b
}

func (b *Builder) Pagination(page, limit int) *Builder {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	b.limit = limit
	b.offset = offset
	return b
}

func (b *Builder) GroupBy(groupBy string) *Builder {
	b.groupBy = groupBy
	return b
}

func (b *Builder) Build(query string, args ...interface{}) (string, []interface{}) {
	var sb strings.Builder

	sb.WriteString(query)
	sb.WriteByte('\n')

	for idx, filter := range b.filters {
		if idx == 0 {
			sb.WriteString("WHERE ")
		} else {
			sb.WriteString("AND ")
		}

		sb.WriteByte('(')
		sb.WriteString(filter.expression)
		sb.WriteByte(')')
		sb.WriteByte('\n')

		args = append(args, filter.args...)
	}

	if b.groupBy != "" {
		sb.WriteString("GROUP BY ")
		sb.WriteString(b.groupBy)
		sb.WriteByte('\n')
	}

	if b.order != "" {
		sb.WriteString("ORDER BY ")
		sb.WriteString(b.order)
		sb.WriteByte('\n')
	}

	if b.limit > 0 {
		sb.WriteString("LIMIT ")
		sb.WriteString(strconv.Itoa(b.limit))
		sb.WriteByte('\n')
	}

	if b.offset > 0 {
		sb.WriteString("OFFSET ")
		sb.WriteString(strconv.Itoa(b.offset))
		sb.WriteByte('\n')
	}

	return sb.String(), args
}
