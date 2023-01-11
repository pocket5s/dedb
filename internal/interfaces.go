package internal

import (
	"context"
	"dedb"
)

type repository interface {
	save(ctx context.Context, events []*dedb.Event) error
	shutdown()
	getDomain(ctx context.Context, domain string, domainId string, offset int64, limit int64) ([]*dedb.Event, error)
	getDomainIds(ctx context.Context, domain string, offset int64, limit int64) ([]string, error)
}

type publisher interface {
	publish(ctx context.Context, events []*dedb.Event)
	shutdown()
}
