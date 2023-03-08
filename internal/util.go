package internal

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Encode(entity proto.Message) (string, error) {
	opts := protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	b, err := opts.Marshal(entity)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func Decode(entity proto.Message, data string) error {
	un := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	return un.Unmarshal([]byte(data), entity)
}

func newPool(useSearch bool, baseConfig Config, log *zerolog.Logger) (*redis.Client, error) {
	type commonConfig struct {
		DbAddress     string
		Password      string
		RedisCa       string
		RedisUserCert string
		RedisUserKey  string
		Index         string
		MinIdle       int
		MaxActive     int
		IdleTimeout   int64
		DbIndex       int
	}

	config := commonConfig(baseConfig.RedisDbConfig)
	if useSearch {
		config = commonConfig(baseConfig.RedisSearchConfig)
	}
	if config.MinIdle == 0 {
		config.MinIdle = 1
	}
	if config.MaxActive == 0 {
		config.MaxActive = 10
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = int64(240)
	}
	log.Info().Msgf("has ca: %v, has cert: %v, has key: %v", config.RedisCa != "", config.RedisUserCert != "", config.RedisUserKey != "")
	if config.RedisUserCert != "" {
		/*
			log.Info().Msgf("loading X509 cert and key pair")
			cert, err := tls.X509KeyPair([]byte(config.RedisUserCert), []byte(config.RedisUserKey))
			if err != nil {
				log.Error().Err(err).Msg("could not load redis keypair")
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM([]byte(config.RedisCa))
			tlscfg := &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      caCertPool,
			}
			tlscfg.InsecureSkipVerify = true
			log.Info().Msgf("setting pool for TLS enabled Redis server: %s, max idle: %d, max active: %d, idle timeout: %d", config.Server, config.MaxIdle, config.MaxActive, config.IdleTimeout)
			return &redis.Pool{
				MaxIdle:     config.MaxIdle,
				MaxActive:   config.MaxActive,
				IdleTimeout: time.Duration(config.IdleTimeout) * time.Second,
				Dial: func() (redis.Conn, error) {
					c, err := redis.Dial("tcp",
						config.Server,
						redis.DialPassword(config.Password),
						redis.DialUseTLS(true),
						redis.DialTLSConfig(tlscfg),
						redis.DialTLSSkipVerify(true))
					if err != nil {
						log.Error().Err(err).Msgf("error dialing redis")
					}

					return c, err
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			}, nil
		*/
	} else {
		log.Info().Msgf("setting pool for non TLS enabled Redis server: %s, min idle: %d, max active: %d, idle timeout: %d", config.DbAddress, config.MinIdle, config.MaxActive, config.IdleTimeout)
		return redis.NewClient(&redis.Options{
			Addr:         config.DbAddress,
			Password:     config.Password,
			MinIdleConns: config.MinIdle,
			PoolSize:     config.MaxActive,
			DB:           config.DbIndex,
			IdleTimeout:  time.Duration(config.IdleTimeout) * time.Second,
		}), nil
	}
	return nil, nil
}

/*
func newSearchClientFromPool(config Config, log *zerolog.Logger) (client *redisearch.Client, err error) {
	if config.RedisSearchConfig.Index == "" {
		return nil, status.Error(codes.InvalidArgument, "redis index is required for search client")
	}
	var (
		tries int
	)
	pool, err := newPool(true, config, log)
	log.Debug().Msg("Creating a new redis search client from pool")
	for client == nil && tries < 2 {
		client = redisearch.NewClientFromPool(pool, config.RedisSearchConfig.Index)
		if client != nil {
			log.Info().Msg("Created a new redis search client from pool")
			break
		}
		if tries > 2 {
			log.Error().Msgf("Could not create a new redis search client from pool after %d tries", tries)
			return nil, status.Error(codes.Internal, "could not create a new redis search client from pool")
		}
	}
	return client, nil
}
*/
