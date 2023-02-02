package internal

import (
	"context"
	"dedb"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRepoSave(t *testing.T) {
	// setup
	ctx := context.Background()
	config := Config{
		RedisDbConfig: RedisDbConfig{
			DbAddress: "redis:6379",
			DbIndex:   0,
		},
	}
	repo, err := NewRedisRepo(config)
	if err != nil {
		panic(err)
	}
	cases := []struct {
		name   string
		events []*dedb.Event
		err    error
	}{
		{
			name: "Simple happy path",
			events: []*dedb.Event{
				{
					Id:       "",
					Name:     "CustomerCreated",
					Domain:   "CUSTOMER",
					DomainId: "testid",
				},
			},
			err: nil,
		},
		{
			name:   "Simple error for empty events list",
			events: []*dedb.Event{},
			err:    fmt.Errorf("no events were supplied to save"),
		},
	}

	// clear the db
	repo.pool.FlushAll(ctx)

	// when / then
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.save(ctx, tc.events)
			if tc.err == nil {
				assert.Nil(t, err)

				// assert each event in the batch
				for _, e := range tc.events {
					eventId := ""
					domainId := e.DomainId
					domain := e.Domain
					// see if the index key exists
					key := redisKey{
						db:     "dedb",
						shard:  0,
						prefix: "domain_events_timestamp_idx",
						key:    domainId,
					}
					s := repo.pool.ZRange(ctx, key.String(), int64(0), int64(1)).Val()
					repo.log.Info().Msgf("s: %v", s)
					assert.Equal(t, len(s), 1)
					eventId = s[0]
					repo.log.Debug().Msgf("event id %s", eventId)

					// see if the domain events has an entry
					key.prefix = "domain_events"
					key.key = domainId
					s = repo.pool.LRange(ctx, key.String(), int64(0), int64(1)).Val()
					assert.Equal(t, len(s), 1)
					payload := s[0]
					repo.log.Debug().Msgf("event id %s", payload)
					assert.NotEqual(t, "", payload)

					// see if the domain types has an entry
					key.prefix = "domain_types"
					key.key = domain
					s = repo.pool.ZRange(ctx, key.String(), int64(0), int64(1)).Val()
					assert.Equal(t, len(s), 1)
					payload = s[0]
					repo.log.Debug().Msgf("domain id %s", payload)
					assert.Equal(t, domainId, payload)
				}
			} else {
				assert.Equal(t, tc.err, err)
			}
		})
	}
}

func TestRepoGetDomain(t *testing.T) {
	// setup
	ctx := context.Background()
	config := Config{
		RedisDbConfig: RedisDbConfig{
			DbAddress: "redis:6379",
			DbIndex:   0,
		},
	}
	repo, err := NewRedisRepo(config)
	if err != nil {
		panic(err)
	}

	type jsonEvent struct {
		eventName string
		id        string
		data      string
	}
	cases := []struct {
		name       string
		domain     string
		domainId   string
		offset     int64
		limit      int64
		jsonEvents []jsonEvent
		err        error
	}{
		{
			name:     "Single event happy path",
			domain:   "CUSTOMER",
			domainId: "testid",
			offset:   0,
			limit:    1,
			jsonEvents: []jsonEvent{
				{
					eventName: "CustomerCreated",
					id:        "01GJ95SGV2492NWNQ3GHR3AZ02",
					data:      `{"id":"01GJ95SGV2492NWNQ3GHR3AZ02", "name":"CustomerCreated", "timestamp":"1668902863713555222", "domain":"CUSTOMER", "domainId":"testid"}`,
				},
			},
			err: nil,
		},
		{
			name:     "Two event happy path",
			domain:   "CUSTOMER",
			domainId: "testid2",
			offset:   0,
			limit:    2,
			jsonEvents: []jsonEvent{
				{
					eventName: "CustomerCreated",
					id:        "01GJ95SGV2492NWNQ3GHR3AZ02",
					data:      `{"id":"01GJ95SGV2492NWNQ3GHR3AZ02", "name":"CustomerCreated", "timestamp":"1668902863713555222", "domain":"CUSTOMER", "domainId":"testid2"}`,
				},
				{
					eventName: "CustomerUpdated",
					id:        "01GJ95SGV2492NWNQ3GHR3AZ01",
					data:      `{"id":"01GJ95SGV2492NWNQ3GHR3AZ01", "name":"CustomerUpdated", "timestamp":"1668902863713555233", "domain":"CUSTOMER", "domainId":"testid2"}`,
				},
			},
			err: nil,
		},
	}

	// clear the db
	repo.pool.FlushAll(ctx)

	// when / then
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// add the test data for the test case
			for _, je := range tc.jsonEvents {
				repo.pool.RPush(ctx, "dedb:domain_events:0:"+tc.domainId, je.data)
			}

			// execute the test case
			events, err := repo.getDomain(ctx, tc.domain, tc.domainId, tc.offset, tc.limit)
			if tc.err == nil {
				assert.Nil(t, err)
				assert.NotNil(t, events)
				assert.Equal(t, len(tc.jsonEvents), len(events))
				if len(events) == 0 {
					t.Fail()
				}
				assert.Equal(t, tc.domain, events[0].Domain)
				assert.Equal(t, tc.domainId, events[0].DomainId)
				// verify the event props match the test data
				for i, e := range tc.jsonEvents {
					assert.Equal(t, e.eventName, events[i].Name)
					assert.Equal(t, e.id, events[i].Id)
				}

			} else {
				assert.Equal(t, tc.err, err)
			}
		})
	}
}

func TestGetDomainIds(t *testing.T) {
	// setup
	ctx := context.Background()
	config := Config{
		RedisDbConfig: RedisDbConfig{
			DbAddress: "redis:6379",
			DbIndex:   0,
		},
	}
	repo, err := NewRedisRepo(config)
	if err != nil {
		panic(err)
	}
	cases := []struct {
		name     string
		domain   string
		domainId string
		offset   int64
		limit    int64
		expected int32
		ids      []string
		err      error
	}{
		{
			name:     "Single domain happy path",
			domain:   "CUSTOMER",
			domainId: "testid",
			offset:   0,
			limit:    2,
			expected: 2,
			ids: []string{
				"test_id",
				"test_id2",
			},
			err: nil,
		},
		{
			name:     "limit 1 id returned",
			domain:   "CUSTOMER",
			domainId: "testid",
			offset:   0,
			limit:    1,
			expected: 1,
			ids: []string{
				"test_id",
				"test_id2",
			},
			err: nil,
		},
	}

	// clear the db
	repo.pool.FlushAll(ctx)

	// when / then
	for _, tc := range cases {
		timestamp := time.Now().UnixMicro()
		t.Run(tc.name, func(t *testing.T) {
			// add the test data for the test case
			for _, id := range tc.ids {
				repo.pool.ZAddNX(ctx, "dedb:domain_types:0:"+tc.domain, &redis.Z{Score: float64(timestamp + 1), Member: id})
			}

			// execute the test case
			ids, err := repo.getDomainIds(ctx, tc.domain, tc.offset, tc.limit)
			if tc.err == nil {
				assert.Nil(t, err)
				assert.NotNil(t, ids)
				assert.NotEqual(t, 0, len(ids))
				assert.Equal(t, tc.expected, int32(len(ids)))
				// make sure the id(s) match
				for idx, i := range ids {
					assert.Equal(t, tc.ids[idx], i)
				}
			} else {
				assert.Equal(t, tc.err, err)
			}
		})
	}

}
