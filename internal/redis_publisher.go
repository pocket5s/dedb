package internal

import (
	"context"
	"dedb"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type redisPublisher struct {
	log    zerolog.Logger
	config Config
	client *redis.Client
}

// NewRedisPublisher function    
func NewRedisPublisher(config Config) (*redisPublisher, error) {
	pub := &redisPublisher{
		log:    log.With().Str("logger", "redisPublisher").Logger(),
		config: config,
	}

	client, err := newPool(false, pub.config, &pub.log)
	if err != nil {
		return nil, err
	}
	pub.client = client

	return pub, nil
}

func (p *redisPublisher) publish(ctx context.Context, events []*dedb.Event) {
	for _, event := range events {
		encoded, err := Encode(event)
		if err != nil {
			p.log.Error().Err(err).Msgf("could not encode event id %s", event.Id)
		} else {
			args := redis.XAddArgs{
				Stream: "dedb:stream:" + event.Domain,
				Values: map[string]interface{}{
					"id":        event.Id,
					"name":      event.Name,
					"timestamp": event.Timestamp,
					"metadata":  event.Metadata,
					"data":      encoded,
				},
			}
			p.client.XAdd(ctx, &args)
		}
	}
}

func (r *redisPublisher) shutdown() {
	r.client.Close()
}
