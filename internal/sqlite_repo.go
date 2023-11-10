package internal

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/libsql/go-libsql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"dedb"
)

type sqliteRepo struct {
	log    zerolog.Logger
	db     *sqlx.DB
	config Config
}

func (s *sqliteRepo) save(ctx context.Context, events []*dedb.Event) error {
	log := s.log.With().Str("op", "save").Logger()
	if len(events) == 0 {
		return fmt.Errorf("no events were supplied to save")
	}
	log.Info().Msgf("saving %d events", len(events))
	// timestamp := time.Now().UnixMicro()
	return nil
}

func (s *sqliteRepo) getDomain(ctx context.Context, domain string, domainId string, offset int64, limit int64) ([]*dedb.Event, error) {
	log := s.log.With().Str("op", "getDomain").Logger()
	log.Info().Msgf("getting domain %s, id %s, offset %d, limit %d", domain, domainId, offset, limit)
	sql := "SELECT id, domain, domain_id, name, timestamp, trace_id, data FROM domain_events WHERE domain_id = ? ORDER BY timestamp ASC LIMIT ?, ?"
	events := []*dedb.Event{}
	err := s.db.Select(&events, sql, domainId, limit, offset)
	if err != nil {
		s.log.Error().Err(err).Msgf("could not query domain events for domain %s, id %s", domain, domainId)
		return events, err
	}
	return events, nil
}

func (s *sqliteRepo) getDomainIds(ctx context.Context, domain string, offset int64, limit int64) ([]string, error) {
	log := s.log.With().Str("op", "getDomainIds").Logger()
	log.Debug().Msgf("getting domain ids for domain %s, offset %d, limit %d", domain, offset, limit)

	sql := "SELECT id FROM domains WHERE domain = ? ORDER BY timestamp ASC LIMIT ?, ?"
	ids := []string{}
	err := s.db.Select(&ids, sql, domain, limit, offset)
	if err != nil {
		s.log.Error().Err(err).Msgf("could not query domain ids for domain %s", domain)
		return ids, err
	}
	return ids, nil
}

func (s *sqliteRepo) shutdown() {
	s.db.Close()
}

func NewSqliteRepo(config Config) (*sqliteRepo, error) {
	r := &sqliteRepo{
		log:    log.With().Str("logger", "sqliteRepo").Logger(),
		config: config,
	}

	r.log.Info().Msgf("connecting to db at %s", config.SqliteDbConfig.DbUrl)
	db, err := sqlx.Connect("libsql", config.SqliteDbConfig.DbUrl)
	if err != nil {
		return nil, err
	}
	r.db = db

	r.log.Info().Msgf("connected to sqlite")
	return r, nil
}
