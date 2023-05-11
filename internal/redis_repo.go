package internal

import (
	"context"
	"dedb"
	b64 "encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	domainTypes = "domain_types"
	domains     = "domains"
)

/*
data structure is as follows:
each domain type has a sorted set of domain ids, sorted by microsecond it was added

	dedb:domain_types:<domain> => SortedSet

each domain instance has a list of events that make up that domain instance

	dedb:domain_events:<domain_id> => List

each domain id has a sorted set reverse index of the timestamp:event_id

	dedb:domain_events_timestamp_idx:0:<domain_id> => SortedSet
*/
type redisRepo struct {
	log    zerolog.Logger
	config Config
	pool   *redis.Client
}

type redisKey struct {
	db     string
	shard  int    // shard number
	prefix string // prefix name like 'domains', 'domain_types', 'domain_events'
	key    string // the actual key value of the entry
}

func (rk redisKey) String() string {
	return rk.db + ":" + rk.prefix + ":" + strconv.FormatInt(int64(rk.shard), 10) + ":" + rk.key
}

func NewRedisRepo(config Config) (*redisRepo, error) {
	repo := &redisRepo{
		log:    log.With().Str("logger", "redisRepo").Logger(),
		config: config,
	}
	err := validateRedisDbConfig(config.RedisDbConfig)
	if err != nil {
		return nil, err
	}

	pool, err := newPool(false, repo.config, &repo.log)
	if err != nil {
		return nil, err
	}
	repo.pool = pool

	return repo, nil
}

func validateRedisDbConfig(config RedisDbConfig) error {
	if config.DbAddress == "" {
		return fmt.Errorf("REDIS_DB config entry required")
	}
	if config.RedisCa != "" {
		if config.RedisUserCert == "" {
			return fmt.Errorf("REDIS_DB_USER_CERT config entry required")
		}
		if config.RedisUserKey == "" {
			return fmt.Errorf("REDIS_DB_USER_KEY config entry required")
		}
	}
	return nil
}

func (r *redisRepo) save(ctx context.Context, events []*dedb.Event) error {
	log := r.log.With().Str("op", "save").Logger()
	if len(events) == 0 {
		return fmt.Errorf("no events were supplied to save")
	}
	log.Debug().Msgf("saving %d events", len(events))
	// TODO: add locking
	timestamp := time.Now().UnixMicro()
	shard := r.getShard(events[0].Domain)
	_, err := r.pool.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// set up each event to save
		for _, e := range events {
			// give each event an id and a timestamp
			timestamp++
			id, _ := generateId()
			e.Id = id.String()
			e.Timestamp = timestamp

			// create the keys used
			domainsKey := redisKey{
				db:     "dedb",
				shard:  shard,
				prefix: "domain_types",
				key:    e.Domain,
			}

			domainEventsKey := redisKey{
				db:     "dedb",
				shard:  shard,
				prefix: "domain_events",
				key:    e.DomainId,
			}

			eventTimestampIndexKey := redisKey{
				db:     "dedb",
				shard:  shard,
				prefix: "domain_events_timestamp_idx",
				key:    e.DomainId,
			}

			sEnc := b64.StdEncoding.EncodeToString([]byte(e.Data))
			e.Data = []byte(sEnc)
			encoded, err := Encode(e)
			if err != nil {
				log.Error().Err(err).Msgf("could not encode event")
				return err
			}

			pipe.ZAddNX(ctx, domainsKey.String(), &redis.Z{Score: float64(timestamp), Member: e.DomainId})
			pipe.RPush(ctx, domainEventsKey.String(), encoded)
			pipe.ZAdd(ctx, eventTimestampIndexKey.String(), &redis.Z{Score: float64(timestamp), Member: e.Id})
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *redisRepo) getDomainIds(ctx context.Context, domain string, offset int64, limit int64) ([]string, error) {
	log := r.log.With().Str("op", "getDomainIds").Logger()
	log.Debug().Msgf("getting domain ids for domain %s, offset %d, limit %d", domain, offset, limit)

	key := redisKey{
		db:     "dedb",
		shard:  0,
		prefix: "domain_types",
		key:    domain,
	}

	args := &redis.ZRangeBy{
		Offset: offset,
		Count:  limit,
		Min:    "-inf",
		Max:    "+inf",
	}
	reply := r.pool.ZRangeByScore(ctx, key.String(), args)

	return reply.Val(), nil
}

func (r *redisRepo) getDomain(ctx context.Context, domain string, domainId string, offset int64, limit int64) ([]*dedb.Event, error) {
	log := r.log.With().Str("op", "getDomain").Logger()

	key := redisKey{
		db:     "dedb",
		shard:  0,
		prefix: "domain_events",
		key:    domainId,
	}

	reply := r.pool.LRange(ctx, key.String(), offset, limit)
	events := make([]*dedb.Event, 0)
	for _, je := range reply.Val() {
		e := &dedb.Event{}
		err := Decode(e, je)
		if err != nil {
			log.Error().Err(err).Msgf("could not decode event")
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

/*
func (r *redisRepo) getDomainByTimeRange(ctx context.Context, domain string, domainId string, from int64, limit int32) ([]*dedb.Event, error) {
	log := r.log.With().Str("op", "getDomainByTimeRange").Logger()
	conn := r.pool.Get()
	defer conn.Close()

	key := redisKey{
		db:     "dedb",
		shard:  0,
		prefix: "domain_events_timestamp_idx",
		key:    domainId,
	}

	var sortedIds []string
	res, err := redis.Values(conn.Do("ZRANGE", key.String(), strconv.FormatInt(from, 10), limit))
	if err != nil {
		return nil, err
	}
	err = redis.ScanSlice(res, &sortedIds)
	if err != nil {
		return nil, err
	}

	// parse the event id off the sorted ids list
	ids := make([]string, 0)
	for _, id := range sortedIds {
		ids = append(ids, id)
	}
	log.Debug().Msgf("ids: %v", ids)

	return nil, nil
}
*/

func (r *redisRepo) shutdown() {
	r.pool.Close()
}

func (r *redisRepo) getShard(domain string) int {
	return 0
}
